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

var builtinCommands = []string{"echo", "exit", "type"}

func main() {

	for {
		printPrompt()
		handleInput()
	}
}

func printPrompt() {
	fmt.Print("$ ")
}

func handleInput() {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading output:", err)
		os.Exit(1)
	}

	splitInput := strings.Split(input[:len(input)-1], " ")
	command := splitInput[0]

	switch command {
	case "exit":
		os.Exit(0)
	case "echo":
		fmt.Print(input[5:])
	case "type":
		if slices.Contains(builtinCommands, splitInput[1]) {
			fmt.Println(splitInput[1] + " is a shell builtin")
		} else {
			fmt.Println(splitInput[1] + ": not found")
		}

	default:
		fmt.Println(command + ": command not found")
	}

}
