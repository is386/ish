package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/is386/ish/internal/parsing"
	"github.com/is386/ish/internal/scanning"
)

var (
	isCmdRunning = false
	errExit      = errors.New("exit")
)

// 6. export builtin

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

		parser := parsing.NewParser(scanner.Tokens)

		err = parser.Parse()
		if err != nil {
			parser.CloseFiles()
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		err = execCmds(parser.Cmds)
		parser.CloseFiles()
		if err != nil {
			if errors.Is(err, errExit) {
				os.Exit(0)
			}
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

func execCmds(cmds []parsing.Command) error {
	isCmdRunning = true
	for _, cmd := range cmds {
		switch cmd.Cmd.Args[0] {
		case "cd":
			var path string
			if len(cmd.Cmd.Args) == 1 {
				path = os.Getenv("HOME")
			} else {
				path = cmd.Cmd.Args[1]
			}
			isCmdRunning = false
			return os.Chdir(path)
		case "exit":
			return errExit
		}
		err := cmd.Cmd.Start()
		if err != nil {
			return err
		}
	}

	for _, cmd := range cmds {
		err := cmd.Cmd.Wait()
		if err != nil {
			return err
		}
	}
	isCmdRunning = false
	return nil
}
