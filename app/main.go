package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage
	scanner := bufio.NewScanner(os.Stdin)
	VALID_COMMANDS := []string{"echo", "exit", "type"}
	for {
		fmt.Print("$ ")
		// fmt.Scan(&input)
		scanner.Scan()
		input := scanner.Text()
		if len(input) == 0 || input == "" {
			fmt.Println("command not found")
		}
		result := strings.Split(input, " ")
		command := result[0]
		text := result[1:]

		exists := slices.Contains(VALID_COMMANDS, command)
		if !exists {
			fmt.Printf("%s: command not found\n", input)
		}
		if len(result) == 1 {
			if command == "exit" {
				break
			}
		}

		if command == "echo" {
			res := strings.Join(text, " ")
			fmt.Println(res)
		}
		if command == "type" {
			fmt.Printf("%s is a shell builtin\n", text[1])
		}

	}
}
