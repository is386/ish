package scanning

import (
	"strings"
)

type TokenType int

const (
	ARG TokenType = iota
	STRING
)

type Token struct {
	Type   TokenType
	Lexeme string
}

type Scanner struct {
	input   string
	start   int
	current int
	Tokens  []Token
}

func NewScanner(input string) *Scanner {
	return &Scanner{input: input}
}

func (scanner *Scanner) Scan() {
	for !scanner.isAtEnd() {
		scanner.start = scanner.current
		scanner.parseToken()
	}
}

func (scanner *Scanner) parseToken() {
	c := scanner.advance()
	switch c {
	case ' ':
		return
	case '"':
		scanner.scanString()
	default:
		scanner.scanArg()
	}
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
	scanner.Tokens = append(scanner.Tokens, Token{Type: ARG, Lexeme: strings.TrimSpace(arg)})
}

func (scanner *Scanner) scanString() {
	for !scanner.isAtEnd() && scanner.peek() != '"' {
		scanner.advance()
	}
	// TODO: Handle case where there is a not closed string
	scanner.advance()

	str := scanner.input[scanner.start:scanner.current]
	scanner.Tokens = append(scanner.Tokens, Token{Type: ARG, Lexeme: str})
}
