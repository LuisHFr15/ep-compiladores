package parser

import (
	"fmt"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/diagnostic"
	"ep-compiladores/internal/token"
)

func errorAt(pos ast.Pos, format string, args ...any) diagnostic.Diagnostic {
	return diagnostic.Diagnostic{
		Severity: diagnostic.Error,
		Line:     pos.Line,
		Col:      pos.Col,
		Message:  fmt.Sprintf(format, args...),
	}
}

func (p *Parser) isFigurative() bool {
	t := p.cur()
	return isWordToken(t) && figurative[t.Literal] != ""
}

func (p *Parser) parseOperand() ast.Expr {
	t := p.cur()
	pos := p.here()
	switch {
	case t.Type == token.NUMBER:
		p.pos++
		return &ast.NumLit{Pos: pos, Text: t.Literal}
	case t.Type == token.STRING:
		p.pos++
		return &ast.StrLit{Pos: pos, Value: unquote(t.Literal)}
	case t.Type == token.LPAREN:
		p.pos++
		e := p.parseExpr()
		p.expect(token.RPAREN)
		return e
	case t.Type == token.OPERATOR && (t.Literal == "-" || t.Literal == "+"):
		p.pos++
		x := p.parseOperand()
		if t.Literal == "-" {
			return &ast.Neg{Pos: pos, X: x}
		}
		return x
	case p.isFigurative():
		p.pos++
		return &ast.Figurative{Pos: pos, Kind: figurative[t.Literal]}
	case p.isWord("FUNCTION"):
		p.pos++
		name := p.expectAnyWord("nome da função")
		call := &ast.FuncCall{Pos: pos, Name: name}
		p.expect(token.LPAREN)
		for p.cur().Type != token.RPAREN {
			call.Args = append(call.Args, p.parseExpr())
			p.accept(token.COMMA)
		}
		p.expect(token.RPAREN)
		return call
	case t.Type == token.IDENT:
		return p.parseRef()
	}
	p.fail("esperado operando, encontrado %s", p.describe(t))
	return nil
}

func (p *Parser) parseRef() *ast.Ref {
	pos := p.here()
	name := p.expectIdent("identificador")
	ref := &ast.Ref{Pos: pos, Name: name.Text}
	for p.isWord("OF") || p.isWord("IN") {
		p.pos++
		ref.Quals = append(ref.Quals, p.expectIdent("qualificador").Text)
	}
	if p.cur().Type == token.LPAREN {
		p.pos++
		for p.cur().Type != token.RPAREN {
			ref.Subs = append(ref.Subs, p.parseExpr())
			p.accept(token.COMMA)
		}
		p.expect(token.RPAREN)
	}
	return ref
}

func (p *Parser) parseExpr() ast.Expr {
	left := p.parseTerm()
	for t := p.cur(); t.Type == token.OPERATOR && (t.Literal == "+" || t.Literal == "-"); t = p.cur() {
		pos := p.here()
		p.pos++
		left = &ast.Binary{Pos: pos, Op: t.Literal, L: left, R: p.parseTerm()}
	}
	return left
}

func (p *Parser) parseTerm() ast.Expr {
	left := p.parsePower()
	for t := p.cur(); t.Type == token.OPERATOR && (t.Literal == "*" || t.Literal == "/"); t = p.cur() {
		pos := p.here()
		p.pos++
		left = &ast.Binary{Pos: pos, Op: t.Literal, L: left, R: p.parsePower()}
	}
	return left
}

func (p *Parser) parsePower() ast.Expr {
	base := p.parseOperand()
	if t := p.cur(); t.Type == token.OPERATOR && t.Literal == "**" {
		pos := p.here()
		p.pos++
		return &ast.Binary{Pos: pos, Op: "**", L: base, R: p.parsePower()}
	}
	return base
}

func (p *Parser) parseCond() ast.Cond {
	p.rel = nil
	return p.parseOr()
}

func (p *Parser) parseOr() ast.Cond {
	left := p.parseAnd()
	for p.isWord("OR") {
		pos := p.here()
		p.pos++
		left = &ast.Or{Pos: pos, L: left, R: p.parseAnd()}
	}
	return left
}

func (p *Parser) parseAnd() ast.Cond {
	left := p.parseNot()
	for p.isWord("AND") {
		pos := p.here()
		p.pos++
		left = &ast.And{Pos: pos, L: left, R: p.parseNot()}
	}
	return left
}

func (p *Parser) parseNot() ast.Cond {
	if p.isWord("NOT") {
		pos := p.here()
		p.pos++
		return &ast.Not{Pos: pos, X: p.parseNot()}
	}
	return p.parsePrimaryCond()
}

func (p *Parser) parsePrimaryCond() ast.Cond {
	if p.cur().Type == token.LPAREN {
		var c ast.Cond
		if p.try(func() {
			p.pos++
			c = p.parseOr()
			p.expect(token.RPAREN)
		}) {
			return c
		}
	}
	return p.parseRelation()
}

func (p *Parser) parseRelation() ast.Cond {
	pos := p.here()
	left := p.parseExpr()

	if p.isWord("IS") || p.isWord("NOT") {
		save := p.pos
		p.acceptWord("IS")
		neg := p.acceptWord("NOT")
		if class := p.cur().Literal; p.isWord("NUMERIC") || p.isWord("ALPHABETIC") {
			p.pos++
			return &ast.ClassCond{Pos: pos, X: left, Class: class, Neg: neg}
		}
		p.pos = save
	}

	if op := p.parseRelOp(); op != "" {
		right := p.parseExpr()
		p.rel = &relCtx{left: left, op: op}
		return &ast.Rel{Pos: pos, L: left, Op: op, R: right}
	}

	if p.rel != nil {
		return &ast.Rel{Pos: pos, L: p.rel.left, Op: p.rel.op, R: left}
	}
	p.fail("esperado operador relacional, encontrado %s", p.describe(p.cur()))
	return nil
}

var negated = map[string]string{"=": "<>", "<>": "=", ">": "<=", "<": ">=", ">=": "<", "<=": ">"}

func (p *Parser) parseRelOp() string {
	save := p.pos
	p.acceptWord("IS")
	neg := p.acceptWord("NOT")

	op := ""
	t := p.cur()
	switch {
	case t.Type == token.OPERATOR && (t.Literal == "=" || t.Literal == ">" || t.Literal == "<" || t.Literal == ">=" || t.Literal == "<="):
		op = t.Literal
		p.pos++
	case p.isWord("EQUAL"):
		op = "="
		p.pos++
		p.acceptWord("TO")
	case p.isWord("GREATER"):
		op = ">"
		p.pos++
		p.acceptWord("THAN")
		if p.seq("OR", "EQUAL") {
			p.pos += 2
			p.acceptWord("TO")
			op = ">="
		}
	case p.isWord("LESS"):
		op = "<"
		p.pos++
		p.acceptWord("THAN")
		if p.seq("OR", "EQUAL") {
			p.pos += 2
			p.acceptWord("TO")
			op = "<="
		}
	}
	if op == "" {
		p.pos = save
		return ""
	}
	if neg {
		op = negated[op]
	}
	return op
}
