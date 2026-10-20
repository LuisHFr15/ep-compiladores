package lexer

import (
	"strings"

	"ep-compiladores/internal/token"
)

type Lexer struct {
	input string
}

func New(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token

	lines := strings.Split(l.input, "\n")
	for i, raw := range lines {
		lineNo := i + 1
		line := strings.TrimRight(raw, "\r")

		if len(line) >= 7 && (line[6] == '*' || line[6] == '/') {
			continue
		}

		code := codeArea(line)
		col := 8
		j := 0
		for j < len(code) {
			c := code[j]
			allowUnderscore := afterProgramID(tokens)
			switch {
			case c == ' ' || c == '\t':
				j++
				col++
			case c == '.':
				tokens = append(tokens, token.Token{Type: token.PERIOD, Literal: ".", Line: lineNo, Col: col})
				j++
				col++
			case isWordChar(c, allowUnderscore):
				start, startCol := j, col
				for j < len(code) && isWordChar(code[j], allowUnderscore) {
					j++
					col++
				}
				word := code[start:j]
				tokens = append(tokens, token.Token{Type: token.LookupWord(word), Literal: word, Line: lineNo, Col: startCol})
			default:
				tokens = append(tokens, token.Token{Type: token.ILLEGAL, Literal: string(c), Line: lineNo, Col: col})
				j++
				col++
			}
		}
	}

	tokens = append(tokens, token.Token{Type: token.EOF, Line: len(lines) + 1, Col: 1})
	return tokens
}

func codeArea(line string) string {
	if len(line) < 8 {
		return ""
	}
	end := 72
	if len(line) < end {
		end = len(line)
	}
	return line[7:end]
}

func isWordChar(c byte, allowUnderscore bool) bool {
	if c == '_' {
		return allowUnderscore
	}
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
}

func afterProgramID(tokens []token.Token) bool {
	n := len(tokens)
	return n >= 2 && tokens[n-2].Type == token.PROGRAM_ID && tokens[n-1].Type == token.PERIOD
}
