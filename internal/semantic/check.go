package semantic

import (
	"ep-compiladores/internal/ast"
)

func (in *Info) checkProcedure() {
	proc := in.Program.Procedure
	if proc == nil {
		return
	}
	for _, l := range proc.Labels {
		in.stmts(l.Body)
	}
}

func (in *Info) stmts(list []ast.Stmt) {
	for _, s := range list {
		in.stmt(s)
	}
}

func (in *Info) branch(b *ast.Branch) {
	if b != nil {
		in.stmts(b.On)
		in.stmts(b.Not)
	}
}

func (in *Info) target(n *ast.Name) {
	if _, ok := in.LabelIndex[n.Text]; !ok {
		in.errorf(n.Pos, "parágrafo ou seção '%s' não declarado na PROCEDURE DIVISION", n.Text)
	}
}

func (in *Info) fileUse(n *ast.Name) *FileInfo {
	fi := in.fileByName[n.Text]
	switch {
	case fi == nil || fi.Select == nil:
		in.errorf(n.Pos, "arquivo '%s' não declarado em FILE-CONTROL (SELECT)", n.Text)
		return nil
	case fi.FD == nil:
		in.errorf(n.Pos, "arquivo '%s' sem descritor FD na FILE SECTION", n.Text)
		return nil
	}
	return fi
}

func (in *Info) stmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.Display:
		for _, e := range s.Items {
			in.expr(e)
		}
	case *ast.Accept:
		in.ref(s.Target, false)
	case *ast.Move:
		in.expr(s.Src)
		for _, t := range s.Targets {
			in.checkMove(s.Src, in.ref(t, false), t)
		}
	case *ast.Add:
		in.exprs(s.Operands)
		in.exprs(s.To)
		in.refs(s.Giving)
		in.branch(s.Size)
	case *ast.Subtract:
		in.exprs(s.Operands)
		in.exprs(s.From)
		in.refs(s.Giving)
		in.branch(s.Size)
	case *ast.Multiply:
		in.expr(s.Operand)
		in.exprs(s.By)
		in.refs(s.Giving)
		in.branch(s.Size)
	case *ast.Divide:
		in.expr(s.A)
		in.expr(s.B)
		in.refs(s.Giving)
		if s.Remainder != nil {
			in.ref(s.Remainder, false)
		}
		in.branch(s.Size)
	case *ast.Compute:
		in.refs(s.Targets)
		in.expr(s.Expr)
		in.branch(s.Size)
	case *ast.If:
		in.cond(s.Cond)
		in.stmts(s.Then)
		in.stmts(s.Else)
	case *ast.Evaluate:
		in.expr(s.Subject)
		for _, w := range s.Whens {
			for _, a := range w.Alts {
				if a.Value != nil {
					in.expr(a.Value)
				}
			}
			in.stmts(w.Body)
		}
		in.stmts(s.Other)
	case *ast.Perform:
		if s.Target != nil {
			in.target(s.Target)
		}
		if s.Thru != nil {
			in.target(s.Thru)
			a, b := in.LabelIndex[s.Target.Text], in.LabelIndex[s.Thru.Text]
			if b < a {
				in.errorf(s.Thru.Pos, "THRU '%s' precede o parágrafo inicial '%s'", s.Thru.Text, s.Target.Text)
			}
		}
		if s.Varying != nil {
			in.ref(s.Varying.Var, false)
			in.expr(s.Varying.From)
			in.expr(s.Varying.By)
		}
		if s.Until != nil {
			in.cond(s.Until)
		}
		in.stmts(s.Body)
	case *ast.Initialize:
		in.refs(s.Targets)
		for _, r := range s.Replacing {
			in.expr(r.Value)
		}
	case *ast.GoTo:
		in.target(s.Target)
	case *ast.Set:
		in.refs(s.Targets)
		in.expr(s.Value)
	case *ast.Search:
		if sym := in.ref(s.Table, true); sym != nil {
			if sym.Occurs == 0 {
				in.errorf(s.Table.Pos, "SEARCH exige item com OCCURS, '%s' não é uma tabela", s.Table.Name)
			} else if len(sym.Item.Indexes) == 0 {
				in.errorf(s.Table.Pos, "tabela '%s' sem INDEXED BY para SEARCH", s.Table.Name)
			}
		}
		in.stmts(s.AtEnd)
		for _, w := range s.Whens {
			in.cond(w.Cond)
			in.stmts(w.Body)
		}
	case *ast.Inspect:
		in.inspect(s)
	case *ast.Open:
		for _, it := range s.Items {
			in.fileUse(it.File)
		}
	case *ast.Close:
		for _, f := range s.Files {
			in.fileUse(f)
		}
	case *ast.Read:
		in.fileUse(s.File)
		if s.Into != nil {
			in.ref(s.Into, false)
		}
		if s.Key != nil {
			in.ref(s.Key, false)
		}
		in.branch(s.AtEnd)
		in.branch(s.Invalid)
	case *ast.Write:
		fi := in.recordFile[s.Record.Text]
		if fi == nil {
			in.errorf(s.Record.Pos, "registro '%s' não pertence a nenhum arquivo (FD)", s.Record.Text)
		} else if fi.Select == nil {
			in.errorf(s.Record.Pos, "arquivo do registro '%s' não declarado em FILE-CONTROL (SELECT)", s.Record.Text)
		}
		if s.From != nil {
			in.expr(s.From)
		}
		in.branch(s.Invalid)
	case *ast.Start:
		in.fileUse(s.File)
		in.ref(s.Key, false)
		in.branch(s.Invalid)
	}
}

