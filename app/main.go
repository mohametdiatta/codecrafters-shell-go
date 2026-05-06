package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

var builtins = []string{"echo", "exit", "type", "pwd", "cd"}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("$ ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		parts := strings.Fields(input) // handles multiple spaces better than Split
		command, args := parts[0], parts[1:]

		switch command {
		case "exit":
			return

		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "pwd":
			path, err := os.Getwd()
			if err == nil {
				fmt.Println(path)
			}

		case "cd":
			var dir = args[0]
			if dir == "~" {
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
				out, _ := cmd.CombinedOutput()
				fmt.Print(string(out))
			} else {
				fmt.Printf("%s: command not found\n", command)
			}

		}
	}
}
