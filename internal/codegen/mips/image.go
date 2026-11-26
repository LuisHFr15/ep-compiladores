package mips

import (
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/semantic"
)

func (g *gen) dataImage() string {
	img := make([]byte, g.info.DataSize+1)
	for _, root := range g.info.Roots {
		if root.IsIndex {
			for i := 0; i < 9; i++ {
				img[root.Base+i] = '0'
			}
			continue
		}
		g.initSym(img, root, root.Base)
	}
	var b strings.Builder
	b.WriteString("DATA:\n")
	for i := 0; i < len(img); i += 16 {
		end := i + 16
		if end > len(img) {
			end = len(img)
		}
		b.WriteString("\t.byte ")
		for j := i; j < end; j++ {
			if j > i {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, "%d", img[j])
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (g *gen) initSym(img []byte, s *semantic.Symbol, base int) {
	n := s.Occurs
	if n < 1 {
		n = 1
	}
	for k := 0; k < n; k++ {
		inst := base + k*s.ElemSize
		if len(s.Children) == 0 {
			g.defaultFill(img[inst:inst+s.ElemSize], s)
		}
		for _, c := range s.Children {
			g.initSym(img, c, inst+c.Rel)
		}
		if s.Item != nil && s.Item.Value != nil {
			g.applyValue(img[inst:inst+s.ElemSize], s, s.Item.Value)
		}
	}
}

func (g *gen) defaultFill(dst []byte, s *semantic.Symbol) {
	fill := byte(' ')
	if s.Cat == semantic.CatNumeric {
		fill = '0'
	}
	for i := range dst {
		dst[i] = fill
	}
}

func figByte(kind string) byte {
	switch kind {
	case "ZERO":
		return '0'
	case "LOW-VALUE":
		return 0
	case "HIGH-VALUE":
		return 255
	case "QUOTE":
		return '"'
	}
	return ' '
}

func (g *gen) applyValue(dst []byte, s *semantic.Symbol, v *ast.Value) {
	switch {
	case v.Kind == ast.ValueFigurative:
		b := figByte(v.Text)
		for i := range dst {
			dst[i] = b
		}
	case s.Cat == semantic.CatNumeric && v.Kind == ast.ValueNumber:
		copy(dst, numField(v.Text, s.Pic))
	default:
		text := []byte(v.Text)
		for i := range dst {
			if i < len(text) {
				dst[i] = text[i]
			} else {
				dst[i] = ' '
			}
		}
	}
}

func parseNum(text string) (neg bool, digits string, scale int) {
	t := text
	if strings.HasPrefix(t, "-") {
		neg = true
		t = t[1:]
	} else if strings.HasPrefix(t, "+") {
		t = t[1:]
	}
	if i := strings.IndexAny(t, ".,"); i >= 0 {
		scale = len(t) - i - 1
		t = t[:i] + t[i+1:]
	}
	t = strings.TrimLeft(t, "0")
	if t == "" {
		t = "0"
		neg = false
	}
	return neg, t, scale
}

func numField(text string, p *semantic.Pic) []byte {
	neg, digits, scale := parseNum(text)
	for scale < p.Scale {
		digits += "0"
		scale++
	}
	digits = digits[:len(digits)-(scale-p.Scale)]
	if digits == "" {
		digits = "0"
	}
	out := make([]byte, p.Size)
	for i := range out {
		out[i] = '0'
	}
	for i := 0; i < len(digits) && i < p.Size; i++ {
		out[p.Size-1-i] = digits[len(digits)-1-i]
	}
	if neg && p.Signed {
		out[p.Size-1] += 0x40
	}
	return out
}
