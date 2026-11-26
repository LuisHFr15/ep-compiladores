package parser

import (
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/token"
)

var verbs = map[string]bool{
	"DISPLAY": true, "ACCEPT": true, "MOVE": true, "ADD": true, "SUBTRACT": true,
	"MULTIPLY": true, "DIVIDE": true, "COMPUTE": true, "IF": true, "EVALUATE": true,
	"PERFORM": true, "GO": true, "STOP": true, "EXIT": true, "CONTINUE": true,
	"GOBACK": true, "SET": true, "SEARCH": true, "INSPECT": true, "OPEN": true,
	"CLOSE": true, "READ": true, "WRITE": true, "START": true, "INITIALIZE": true,
}

var stopWords = map[string]bool{
	"TO": true, "GIVING": true, "FROM": true, "BY": true, "INTO": true,
	"REMAINDER": true, "ON": true, "NOT": true, "AT": true, "WHEN": true,
	"WITH": true, "UNTIL": true, "INVALID": true, "ROUNDED": true, "ELSE": true,
	"END": true, "THEN": true, "KEY": true, "INDEXED": true, "OTHER": true,
	"REPLACING": true,
}

func (p *Parser) isVerb() bool {
	t := p.cur()
	return isWordToken(t) && verbs[t.Literal]
}

func (p *Parser) isOperandStart() bool {
	t := p.cur()
	switch {
	case t.Type == token.STRING || t.Type == token.NUMBER:
		return true
	case t.Type == token.OPERATOR && (t.Literal == "-" || t.Literal == "+"):
		return true
	case !isWordToken(t):
		return false
	}
	w := t.Literal
	return !verbs[w] && !stopWords[w] && !strings.HasPrefix(w, "END-")
}

func (p *Parser) parseProcedure() *ast.ProcedureDivision {
	proc := &ast.ProcedureDivision{Pos: p.here()}
	p.guard(func() {
		p.expectWord("PROCEDURE")
		p.expect(token.DIVISION)
		if p.acceptWord("USING") {
			for p.cur().Type == token.IDENT {
				proc.Using = append(proc.Using, p.cur().Literal)
				p.pos++
			}
		}
		p.expect(token.PERIOD)
	})

	var current *ast.Label
	for !p.atEOF() {
		if p.accept(token.PERIOD) {
			continue
		}
		if p.isWord("END") && p.isWordAt(1, "PROGRAM") {
			p.guard(func() {
				p.pos += 2
				p.expectIdent("nome do programa")
				p.expect(token.PERIOD)
			})
			break
		}
		if label := p.tryLabel(); label != nil {
			current = label
			proc.Labels = append(proc.Labels, label)
			continue
		}
		if current == nil {
			current = &ast.Label{Pos: p.here()}
			proc.Labels = append(proc.Labels, current)
		}
		p.guard(func() {
			stmts := p.parseStatements()
			if len(stmts) == 0 {
				p.fail("comando desconhecido: %s", p.describe(p.cur()))
			}
			current.Body = append(current.Body, stmts...)
			if !p.accept(token.PERIOD) {
				p.fail("esperado '.' ao final da sentença, encontrado %s", p.describe(p.cur()))
			}
		})
	}
	return proc
}

func (p *Parser) tryLabel() *ast.Label {
	t := p.cur()
	if t.Type != token.IDENT || verbs[t.Literal] {
		return nil
	}
	next := p.peek(1)
	switch {
	case next.Type == token.PERIOD:
		p.pos += 2
		return &ast.Label{Pos: ast.Pos{Line: t.Line, Col: t.Col}, Name: t.Literal}
	case next.Type == token.SECTION && p.peek(2).Type == token.PERIOD:
		p.pos += 3
		return &ast.Label{Pos: ast.Pos{Line: t.Line, Col: t.Col}, Name: t.Literal, Section: true}
	}
	return nil
}

func (p *Parser) parseStatements() []ast.Stmt {
	var out []ast.Stmt
	for p.isVerb() {
		out = append(out, p.parseStatement())
	}
	return out
}

func (p *Parser) requireStatements(what string) []ast.Stmt {
	stmts := p.parseStatements()
	if len(stmts) == 0 {
		p.fail("esperado comando após %s, encontrado %s", what, p.describe(p.cur()))
	}
	return stmts
}

