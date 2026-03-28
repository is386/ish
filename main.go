package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// 1. Signal handling (Ctrl+C) — Currently Ctrl+C kills the shell itself. You need to trap
// SIGINT so it cancels the running command but returns you to the prompt.
// 2. Quoted string parsing — strings.Split(input, " ") breaks on any space, so echo "hello
// world" passes two args instead of one. You need basic quote-aware tokenization.
// 3. Pipes (cmd1 | cmd2) — This is the single most-used shell feature. Without it you can't
// compose commands at all.
// 4. I/O redirection (>, >>, <) — Writing output to files and reading input from files is
// essential for real work.
// 5. Environment variable expansion ($VAR, $HOME) — Commands like echo $PATH currently pass
// the literal string $PATH.
// 6. export builtin — Needed to set env vars for child processes (e.g., export
// GOPATH=/home/user/go).
func main() {
	inputReader := bufio.NewReader(os.Stdin)

	for {
		prompt()

		input := getInput(inputReader)
		if input == "" {
			continue
		}

		err := execCmd(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
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

	fmt.Printf("%s %s $ ", user, path)
}

func getInput(reader *bufio.Reader) string {
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ""
	}

	input = strings.TrimSpace(input)
	return input
}

func execCmd(input string) error {
	args := strings.Split(input, " ")

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
	return cmd.Run()
}
