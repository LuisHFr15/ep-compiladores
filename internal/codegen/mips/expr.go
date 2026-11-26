package mips

import (
	"fmt"
	"strconv"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/semantic"
)

const divScale = 18

func (g *gen) symOf(r *ast.Ref) *semantic.Symbol { return g.info.Refs[r] }

func flagsOf(sym *semantic.Symbol) int {
	f := 0
	if sym.Pic != nil {
		f = sym.Pic.Scale
		if sym.Pic.Signed {
			f |= 0x100
		}
	}
	return f
}

func (g *gen) addrOf(reg string, r *ast.Ref) {
	sym := g.symOf(r)
	off := sym.Addr()
	dynamic := false
	for k, sub := range r.Subs {
		if lit, ok := sub.(*ast.NumLit); ok {
			v, _ := strconv.Atoi(lit.Text)
			off += (v - 1) * sym.Strides[k]
		} else {
			dynamic = true
		}
	}
	if !dynamic {
		g.addrReg(reg, off)
		return
	}
	g.ins("addiu $sp, $sp, -4")
	g.ins("sw $zero, 0($sp)")
	for k, sub := range r.Subs {
		if _, ok := sub.(*ast.NumLit); ok {
			continue
		}
		save := g.tmp
		num := g.evalNum(sub)
		g.tmp = save
		g.ins("la $a0, %s", num)
		g.call("rt_num_to_int")
		g.ins("addiu $v0, $v0, -1")
		g.ins("li $t0, %d", sym.Strides[k])
		g.ins("mul $v0, $v0, $t0")
		g.ins("lw $t1, 0($sp)")
		g.ins("addu $t1, $t1, $v0")
		g.ins("sw $t1, 0($sp)")
	}
	g.ins("lw $t1, 0($sp)")
	g.ins("addiu $sp, $sp, 4")
	g.addrReg(reg, off)
	g.ins("addu %s, %s, $t1", reg, reg)
}

func (g *gen) numConst(text string) string {
	if l, ok := g.nums[text]; ok {
		return l
	}
	neg, digits, scale := parseNum(text)
	l := g.newLabel("NC")
	g.nums[text] = l
	if len(digits) > 40 {
		digits = digits[len(digits)-40:]
	}
	var b strings.Builder
	sign := 0
	if neg {
		sign = 1
	}
	fmt.Fprintf(&b, "\t.align 2\n%s:\t.byte %d,%d,0,0", l, sign, scale)
	for i := 0; i < 40; i++ {
		d := 0
		if j := i - (40 - len(digits)); j >= 0 {
			d = int(digits[j] - '0')
		}
		fmt.Fprintf(&b, ",%d", d)
	}
	b.WriteString("\n")
	g.data.WriteString(b.String())
	return l
}

func (g *gen) isNumeric(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.NumLit, *ast.Binary, *ast.Neg:
		return true
	case *ast.Ref:
		if sym := g.symOf(e); sym != nil {
			return sym.Cat == semantic.CatNumeric
		}
	}
	return false
}

func (g *gen) loadNum(dst string, r *ast.Ref) {
	sym := g.symOf(r)
	if sym.Cat == semantic.CatEdited {
		g.errf(r.Pos, "item numérico editado '%s' não pode ser usado em operação aritmética", r.Name)
		return
	}
	g.addrOf("$a1", r)
	g.ins("la $a0, %s", dst)
	size := sym.ElemSize
	flags := 0
	if sym.Pic != nil {
		size = sym.Pic.Size
		flags = flagsOf(sym)
	}
	g.ins("li $a2, %d", size)
	g.ins("li $a3, %d", flags)
	g.call("rt_load")
}

func (g *gen) op3(fn, dst, a, b string) {
	g.ins("la $a0, %s", dst)
	g.ins("la $a1, %s", a)
	g.ins("la $a2, %s", b)
	g.call(fn)
}

func (g *gen) evalNum(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.NumLit:
		return g.numConst(e.Text)
	case *ast.Figurative:
		if e.Kind == "ZERO" {
			return g.numConst("0")
		}
	case *ast.Ref:
		t := g.newTmp()
		g.loadNum(t, e)
		return t
	case *ast.Neg:
		x := g.evalNum(e.X)
		t := g.newTmp()
		g.ins("la $a0, %s", t)
		g.ins("la $a1, %s", x)
		g.call("rt_neg")
		return t
	case *ast.Binary:
		l := g.evalNum(e.L)
		r := g.evalNum(e.R)
		t := g.newTmp()
		switch e.Op {
		case "+":
			g.op3("rt_add", t, l, r)
		case "-":
			g.op3("rt_sub", t, l, r)
		case "*":
			g.op3("rt_mul", t, l, r)
		case "/":
			g.ins("li $a3, %d", divScale)
			g.op3Keep("rt_div", t, l, r)
		case "**":
			g.op3("rt_pow", t, l, r)
		}
		return t
	}
	g.errf(e.Position(), "expressão não numérica em contexto aritmético")
	return g.numConst("0")
}

