package parser

import (
	"strconv"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/token"
)

func (p *Parser) expectAnyWord(what string) string {
	cur := p.cur()
	if !isWordToken(cur) {
		p.fail("esperado %s, encontrado %s", what, p.describe(cur))
	}
	p.pos++
	return cur.Literal
}

func (p *Parser) parseEnvironment() *ast.EnvironmentDivision {
	env := &ast.EnvironmentDivision{}
	p.guard(func() {
		p.expectWord("ENVIRONMENT")
		p.expect(token.DIVISION)
		p.expect(token.PERIOD)
	})
	for !p.atEOF() && !p.isDivision("DATA") && !p.isDivision("PROCEDURE") {
		p.guard(func() { p.envSentence(env) })
	}
	return env
}

func (p *Parser) envSentence(env *ast.EnvironmentDivision) {
	switch {
	case (p.isWord("CONFIGURATION") || p.isWord("INPUT-OUTPUT")) && p.peek(1).Type == token.SECTION:
		p.pos += 2
		p.expect(token.PERIOD)
	case p.isWord("SOURCE-COMPUTER") || p.isWord("OBJECT-COMPUTER"):
		p.syncPeriod()
		p.syncPeriod()
	case p.isWord("SPECIAL-NAMES") || p.isWord("FILE-CONTROL"):
		p.pos++
		p.expect(token.PERIOD)
	case p.isWord("DECIMAL-POINT"):
		p.pos++
		p.acceptWord("IS")
		p.expectWord("COMMA")
		p.expect(token.PERIOD)
		env.DecimalComma = true
	case p.isWord("SELECT"):
		env.Files = append(env.Files, p.parseSelect())
	default:
		p.fail("sentença inesperada na ENVIRONMENT DIVISION: %s", p.describe(p.cur()))
	}
}

func (p *Parser) parseSelect() *ast.Select {
	sel := &ast.Select{Pos: p.here()}
	p.pos++
	sel.Name = p.expectIdent("nome do arquivo").Text
	p.expectWord("ASSIGN")
	p.acceptWord("TO")
	switch cur := p.cur(); cur.Type {
	case token.STRING:
		sel.Assign = unquote(cur.Literal)
	case token.IDENT:
		sel.Assign = cur.Literal
		sel.AssignIsName = true
	default:
		p.fail("esperado destino físico do arquivo, encontrado %s", p.describe(cur))
	}
	p.pos++

	for p.cur().Type != token.PERIOD {
		switch {
		case p.acceptWord("ORGANIZATION"):
			p.acceptWord("IS")
			if p.acceptWord("LINE") {
				p.expectWord("SEQUENTIAL")
				sel.Organization = "LINE SEQUENTIAL"
			} else {
				sel.Organization = p.expectAnyWord("organização do arquivo")
			}
		case p.acceptWord("ACCESS"):
			p.acceptWord("MODE")
			p.acceptWord("IS")
			sel.Access = p.expectAnyWord("modo de acesso")
		case p.acceptWord("RECORD"):
			p.acceptWord("KEY")
			p.acceptWord("IS")
			n := p.expectIdent("chave do registro")
			sel.RecordKey = &n
		case p.acceptWord("ALTERNATE"):
			p.expectWord("RECORD")
			p.acceptWord("KEY")
			p.acceptWord("IS")
			n := p.expectIdent("chave alternativa")
			sel.AltKeys = append(sel.AltKeys, &n)
			if p.acceptWord("WITH") {
				p.expectWord("DUPLICATES")
			}
		case p.acceptWord("FILE"):
			p.expectWord("STATUS")
			p.acceptWord("IS")
			n := p.expectIdent("variável de FILE STATUS")
			sel.Status = &n
		default:
			p.fail("cláusula inesperada em SELECT: %s", p.describe(p.cur()))
		}
	}
	p.expect(token.PERIOD)
	return sel
}

type dataState struct {
	section string
	fd      *ast.FileDesc
	stack   []*ast.DataItem
}

func (p *Parser) parseData() *ast.DataDivision {
	d := &ast.DataDivision{}
	st := &dataState{}
	p.guard(func() {
		p.expectWord("DATA")
		p.expect(token.DIVISION)
		p.expect(token.PERIOD)
	})
	for !p.atEOF() && !p.isDivision("PROCEDURE") {
		p.guard(func() { p.dataSentence(d, st) })
	}
	return d
}

func (p *Parser) dataSentence(d *ast.DataDivision, st *dataState) {
	switch {
	case (p.isWord("FILE") || p.isWord("WORKING-STORAGE") || p.isWord("LINKAGE")) && p.peek(1).Type == token.SECTION:
		st.section = p.cur().Literal
		st.stack = nil
		st.fd = nil
		p.pos += 2
		p.expect(token.PERIOD)
	case p.isWord("FD"):
		fd := &ast.FileDesc{Pos: p.here()}
		p.pos++
		fd.Name = p.expectIdent("nome do arquivo").Text
		st.fd = fd
		st.stack = nil
		d.Files = append(d.Files, fd)
		p.syncPeriod()
	case p.cur().Type == token.NUMBER:
		item := p.parseDataItem()
		p.attach(d, st, item)
	default:
		p.fail("sentença inesperada na DATA DIVISION: %s", p.describe(p.cur()))
	}
}

