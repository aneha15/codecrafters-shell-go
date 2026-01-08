package main

import (
	"bufio"
	"fmt"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	for {
		fmt.Print("$ ")
		reader := bufio.NewReader(os.Stdin)
		command, err := reader.ReadString('\n')

		if command == "exit\n" {
			os.Exit(0)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading output:", err)
			os.Exit(1)
		}

		fmt.Println(command[:len(command)-1] + ": command not found")

	}

}
