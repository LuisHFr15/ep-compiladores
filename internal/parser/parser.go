package parser

import (
	"fmt"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/token"
)

type Parser struct {
	tokens []token.Token
	pos    int
}

func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) current() token.Token {
	return p.tokens[p.pos]
}

func (p *Parser) expect(t token.Type) (token.Token, error) {
	cur := p.current()
	if cur.Type != t {
		return cur, fmt.Errorf("erro sintático (linha %d): esperado %s, encontrado %q", cur.Line, t, cur.Literal)
	}
	p.pos++
	return cur, nil
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	id, err := p.parseIdentificationDivision()
	if err != nil {
		return nil, err
	}
	return &ast.Program{Identification: id}, nil
}

func (p *Parser) parseIdentificationDivision() (*ast.IdentificationDivision, error) {
	for _, t := range []token.Type{token.IDENTIFICATION, token.DIVISION, token.PERIOD, token.PROGRAM_ID, token.PERIOD} {
		if _, err := p.expect(t); err != nil {
			return nil, err
		}
	}

	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.PERIOD); err != nil {
		return nil, err
	}

	return &ast.IdentificationDivision{ProgramID: name.Literal}, nil
}
