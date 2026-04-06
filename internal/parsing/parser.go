package parsing

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/is386/ish/internal/scanning"
)

type Parser struct {
	Cmds    []*exec.Cmd
	tokens  []scanning.Token
	current int
	stdin   io.ReadCloser
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
		parser.Cmds[len(parser.Cmds)-1].Stdout = os.Stdout
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
		}
	}

	if (parser.isAtEnd() || parser.peek().Type == scanning.Pipe) && arg != "" {
		args = append(args, arg)
	}

	cmd, err := buildCmd(args, parser.stdin)
	if err != nil {
		return err
	}

	parser.Cmds = append(parser.Cmds, cmd)
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

	pipe, err := parser.Cmds[len(parser.Cmds)-1].StdoutPipe()
	if err != nil {
		return err
	}

	parser.stdin = pipe

	return nil
}

func buildCmd(args []string, stdin io.ReadCloser) (*exec.Cmd, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if stdin != nil {
		cmd.Stdin = stdin
	}

	return cmd, nil
}
