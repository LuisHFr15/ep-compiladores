package semantic

import (
	"fmt"
	"sort"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/diagnostic"
)

type Symbol struct {
	Name     string
	Item     *ast.DataItem
	Parent   *Symbol
	Children []*Symbol
	Root     *Symbol
	Pic      *Pic
	Cat      Category
	ElemSize int
	Occurs   int
	Offset   int
	Base     int
	IsIndex  bool
	Strides  []int
	Rel      int
}

func (s *Symbol) Total() int {
	if s.Occurs > 0 {
		return s.ElemSize * s.Occurs
	}
	return s.ElemSize
}

func (s *Symbol) Addr() int { return s.Root.Base + s.Offset }

type FileInfo struct {
	Name    string
	Index   int
	Select  *ast.Select
	FD      *ast.FileDesc
	Org     string
	Records []*Symbol
	RecSize int
	RecBase int
	Key     *Symbol
	Status  *Symbol
	Assign  *Symbol
}

type Info struct {
	Program      *ast.Program
	DecimalComma bool
	Roots        []*Symbol
	Symbols      []*Symbol
	Files        []*FileInfo
	Labels       []*ast.Label
	LabelIndex   map[string]int
	Refs         map[*ast.Ref]*Symbol
	DataSize     int
	Diags        []diagnostic.Diagnostic

	byName     map[string][]*Symbol
	fileByName map[string]*FileInfo
	recordFile map[string]*FileInfo
	seen       map[string]bool
	indexNames []string
}

func (in *Info) errorf(pos ast.Pos, format string, args ...any) {
	in.Diags = append(in.Diags, diagnostic.Diagnostic{
		Severity: diagnostic.Error,
		Line:     pos.Line,
		Col:      pos.Col,
		Message:  fmt.Sprintf(format, args...),
	})
}

func Analyze(prog *ast.Program) *Info {
	in := &Info{
		Program:    prog,
		LabelIndex: map[string]int{},
		Refs:       map[*ast.Ref]*Symbol{},
		byName:     map[string][]*Symbol{},
		fileByName: map[string]*FileInfo{},
		recordFile: map[string]*FileInfo{},
		seen:       map[string]bool{},
	}
	in.buildData()
	in.buildFiles()
	in.buildLabels()
	in.checkProcedure()

	sort.SliceStable(in.Diags, func(i, j int) bool {
		if in.Diags[i].Line != in.Diags[j].Line {
			return in.Diags[i].Line < in.Diags[j].Line
		}
		return in.Diags[i].Col < in.Diags[j].Col
	})
	return in
}

func (in *Info) fileFor(name string) *FileInfo {
	fi := in.fileByName[name]
	if fi == nil {
		fi = &FileInfo{Name: name, Index: len(in.Files)}
		in.Files = append(in.Files, fi)
		in.fileByName[name] = fi
	}
	return fi
}

func (in *Info) buildData() {
	prog := in.Program
	if prog.Environment != nil {
		in.DecimalComma = prog.Environment.DecimalComma
		for _, sel := range prog.Environment.Files {
			if in.fileByName[sel.Name] != nil && in.fileByName[sel.Name].Select != nil {
				in.errorf(sel.Pos, "arquivo '%s' declarado mais de uma vez em FILE-CONTROL", sel.Name)
				continue
			}
			in.fileFor(sel.Name).Select = sel
		}
	}

	offset := 0
	d := prog.Data
	if d != nil {
		for _, fd := range d.Files {
			fi := in.fileFor(fd.Name)
			if fi.FD != nil {
				in.errorf(fd.Pos, "FD '%s' declarado mais de uma vez", fd.Name)
				continue
			}
			fi.FD = fd
			fi.RecBase = offset
			for _, rec := range fd.Records {
				sym := in.addItem(rec, nil)
				sym.Base = offset
				if sym.Total() > fi.RecSize {
					fi.RecSize = sym.Total()
				}
				fi.Records = append(fi.Records, sym)
				in.recordFile[rec.Name] = fi
			}
			offset += fi.RecSize
		}
		for _, list := range [][]*ast.DataItem{d.Working, d.Linkage} {
			for _, it := range list {
				sym := in.addItem(it, nil)
				sym.Base = offset
				offset += sym.Total()
			}
		}
	}

	for _, name := range in.indexNames {
		sym := &Symbol{Name: name, IsIndex: true, Cat: CatNumeric, ElemSize: 9, Base: offset,
			Pic: &Pic{Cat: CatNumeric, Size: 9, Digits: 9, Mask: "9(9)"}}
		sym.Root = sym
		offset += 9
		in.Roots = append(in.Roots, sym)
		in.Symbols = append(in.Symbols, sym)
		in.byName[name] = append(in.byName[name], sym)
	}
	in.DataSize = offset
}

