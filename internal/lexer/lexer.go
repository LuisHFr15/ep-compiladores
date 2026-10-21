package lexer

import (
	"fmt"
	"sort"
	"strings"

	"ep-compiladores/internal/diagnostic"
	"ep-compiladores/internal/token"
)

const (
	codeStart     = 8
	codeEnd       = 72
	maxIdentifier = 30
)

type openLiteral struct {
	quote rune
	text  []rune
	line  int
	col   int
}

type Lexer struct {
	lines   []string
	stream  *TokenStream
	picture bool
	debug   bool
	open    *openLiteral
}

func New(input string, debug bool) *Lexer {
	return &Lexer{lines: strings.Split(input, "\n"), stream: &TokenStream{}, debug: debug}
}

func (l *Lexer) Tokenize() *TokenStream {
	for i, raw := range l.lines {
		lineNo := i + 1
		line := []rune(strings.TrimRight(raw, "\r"))

		indicator := ' '
		if len(line) >= 7 {
			indicator = line[6]
		}
		switch indicator {
		case '*', '/':
			continue
		case 'D', 'd':
			if !l.debug {
				continue
			}
		case ' ', '-':
		default:
			l.errorf(lineNo, 7, "indicador inválido na coluna 7: '%c'", indicator)
		}

		code := codeArea(line)
		pos := 0
		if l.open != nil {
			if indicator == '-' {
				pos = l.resumeLiteral(code, lineNo)
			} else {
				l.closeUnterminated()
			}
		}
		l.scan(code, pos, lineNo)
	}

	if l.open != nil {
		l.closeUnterminated()
	}
	l.emit(token.EOF, "", len(l.lines), 1)

	d := l.stream.Diagnostics
	sort.SliceStable(d, func(i, j int) bool {
		if d[i].Line != d[j].Line {
			return d[i].Line < d[j].Line
		}
		return d[i].Col < d[j].Col
	})
	return l.stream
}

func (l *Lexer) scan(code []rune, i, lineNo int) {
	for i < len(code) {
		c := code[i]
		col := i + codeStart

		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '"' || c == '\'':
			l.open = &openLiteral{quote: c, text: []rune{c}, line: lineNo, col: col}
			i = l.readLiteral(code, i+1)
		case l.picture:
			i = l.readPicture(code, i, lineNo)
		case isDigit(c):
			i = l.readWord(code, i, lineNo)
		case isLetter(c) || (c == '_' && l.programNameContext()):
			i = l.readWord(code, i, lineNo)
		case (c == '+' || c == '-') && i+1 < len(code) && isDigit(code[i+1]) && !l.prevIsOperand():
			i = l.readWord(code, i, lineNo)
		case c == '-' && i+1 < len(code) && isLetter(code[i+1]):
			i = l.readWord(code, i, lineNo)
		case strings.ContainsRune("+-*/=<>", c):
			op := string(c)
			if i+1 < len(code) && ((c == '*' && code[i+1] == '*') || ((c == '<' || c == '>') && code[i+1] == '=')) {
				op += string(code[i+1])
			}
			l.emit(token.OPERATOR, op, lineNo, col)
			i += len(op)
		case c == '.':
			l.emit(token.PERIOD, ".", lineNo, col)
			i++
		case c == ',':
			l.emit(token.COMMA, ",", lineNo, col)
			i++
		case c == ';':
			l.emit(token.SEMICOLON, ";", lineNo, col)
			i++
		case c == '(':
			l.emit(token.LPAREN, "(", lineNo, col)
			i++
		case c == ')':
			l.emit(token.RPAREN, ")", lineNo, col)
			i++
		default:
			l.errorf(lineNo, col, "caractere inválido '%c'", c)
			i++
		}
	}
}

func (l *Lexer) readWord(code []rune, i, lineNo int) int {
	start := i
	col := i + codeStart
	allowUnderscore := l.programNameContext()

	i++
	for i < len(code) && isWordChar(code[i], allowUnderscore) {
		i++
	}

	if isNumber(code[start:i]) {
		dots := 0
		for i+1 < len(code) && code[i] == '.' && isDigit(code[i+1]) {
			dots++
			i++
			for i < len(code) && isDigit(code[i]) {
				i++
			}
		}
		lexeme := string(code[start:i])
		if dots > 1 {
			l.errorf(lineNo, col, "literal numérico inválido '%s': mais de um ponto decimal", lexeme)
		}
		l.emit(token.NUMBER, lexeme, lineNo, col)
		return i
	}

	word := strings.ToUpper(string(code[start:i]))
	t := token.LookupWord(word)
	if t == token.IDENT {
		l.validateIdentifier(word, lineNo, col)
	}
	l.emit(t, word, lineNo, col)
	if t == token.PIC || t == token.PICTURE {
		l.picture = true
	}
	return i
}