func (g *gen) op3Keep(fn, dst, a, b string) {
	g.ins("la $a0, %s", dst)
	g.ins("la $a1, %s", a)
	g.ins("la $a2, %s", b)
	g.call(fn)
}

type strOpnd struct {
	label string
	ref   *ast.Ref
	size  int
}

func (g *gen) strOf(e ast.Expr) strOpnd {
	switch e := e.(type) {
	case *ast.StrLit:
		return strOpnd{label: g.strLabel(e.Value), size: len([]byte(e.Value))}
	case *ast.NumLit:
		text := strings.TrimLeft(e.Text, "+-")
		return strOpnd{label: g.strLabel(text), size: len(text)}
	case *ast.Ref:
		return strOpnd{ref: e, size: g.symOf(e).ElemSize}
	case *ast.FuncCall:
		arg := g.strOf(e.Args[0])
		buf := g.newLabel("UP")
		fmt.Fprintf(&g.data, "%s:\t.space %d\n", buf, arg.size+1)
		g.loadStr("$a1", arg)
		g.ins("la $a0, %s", buf)
		g.ins("li $a2, %d", arg.size)
		g.call("rt_upper")
		return strOpnd{label: buf, size: arg.size}
	}
	g.errf(e.Position(), "operando alfanumérico não suportado")
	return strOpnd{label: g.strLabel(" "), size: 1}
}

func (g *gen) loadStr(reg string, s strOpnd) {
	if s.label != "" {
		g.ins("la %s, %s", reg, s.label)
		return
	}
	g.addrOf(reg, s.ref)
}

func asFig(e ast.Expr) *ast.Figurative {
	f, _ := e.(*ast.Figurative)
	return f
}

func (g *gen) cond(c ast.Cond, tl, fl string) {
	switch c := c.(type) {
	case *ast.And:
		mid := g.newLabel("C")
		g.cond(c.L, mid, fl)
		g.label(mid)
		g.cond(c.R, tl, fl)
	case *ast.Or:
		mid := g.newLabel("C")
		g.cond(c.L, tl, mid)
		g.label(mid)
		g.cond(c.R, tl, fl)
	case *ast.Not:
		g.cond(c.X, fl, tl)
	case *ast.ClassCond:
		g.classCond(c, tl, fl)
	case *ast.Rel:
		g.rel(c)
		switch c.Op {
		case "=":
			g.ins("beqz $v0, %s", tl)
		case "<>":
			g.ins("bnez $v0, %s", tl)
		case "<":
			g.ins("bltz $v0, %s", tl)
		case ">":
			g.ins("bgtz $v0, %s", tl)
		case "<=":
			g.ins("blez $v0, %s", tl)
		case ">=":
			g.ins("bgez $v0, %s", tl)
		}
		g.ins("j %s", fl)
	}
}

func (g *gen) classCond(c *ast.ClassCond, tl, fl string) {
	if c.Class != "NUMERIC" {
		g.errf(c.Pos, "condição de classe %s não suportada", c.Class)
		return
	}
	s := g.strOf(c.X)
	g.loadStr("$a0", s)
	g.ins("li $a1, %d", s.size)
	g.call("rt_is_numeric")
	if c.Neg {
		g.ins("beqz $v0, %s", tl)
	} else {
		g.ins("bnez $v0, %s", tl)
	}
	g.ins("j %s", fl)
}

func (g *gen) rel(c *ast.Rel) {
	figL, figR := asFig(c.L), asFig(c.R)
	switch {
	case figL != nil && figR != nil:
		g.errf(c.Pos, "comparação entre duas constantes figurativas")
	case figL != nil || figR != nil:
		fig, other, negate := figR, c.L, false
		if figL != nil {
			fig, other, negate = figL, c.R, true
		}
		if fig.Kind == "ZERO" && g.isNumeric(other) {
			x, y := g.evalNum(c.L), g.evalNum(c.R)
			g.ins("la $a0, %s", x)
			g.ins("la $a1, %s", y)
			g.call("rt_cmp")
			return
		}
		s := g.strOf(other)
		g.loadStr("$a0", s)
		g.ins("li $a1, %d", s.size)
		g.ins("li $a2, %d", figByte(fig.Kind))
		g.call("rt_cmp_fill")
		if negate {
			g.ins("subu $v0, $zero, $v0")
		}
	case g.isNumeric(c.L) && g.isNumeric(c.R):
		x, y := g.evalNum(c.L), g.evalNum(c.R)
		g.ins("la $a0, %s", x)
		g.ins("la $a1, %s", y)
		g.call("rt_cmp")
	default:
		ls := g.strOf(c.L)
		rs := g.strOf(c.R)
		g.loadStr("$a0", ls)
		g.ins("addiu $sp, $sp, -4")
		g.ins("sw $a0, 0($sp)")
		g.loadStr("$a2", rs)
		g.ins("lw $a0, 0($sp)")
		g.ins("addiu $sp, $sp, 4")
		g.ins("li $a1, %d", ls.size)
		g.ins("li $a3, %d", rs.size)
		g.call("rt_cmp_alnum")
	}
}
