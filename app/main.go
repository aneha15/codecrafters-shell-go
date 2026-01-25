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
	Redirections []Redirection
}

type Redirection struct {
	FD       int // 1 for stdout, 2 for stderr, 0 for both
	Filename string
	Append   bool // true for >>, false for >
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
	var tokens []string
	var current strings.Builder
	var active rune = 0
	escaped := false
	inToken := false

	specialChars := []rune{'"', '$', '`', '\\'}
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		// handle redirection operators (outside quotes)
		if active == 0 && !escaped {
			redirect := ""

			// check for file descriptor notation (1>, 2>, 1>>, 2>>)
			if (char == '1' || char == '2') && i+1 < len(runes) && runes[i+1] == '>' {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
					inToken = false
				}
				if i+2 < len(runes) && runes[i+2] == '>' {
					redirect = string(char) + ">>"
					i += 2
				} else {
					redirect = string(char) + ">"
					i += 1
				}

				tokens = append(tokens, redirect)
				continue
			}

			// check for >, >>
			if char == '>' {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
					inToken = false
				}

				if i+1 < len(runes) && runes[i+1] == '>' {
					redirect = ">>"
					i += 1
				} else {
					redirect = ">"
				}

				tokens = append(tokens, redirect)
				continue
			}

			// check for &> and &>>
			if char == '&' && i+1 < len(runes) && runes[i+1] == '>' {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
					inToken = false
				}

				if i+2 < len(runes) && runes[i+2] == '>' {
					redirect = "&>>"
					i += 2
				} else {
					redirect = "&>"
					i += 1
				}

				tokens = append(tokens, redirect)
				continue
			}
		}

		// handle escaped chars
		if escaped {
			// double quotes => only special chars escaped
			if active == '"' && slices.Contains(specialChars, char) {
				current.WriteRune(char)
			} else if active == '"' {
				// not a special char => keep backslash
				current.WriteRune('\\')
				current.WriteRune(char)
			} else {
				// single quotes or outside quotes => escape everything
				current.WriteRune(char)
			}
			escaped = false
			inToken = true
			continue
		}

		// handle backslash
		if char == '\\' {
			if active == '\'' {
				// single quotes => backslash is literal
				current.WriteRune(char)
			} else {
				// double quotes or outside quotes => escaped
				escaped = true
			}
			inToken = true
			continue
		}

		// handle quotes
		if char == '\'' || char == '"' {
			if active == 0 {
				// opening quote
				active = char
			} else if active == char {
				// closing quote
				active = 0
			} else {
				// different quote inside active quotes
				current.WriteRune(char)
			}
			inToken = true
			continue
		}

		// handle spaces
		if char == ' ' || char == '\t' {
			if active != 0 {
				// space inside quotes
				current.WriteRune(char)
			} else if inToken {
				// space outside quotes => end of current token
				tokens = append(tokens, current.String())
				current.Reset()
				inToken = false
			}
			continue
		}

		// regular character
		current.WriteRune(char)
		inToken = true
	}

	// last token
	if inToken {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func handleRedirection(tokens []string) (cleanedArgs []string, redirects []Redirection, err error) {
	cleanedArgs = []string{}
	redirects = []Redirection{}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		redirect := Redirection{}
		isRedirect := false

		switch token {
		case ">", "1>":
			redirect.FD = 1
			redirect.Append = false
			isRedirect = true
		case ">>", "1>>":
			redirect.FD = 1
			redirect.Append = true
			isRedirect = true
		case "2>":
			redirect.FD = 2
			redirect.Append = false
			isRedirect = true
		case "2>>":
			redirect.FD = 2
			redirect.Append = true
			isRedirect = true
		case "&>":
			redirect.FD = 0
			redirect.Append = false
			isRedirect = true
		case "&>>":
			redirect.FD = 0
			redirect.Append = true
			isRedirect = true
		}

		if isRedirect {
			// next token should be filename
			if i+1 >= len(tokens) {
				return nil, nil, fmt.Errorf("syntax error: missing filename after %s", token)
			}

			redirect.Filename = tokens[i+1]
			redirects = append(redirects, redirect)
			i++ // skip next token (filename)
			continue
		}

		// not a redirect, add to cleaned args
		cleanedArgs = append(cleanedArgs, token)
	}

	return cleanedArgs, redirects, nil
}

func (s *Shell) readCommand() (*Command, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')

	if err != nil {
		return nil, err
	}

	input = strings.TrimSuffix(input, "\n")

	if input == "" {
		return nil, nil
	}

	tokens := parseInput(input)

	if len(tokens) == 0 {
		return nil, nil
	}

	// handle redirection
	cleanedArgs, redirects, err := handleRedirection(tokens)

	if err != nil {
		return nil, err
	}

	if len(cleanedArgs) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	cmd := &Command{
		Name:         cleanedArgs[0],
		Args:         cleanedArgs[1:],
		Redirections: redirects,
	}

	return cmd, nil
}

func (s *Shell) executeCommand(cmd *Command) {
	// setup default output writers
	stdoutWriter := io.Writer(os.Stdout)
	stderrWriter := io.Writer(os.Stderr)
	var filesToClose []*os.File // slice to track all files we open

	// close files when done
	defer func() {
		for _, f := range filesToClose {
			f.Close()
		}
	}()

	// process all redirections
	for _, redir := range cmd.Redirections {
		var file *os.File
		var err error

		if redir.Append {
			file, err = os.OpenFile(redir.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		} else {
			file, err = os.Create(redir.Filename)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", redir.Filename, err)
			return
		}

		filesToClose = append(filesToClose, file)

		// based on file descriptor, assign to appropriate stream(s)
		switch redir.FD {
		case 1:
			stdoutWriter = file
		case 2:
			stderrWriter = file
		case 0:
			stdoutWriter = file
			stderrWriter = file
		}
	}

	// try builtins first
	if handler, exists := s.builtins[cmd.Name]; exists {
		if err := handler(cmd, stdoutWriter); err != nil {
			fmt.Fprintf(stderrWriter, "%s: %v\n", cmd.Name, err)
		}
		return
	}

	// execute external cmd
	s.executeExternal(cmd, stdoutWriter, stderrWriter)
}

func (s *Shell) executeExternal(cmd *Command, stdout, stderr io.Writer) {
	_, err := exec.LookPath(cmd.Name)

	if err != nil {
		fmt.Fprintf(stderr, "%s: command not found\n", cmd.Name)
		return
	}

	extCmd := exec.Command(cmd.Name, cmd.Args...)

	// connect to terminal
	extCmd.Stdin = os.Stdin
	extCmd.Stdout = stdout
	extCmd.Stderr = stderr

	extCmd.Run()
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
		}
	}
	// change dir
	err := os.Chdir(path)
	if err != nil {
		fmt.Println("cd: " + path + ": No such file or directory")
	}
	return nil
}

func (s *Shell) cmdType(cmd *Command, w io.Writer) error {

	if _, exists := s.builtins[cmd.Args[0]]; exists {
		fmt.Fprintln(w, cmd.Args[0], "is a shell builtin")
		return nil
	}

	path, err := exec.LookPath(cmd.Args[0])
	if err == nil {
		fmt.Fprintln(w, cmd.Args[0], "is", path)
		return nil
	}
	fmt.Fprintf(w, "%s: not found\n", cmd.Args[0])
	return nil
}

func main() {
	shell := NewShell()
	shell.Run()
}
