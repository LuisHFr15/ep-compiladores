package semantic

import (
	"fmt"
	"strconv"
	"strings"
)

type Category int

const (
	CatGroup Category = iota
	CatNumeric
	CatAlpha
	CatAlnum
	CatEdited
)

type Pic struct {
	Cat    Category
	Size   int
	Digits int
	Scale  int
	Signed bool
	Mask   string
}

func expandPic(s string) (string, error) {
	var b strings.Builder
	r := []rune(strings.ToUpper(s))
	for i := 0; i < len(r); i++ {
		c := r[i]
		if c == '(' {
			j := i
			for j < len(r) && r[j] != ')' {
				j++
			}
			if j == len(r) || i == 0 {
				return "", fmt.Errorf("máscara PIC mal formada '%s'", s)
			}
			n, err := strconv.Atoi(string(r[i+1 : j]))
			if err != nil || n < 1 {
				return "", fmt.Errorf("repetição inválida na máscara PIC '%s'", s)
			}
			b.WriteString(strings.Repeat(string(r[i-1]), n-1))
			i = j
			continue
		}
		b.WriteRune(c)
	}
	return b.String(), nil
}

func ParsePic(s string, decimalComma bool) (*Pic, error) {
	mask, err := expandPic(s)
	if err != nil {
		return nil, err
	}
	if mask == "" {
		return nil, fmt.Errorf("máscara PIC vazia")
	}

	editChars := "Z*+-.,B0/$"
	edited := strings.ContainsAny(mask, editChars) || strings.Contains(mask, "CR") || strings.Contains(mask, "DB")
	if edited {
		return parseEdited(mask, decimalComma)
	}

	p := &Pic{Mask: mask}
	seenV := false
	hasX, hasA, has9 := false, false, false
	for i, c := range mask {
		switch c {
		case 'S':
			if i != 0 {
				return nil, fmt.Errorf("'S' deve ser o primeiro caractere da máscara PIC '%s'", s)
			}
			p.Signed = true
		case 'V':
			if seenV {
				return nil, fmt.Errorf("mais de um 'V' na máscara PIC '%s'", s)
			}
			seenV = true
		case '9':
			has9 = true
			p.Digits++
			p.Size++
			if seenV {
				p.Scale++
			}
		case 'X':
			hasX = true
			p.Size++
		case 'A':
			hasA = true
			p.Size++
		default:
			return nil, fmt.Errorf("caractere '%c' não suportado na máscara PIC '%s'", c, s)
		}
	}
	switch {
	case hasX || (hasA && has9):
		p.Cat = CatAlnum
		p.Digits, p.Scale = 0, 0
	case hasA:
		p.Cat = CatAlpha
	default:
		p.Cat = CatNumeric
	}
	return p, nil
}

func parseEdited(mask string, decimalComma bool) (*Pic, error) {
	dp, sep := '.', ','
	if decimalComma {
		dp, sep = ',', '.'
	}
	p := &Pic{Cat: CatEdited, Mask: mask, Size: len([]rune(mask))}
	r := []rune(mask)
	afterDP := false
	for i, c := range r {
		switch c {
		case '9', 'Z', '*':
			p.Digits++
			if afterDP {
				p.Scale++
			}
		case dp:
			afterDP = true
		case sep, 'B', '0', '/', '$':
		case '-', '+':
			if i != 0 && i != len(r)-1 {
				return nil, fmt.Errorf("inserção flutuante '%c' não suportada na máscara PIC '%s'", c, mask)
			}
			p.Signed = true
		case 'C', 'D', 'R':
			p.Signed = true
		default:
			return nil, fmt.Errorf("caractere '%c' não suportado na máscara PIC '%s'", c, mask)
		}
	}
	return p, nil
}
