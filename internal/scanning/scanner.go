package scanning

import (
	"errors"
	"fmt"
	"os"
)

type TokenType int

const (
	Arg TokenType = iota
	String
	EnvVar
	Space
	Eol
	Pipe
	RedirectStdout
	RedirectStdoutAppend
	RedirectStdin
	Export
	Equals
)

type Token struct {
	Type   TokenType
	Lexeme string
	Value  string
}

type Scanner struct {
	input   string
	start   int
	current int
	Tokens  []Token
}

func isAlphaNumeric(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isSpecial(r rune) bool {
	return r == ' ' || r == '>' || r == '<' || r == '|' || r == '"' || r == '$'
}

func NewScanner(input string) *Scanner {
	return &Scanner{input: input}
}

func (scanner *Scanner) Scan() error {
	for !scanner.isAtEnd() {
		scanner.start = scanner.current
		err := scanner.scanToken()
		if err != nil {
			return err
		}
	}
	scanner.Tokens = append(scanner.Tokens, Token{Type: Eol, Lexeme: "\n"})
	return nil
}

func (scanner *Scanner) scanToken() error {
	c := scanner.advance()
	switch c {
	case ' ':
		scanner.Tokens = append(scanner.Tokens, Token{Type: Space, Lexeme: " "})
	case '"':
		return scanner.scanString()
	case '$':
		scanner.scanEnvVar()
	case '|':
		return scanner.scanPipe()
	case '>':
		return scanner.scanRedirectStdout()
	case '<':
		return scanner.scanRedirectStdin()
	case '=':
		return scanner.scanEquals()
	default:
		scanner.scanArg()
	}

	return nil
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.current >= len(scanner.input)
}

func (scanner *Scanner) advance() rune {
	c := scanner.input[scanner.current]
	scanner.current++
	return rune(c)
}

func (scanner *Scanner) peek() rune {
	return rune(scanner.input[scanner.current])
}

func (scanner *Scanner) scanArg() {
	t := Arg
	for !scanner.isAtEnd() && !isSpecial(scanner.peek()) {
		scanner.advance()
	}

	arg := scanner.input[scanner.start:scanner.current]

	if arg == "export" && len(scanner.Tokens) == 0 {
		t = Export
	}
	if arg == "export" && len(scanner.Tokens) > 0 {
		lastToken := scanner.Tokens[len(scanner.Tokens)-1].Type
		if lastToken == Pipe {
			t = Export
		}
	}

	scanner.Tokens = append(scanner.Tokens, Token{Type: t, Lexeme: arg})
}

func (scanner *Scanner) scanString() error {
	for !scanner.isAtEnd() && scanner.peek() != '"' {
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return errors.New("missing \" to close string")
	}

	scanner.advance()

	str := scanner.input[scanner.start+1 : scanner.current-1]
	scanner.Tokens = append(scanner.Tokens, Token{Type: String, Lexeme: str})
	return nil
}

func (scanner *Scanner) scanEnvVar() {
	for !scanner.isAtEnd() && (isAlphaNumeric(scanner.peek()) || scanner.peek() == '_') {
		scanner.advance()
	}

	if scanner.input[scanner.current-1] == '$' {
		scanner.Tokens = append(scanner.Tokens, Token{Type: Arg, Lexeme: "$"})
		return
	}

	envVar := scanner.input[scanner.start+1 : scanner.current]
	scanner.Tokens = append(
		scanner.Tokens,
		Token{Type: EnvVar, Lexeme: envVar, Value: os.Getenv(envVar)},
	)
}

func (scanner *Scanner) scanEquals() error {
	for !scanner.isAtEnd() && scanner.peek() == ' ' {
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return errors.New("error near '='")
	}

	scanner.Tokens = append(scanner.Tokens, Token{Type: Pipe, Lexeme: "="})
	return nil
}

func (scanner *Scanner) scanPipe() error {
	for !scanner.isAtEnd() && scanner.peek() == ' ' {
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return errors.New("error near '|'")
	}

	scanner.Tokens = append(scanner.Tokens, Token{Type: Pipe, Lexeme: "|"})
	return nil
}

func (scanner *Scanner) scanRedirectStdout() error {
	lexeme := ">"
	t := RedirectStdout

	if !scanner.isAtEnd() && scanner.peek() == '>' {
		lexeme += ">"
		t = RedirectStdoutAppend
		scanner.advance()
	}

	for !scanner.isAtEnd() && scanner.peek() == ' ' {
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return fmt.Errorf("error near '%s'", lexeme)
	}

	scanner.Tokens = append(scanner.Tokens, Token{Type: t, Lexeme: lexeme})
	return nil
}

func (scanner *Scanner) scanRedirectStdin() error {
	for !scanner.isAtEnd() && scanner.peek() == ' ' {
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return errors.New("error near '<'")
	}

	scanner.Tokens = append(scanner.Tokens, Token{Type: RedirectStdin, Lexeme: "<"})
	return nil
}
