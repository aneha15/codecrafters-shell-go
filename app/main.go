package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
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

	builtinCommands := []string{"echo", "exit", "type"}

	switch command {
	case "exit":
		os.Exit(0)
	case "echo":
		fmt.Print(input[5:])
	case "type":
		c := splitInput[1]
		if slices.Contains(builtinCommands, c) {
			fmt.Println(c + " is a shell builtin")
			return
		}

		fullPath, err := exec.LookPath(c)
		if err == nil {
			fmt.Println(c + " is " + fullPath)
			return
		}
		fmt.Println(c + ": not found")
	default:
		_, err := exec.LookPath(command)
		if err == nil {
			cmd := exec.Command(command, splitInput[1:]...)

			// connect to terminal
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			cmd.Run()
			return
		}
		fmt.Println(command + ": command not found")
	}
}
