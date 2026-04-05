package scanning

import (
	"errors"
	"os"
)

type TokenType int

const (
	ARG TokenType = iota
	STRING
	ENVVAR
	SPACE
	EOL
	PIPE
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

func NewScanner(input string) *Scanner {
	return &Scanner{input: input}
}

func (scanner *Scanner) Scan() error {
	for !scanner.isAtEnd() {
		scanner.start = scanner.current
		err := scanner.parseToken()
		if err != nil {
			return err
		}
	}
	scanner.Tokens = append(scanner.Tokens, Token{Type: EOL, Lexeme: "\n"})
	return nil
}

func (scanner *Scanner) parseToken() error {
	c := scanner.advance()
	switch c {
	case ' ':
		scanner.Tokens = append(scanner.Tokens, Token{Type: SPACE, Lexeme: " "})
	case '"':
		return scanner.scanString()
	case '$':
		scanner.scanEnvVar()
	case '|':
		scanner.Tokens = append(scanner.Tokens, Token{Type: PIPE, Lexeme: "|"})
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
	for !scanner.isAtEnd() && scanner.peek() != ' ' {
		scanner.advance()
	}

	arg := scanner.input[scanner.start:scanner.current]
	scanner.Tokens = append(scanner.Tokens, Token{Type: ARG, Lexeme: arg})
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
	scanner.Tokens = append(scanner.Tokens, Token{Type: STRING, Lexeme: str})
	return nil
}

func (scanner *Scanner) scanEnvVar() {
	for !scanner.isAtEnd() && (isAlphaNumeric(scanner.peek()) || scanner.peek() == '_') {
		scanner.advance()
	}

	if scanner.input[scanner.current-1] == '$' {
		scanner.Tokens = append(scanner.Tokens, Token{Type: ARG, Lexeme: "$"})
		return
	}

	envVar := scanner.input[scanner.start+1 : scanner.current]
	scanner.Tokens = append(scanner.Tokens, Token{Type: ENVVAR, Lexeme: envVar, Value: os.Getenv(envVar)})
}