func (l *Lexer) readPicture(code []rune, i, lineNo int) int {
	start := i
	for i < len(code) && code[i] != ' ' && code[i] != '\t' {
		i++
	}
	end := i
	if end-start > 1 && code[end-1] == '.' {
		end--
	}

	l.picture = false
	if end > start && !(end-start == 1 && code[start] == '.') {
		l.emit(token.PICTURE_STRING, strings.ToUpper(string(code[start:end])), lineNo, start+codeStart)
	} else {
		end = start
	}
	if end < i {
		l.emit(token.PERIOD, ".", lineNo, end+codeStart)
	}
	return i
}

func (l *Lexer) readLiteral(code []rune, i int) int {
	for i < len(code) {
		c := code[i]
		l.open.text = append(l.open.text, c)
		i++
		if c == l.open.quote {
			if i < len(code) && code[i] == l.open.quote {
				l.open.text = append(l.open.text, code[i])
				i++
				continue
			}
			l.emit(token.STRING, string(l.open.text), l.open.line, l.open.col)
			l.open = nil
			return i
		}
	}
	for n := len(code); n < codeEnd-codeStart+1; n++ {
		l.open.text = append(l.open.text, ' ')
	}
	return i
}

func (l *Lexer) resumeLiteral(code []rune, lineNo int) int {
	i := 0
	for i < len(code) && (code[i] == ' ' || code[i] == '\t') {
		i++
	}
	if i < len(code) && code[i] == l.open.quote {
		return l.readLiteral(code, i+1)
	}
	l.errorf(lineNo, i+codeStart, "continuação de literal deve iniciar com aspa (%c)", l.open.quote)
	l.closeUnterminated()
	return i
}

func (l *Lexer) closeUnterminated() {
	o := l.open
	l.open = nil
	l.errorf(o.line, o.col, "literal alfanumérico não terminado: falta a aspa de fechamento (%c)", o.quote)
	l.emit(token.STRING, strings.TrimRight(string(o.text), " "), o.line, o.col)
}

func (l *Lexer) validateIdentifier(word string, lineNo, col int) {
	if n := len([]rune(word)); n > maxIdentifier {
		l.errorf(lineNo, col, "identificador '%s' excede %d caracteres (tem %d)", word, maxIdentifier, n)
	}
	if strings.HasPrefix(word, "-") {
		l.errorf(lineNo, col, "identificador '%s' não pode iniciar com hífen", word)
	}
	if strings.HasSuffix(word, "-") {
		l.errorf(lineNo, col, "identificador '%s' não pode terminar com hífen", word)
	}
}

func (l *Lexer) programNameContext() bool {
	t := l.stream.Tokens
	n := len(t)
	if n < 2 {
		return false
	}
	a, b := t[n-2], t[n-1]
	return (a.Type == token.PROGRAM_ID && b.Type == token.PERIOD) || (a.Literal == "END" && b.Literal == "PROGRAM")
}

func (l *Lexer) prevIsOperand() bool {
	t := l.stream.Tokens
	if len(t) == 0 {
		return false
	}
	switch t[len(t)-1].Type {
	case token.IDENT, token.NUMBER, token.STRING, token.RPAREN:
		return true
	}
	return false
}

func (l *Lexer) emit(t token.Type, literal string, line, col int) {
	l.stream.Tokens = append(l.stream.Tokens, token.Token{Type: t, Literal: literal, Line: line, Col: col})
}

func (l *Lexer) errorf(line, col int, format string, args ...any) {
	l.stream.Diagnostics = append(l.stream.Diagnostics, diagnostic.Diagnostic{
		Severity: diagnostic.Error,
		Line:     line,
		Col:      col,
		Message:  fmt.Sprintf(format, args...),
	})
}

func codeArea(line []rune) []rune {
	if len(line) < codeStart {
		return nil
	}
	end := codeEnd
	if len(line) < end {
		end = len(line)
	}
	return line[codeStart-1 : end]
}

func isNumber(r []rune) bool {
	start := 0
	if r[0] == '+' || r[0] == '-' {
		start = 1
	}
	if start == len(r) {
		return false
	}
	for _, c := range r[start:] {
		if !isDigit(c) {
			return false
		}
	}
	return true
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isWordChar(c rune, allowUnderscore bool) bool {
	if c == '_' {
		return allowUnderscore
	}
	return isLetter(c) || isDigit(c) || c == '-'
}
