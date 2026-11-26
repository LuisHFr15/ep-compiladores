package mips

import (
	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/semantic"
)

func (g *gen) stmts(list []ast.Stmt) {
	for _, s := range list {
		g.stmt(s)
	}
}

func (g *gen) stmt(s ast.Stmt) {
	g.tmp = 0
	switch s := s.(type) {
	case *ast.Display:
		g.display(s)
	case *ast.Accept:
		g.accept(s)
	case *ast.Move:
		for _, t := range s.Targets {
			g.tmp = 0
			g.move(s.Src, t)
		}
	case *ast.Add:
		g.add(s)
	case *ast.Subtract:
		g.subtract(s)
	case *ast.Multiply:
		g.multiply(s)
	case *ast.Divide:
		g.divide(s)
	case *ast.Compute:
		g.compute(s)
	case *ast.If:
		g.ifStmt(s)
	case *ast.Evaluate:
		g.evaluate(s)
	case *ast.Perform:
		g.perform(s)
	case *ast.GoTo:
		g.ins("j %s", g.stmtLabelP(g.info.LabelIndex[s.Target.Text]))
	case *ast.StopRun, *ast.Goback:
		g.ins("li $v0, 10")
		g.ins("syscall")
	case *ast.Exit, *ast.Continue:
	case *ast.Set:
		g.set(s)
	case *ast.Search:
		g.search(s)
	case *ast.Inspect:
		g.inspect(s)
	case *ast.Open:
		g.open(s)
	case *ast.Close:
		g.closeFiles(s)
	case *ast.Read:
		g.read(s)
	case *ast.Write:
		g.write(s)
	case *ast.Start:
		g.start(s)
	default:
		g.errf(s.Position(), "comando não suportado na geração de código")
	}
}

func (g *gen) display(d *ast.Display) {
	for _, item := range d.Items {
		g.tmp = 0
		g.displayItem(item)
	}
	if !d.NoAdvancing {
		g.call("rt_print_nl")
	}
}

func (g *gen) printStr(s strOpnd) {
	g.loadStr("$a0", s)
	g.ins("li $a1, %d", s.size)
	g.call("rt_print")
}

func (g *gen) displayItem(e ast.Expr) {
	switch e := e.(type) {
	case *ast.StrLit, *ast.NumLit, *ast.FuncCall:
		g.printStr(g.strOf(e))
	case *ast.Figurative:
		g.printStr(strOpnd{label: g.bytesLabel("S", []byte{figByte(e.Kind)}), size: 1})
	case *ast.Ref:
		sym := g.symOf(e)
		g.addrOf("$a0", e)
		switch {
		case sym.IsIndex:
			g.call("rt_print_index")
		case sym.Cat == semantic.CatNumeric && sym.Pic.Signed:
			g.ins("li $a1, %d", sym.ElemSize)
			g.call("rt_print_signed")
		default:
			g.ins("li $a1, %d", sym.ElemSize)
			g.call("rt_print")
		}
	default:
		g.errf(e.Position(), "operando de DISPLAY não suportado")
	}
}

func (g *gen) accept(a *ast.Accept) {
	sym := g.symOf(a.Target)
	g.addrOf("$a0", a.Target)
	if sym.Cat == semantic.CatNumeric {
		flags := flagsOf(sym)
		if g.info.DecimalComma {
			flags |= 0x400
		}
		g.ins("li $a1, %d", sym.Pic.Size)
		g.ins("li $a2, %d", flags)
		g.call("rt_accept_num")
		return
	}
	g.ins("li $a1, %d", sym.ElemSize)
	g.call("rt_accept_alnum")
}

func (g *gen) fillTarget(t *ast.Ref, b byte) {
	sym := g.symOf(t)
	g.addrOf("$a0", t)
	g.ins("li $a1, %d", sym.ElemSize)
	g.ins("li $a2, %d", b)
	g.call("rt_fill")
}

func (g *gen) move(src ast.Expr, t *ast.Ref) {
	dst := g.symOf(t)
	if dst == nil {
		return
	}
	if f := asFig(src); f != nil {
		g.fillTarget(t, figByte(f.Kind))
		return
	}
	numSrc := g.isNumeric(src)
	switch {
	case dst.Cat == semantic.CatNumeric && (numSrc || g.isAlnumRef(src)):
		n := g.evalNum(src)
		g.storeNum(n, t, 0)
	case dst.Cat == semantic.CatEdited && numSrc:
		n := g.evalNum(src)
		g.addrOf("$a1", t)
		g.ins("la $a0, %s", n)
		g.ins("la $a2, %s", g.strLabel(dst.Pic.Mask))
		flags := dst.Pic.Size
		if g.info.DecimalComma {
			flags |= 1 << 16
		}
		g.ins("li $a3, %d", flags)
		g.call("rt_edit")
	case dst.Cat == semantic.CatNumeric:
		g.errf(t.Pos, "MOVE de literal alfanumérico para item numérico '%s' não suportado", t.Name)
	default:
		g.moveAlnum(src, t, dst)
	}
}

