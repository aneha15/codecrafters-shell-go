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

func handleCd(args []string) {
	path := "~"
	if args != nil {
		path = args[0]
	}

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
		fmt.Println("cd: " + path + ": No such file or directory")
	}
}

func parseInput(input string) []string {
	var parts []string
	var current strings.Builder
	var active rune = 0
	escaped := false
	inToken := false

	for _, char := range input {
		if escaped {
			current.WriteRune(char)
			escaped = false
			inToken = true
			continue
		}

		if char == '\\' {
			escaped = true
			inToken = true
			continue
		}

		quote := char == '\'' || char == '"'
		if quote {
			switch active {
			case 0:
				// opening quote
				active = char
				inToken = true
			case char:
				// closing quote
				active = 0
				inToken = true
			default:
				current.WriteRune(char)
			}
		} else if char == ' ' {
			if active != 0 {
				current.WriteRune(char)
			} else if inToken {
				// at a space outside quotes => end of word
				parts = append(parts, current.String())
				current.Reset()
				inToken = false
			}
		} else {
			current.WriteRune(char)
			inToken = true
		}
	}

	if inToken {
		parts = append(parts, current.String())
	}
	return parts
}

func handleInput() {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading output:", err)
		os.Exit(1)
	}

	parts := parseInput(input[:len(input)-1])
	command := parts[0]
	args := parts[1:]

	switch command {
	case "exit":
		os.Exit(0)
	case "echo":
		fmt.Println(strings.Join(args, " "))
	case "type":
		handleType(args[0])
	case "pwd":
		dir, _ := os.Getwd()
		fmt.Println(dir)
	case "cd":
		handleCd(args)
	default:
		handleExternal(command, args)
	}
}