func (p *Parser) parseStatement() ast.Stmt {
	pos := p.here()
	verb := p.cur().Literal
	p.pos++

	switch verb {
	case "DISPLAY":
		return p.parseDisplay(pos)
	case "ACCEPT":
		return &ast.Accept{Pos: pos, Target: p.parseRef()}
	case "MOVE":
		return p.parseMove(pos)
	case "ADD":
		return p.parseAdd(pos)
	case "SUBTRACT":
		return p.parseSubtract(pos)
	case "MULTIPLY":
		return p.parseMultiply(pos)
	case "DIVIDE":
		return p.parseDivide(pos)
	case "COMPUTE":
		return p.parseCompute(pos)
	case "IF":
		return p.parseIf(pos)
	case "EVALUATE":
		return p.parseEvaluate(pos)
	case "PERFORM":
		return p.parsePerform(pos)
	case "GO":
		p.acceptWord("TO")
		n := p.expectIdent("nome de parágrafo")
		return &ast.GoTo{Pos: pos, Target: &n}
	case "STOP":
		p.expect(token.RUN)
		return &ast.StopRun{Pos: pos}
	case "GOBACK":
		return &ast.Goback{Pos: pos}
	case "EXIT":
		return &ast.Exit{Pos: pos}
	case "CONTINUE":
		return &ast.Continue{Pos: pos}
	case "SET":
		return p.parseSet(pos)
	case "SEARCH":
		return p.parseSearch(pos)
	case "INSPECT":
		return p.parseInspect(pos)
	case "INITIALIZE":
		return p.parseInitialize(pos)
	case "OPEN":
		return p.parseOpen(pos)
	case "CLOSE":
		return p.parseClose(pos)
	case "READ":
		return p.parseRead(pos)
	case "WRITE":
		return p.parseWrite(pos)
	case "START":
		return p.parseStart(pos)
	}
	p.pos--
	p.fail("comando não suportado: %s", verb)
	return nil
}

func (p *Parser) parseDisplay(pos ast.Pos) ast.Stmt {
	d := &ast.Display{Pos: pos}
	for p.isOperandStart() {
		d.Items = append(d.Items, p.parseOperand())
	}
	if len(d.Items) == 0 {
		p.fail("esperado operando em DISPLAY, encontrado %s", p.describe(p.cur()))
	}
	if p.acceptSeq("WITH", "NO", "ADVANCING") {
		d.NoAdvancing = true
	}
	return d
}

func (p *Parser) parseMove(pos ast.Pos) ast.Stmt {
	m := &ast.Move{Pos: pos, Src: p.parseOperand()}
	p.expectWord("TO")
	m.Targets = p.parseRefs()
	return m
}

func (p *Parser) parseRefs() []*ast.Ref {
	var refs []*ast.Ref
	for p.cur().Type == token.IDENT && p.isOperandStart() {
		refs = append(refs, p.parseRef())
	}
	if len(refs) == 0 {
		p.fail("esperado identificador, encontrado %s", p.describe(p.cur()))
	}
	return refs
}

func (p *Parser) parseOperands() []ast.Expr {
	var ops []ast.Expr
	for p.isOperandStart() {
		ops = append(ops, p.parseOperand())
	}
	if len(ops) == 0 {
		p.fail("esperado operando, encontrado %s", p.describe(p.cur()))
	}
	return ops
}

func (p *Parser) parseSize(endWord string) *ast.Branch {
	b := p.parseBranch("ON", "SIZE", "ERROR")
	p.acceptWord(endWord)
	return b
}

func (p *Parser) parseBranch(words ...string) *ast.Branch {
	var b *ast.Branch
	not := append([]string{"NOT"}, words...)
	for {
		switch {
		case p.acceptSeq(words...):
			if b == nil {
				b = &ast.Branch{}
			}
			b.On = p.requireStatements(strings.Join(words, " "))
		case p.acceptSeq(not...):
			if b == nil {
				b = &ast.Branch{}
			}
			b.Not = p.requireStatements(strings.Join(not, " "))
		default:
			return b
		}
	}
}

func (p *Parser) parseAdd(pos ast.Pos) ast.Stmt {
	a := &ast.Add{Pos: pos, Operands: p.parseOperands()}
	if p.acceptWord("TO") {
		a.To = p.parseOperands()
	}
	if p.acceptWord("GIVING") {
		a.Giving = p.parseRefs()
	}
	a.Size = p.parseSize("END-ADD")
	return a
}