func (g *gen) isAlnumRef(e ast.Expr) bool {
	r, ok := e.(*ast.Ref)
	if !ok {
		return false
	}
	sym := g.symOf(r)
	return sym != nil && (sym.Cat == semantic.CatAlnum || sym.Cat == semantic.CatAlpha)
}

func (g *gen) moveAlnum(src ast.Expr, t *ast.Ref, dst *semantic.Symbol) {
	s := g.strOf(src)
	g.addrOf("$a0", t)
	g.ins("addiu $sp, $sp, -4")
	g.ins("sw $a0, 0($sp)")
	g.loadStr("$a2", s)
	g.ins("lw $a0, 0($sp)")
	g.ins("addiu $sp, $sp, 4")
	g.ins("li $a1, %d", dst.ElemSize)
	g.ins("li $a3, %d", s.size)
	g.call("rt_move_alnum")
}

func (g *gen) storeNum(num string, t *ast.Ref, extra int) {
	sym := g.symOf(t)
	if sym.Cat != semantic.CatNumeric {
		g.errf(t.Pos, "destino '%s' não é numérico", t.Name)
		return
	}
	g.addrOf("$a1", t)
	g.ins("la $a0, %s", num)
	g.ins("li $a2, %d", sym.Pic.Size)
	g.ins("li $a3, %d", flagsOf(sym)|extra)
	g.call("rt_store")
	if extra != 0 {
		g.ins("or $s1, $s1, $v0")
	}
}

func (g *gen) withSize(size *ast.Branch, body func(extra int)) {
	if size == nil {
		body(0)
		return
	}
	notL, endL := g.newLabel("Z"), g.newLabel("Z")
	g.ins("li $s1, 0")
	body(0x200)
	g.ins("beqz $s1, %s", notL)
	g.stmts(size.On)
	g.ins("j %s", endL)
	g.label(notL)
	g.stmts(size.Not)
	g.label(endL)
}

func (g *gen) sum(ops []ast.Expr) string {
	acc := g.evalNum(ops[0])
	for _, o := range ops[1:] {
		r := g.evalNum(o)
		t := g.newTmp()
		g.op3("rt_add", t, acc, r)
		acc = t
	}
	return acc
}

func (g *gen) refOf(e ast.Expr, what string) *ast.Ref {
	r, ok := e.(*ast.Ref)
	if !ok {
		g.errf(e.Position(), "esperado identificador como destino de %s", what)
	}
	return r
}

func (g *gen) add(s *ast.Add) {
	g.withSize(s.Size, func(extra int) {
		total := g.sum(s.Operands)
		if len(s.Giving) == 0 {
			for _, to := range s.To {
				g.tmp = g.tmpAfter(total)
				r := g.refOf(to, "ADD")
				if r == nil {
					continue
				}
				cur := g.evalNum(r)
				res := g.newTmp()
				g.op3("rt_add", res, cur, total)
				g.storeNum(res, r, extra)
			}
			return
		}
		if len(s.To) > 0 {
			base := g.evalNum(s.To[0])
			t := g.newTmp()
			g.op3("rt_add", t, total, base)
			total = t
		}
		for _, t := range s.Giving {
			g.storeNum(total, t, extra)
		}
	})
}

func (g *gen) tmpAfter(label string) int {
	var n int
	if _, err := sscanTmp(label, &n); err == nil {
		return n + 1
	}
	return g.tmp
}

func (g *gen) subtract(s *ast.Subtract) {
	g.withSize(s.Size, func(extra int) {
		total := g.sum(s.Operands)
		if len(s.Giving) == 0 {
			for _, from := range s.From {
				g.tmp = g.tmpAfter(total)
				r := g.refOf(from, "SUBTRACT")
				if r == nil {
					continue
				}
				cur := g.evalNum(r)
				res := g.newTmp()
				g.op3("rt_sub", res, cur, total)
				g.storeNum(res, r, extra)
			}
			return
		}
		minuend := g.evalNum(s.From[0])
		res := g.newTmp()
		g.op3("rt_sub", res, minuend, total)
		for _, t := range s.Giving {
			g.storeNum(res, t, extra)
		}
	})
}