func (in *Info) addItem(item *ast.DataItem, parent *Symbol) *Symbol {
	sym := &Symbol{Name: item.Name, Item: item, Parent: parent, Occurs: item.Occurs}
	if parent == nil {
		sym.Root = sym
		in.Roots = append(in.Roots, sym)
	} else {
		sym.Root = parent.Root
	}
	in.Symbols = append(in.Symbols, sym)

	switch item.Usage {
	case "", "DISPLAY":
	default:
		in.errorf(item.Pos, "USAGE %s não suportado (apenas DISPLAY)", item.Usage)
	}

	if item.Name != "" {
		key := fmt.Sprintf("%p/%s", parent, item.Name)
		if in.seen[key] {
			in.errorf(item.Pos, "identificador '%s' declarado mais de uma vez no mesmo escopo", item.Name)
		}
		in.seen[key] = true
		in.byName[item.Name] = append(in.byName[item.Name], sym)
	}
	in.indexNames = append(in.indexNames, item.Indexes...)

	rel := 0
	for _, c := range item.Children {
		cs := in.addItem(c, sym)
		sym.Children = append(sym.Children, cs)
		cs.Rel = rel
		rel += cs.Total()
	}

	switch {
	case item.Pic != "":
		pic, err := ParsePic(item.Pic, in.DecimalComma)
		if err != nil {
			in.errorf(item.Pos, "%s", err.Error())
			pic = &Pic{Cat: CatAlnum, Size: 1}
		}
		if len(item.Children) > 0 {
			in.errorf(item.Pos, "item '%s' com PIC não pode ter itens subordinados", item.Name)
		}
		sym.Pic, sym.Cat, sym.ElemSize = pic, pic.Cat, pic.Size
		in.checkValue(item, pic)
	case len(item.Children) > 0:
		sym.Cat, sym.ElemSize = CatGroup, rel
	default:
		in.errorf(item.Pos, "item '%s' sem PIC e sem itens subordinados", item.Name)
		sym.Cat, sym.ElemSize = CatAlnum, 1
	}

	if parent == nil {
		in.assignOffsets(sym, 0, nil)
	}
	return sym
}

func (in *Info) assignOffsets(sym *Symbol, parentOff int, strides []int) {
	sym.Offset = parentOff + sym.Rel
	sym.Strides = append([]int(nil), strides...)
	if sym.Occurs > 0 {
		sym.Strides = append(sym.Strides, sym.ElemSize)
	}
	for _, c := range sym.Children {
		in.assignOffsets(c, sym.Offset, sym.Strides)
	}
}

func (in *Info) checkValue(item *ast.DataItem, pic *Pic) {
	v := item.Value
	if v == nil {
		return
	}
	numeric := pic.Cat == CatNumeric
	switch {
	case numeric && v.Kind == ast.ValueString:
		in.errorf(item.Pos, "VALUE alfanumérico em item numérico '%s'", item.Name)
	case numeric && v.Kind == ast.ValueFigurative && v.Text != "ZERO":
		in.errorf(item.Pos, "VALUE %s inválido em item numérico '%s'", v.Text, item.Name)
	case !numeric && v.Kind == ast.ValueNumber && pic.Cat != CatEdited:
	}
}

func (in *Info) buildFiles() {
	for _, fi := range in.Files {
		switch {
		case fi.Select == nil:
			fi.Org = "SEQUENTIAL"
			continue
		}
		sel := fi.Select
		fi.Org = strings.ToUpper(sel.Organization)
		switch fi.Org {
		case "", "SEQUENTIAL":
			fi.Org = "SEQUENTIAL"
		case "LINE SEQUENTIAL", "INDEXED":
		default:
			in.errorf(sel.Pos, "organização '%s' não suportada no arquivo '%s'", sel.Organization, sel.Name)
		}
		if sel.AssignIsName {
			fi.Assign = in.lookup(&ast.Ref{Pos: sel.Pos, Name: sel.Assign})
		}
		if sel.Status != nil {
			fi.Status = in.lookup(&ast.Ref{Pos: sel.Status.Pos, Name: sel.Status.Text})
		}
		if sel.RecordKey != nil {
			fi.Key = in.lookup(&ast.Ref{Pos: sel.RecordKey.Pos, Name: sel.RecordKey.Text})
		}
		for _, a := range sel.AltKeys {
			in.lookup(&ast.Ref{Pos: a.Pos, Name: a.Text})
		}
		if fi.Org == "INDEXED" && sel.RecordKey == nil {
			in.errorf(sel.Pos, "arquivo indexado '%s' sem RECORD KEY", sel.Name)
		}
	}
}

func (in *Info) buildLabels() {
	proc := in.Program.Procedure
	if proc == nil {
		return
	}
	for _, l := range proc.Labels {
		if l.Name == "" {
			in.Labels = append(in.Labels, l)
			continue
		}
		if _, dup := in.LabelIndex[l.Name]; dup {
			in.errorf(l.Pos, "parágrafo ou seção '%s' declarado mais de uma vez", l.Name)
		}
		in.LabelIndex[l.Name] = len(in.Labels)
		in.Labels = append(in.Labels, l)
	}
}

func (in *Info) SectionEnd(idx int) int {
	if !in.Labels[idx].Section {
		return idx
	}
	end := idx
	for j := idx + 1; j < len(in.Labels) && !in.Labels[j].Section; j++ {
		end = j
	}
	return end
}

func qualifies(c *Symbol, quals []string) bool {
	p := c.Parent
	for _, q := range quals {
		for p != nil && p.Name != q {
			p = p.Parent
		}
		if p == nil {
			return false
		}
		p = p.Parent
	}
	return true
}

func (in *Info) lookup(r *ast.Ref) *Symbol {
	var match []*Symbol
	for _, c := range in.byName[r.Name] {
		if qualifies(c, r.Quals) {
			match = append(match, c)
		}
	}
	switch {
	case len(match) == 1:
		return match[0]
	case len(in.byName[r.Name]) == 0:
		in.errorf(r.Pos, "identificador '%s' não declarado", r.Name)
	case len(match) == 0:
		in.errorf(r.Pos, "qualificação inválida para '%s'", r.Name)
	default:
		in.errorf(r.Pos, "referência ambígua a '%s'; qualifique com OF", r.Name)
	}
	return nil
}

func (in *Info) ByName(name string) *Symbol {
	if list := in.byName[name]; len(list) > 0 {
		return list[0]
	}
	return nil
}

func (in *Info) FileByName(name string) *FileInfo {
	return in.fileByName[name]
}

func (in *Info) RecordFile(name string) *FileInfo {
	return in.recordFile[name]
}