func (p *Parser) parseSubtract(pos ast.Pos) ast.Stmt {
	s := &ast.Subtract{Pos: pos, Operands: p.parseOperands()}
	p.expectWord("FROM")
	s.From = p.parseOperands()
	if p.acceptWord("GIVING") {
		s.Giving = p.parseRefs()
	}
	s.Size = p.parseSize("END-SUBTRACT")
	return s
}

func (p *Parser) parseMultiply(pos ast.Pos) ast.Stmt {
	m := &ast.Multiply{Pos: pos, Operand: p.parseOperand()}
	p.expectWord("BY")
	m.By = p.parseOperands()
	if p.acceptWord("GIVING") {
		m.Giving = p.parseRefs()
	}
	m.Size = p.parseSize("END-MULTIPLY")
	return m
}

func (p *Parser) parseDivide(pos ast.Pos) ast.Stmt {
	d := &ast.Divide{Pos: pos, A: p.parseOperand()}
	switch {
	case p.acceptWord("INTO"):
		d.Into = true
	case p.acceptWord("BY"):
	default:
		p.fail("esperado INTO ou BY, encontrado %s", p.describe(p.cur()))
	}
	d.B = p.parseOperand()
	if p.acceptWord("GIVING") {
		d.Giving = p.parseRefs()
	}
	if p.acceptWord("REMAINDER") {
		d.Remainder = p.parseRef()
	}
	d.Size = p.parseSize("END-DIVIDE")
	return d
}

func (p *Parser) parseCompute(pos ast.Pos) ast.Stmt {
	c := &ast.Compute{Pos: pos}
	for p.cur().Type == token.IDENT && !p.isWord("ROUNDED") {
		c.Targets = append(c.Targets, p.parseRef())
		p.acceptWord("ROUNDED")
	}
	if len(c.Targets) == 0 {
		p.fail("esperado identificador em COMPUTE, encontrado %s", p.describe(p.cur()))
	}
	if t := p.cur(); t.Type != token.OPERATOR || t.Literal != "=" {
		p.fail("esperado '=', encontrado %s", p.describe(t))
	}
	p.pos++
	c.Expr = p.parseExpr()
	c.Size = p.parseSize("END-COMPUTE")
	return c
}

func (p *Parser) parseIf(pos ast.Pos) ast.Stmt {
	s := &ast.If{Pos: pos, Cond: p.parseCond()}
	p.acceptWord("THEN")
	s.Then = p.requireStatements("IF")
	if p.acceptWord("ELSE") {
		s.Else = p.requireStatements("ELSE")
	}
	p.acceptWord("END-IF")
	return s
}

func (p *Parser) parseEvaluate(pos ast.Pos) ast.Stmt {
	e := &ast.Evaluate{Pos: pos, Subject: p.parseOperand()}
	for p.isWord("WHEN") {
		wpos := p.here()
		p.pos++
		if p.acceptWord("OTHER") {
			e.HasOther = true
			e.Other = p.requireStatements("WHEN OTHER")
			continue
		}
		w := &ast.When{Pos: wpos}
		w.Alts = append(w.Alts, p.parseWhenAlt())
		for p.acceptWord("OR") {
			w.Alts = append(w.Alts, p.parseWhenAlt())
		}
		w.Body = p.requireStatements("WHEN")
		e.Whens = append(e.Whens, w)
	}
	if len(e.Whens) == 0 && !e.HasOther {
		p.fail("esperado WHEN em EVALUATE, encontrado %s", p.describe(p.cur()))
	}
	p.acceptWord("END-EVALUATE")
	return e
}

func (p *Parser) parseWhenAlt() ast.WhenAlt {
	save := p.pos
	if p.isWord("IS") || p.isWord("NOT") {
		p.acceptWord("IS")
		neg := p.acceptWord("NOT")
		if p.isWord("NUMERIC") || p.isWord("ALPHABETIC") {
			class := p.cur().Literal
			p.pos++
			return ast.WhenAlt{Class: class, Neg: neg}
		}
		p.pos = save
	}
	if op := p.parseRelOp(); op != "" {
		return ast.WhenAlt{Op: op, Value: p.parseExpr()}
	}
	return ast.WhenAlt{Op: "=", Value: p.parseOperand()}
}

