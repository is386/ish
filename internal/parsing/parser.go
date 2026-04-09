package parsing

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/is386/ish/internal/scanning"
)

type Command struct {
	Cmd        *exec.Cmd
	IsRedirect bool
}

type Parser struct {
	Cmds    []Command
	files   []*os.File
	tokens  []scanning.Token
	current int
	stdin   io.ReadCloser
	stdout  io.Writer
}

func NewParser(tokens []scanning.Token) Parser {
	return Parser{tokens: tokens}
}

func (parser *Parser) Parse() error {
	for !parser.isAtEnd() {
		err := parser.parseToken()
		if err != nil {
			return err
		}
	}

	if len(parser.Cmds) > 0 {
		lastCmd := parser.Cmds[len(parser.Cmds)-1]
		if lastCmd.Cmd.Stdout == nil {
			lastCmd.Cmd.Stdout = os.Stdout
		}
	}

	return nil
}

func (parser *Parser) parseToken() error {
	t := parser.advance()
	switch t.Type {
	case scanning.Arg:
		fallthrough
	case scanning.EnvVar:
		fallthrough
	case scanning.String:
		return parser.parseArgs(t)
	case scanning.Pipe:
		return parser.parsePipe()
	}
	return nil
}

func (parser *Parser) isAtEnd() bool {
	return parser.peek().Type == scanning.Eol
}

func (parser *Parser) advance() scanning.Token {
	t := parser.tokens[parser.current]
	parser.current++
	return t
}

func (parser *Parser) peek() scanning.Token {
	return parser.tokens[parser.current]
}

func (parser *Parser) parseArgs(t scanning.Token) error {
	isRedirect := false
	args := []string{}
	arg := t.Lexeme
	if t.Type == scanning.EnvVar {
		arg = t.Value
	}

	for !parser.isAtEnd() && parser.peek().Type != scanning.Pipe {
		if parser.peek().Type == scanning.Space && arg != "" {
			args = append(args, arg)
			arg = ""
		}

		t = parser.advance()
		switch t.Type {
		case scanning.EnvVar:
			arg += t.Value
		case scanning.String:
			fallthrough
		case scanning.Arg:
			arg += t.Lexeme
		case scanning.RedirectStdout:
			fallthrough
		case scanning.RedirectStdoutAppend:
			err := parser.parseRedirectStdout(t)
			if err != nil {
				return err
			}
			isRedirect = true
		}
	}

	if (parser.isAtEnd() || parser.peek().Type == scanning.Pipe) && arg != "" {
		args = append(args, arg)
	}

	cmd, err := buildCmd(args, parser.stdin, parser.stdout)
	if err != nil {
		return err
	}

	parser.Cmds = append(parser.Cmds, Command{Cmd: cmd, IsRedirect: isRedirect})
	parser.stdout = nil
	return nil
}

func (parser *Parser) parsePipe() error {
	if len(parser.Cmds) == 0 {
		return errors.New("left side of '|' empty")
	}

	for !parser.isAtEnd() && parser.peek().Type == scanning.Space {
		parser.advance()
	}

	if parser.isAtEnd() {
		return errors.New("right side of '|' empty")
	}

	lastCmd := parser.Cmds[len(parser.Cmds)-1]
	if lastCmd.IsRedirect {
		return nil
	}

	pipe, err := lastCmd.Cmd.StdoutPipe()
	if err != nil {
		return err
	}

	parser.stdin = pipe

	return nil
}

func (parser *Parser) parseRedirectStdout(t scanning.Token) error {
	for !parser.isAtEnd() && parser.peek().Type == scanning.Space {
		parser.advance()
	}

	if parser.isAtEnd() {
		return fmt.Errorf("right side of '%s' empty", t.Lexeme)
	}

	tNext := parser.advance()

	var filename string
	switch tNext.Type {
	case scanning.EnvVar:
		filename = tNext.Value
	case scanning.String:
		fallthrough
	case scanning.Arg:
		filename = tNext.Lexeme
	default:
		return fmt.Errorf("parse error near '%s'", tNext.Lexeme)
	}

	flag := os.O_WRONLY | os.O_CREATE
	if t.Type == scanning.RedirectStdoutAppend {
		flag = os.O_APPEND | flag
	} else {
		flag = os.O_TRUNC | flag
	}

	outfile, err := os.OpenFile(filename, flag, 0o600)
	if err != nil {
		return err
	}

	parser.files = append(parser.files, outfile)
	parser.stdout = outfile

	return nil
}

func (parser *Parser) CloseFiles() {
	for _, file := range parser.files {
		file.Close()
	}
}

func buildCmd(args []string, stdin io.ReadCloser, stdout io.Writer) (*exec.Cmd, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if stdin != nil {
		cmd.Stdin = stdin
	}

	return cmd, nil
}
