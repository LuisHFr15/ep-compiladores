package mips

import (
	"ep-compiladores/internal/ast"
)

func (g *gen) search(s *ast.Search) {
	tsym := g.symOf(s.Table)
	if tsym == nil || len(tsym.Item.Indexes) == 0 {
		return
	}
	idx := g.info.ByName(tsym.Item.Indexes[0])
	loop, atEnd, end := g.newLabel("Q"), g.newLabel("Q"), g.newLabel("Q")

	g.label(loop)
	g.addrReg("$a0", idx.Addr())
	g.ins("li $a1, 9")
	g.call("rt_get_int")
	g.ins("blez $v0, %s", atEnd)
	g.ins("li $t0, %d", tsym.Occurs)
	g.ins("bgt $v0, $t0, %s", atEnd)
	for _, w := range s.Whens {
		body, next := g.newLabel("Q"), g.newLabel("Q")
		g.tmp = 0
		g.cond(w.Cond, body, next)
		g.label(body)
		g.stmts(w.Body)
		g.ins("j %s", end)
		g.label(next)
	}
	g.addrReg("$a0", idx.Addr())
	g.ins("li $a1, 9")
	g.call("rt_get_int")
	g.ins("addiu $a2, $v0, 1")
	g.addrReg("$a0", idx.Addr())
	g.ins("li $a1, 9")
	g.call("rt_put_int")
	g.ins("j %s", loop)
	g.label(atEnd)
	g.stmts(s.AtEnd)
	g.label(end)
}

const (
	insTallyChars = 1
	insTallyAll   = 2
	insTallyLead  = 3
	insReplChars  = 4
	insReplAll    = 5
	insReplFirst  = 6
	insReplLead   = 7
	insConvert    = 8
)

func (g *gen) insStore(off int, e ast.Expr) {
	s := g.strOf(e)
	g.loadStr("$a0", s)
	g.ins("la $t9, RT_INS")
	g.ins("sw $a0, %d($t9)", off)
	g.ins("li $t0, %d", s.size)
	g.ins("sw $t0, %d($t9)", off+4)
}

func (g *gen) insCall(s *ast.Inspect, kind int, pat, repl ast.Expr, r ast.Region) {
	g.tmp = 0
	tsym := g.symOf(s.Target)
	g.addrOf("$a0", s.Target)
	g.ins("la $t9, RT_INS")
	g.ins("sw $a0, 0($t9)")
	g.ins("li $t0, %d", tsym.ElemSize)
	g.ins("sw $t0, 4($t9)")
	g.ins("li $t0, %d", kind)
	g.ins("sw $t0, 8($t9)")
	g.ins("sw $zero, 16($t9)")
	g.ins("sw $zero, 24($t9)")
	if pat != nil {
		g.insStore(12, pat)
	}
	if repl != nil {
		g.insStore(20, repl)
	}
	g.ins("la $t9, RT_INS")
	g.ins("li $t0, -1")
	g.ins("sw $t0, 32($t9)")
	g.ins("sw $t0, 40($t9)")
	if r.After != nil {
		g.insStore(28, r.After)
	}
	if r.Before != nil {
		g.insStore(36, r.Before)
	}
	g.call("rt_inspect")
}

func (g *gen) inspect(s *ast.Inspect) {
	for _, t := range s.Tallying {
		kind := insTallyChars
		switch t.Kind {
		case "ALL":
			kind = insTallyAll
		case "LEADING":
			kind = insTallyLead
		}
		g.insCall(s, kind, t.Pattern, nil, t.Region)
		n := g.newTmp()
		g.ins("move $a1, $v0")
		g.ins("la $a0, %s", n)
		g.ins("li $a2, 0")
		g.call("rt_num_from_int")
		cur := g.evalNum(t.Counter)
		sum := g.newTmp()
		g.op3("rt_add", sum, cur, n)
		g.storeNum(sum, t.Counter, 0)
	}
	for _, r := range s.Replacing {
		kind := insReplChars
		switch r.Kind {
		case "ALL":
			kind = insReplAll
		case "FIRST":
			kind = insReplFirst
		case "LEADING":
			kind = insReplLead
		}
		g.insCall(s, kind, r.From, r.To, r.Region)
	}
	if c := s.Convert; c != nil {
		g.insCall(s, insConvert, c.From, c.To, c.Region)
	}
}
