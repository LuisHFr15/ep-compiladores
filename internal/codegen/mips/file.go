package mips

import (
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/semantic"
)

const tableSlots = 256

func (g *gen) fcb(fi *semantic.FileInfo) string { return fmt.Sprintf("FCB%d", fi.Index) }

func (g *gen) fileInit() {
	for _, fi := range g.info.Files {
		if fi.Select == nil || fi.FD == nil {
			continue
		}
		g.ins("la $t0, %s", g.fcb(fi))
		g.addrReg("$t1", fi.RecBase)
		g.ins("sw $t1, 20($t0)")
	}
}

func orgCode(fi *semantic.FileInfo) int {
	switch fi.Org {
	case "LINE SEQUENTIAL":
		return 2
	case "INDEXED":
		return 3
	}
	return 1
}

func (g *gen) fileData() string {
	var b strings.Builder
	for _, fi := range g.info.Files {
		if fi.Select == nil || fi.FD == nil {
			continue
		}
		keyOff, keyLen := 0, 0
		if fi.Key != nil {
			keyOff, keyLen = fi.Key.Offset, fi.Key.ElemSize
		}
		tbl := "0"
		if fi.Org == "INDEXED" {
			tbl = fmt.Sprintf("TBL%d", fi.Index)
		}
		fmt.Fprintf(&b, "\t.align 2\n%s:\t.word -1, 0, 0, %d, %d, 0, %d, %d, %s, 0, 0, 0\n\t.space 16\n\t.space 128\n",
			g.fcb(fi), orgCode(fi), fi.RecSize, keyOff, keyLen, tbl)
		if fi.Org == "INDEXED" {
			fmt.Fprintf(&b, "TBL%d:\t.space %d\n", fi.Index, tableSlots*fi.RecSize)
		}
	}
	return b.String()
}

func (g *gen) fileStatus(fi *semantic.FileInfo) {
	g.need["file"] = true
	g.need["num"] = true
	if fi.Status == nil {
		return
	}
	g.ins("la $t0, %s", g.fcb(fi))
	g.ins("lw $a2, 8($t0)")
	g.addrReg("$a0", fi.Status.Addr())
	g.ins("li $a1, %d", fi.Status.ElemSize)
	g.call("rt_put_int")
}

func (g *gen) statusIs(fi *semantic.FileInfo) {
	g.ins("la $t0, %s", g.fcb(fi))
	g.ins("lw $t0, 8($t0)")
}

func (g *gen) open(s *ast.Open) {
	for _, it := range s.Items {
		fi := g.info.FileByName(it.File.Text)
		if fi == nil || fi.FD == nil {
			continue
		}
		mode := map[string]int{"INPUT": 1, "OUTPUT": 2, "I-O": 3, "EXTEND": 4}[it.Mode]
		g.ins("la $a0, %s", g.fcb(fi))
		if fi.Assign != nil {
			g.addrReg("$a1", fi.Assign.Addr())
			g.ins("li $a2, %d", fi.Assign.ElemSize)
		} else {
			name := []byte(fi.Select.Assign)
			g.ins("la $a1, %s", g.bytesLabel("S", name))
			g.ins("li $a2, %d", len(name))
		}
		g.ins("li $a3, %d", mode)
		g.need["file"] = true
		g.call("rt_f_open")
		g.fileStatus(fi)
	}
}

func (g *gen) closeFiles(s *ast.Close) {
	for _, f := range s.Files {
		fi := g.info.FileByName(f.Text)
		if fi == nil || fi.FD == nil {
			continue
		}
		g.ins("la $a0, %s", g.fcb(fi))
		g.need["file"] = true
		g.call("rt_f_close")
		g.fileStatus(fi)
	}
}

func (g *gen) fileBranches(fi *semantic.FileInfo, atEnd, invalid *ast.Branch) {
	if atEnd == nil && invalid == nil {
		return
	}
	end := g.newLabel("B")
	onL := g.newLabel("B")
	notL := g.newLabel("B")
	b := atEnd
	if b == nil {
		b = invalid
	}
	g.statusIs(fi)
	if atEnd != nil {
		g.ins("li $t1, 10")
		g.ins("beq $t0, $t1, %s", onL)
	} else {
		g.ins("li $t1, 21")
		g.ins("blt $t0, $t1, %s", notL)
		g.ins("li $t1, 24")
		g.ins("ble $t0, $t1, %s", onL)
	}
	g.label(notL)
	g.ins("bnez $t0, %s", end)
	g.stmts(b.Not)
	g.ins("j %s", end)
	g.label(onL)
	g.stmts(b.On)
	g.label(end)
}

func (g *gen) read(s *ast.Read) {
	fi := g.info.FileByName(s.File.Text)
	if fi == nil || fi.FD == nil {
		return
	}
	g.need["file"] = true
	g.ins("la $a0, %s", g.fcb(fi))
	if fi.Org == "INDEXED" && s.Key != nil {
		g.call("rt_f_readkey")
	} else {
		g.call("rt_f_read")
	}
	g.fileStatus(fi)
	if s.Into != nil {
		skip := g.newLabel("B")
		g.statusIs(fi)
		g.ins("bnez $t0, %s", skip)
		dst := g.symOf(s.Into)
		g.addrOf("$a0", s.Into)
		g.ins("addiu $sp, $sp, -4")
		g.ins("sw $a0, 0($sp)")
		g.addrReg("$a2", fi.RecBase)
		g.ins("lw $a0, 0($sp)")
		g.ins("addiu $sp, $sp, 4")
		g.ins("li $a1, %d", dst.ElemSize)
		g.ins("li $a3, %d", fi.Records[0].ElemSize)
		g.call("rt_move_alnum")
		g.label(skip)
	}
	g.fileBranches(fi, s.AtEnd, s.Invalid)
}

func (g *gen) write(s *ast.Write) {
	fi := g.info.RecordFile(s.Record.Text)
	if fi == nil || fi.FD == nil || fi.Select == nil {
		return
	}
	var rec *semantic.Symbol
	for _, r := range fi.Records {
		if r.Name == s.Record.Text {
			rec = r
		}
	}
	if rec == nil {
		return
	}
	g.need["file"] = true
	if s.From != nil {
		ref := &ast.Ref{Pos: s.Pos, Name: rec.Name}
		g.info.Refs[ref] = rec
		g.moveAlnumTo(s.From, ref, rec)
	}
	g.ins("la $a0, %s", g.fcb(fi))
	g.addrReg("$a1", rec.Addr())
	g.ins("li $a2, %d", rec.ElemSize)
	g.call("rt_f_write")
	g.fileStatus(fi)
	g.fileBranches(fi, nil, s.Invalid)
}

func (g *gen) moveAlnumTo(src ast.Expr, t *ast.Ref, dst *semantic.Symbol) {
	if f := asFig(src); f != nil {
		g.fillTarget(t, figByte(f.Kind))
		return
	}
	g.moveAlnum(src, t, dst)
}

func (g *gen) start(s *ast.Start) {
	fi := g.info.FileByName(s.File.Text)
	if fi == nil || fi.FD == nil {
		return
	}
	if s.Op != "=" {
		g.errf(s.Pos, "START suporta apenas KEY EQUAL")
		return
	}
	g.need["file"] = true
	g.ins("la $a0, %s", g.fcb(fi))
	g.call("rt_f_start")
	g.fileStatus(fi)
	g.fileBranches(fi, nil, s.Invalid)
}
