package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/is386/ish/internal/scanning"
)

var isCmdRunning = false

// 3. Pipes (cmd1 | cmd2)
// 4. I/O redirection (>, >>, <)
// 5. Environment variable expansion ($VAR, $HOME)
// 6. export builtin

// TODO: Characters from long running program appearing in next prompt. git log is an example
func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		for range signalChannel {
			if isCmdRunning {
				return
			}
			fmt.Println()
			prompt()
		}
	}()

	inputReader := bufio.NewReader(os.Stdin)
	for {
		prompt()

		input := getInput(inputReader)
		if input == "" {
			continue
		}

		scanner := scanning.NewScanner(input)
		err := scanner.Scan()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		args := buildArgs(scanner.Tokens)
		err = execCmd(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func prompt() {
	home := os.Getenv("HOME")
	user := os.Getenv("USER")

	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find current working directory")
		os.Exit(1)
	}
	path = strings.Replace(path, home, "~", 1)

	fmt.Printf("\033[94m%s %s $\033[0m ", user, path)
}

func getInput(inputReader *bufio.Reader) string {
	input, err := inputReader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ""
	}
	return strings.TrimSpace(input)
}

func execCmd(args []string) error {
	switch args[0] {
	case "cd":
		var path string
		if len(args) == 1 {
			path = os.Getenv("HOME")
		} else {
			path = args[1]
		}
		return os.Chdir(path)
	case "exit":
		os.Exit(0)
	}

	_, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("ish: %s: command not found", args[0])
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	isCmdRunning = true
	cmdOut := cmd.Run()
	isCmdRunning = false
	return cmdOut
}

func buildArgs(tokens []scanning.Token) []string {
	args := []string{}
	arg := ""

	for _, t := range tokens {
		switch t.Type {
		case scanning.EOL:
			fallthrough
		case scanning.SPACE:
			if arg == "" {
				continue
			}
			args = append(args, arg)
			arg = ""
		case scanning.ENVVAR:
			arg += t.Value
		default:
			arg += t.Lexeme
		}
	}

	return args
}
