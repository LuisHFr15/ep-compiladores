package mips

import (
	"errors"
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/semantic"
)

var routineGroup = map[string]string{
	"rt_print": "core", "rt_print_nl": "core", "rt_print_signed": "core", "rt_print_index": "core",
	"rt_readline": "core", "rt_exit": "core",
	"rt_move_alnum": "str", "rt_fill": "str", "rt_cmp_alnum": "str", "rt_cmp_fill": "str",
	"rt_upper": "str", "rt_accept_alnum": "str", "rt_is_numeric": "str",
	"rt_load": "num", "rt_store": "num", "rt_add": "num", "rt_sub": "num", "rt_mul": "num",
	"rt_div": "num", "rt_pow": "num", "rt_cmp": "num", "rt_neg": "num", "rt_num_to_int": "num",
	"rt_num_from_int": "num", "rt_accept_num": "num", "rt_copy": "num", "rt_get_int": "num",
	"rt_put_int": "num",
	"rt_edit":    "edit",
}

type gen struct {
	info   *semantic.Info
	text   strings.Builder
	data   strings.Builder
	n      int
	lits   map[string]string
	nums   map[string]string
	need   map[string]bool
	tmp    int
	maxTmp int
	errs   []string
	files  *fileGen
}

func Generate(info *semantic.Info) (string, error) {
	g := &gen{
		info: info,
		lits: map[string]string{},
		nums: map[string]string{},
		need: map[string]bool{"core": true},
	}
	g.program()
	if len(g.errs) > 0 {
		return "", errors.New(strings.Join(g.errs, "\n"))
	}
	return g.assemble(), nil
}

func (g *gen) errf(pos ast.Pos, format string, args ...any) {
	g.errs = append(g.errs, fmt.Sprintf("[E] linha %d, col %d: %s", pos.Line, pos.Col, fmt.Sprintf(format, args...)))
}

func (g *gen) ins(format string, args ...any) {
	g.text.WriteString("\t")
	fmt.Fprintf(&g.text, format, args...)
	g.text.WriteString("\n")
}

func (g *gen) label(l string) {
	g.text.WriteString(l + ":\n")
}

func (g *gen) newLabel(prefix string) string {
	g.n++
	return fmt.Sprintf("%s%d", prefix, g.n)
}

func (g *gen) call(name string) {
	grp := routineGroup[name]
	if grp == "" {
		grp = "core"
	}
	g.need[grp] = true
	if grp == "edit" {
		g.need["num"] = true
	}
	g.ins("jal %s", name)
}

func (g *gen) stmtLabelP(idx int) string { return fmt.Sprintf("P%d", idx) }
func (g *gen) stmtLabelE(idx int) string { return fmt.Sprintf("E%d", idx) }

func (g *gen) program() {
	g.label("main")
	g.ins("la $s0, DATA")
	g.fileInit()
	for idx, l := range g.info.Labels {
		g.label(g.stmtLabelP(idx))
		g.stmts(l.Body)
		if g.info.SectionEnd(idx) == idx && l.Name != "" {
			g.endCheck(idx)
		}
	}
	g.ins("li $v0, 10")
	g.ins("syscall")
}

func (g *gen) endCheck(idx int) {
	skip := g.newLabel("K")
	e := g.stmtLabelE(idx)
	g.label(e)
	g.ins("lw $t2, RT_PSP")
	g.ins("beqz $t2, %s", skip)
	g.ins("la $t3, RT_PSTK")
	g.ins("addu $t3, $t3, $t2")
	g.ins("lw $t4, -4($t3)")
	g.ins("la $t5, %s", e)
	g.ins("bne $t4, $t5, %s", skip)
	g.ins("lw $t4, -8($t3)")
	g.ins("addiu $t2, $t2, -8")
	g.ins("sw $t2, RT_PSP")
	g.ins("jr $t4")
	g.label(skip)
}

func (g *gen) assemble() string {
	var b strings.Builder
	b.WriteString("# gerado por cobolc\n\t.data\n\t.align 2\n")
	b.WriteString(g.dataImage())
	b.WriteString(g.data.String())
	b.WriteString("\t.align 2\n")
	for i := 0; i < g.maxTmp; i++ {
		fmt.Fprintf(&b, "TN%d:\t.space 44\n", i)
	}
	b.WriteString(g.fileData())
	b.WriteString("\n\t.text\n\t.globl main\n")
	b.WriteString(g.text.String())
	b.WriteString("\n")
	need := g.need
	if need["edit"] {
		need["num"] = true
	}
	b.WriteString(runtimeText(need))
	return b.String()
}

func (g *gen) bytesLabel(prefix string, data []byte) string {
	key := prefix + string(data)
	if l, ok := g.lits[key]; ok {
		return l
	}
	l := g.newLabel(prefix)
	g.lits[key] = l
	g.data.WriteString(l + ":\t")
	if len(data) == 0 {
		g.data.WriteString(".space 1\n")
		return l
	}
	g.data.WriteString(".byte ")
	for i, c := range data {
		if i > 0 {
			g.data.WriteString(",")
		}
		fmt.Fprintf(&g.data, "%d", c)
	}
	g.data.WriteString("\n")
	return l
}

func (g *gen) strLabel(s string) string {
	return g.bytesLabel("S", []byte(s))
}

func (g *gen) newTmp() string {
	l := fmt.Sprintf("TN%d", g.tmp)
	g.tmp++
	if g.tmp > g.maxTmp {
		g.maxTmp = g.tmp
	}
	return l
}

func (g *gen) addrReg(reg string, off int) {
	if off < 32000 {
		g.ins("addiu %s, $s0, %d", reg, off)
		return
	}
	g.ins("li %s, %d", reg, off)
	g.ins("addu %s, %s, $s0", reg, reg)
}
