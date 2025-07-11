package main

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	COMMENT
	NEWLINE
	INST
	INTEGER
	FLOAT
	ARGS
	STRING
	FUNCTION
)

type Token struct {
	Type    TokenType
	Literal interface{}
	Line    int
}

type Lexer struct {
	lines []string
	line  int
	pos   int
	char  rune
}

func NewLexer(input string) *Lexer {
	lines := strings.Split(input, "\n")
	l := &Lexer{lines: lines, line: 0, pos: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.line >= len(l.lines) {
		l.char = 0
		return
	}
	if l.pos >= len(l.lines[l.line]) {
		l.char = '\n'
	} else {
		l.char = rune(l.lines[l.line][l.pos])
	}
}

func (l *Lexer) advance() {
	if l.char == '\n' {
		l.line++
		l.pos = 0
	} else {
		l.pos++
	}
	l.readChar()
}

func (l *Lexer) skipWhitespace() {
	for l.char == ' ' || l.char == '\t' {
		l.advance()
	}
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.char {
	case '\n':
		tok = Token{Type: NEWLINE, Literal: "", Line: l.line}
	case 0:
		tok = Token{Type: EOF, Literal: "", Line: l.line}
	case ';':
		comment := l.readComment()
		tok = Token{Type: COMMENT, Literal: comment, Line: l.line}

	default:
		if unicode.IsLetter(l.char) {
			literal := l.readIdentifier()
			tok = Token{Type: INST, Literal: literal, Line: l.line}
			return tok
		} else if unicode.IsDigit(l.char) || l.char == '-' {
			literal, tokType := l.readNumber()
			tok = Token{Type: tokType, Literal: literal, Line: l.line}
			return tok
		} else if l.char == '"' {
			read_string := l.readString()
			tok = Token{Type: STRING, Literal: read_string, Line: l.line}
		} else if l.char == '%' {
			tok = Token{Type: ARGS, Literal: "PAR", Line: l.line}
		} else if l.char == '_' {
			function := l.generateFunction()
			tok = Token{Type: FUNCTION, Literal: function, Line: l.line}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.char), Line: l.line}
			panic("illegal")
		}
	}
	l.advance()
	return tok
}
func (l *Lexer) readString() string {
	//adv to ign first quote
	l.advance()
	start := l.pos

	for l.char != '"' && l.char != '\n' {
		l.advance()
	}
	return l.lines[l.line][start:l.pos]
}
func (l *Lexer) readIdentifier() string {
	start := l.pos
	for unicode.IsLetter(l.char) {
		l.advance()
	}
	return l.lines[l.line][start:l.pos]
}

func (l *Lexer) readNumber() (string, TokenType) {
	start := l.pos
	tokType := INTEGER
	if l.char == '-' {
		l.advance()
	}
	for unicode.IsDigit(l.char) {
		l.advance()
	}
	if l.char == '.' {
		tokType = FLOAT
		l.advance()
		for unicode.IsDigit(l.char) {
			l.advance()
		}
	}
	return l.lines[l.line][start:l.pos], tokType
}

func (l *Lexer) readComment() string {
	start := l.pos
	for l.char != '\n' && l.char != 0 {
		l.advance()
	}
	return l.lines[l.line][start:l.pos]
}

func (l *Lexer) generateFunction() []string {
	l.advance()
	start := l.line
	for l.char != '_' && l.char != '"' && l.char != 0 {
		l.NextToken()
	}
	return l.lines[start:l.line]
}
