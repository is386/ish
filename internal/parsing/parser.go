package parsing

import (
	"errors"
	"os"
	"os/exec"

	"github.com/is386/ish/internal/scanning"
)

func Parse(tokens []scanning.Token) ([]*exec.Cmd, error) {
	cmds := []*exec.Cmd{}
	args := []string{}
	arg := ""
	var prevCmd *exec.Cmd

	for _, t := range tokens {
		switch t.Type {
		case scanning.Eol:
			fallthrough
		case scanning.Space:
			if arg == "" {
				continue
			}
			args = append(args, arg)
			arg = ""
		case scanning.Pipe:
			if len(args) == 0 {
				return nil, errors.New("error near '|'")
			}
			cmd, err := buildCmd(args, prevCmd)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
			prevCmd = cmd
			args = nil
			arg = ""
		case scanning.EnvVar:
			arg += t.Value
		default:
			arg += t.Lexeme
		}
	}

	cmd, err := buildCmd(args, prevCmd)
	if err != nil {
		return nil, err
	}
	cmd.Stdout = os.Stdout

	return append(cmds, cmd), nil
}

func buildCmd(args []string, prevCmd *exec.Cmd) (*exec.Cmd, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if prevCmd != nil {
		pipe, err := prevCmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		cmd.Stdin = pipe
	}

	return cmd, nil
}
