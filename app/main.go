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

func handleType(c string) {
	builtinCommands := []string{"echo", "exit", "type", "pwd", "cd"}
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
}

func handleExternal(c string, args []string) {
	_, err := exec.LookPath(c)
	if err == nil {
		cmd := exec.Command(c, args...)

		// connect to terminal
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Run()
		return
	}
	fmt.Println(c + ": command not found")
}

func handleCd(dir string) {
	path := dir
	// find home dir
	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = homeDir
		}
	}

	// change dir
	err := os.Chdir(path)
	if err != nil {
		fmt.Println("cd: " + dir + ": No such file or directory")
	}
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
		handleType(splitInput[1])
	case "pwd":
		dir, _ := os.Getwd()
		fmt.Println(dir)
	case "cd":
		if len(splitInput) == 1 {
			handleCd("~")
		} else {
			handleCd(splitInput[1])
		}
	default:
		handleExternal(command, splitInput[1:])
	}
}