func (p *Parser) attach(d *ast.DataDivision, st *dataState, item *ast.DataItem) {
	if item.Level == 1 || item.Level == 77 {
		st.stack = []*ast.DataItem{item}
		switch {
		case st.section == "FILE" && st.fd != nil:
			st.fd.Records = append(st.fd.Records, item)
		case st.section == "LINKAGE":
			d.Linkage = append(d.Linkage, item)
		default:
			d.Working = append(d.Working, item)
		}
		return
	}
	for len(st.stack) > 0 && st.stack[len(st.stack)-1].Level >= item.Level {
		st.stack = st.stack[:len(st.stack)-1]
	}
	if len(st.stack) == 0 {
		p.diags = append(p.diags, errorAt(item.Pos, "item de nível %02d sem grupo pai (nível 01)", item.Level))
		return
	}
	parent := st.stack[len(st.stack)-1]
	parent.Children = append(parent.Children, item)
	st.stack = append(st.stack, item)
}

var clauseWords = map[string]bool{
	"VALUE": true, "VALUES": true, "USAGE": true, "OCCURS": true, "INDEXED": true,
	"COMP": true, "COMP-3": true, "COMPUTATIONAL": true, "BINARY": true,
}

var figurative = map[string]string{
	"ZERO": "ZERO", "ZEROS": "ZERO", "ZEROES": "ZERO",
	"SPACE": "SPACE", "SPACES": "SPACE",
	"LOW-VALUE": "LOW-VALUE", "LOW-VALUES": "LOW-VALUE",
	"HIGH-VALUE": "HIGH-VALUE", "HIGH-VALUES": "HIGH-VALUE",
	"QUOTE": "QUOTE", "QUOTES": "QUOTE",
}

func (p *Parser) parseDataItem() *ast.DataItem {
	pos := p.here()
	lvlTok := p.expect(token.NUMBER)
	level, err := strconv.Atoi(lvlTok.Literal)
	if err != nil || level < 1 || (level > 49 && level != 77) {
		p.pos--
		p.fail("nível de dado não suportado: %s", lvlTok.Literal)
	}
	item := &ast.DataItem{Pos: pos, Level: level}

	if cur := p.cur(); cur.Type == token.IDENT && !clauseWords[cur.Literal] {
		if cur.Literal != "FILLER" {
			item.Name = cur.Literal
		}
		p.pos++
	}

	for p.cur().Type != token.PERIOD {
		cur := p.cur()
		switch {
		case cur.Type == token.PIC || cur.Type == token.PICTURE:
			p.pos++
			p.acceptWord("IS")
			item.Pic = p.expect(token.PICTURE_STRING).Literal
		case p.acceptWord("USAGE"):
			p.acceptWord("IS")
			item.Usage = p.expectAnyWord("tipo de uso")
		case p.isWord("COMP") || p.isWord("COMP-3") || p.isWord("COMPUTATIONAL") || p.isWord("BINARY"):
			item.Usage = cur.Literal
			p.pos++
		case p.isWord("VALUE") || p.isWord("VALUES"):
			p.pos++
			if !p.acceptWord("IS") {
				p.acceptWord("ARE")
			}
			item.Value = p.parseValue()
		case p.acceptWord("OCCURS"):
			n, err := strconv.Atoi(p.expect(token.NUMBER).Literal)
			if err != nil {
				p.pos--
				p.fail("quantidade inválida em OCCURS")
			}
			item.Occurs = n
			p.acceptWord("TIMES")
			if p.acceptWord("INDEXED") {
				p.acceptWord("BY")
				for p.cur().Type == token.IDENT && !clauseWords[p.cur().Literal] {
					item.Indexes = append(item.Indexes, p.cur().Literal)
					p.pos++
				}
			}
		default:
			p.fail("cláusula inesperada na declaração de dados: %s", p.describe(cur))
		}
	}
	p.expect(token.PERIOD)
	return item
}

func (p *Parser) parseValue() *ast.Value {
	cur := p.cur()
	switch {
	case cur.Type == token.NUMBER:
		p.pos++
		return &ast.Value{Kind: ast.ValueNumber, Text: cur.Literal}
	case cur.Type == token.STRING:
		p.pos++
		return &ast.Value{Kind: ast.ValueString, Text: unquote(cur.Literal)}
	case isWordToken(cur) && figurative[cur.Literal] != "":
		p.pos++
		return &ast.Value{Kind: ast.ValueFigurative, Text: figurative[cur.Literal]}
	}
	p.fail("esperado valor inicial, encontrado %s", p.describe(cur))
	return nil
}
