package main

import (
	"fmt"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage

	var command string

	fmt.Print("$ ")
	fmt.Scan(&command)
	fmt.Printf("%s: command not found", command)
}
