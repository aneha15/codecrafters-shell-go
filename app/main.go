package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

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
	default:
		fmt.Println(command + ": command not found")
	}

}
