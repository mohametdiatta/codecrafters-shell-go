package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/google/shlex"
)

var builtins = []string{"echo", "exit", "type", "pwd", "cd"}
var operators = []string{">", "1>"}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("$ ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		parts, _ := shlex.Split(input)
		command, args := parts[0], parts[1:]

		switch command {
		case "exit":
			return

		case "echo":
			operator := args[1]
			if slices.Contains(operators, operator) {
				fileName := args[2]
				fileContent := []byte(args[0])
				err := os.WriteFile(fileName, fileContent, 0644)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println(strings.Join(args, " "))
			}

		case "pwd":
			path, err := os.Getwd()
			if err == nil {
				fmt.Println(path)
			}

		case "cd":
			var dir = args[0]
			if dir == "~" || dir == "" {
				home, _ := os.UserHomeDir()
				dir = home
			}
			err := os.Chdir(dir)
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", dir)
			}

		case "type":
			if len(args) == 0 {
				fmt.Println("type: missing argument")
				continue
			}
			arg := args[0]
			if slices.Contains(builtins, arg) {
				fmt.Printf("%s is a shell builtin\n", arg)
			} else if path, err := exec.LookPath(arg); err == nil {
				fmt.Printf("%s is %s\n", arg, path)
			} else {
				fmt.Printf("%s: not found\n", arg)
			}

		default:
			if path, err := exec.LookPath(command); err == nil {
				cmd := exec.Command(path, args...)
				cmd.Args[0] = command
				out, err := cmd.CombinedOutput()
				if err != nil {
					continue
					// fmt.Println(err.Error())
				}
				if slices.Contains(args, ">") || slices.Contains(args, "1>") {
					fileName := args[len(args)-1]
					fileContent := []byte(out)
					err := os.WriteFile(fileName, fileContent, 0644)
					if err != nil {
						fmt.Println(err.Error())
					}
				} else {
					fmt.Print(string(out))
				}
			} else {
				fmt.Printf("%s: command not found\n", command)
			}

		}
	}
}