func (g *gen) multiply(s *ast.Multiply) {
	g.withSize(s.Size, func(extra int) {
		a := g.evalNum(s.Operand)
		if len(s.Giving) == 0 {
			for _, by := range s.By {
				g.tmp = g.tmpAfter(a)
				r := g.refOf(by, "MULTIPLY")
				if r == nil {
					continue
				}
				cur := g.evalNum(r)
				res := g.newTmp()
				g.op3("rt_mul", res, a, cur)
				g.storeNum(res, r, extra)
			}
			return
		}
		b := g.evalNum(s.By[0])
		res := g.newTmp()
		g.op3("rt_mul", res, a, b)
		for _, t := range s.Giving {
			g.storeNum(res, t, extra)
		}
	})
}

func (g *gen) divide(s *ast.Divide) {
	dividend, divisor := s.A, s.B
	if s.Into {
		dividend, divisor = s.B, s.A
	}
	targets := s.Giving
	if len(targets) == 0 {
		r := g.refOf(s.B, "DIVIDE")
		if r == nil {
			return
		}
		targets = []*ast.Ref{r}
	}
	g.withSize(s.Size, func(extra int) {
		x := g.evalNum(dividend)
		y := g.evalNum(divisor)
		for i, t := range targets {
			sym := g.symOf(t)
			if sym.Cat != semantic.CatNumeric {
				g.errf(t.Pos, "destino '%s' não é numérico", t.Name)
				continue
			}
			q := g.newTmp()
			g.ins("li $a3, %d", sym.Pic.Scale)
			g.op3Keep("rt_div", q, x, y)
			if extra != 0 {
				g.ins("or $s1, $s1, $v0")
			}
			g.storeNum(q, t, extra)
			if i == 0 && s.Remainder != nil {
				m := g.newTmp()
				g.op3("rt_mul", m, q, y)
				r := g.newTmp()
				g.op3("rt_sub", r, x, m)
				g.storeNum(r, s.Remainder, 0)
			}
		}
	})
}

func (g *gen) compute(s *ast.Compute) {
	g.withSize(s.Size, func(extra int) {
		n := g.evalNum(s.Expr)
		for _, t := range s.Targets {
			g.storeNum(n, t, extra)
		}
	})
}

func (g *gen) ifStmt(s *ast.If) {
	thenL, elseL, endL := g.newLabel("I"), g.newLabel("I"), g.newLabel("I")
	g.cond(s.Cond, thenL, elseL)
	g.label(thenL)
	g.stmts(s.Then)
	g.ins("j %s", endL)
	g.label(elseL)
	g.stmts(s.Else)
	g.label(endL)
}

func (g *gen) evaluate(s *ast.Evaluate) {
	endL := g.newLabel("V")
	for _, w := range s.Whens {
		body, next := g.newLabel("V"), g.newLabel("V")
		for _, v := range w.Values {
			nxt := g.newLabel("V")
			g.cond(&ast.Rel{Pos: w.Pos, L: s.Subject, Op: "=", R: v}, body, nxt)
			g.label(nxt)
		}
		g.ins("j %s", next)
		g.label(body)
		g.stmts(w.Body)
		g.ins("j %s", endL)
		g.label(next)
	}
	g.stmts(s.Other)
	g.label(endL)
}

func (g *gen) perform(s *ast.Perform) {
	idx := g.info.LabelIndex[s.Target.Text]
	last := idx
	if s.Thru != nil {
		last = g.info.LabelIndex[s.Thru.Text]
	}
	end := g.info.SectionEnd(last)

	test, body, done := g.newLabel("F"), g.newLabel("F"), g.newLabel("F")
	if s.Until != nil {
		g.label(test)
		g.cond(s.Until, done, body)
		g.label(body)
	}
	ret := g.newLabel("R")
	g.ins("la $t0, %s", ret)
	g.ins("la $t1, %s", g.stmtLabelE(end))
	g.ins("lw $t2, RT_PSP")
	g.ins("la $t3, RT_PSTK")
	g.ins("addu $t3, $t3, $t2")
	g.ins("sw $t0, 0($t3)")
	g.ins("sw $t1, 4($t3)")
	g.ins("addiu $t2, $t2, 8")
	g.ins("sw $t2, RT_PSP")
	g.ins("j %s", g.stmtLabelP(idx))
	g.label(ret)
	if s.Until != nil {
		g.ins("j %s", test)
		g.label(done)
	}
}

func (g *gen) set(s *ast.Set) {
	for _, t := range s.Targets {
		n := g.evalNum(s.Value)
		g.storeNum(n, t, 0)
		g.tmp = 0
	}
}
