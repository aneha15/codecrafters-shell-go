package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// parsed command
type Command struct {
	Name         string
	Args         []string
	RedirectFile string
}

func main() {

}

// shell
type Shell struct {
	builtins map[string]func(*Command, io.Writer) error
}

func NewShell() *Shell {
	s := &Shell{
		builtins: make(map[string]func(*Command, io.Writer) error),
	}
	s.registerBuiltins()
	return s
}

func (s *Shell) registerBuiltins() {
	s.builtins["exit"] = s.cmdExit
	s.builtins["echo"] = s.cmdEcho
	s.builtins["type"] = s.cmdType
	s.builtins["pwd"] = s.cmdPwd
	s.builtins["cd"] = s.cmdCd
}

func (s *Shell) Run() {
	for {
		fmt.Print("$ ")
		cmd, err := s.readCommand()

		// error present
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}
		// empty input
		if cmd == nil {
			continue
		}

		s.executeCommand(cmd)
	}
}

func parseInput(input string) []string {
	var parts []string
	var current strings.Builder
	var active rune = 0
	escaped := false
	inToken := false

	specialChars := []rune{'"', '$', '`', '\\'}

	for _, char := range input {

		if char == '>' && active == 0 {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			parts = append(parts, ">")
			continue
		}

		if escaped {
			if active == '"' && slices.Contains(specialChars, char) {
				current.WriteRune(char)
			} else if active == '"' {
				current.WriteRune('\\')
				current.WriteRune(char)
			} else {
				current.WriteRune(char)
			}
			escaped = false
			inToken = true
			continue
		}

		if char == '\\' {
			if active == '\'' {
				current.WriteRune(char)
			} else {
				escaped = true
			}
			inToken = true
			continue
		}

		if char == '\'' || char == '"' {
			if active == 0 {
				active = char
			} else if active == char {
				active = 0
			} else {
				current.WriteRune(char)
			}
			inToken = true
		} else if char == ' ' {
			if active != 0 {
				current.WriteRune(char)
			} else if inToken {
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

func handleRedirection(args []string) (cleanedArgs []string, file string, err error) {
	for i := 0; i < len(args); i++ {
		if args[i] == ">" {
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("redirection error: no output file specified")
			}
			file = args[i+1]
			cleanedArgs = append(args[:i], args[i+2:]...)
			return cleanedArgs, file, nil
		}
	}
	return args, "", nil // no > found
}

func (s *Shell) readCommand() (*Command, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')

	if err != nil {
		return nil, err
	}

	parts := parseInput(strings.TrimSuffix(input, "\n"))

	if len(parts) == 0 {
		return nil, nil
	}

	cmd := &Command{
		Name: parts[0],
		Args: parts[1:],
	}

	// handle redirection
	cleanedArgs, file, err := handleRedirection(parts)

	if err != nil {
		return nil, err
	}

	cmd.Args = cleanedArgs[1:]
	cmd.RedirectFile = file

	return cmd, nil
}

func (s *Shell) executeCommand(cmd *Command) {
	output := io.Writer(os.Stdout)

	if cmd.RedirectFile != "" {
		file, err := os.Create(cmd.RedirectFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating file:", err)
			return
		}
		defer file.Close()
		output = file
	}

	// check builtins and execute if it exists
	if handler, exists := s.builtins[cmd.Name]; exists {
		if err := handler(cmd, output); err != nil {
			fmt.Fprintln(os.Stderr, cmd.Name, ":", err)
		}
		return
	}
	// if not builtin, try executing external
	s.executeExternal(cmd, output)
}

func (s *Shell) executeExternal(cmd *Command, w io.Writer) {
	path, err := exec.LookPath(cmd.Name)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", cmd.Name)
		return
	}

	extCmd := exec.Command(path, cmd.Args...)

	// connect to terminal
	extCmd.Stdin, extCmd.Stdout, extCmd.Stderr = os.Stdin, w, os.Stderr

	if err := extCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing %s: %v\n", cmd.Name, err)
	}
}

func (s *Shell) cmdEcho(cmd *Command, w io.Writer) error {
	fmt.Fprintln(w, strings.Join(cmd.Args, " "))
	return nil
}

func (s *Shell) cmdExit(cmd *Command, w io.Writer) error {
	os.Exit(0)
	return nil
}

func (s *Shell) cmdPwd(cmd *Command, w io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, dir)
	return nil
}

func (s *Shell) cmdCd(cmd *Command, w io.Writer) error {
	path := "~"
	if cmd.Args != nil {
		path = cmd.Args[0]
	}

	// find home dir
	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = homeDir
		} else {
			return fmt.Errorf("cd: cannot determine home directory: %v", err)
		}
	}
	// change dir
	err := os.Chdir(path)
	if err != nil {
		fmt.Println("cd: " + path + ": No such file or directory")
	} else {
		return fmt.Errorf("cd: %v", err)
	}
	return nil
}

func (s *Shell) cmdType(cmd *Command, w io.Writer) error {

	if _, exists := s.builtins[cmd.Args[0]]; exists {
		fmt.Fprintln(w, cmd.Args[0], "is a shell builtin")
	}

	path, err := exec.LookPath(cmd.Args[0])
	if err == nil {
		fmt.Fprintln(w, cmd.Args[0], "is", path)
		return nil
	}
	fmt.Fprintln(w, cmd.Args[0], ": not found")
	return err
}