func (p *Parser) parsePerform(pos ast.Pos) ast.Stmt {
	s := &ast.Perform{Pos: pos}
	if p.cur().Type == token.IDENT && !p.isWord("UNTIL") && !p.isWord("VARYING") {
		target := p.expectIdent("nome de parágrafo")
		s.Target = &target
		if p.acceptWord("THRU") || p.acceptWord("THROUGH") {
			thru := p.expectIdent("nome de parágrafo")
			s.Thru = &thru
		}
	}
	if p.acceptWord("VARYING") {
		v := &ast.Varying{Var: p.parseRef()}
		p.expectWord("FROM")
		v.From = p.parseExpr()
		p.expectWord("BY")
		v.By = p.parseExpr()
		s.Varying = v
		if !p.isWord("UNTIL") {
			p.fail("esperado UNTIL em PERFORM VARYING, encontrado %s", p.describe(p.cur()))
		}
	}
	if p.acceptWord("UNTIL") {
		s.Until = p.parseCond()
	}
	if s.Target == nil {
		if s.Until == nil {
			p.fail("PERFORM em linha exige UNTIL ou VARYING; esperado nome de parágrafo, encontrado %s", p.describe(p.cur()))
		}
		s.Body = p.requireStatements("PERFORM")
		p.expectWord("END-PERFORM")
	}
	return s
}

func (p *Parser) parseInitialize(pos ast.Pos) ast.Stmt {
	s := &ast.Initialize{Pos: pos, Targets: p.parseRefs()}
	if !p.acceptWord("REPLACING") {
		return s
	}
	for p.isWord("ALPHANUMERIC") || p.isWord("NUMERIC") || p.isWord("ALPHABETIC") {
		cat := p.cur().Literal
		p.pos++
		p.acceptWord("DATA")
		p.expectWord("BY")
		s.Replacing = append(s.Replacing, ast.InitReplace{Category: cat, Value: p.parseOperand()})
	}
	if len(s.Replacing) == 0 {
		p.fail("esperado ALPHANUMERIC, NUMERIC ou ALPHABETIC em INITIALIZE REPLACING, encontrado %s", p.describe(p.cur()))
	}
	return s
}

func (p *Parser) parseSet(pos ast.Pos) ast.Stmt {
	s := &ast.Set{Pos: pos, Targets: p.parseRefs()}
	p.expectWord("TO")
	s.Value = p.parseOperand()
	return s
}

func (p *Parser) parseSearch(pos ast.Pos) ast.Stmt {
	s := &ast.Search{Pos: pos, Table: p.parseRef()}
	if p.acceptSeq("AT", "END") {
		s.HasAtEnd = true
		s.AtEnd = p.requireStatements("AT END")
	}
	for p.isWord("WHEN") {
		wpos := p.here()
		p.pos++
		w := &ast.SearchWhen{Pos: wpos, Cond: p.parseCond()}
		w.Body = p.requireStatements("WHEN")
		s.Whens = append(s.Whens, w)
		if !s.HasAtEnd && p.acceptSeq("AT", "END") {
			s.HasAtEnd = true
			s.AtEnd = p.requireStatements("AT END")
		}
	}
	if len(s.Whens) == 0 {
		p.fail("esperado WHEN em SEARCH, encontrado %s", p.describe(p.cur()))
	}
	p.acceptWord("END-SEARCH")
	return s
}

func (p *Parser) parseRegion() ast.Region {
	var r ast.Region
	for {
		switch {
		case p.acceptWord("AFTER"):
			p.acceptWord("INITIAL")
			r.After = p.parseOperand()
		case p.acceptWord("BEFORE"):
			p.acceptWord("INITIAL")
			r.Before = p.parseOperand()
		default:
			return r
		}
	}
}