func (in *Info) inspect(s *ast.Inspect) {
	if s.Func != nil {
		in.expr(s.Func)
		if len(s.Replacing) > 0 || s.Convert != nil {
			in.errorf(s.Pos, "INSPECT com FUNCTION só aceita TALLYING")
		}
	} else {
		in.ref(s.Target, false)
	}
	region := func(r ast.Region) {
		if r.After != nil {
			in.expr(r.After)
		}
		if r.Before != nil {
			in.expr(r.Before)
		}
	}
	for _, t := range s.Tallying {
		in.ref(t.Counter, false)
		if t.Pattern != nil {
			in.expr(t.Pattern)
		}
		region(t.Region)
	}
	for _, r := range s.Replacing {
		if r.From != nil {
			in.expr(r.From)
		}
		in.expr(r.To)
		region(r.Region)
	}
	if c := s.Convert; c != nil {
		in.expr(c.From)
		in.expr(c.To)
		region(c.Region)
	}
}

func (in *Info) exprs(list []ast.Expr) {
	for _, e := range list {
		in.expr(e)
	}
}

func (in *Info) refs(list []*ast.Ref) {
	for _, r := range list {
		in.ref(r, false)
	}
}

func (in *Info) expr(e ast.Expr) {
	switch e := e.(type) {
	case *ast.Ref:
		in.ref(e, false)
	case *ast.FuncCall:
		if e.Name != "UPPER-CASE" && e.Name != "REVERSE" {
			in.errorf(e.Pos, "função '%s' não suportada", e.Name)
		}
		in.exprs(e.Args)
	case *ast.LengthOf:
		in.ref(e.Ref, true)
	case *ast.Binary:
		in.expr(e.L)
		in.expr(e.R)
	case *ast.Neg:
		in.expr(e.X)
	}
}

func (in *Info) cond(c ast.Cond) {
	switch c := c.(type) {
	case *ast.Rel:
		in.expr(c.L)
		in.expr(c.R)
	case *ast.And:
		in.cond(c.L)
		in.cond(c.R)
	case *ast.Or:
		in.cond(c.L)
		in.cond(c.R)
	case *ast.Not:
		in.cond(c.X)
	case *ast.ClassCond:
		in.expr(c.X)
	}
}

func (in *Info) ref(r *ast.Ref, allowTable bool) *Symbol {
	sym := in.lookup(r)
	for _, sub := range r.Subs {
		in.expr(sub)
	}
	if sym == nil {
		return nil
	}
	in.Refs[r] = sym
	if r.Start != nil {
		in.expr(r.Start)
		if n, ok := r.Start.(*ast.NumLit); ok && n.Text == "0" {
			in.errorf(r.Pos, "posição inicial de '%s' deve ser maior que zero", r.Name)
		}
		if r.Len != nil {
			if _, ok := r.Len.(*ast.NumLit); !ok {
				in.errorf(r.Pos, "tamanho da referência por posição de '%s' deve ser uma constante numérica", r.Name)
			}
		} else if _, ok := r.Start.(*ast.NumLit); !ok {
			in.errorf(r.Pos, "referência por posição de '%s' sem tamanho exige posição inicial constante", r.Name)
		}
	}

	need, got := len(sym.Strides), len(r.Subs)
	switch {
	case got > need:
		in.errorf(r.Pos, "'%s' recebe subscritos demais (%d, esperado %d)", r.Name, got, need)
	case got < need && !(got == 0 && allowTable):
		in.errorf(r.Pos, "'%s' exige %d subscrito(s), encontrado %d", r.Name, need, got)
	}
	return sym
}

func (in *Info) category(e ast.Expr) Category {
	switch e := e.(type) {
	case *ast.NumLit, *ast.Binary, *ast.Neg, *ast.LengthOf:
		return CatNumeric
	case *ast.Figurative:
		if e.Kind == "ZERO" {
			return CatNumeric
		}
	case *ast.Ref:
		if e.Start != nil {
			return CatAlnum
		}
		if sym := in.Refs[e]; sym != nil {
			return sym.Cat
		}
	case *ast.FuncCall:
		if len(e.Args) > 0 {
			return in.category(e.Args[0])
		}
	}
	return CatAlnum
}

func (in *Info) checkMove(src ast.Expr, dst *Symbol, at *ast.Ref) {
	if dst == nil {
		return
	}
	if in.category(src) == CatAlpha && (dst.Cat == CatNumeric || dst.Cat == CatEdited) {
		in.errorf(at.Pos, "MOVE inválido: item alfabético (PIC A) não pode ser movido para o item numérico '%s'", dst.Name)
	}
}
