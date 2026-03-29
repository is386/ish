package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

// 2. Quoted string parsing
// 3. Pipes (cmd1 | cmd2)
// 4. I/O redirection (>, >>, <)
// 5. Environment variable expansion ($VAR, $HOME)
// 6. export builtin
func main() {
	inputChannel := make(chan string)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	inputReader := bufio.NewReader(os.Stdin)
	go readInput(inputReader, inputChannel)

	for {
		prompt()

		select {
		case input := <-inputChannel:
			if input == "" {
				continue
			}

			err := execCmd(input)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			select {
			case <-signalChannel:
			default:
			}
		case <-signalChannel:
			fmt.Println()
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

func readInput(inputReader *bufio.Reader, inputChannel chan<- string) {
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			inputChannel <- ""
			continue
		}
		inputChannel <- strings.TrimSpace(input)
	}
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