func (p *Parser) parseInspect(pos ast.Pos) ast.Stmt {
	in := &ast.Inspect{Pos: pos}
	if p.isWord("FUNCTION") {
		call, ok := p.parseOperand().(*ast.FuncCall)
		if !ok {
			p.fail("esperado função em INSPECT")
		}
		in.Func = call
	} else {
		in.Target = p.parseRef()
	}
	switch {
	case p.acceptWord("TALLYING"):
		counter := p.parseRef()
		p.expectWord("FOR")
		for {
			switch {
			case p.isWord("CHARACTERS"):
				p.pos++
				in.Tallying = append(in.Tallying, ast.InspectTally{Counter: counter, Kind: "CHARACTERS", Region: p.parseRegion()})
			case p.isWord("ALL") || p.isWord("LEADING"):
				kind := p.cur().Literal
				p.pos++
				pat := p.parseOperand()
				in.Tallying = append(in.Tallying, ast.InspectTally{Counter: counter, Kind: kind, Pattern: pat, Region: p.parseRegion()})
			case p.cur().Type == token.IDENT && p.isWordAt(1, "FOR") && p.isOperandStart():
				counter = p.parseRef()
				p.expectWord("FOR")
			default:
				if len(in.Tallying) == 0 {
					p.fail("esperado CHARACTERS, ALL ou LEADING, encontrado %s", p.describe(p.cur()))
				}
				return in
			}
		}
	case p.acceptWord("REPLACING"):
		for {
			var r ast.InspectReplace
			switch {
			case p.acceptWord("CHARACTERS"):
				r.Kind = "CHARACTERS"
			case p.isWord("ALL") || p.isWord("FIRST") || p.isWord("LEADING"):
				r.Kind = p.cur().Literal
				p.pos++
				r.From = p.parseOperand()
			default:
				if len(in.Replacing) == 0 {
					p.fail("esperado CHARACTERS, ALL, FIRST ou LEADING, encontrado %s", p.describe(p.cur()))
				}
				return in
			}
			p.expectWord("BY")
			r.To = p.parseOperand()
			r.Region = p.parseRegion()
			in.Replacing = append(in.Replacing, r)
		}
	case p.acceptWord("CONVERTING"):
		c := &ast.InspectConvert{From: p.parseOperand()}
		p.expectWord("TO")
		c.To = p.parseOperand()
		c.Region = p.parseRegion()
		in.Convert = c
		return in
	}
	p.fail("esperado TALLYING, REPLACING ou CONVERTING, encontrado %s", p.describe(p.cur()))
	return nil
}

func (p *Parser) parseOpen(pos ast.Pos) ast.Stmt {
	o := &ast.Open{Pos: pos}
	for {
		mode := p.cur().Literal
		if !(p.isWord("INPUT") || p.isWord("OUTPUT") || p.isWord("I-O") || p.isWord("EXTEND")) {
			if len(o.Items) == 0 {
				p.fail("esperado INPUT, OUTPUT, I-O ou EXTEND, encontrado %s", p.describe(p.cur()))
			}
			return o
		}
		p.pos++
		for p.cur().Type == token.IDENT && p.isOperandStart() {
			n := p.expectIdent("nome de arquivo")
			o.Items = append(o.Items, ast.OpenItem{Mode: mode, File: &n})
		}
	}
}

func (p *Parser) parseClose(pos ast.Pos) ast.Stmt {
	c := &ast.Close{Pos: pos}
	for p.cur().Type == token.IDENT && p.isOperandStart() {
		n := p.expectIdent("nome de arquivo")
		c.Files = append(c.Files, &n)
	}
	if len(c.Files) == 0 {
		p.fail("esperado nome de arquivo em CLOSE, encontrado %s", p.describe(p.cur()))
	}
	return c
}

func (p *Parser) parseRead(pos ast.Pos) ast.Stmt {
	file := p.expectIdent("nome de arquivo")
	r := &ast.Read{Pos: pos, File: &file}
	p.acceptWord("NEXT")
	p.acceptWord("RECORD")
	if p.acceptWord("INTO") {
		r.Into = p.parseRef()
	}
	if p.acceptWord("KEY") {
		p.acceptWord("IS")
		r.Key = p.parseRef()
	}
	r.AtEnd = p.parseBranch("AT", "END")
	r.Invalid = p.parseBranch("INVALID", "KEY")
	p.acceptWord("END-READ")
	return r
}

func (p *Parser) parseWrite(pos ast.Pos) ast.Stmt {
	rec := p.expectIdent("nome de registro")
	w := &ast.Write{Pos: pos, Record: &rec}
	if p.acceptWord("FROM") {
		w.From = p.parseOperand()
	}
	w.Invalid = p.parseBranch("INVALID", "KEY")
	p.acceptWord("END-WRITE")
	return w
}

func (p *Parser) parseStart(pos ast.Pos) ast.Stmt {
	file := p.expectIdent("nome de arquivo")
	s := &ast.Start{Pos: pos, File: &file}
	p.expectWord("KEY")
	op := p.parseRelOp()
	if op == "" {
		p.fail("esperado operador relacional em START, encontrado %s", p.describe(p.cur()))
	}
	s.Op = op
	s.Key = p.parseRef()
	s.Invalid = p.parseBranch("INVALID", "KEY")
	p.acceptWord("END-START")
	return s
}
