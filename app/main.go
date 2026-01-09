package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
		c := splitInput[1]
		if slices.Contains(builtinCommands, c) {
			fmt.Println(c + " is a shell builtin")
			return
		}

		path := os.Getenv(("PATH"))
		directories := strings.Split(path, string(os.PathListSeparator))

		// for range loop -> used to iterate over collections
		// range returns index and value for every item
		for _, dir := range directories {
			fullPath := filepath.Join(dir, c)

			// check if file with command name exists and its permissions
			info, err := os.Stat(fullPath)
			if err == nil {
				if !info.IsDir() && info.Mode().Perm()&0111 != 0 {
					fmt.Println(c + " is " + fullPath)
					return
				}
			}

		}

		fmt.Println(c + ": not found")

	default:
		fmt.Println(command + ": command not found")
	}

}
