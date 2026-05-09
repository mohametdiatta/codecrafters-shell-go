package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/chzyer/readline"
)

type Command struct {
	Stdout io.Writer
	Stderr io.Writer
	exec   func(cmd Command, args []string)
}

func (cmd Command) Start(args []string) {
	cmd.exec(cmd, args)
}

func handleExit(cmd Command, args []string) {
	os.Exit(0)
}

func handleEcho(cmd Command, args []string) {
	fmt.Fprintln(cmd.Stdout, strings.Join(args[1:], " "))
}

func handleType(cmd Command, args []string) {
	var _, exists = builtins[args[1]]
	if exists {
		fmt.Fprintf(cmd.Stdout, "%s is a shell builtin\n", args[1])
	} else if path, err := exec.LookPath(args[1]); nil == err {
		fmt.Fprintln(cmd.Stdout, args[1], "is", path)
	} else {
		fmt.Fprintf(cmd.Stdout, "%s: not found\n", args[1])
	}
}

func handlePwd(cmd Command, args []string) {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(cmd.Stderr, "error retrieving working directory: %v\n", err)
	}
	fmt.Fprintln(cmd.Stdout, pwd)
}

func handleCd(cmd Command, args []string) {
	var dir string
	if args[1] == "~" {
		dir = os.Getenv("HOME")
	} else {
		dir = args[1]
	}
	err := os.Chdir(dir)
	if err != nil {
		fmt.Fprintf(cmd.Stderr, "cd: %s: No such file or directory\n", args[1])
	}
}

var builtins map[string]Command

func init() {
	builtins = map[string]Command{
		"exit": {exec: handleExit},
		"echo": {exec: handleEcho},
		"type": {exec: handleType},
		"pwd":  {exec: handlePwd},
		"cd":   {exec: handleCd},
	}
}

func findMatches(line string) [][]rune {
	// Ta liste de commandes disponibles
	commands := []string{"echo", "exit"}
	var matches [][]rune

	for _, cmd := range commands {
		// Vérifie si la commande commence par la saisie actuelle
		if strings.HasPrefix(cmd, line) {
			// On convertit en []rune car c'est ce qu'attend readline
			matches = append(matches, []rune(cmd[len(line):]+" "))
		}
	}
	return matches
}

type myCompleter struct{}

func (c *myCompleter) Do(line []rune, pos int) ([][]rune, int) {
	suggestions := findMatches(string(line[:pos]))
	if len(suggestions) == 0 {
		fmt.Printf("%c", 0x07)
		return nil, 0
	}
	return suggestions, len(line[:pos])
}

func readCommand() []string {

	var completer readline.AutoCompleter = &myCompleter{}
	cmd, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		InterruptPrompt: "^C",
		AutoComplete:    completer,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}
	defer cmd.Close()
	line, _ := cmd.Readline()
	args, err := Parse(string(line))

	return args
}

func evalCommand(args []string) {
	var stdout *os.File = os.Stdout
	if len(args) > 2 && (args[len(args)-2] == ">" || args[len(args)-2] == "1>") {
		outputFile, err := os.Create(args[len(args)-1])
		if err != nil {
			panic(err)
		}
		defer outputFile.Close()
		stdout = outputFile
		args = args[:len(args)-2]
	}
	var stderr *os.File = os.Stderr

	if len(args) > 2 && args[len(args)-2] == "2>" {
		outputFile, err := os.Create(args[len(args)-1])
		if err == nil {
			stderr = outputFile
			defer stderr.Close()
		}

		args = args[:len(args)-2]
	}
	if len(args) > 2 && (args[len(args)-2] == ">>" || args[len(args)-2] == "1>>") {
		outputFile, err := os.OpenFile(args[len(args)-1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		stdout = outputFile

		defer outputFile.Close()
		args = args[:len(args)-2]
	}
	if len(args) > 2 && (args[len(args)-2] == ">>" || args[len(args)-2] == "2>>") {
		outputFile, err := os.OpenFile(args[len(args)-1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		stderr = outputFile
		defer stderr.Close()

		defer outputFile.Close()
		args = args[:len(args)-2]
	}

	if cmd, builtin := builtins[args[0]]; builtin {
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		cmd.Start(args)
	} else if _, err := exec.LookPath(args[0]); nil == err {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		cmd.Stdin = os.Stdin
		cmd.Start()
		cmd.Wait()
	} else {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", args[0])
	}
}

type argType int

const (
	argNo argType = iota
	argSingle
	argQuoted
)

func Parse(line string) ([]string, error) {
	args := []string{}
	buf := ""
	var escaped, doubleQuoted, singleQuoted bool

	got := argNo

	for _, r := range line {
		if escaped {
			if doubleQuoted && !slices.Contains([]rune{'"', '\\', '$', '`', '\n'}, r) {
				buf += string('\\')
			}
			buf += string(r)
			escaped = false
			got = argSingle
			continue
		}

		switch r {
		case ' ':
			if singleQuoted || doubleQuoted {
				buf += string(r)
			} else if got != argNo {
				args = append(args, buf)
				buf = ""
				got = argNo
			}
			continue
		case '\\':
			if singleQuoted {
				buf += string(r)
			} else {
				escaped = true
			}
			continue
		case '"':
			if !singleQuoted {
				if doubleQuoted {
					got = argQuoted
				}
				doubleQuoted = !doubleQuoted
				continue
			}
		case '\'':
			if !doubleQuoted {
				if singleQuoted {
					got = argQuoted
				}
				singleQuoted = !singleQuoted
				continue
			}
		}

		got = argSingle
		buf += string(r)
	}

	if got != argNo {
		args = append(args, buf)
	}

	if escaped || singleQuoted || doubleQuoted {
		return nil, errors.New("invalid command line string")
	}
	// fmt.Fprintf(os.Stderr, "args %s.\n", strings.Join(args, ","))
	return args, nil
}

func main() {
	for {
		args := readCommand()
		evalCommand(args)
	}
}
