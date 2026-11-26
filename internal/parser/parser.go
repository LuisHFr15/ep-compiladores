package parser

import (
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/diagnostic"
	"ep-compiladores/internal/token"
)

type bailout struct{}

type relCtx struct {
	left ast.Expr
	op   string
}

type Parser struct {
	tokens []token.Token
	pos    int
	diags  []diagnostic.Diagnostic
	rel    *relCtx
}

func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) cur() token.Token {
	return p.peek(0)
}

func (p *Parser) peek(n int) token.Token {
	i := p.pos + n
	if i >= len(p.tokens) {
		i = len(p.tokens) - 1
	}
	return p.tokens[i]
}

func (p *Parser) atEOF() bool {
	return p.cur().Type == token.EOF
}

func (p *Parser) here() ast.Pos {
	t := p.cur()
	return ast.Pos{Line: t.Line, Col: t.Col}
}

func (p *Parser) fail(format string, args ...any) {
	t := p.cur()
	p.diags = append(p.diags, diagnostic.Diagnostic{
		Severity: diagnostic.Error,
		Line:     t.Line,
		Col:      t.Col,
		Message:  fmt.Sprintf(format, args...),
	})
	panic(bailout{})
}

func (p *Parser) describe(t token.Token) string {
	if t.Type == token.EOF {
		return "fim do arquivo"
	}
	return fmt.Sprintf("%q", t.Literal)
}

func (p *Parser) guard(f func()) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(bailout); !ok {
				panic(r)
			}
			p.syncPeriod()
		}
	}()
	f()
}

func (p *Parser) syncPeriod() {
	for !p.atEOF() && p.cur().Type != token.PERIOD {
		p.pos++
	}
	if p.cur().Type == token.PERIOD {
		p.pos++
	}
}

func (p *Parser) try(f func()) (ok bool) {
	savePos, saveDiags, saveRel := p.pos, len(p.diags), p.rel
	defer func() {
		if r := recover(); r != nil {
			if _, isBail := r.(bailout); !isBail {
				panic(r)
			}
			p.pos, p.diags, p.rel = savePos, p.diags[:saveDiags], saveRel
			ok = false
		}
	}()
	f()
	return true
}

func isWordToken(t token.Token) bool {
	return t.Type == token.IDENT || token.IsKeyword(t.Type)
}

func (p *Parser) isWord(w string) bool {
	t := p.cur()
	return isWordToken(t) && t.Literal == w
}

func (p *Parser) isWordAt(n int, w string) bool {
	t := p.peek(n)
	return isWordToken(t) && t.Literal == w
}

func (p *Parser) acceptWord(w string) bool {
	if p.isWord(w) {
		p.pos++
		return true
	}
	return false
}

func (p *Parser) expectWord(w string) {
	if !p.acceptWord(w) {
		p.fail("esperado %s, encontrado %s", w, p.describe(p.cur()))
	}
}

func (p *Parser) seq(words ...string) bool {
	for i, w := range words {
		if !p.isWordAt(i, w) {
			return false
		}
	}
	return true
}

func (p *Parser) acceptSeq(words ...string) bool {
	if p.seq(words...) {
		p.pos += len(words)
		return true
	}
	return false
}

func (p *Parser) accept(t token.Type) bool {
	if p.cur().Type == t {
		p.pos++
		return true
	}
	return false
}

func (p *Parser) expect(t token.Type) token.Token {
	cur := p.cur()
	if cur.Type != t {
		p.fail("esperado %s, encontrado %s", t, p.describe(cur))
	}
	p.pos++
	return cur
}

func (p *Parser) expectIdent(what string) ast.Name {
	cur := p.cur()
	if cur.Type != token.IDENT {
		p.fail("esperado %s, encontrado %s", what, p.describe(cur))
	}
	p.pos++
	return ast.Name{Pos: ast.Pos{Line: cur.Line, Col: cur.Col}, Text: cur.Literal}
}

func (p *Parser) isDivision(name string) bool {
	return p.isWord(name) && p.peek(1).Type == token.DIVISION
}

func (p *Parser) ParseProgram() (*ast.Program, []diagnostic.Diagnostic) {
	prog := &ast.Program{}

	p.guard(func() { prog.Identification = p.parseIdentification() })

	if p.isDivision("ENVIRONMENT") {
		prog.Environment = p.parseEnvironment()
	}
	if p.isDivision("DATA") {
		prog.Data = p.parseData()
	}
	if p.isDivision("PROCEDURE") {
		prog.Procedure = p.parseProcedure()
	} else if !p.atEOF() || prog.Identification != nil {
		p.guard(func() { p.fail("esperado PROCEDURE DIVISION, encontrado %s", p.describe(p.cur())) })
	}

	return prog, p.diags
}

var identParagraphs = map[string]bool{
	"AUTHOR": true, "INSTALLATION": true, "DATE-WRITTEN": true,
	"DATE-COMPILED": true, "SECURITY": true, "REMARKS": true,
}

func (p *Parser) parseIdentification() *ast.IdentificationDivision {
	pos := p.here()
	p.expect(token.IDENTIFICATION)
	p.expect(token.DIVISION)
	p.expect(token.PERIOD)
	p.expect(token.PROGRAM_ID)
	p.expect(token.PERIOD)
	name := p.expectIdent("nome do programa")
	p.expect(token.PERIOD)

	for p.cur().Type == token.IDENT && identParagraphs[p.cur().Literal] && p.peek(1).Type == token.PERIOD {
		p.syncPeriod()
		p.syncPeriod()
	}
	return &ast.IdentificationDivision{Pos: pos, ProgramID: name.Text}
}

func unquote(lexeme string) string {
	if lexeme == "" {
		return ""
	}
	q := lexeme[:1]
	body := lexeme[1:]
	body = strings.TrimSuffix(body, q)
	return strings.ReplaceAll(body, q+q, q)
}
