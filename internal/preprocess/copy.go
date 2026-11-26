package preprocess

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ep-compiladores/internal/diagnostic"
)

const (
	maxDepth = 10
	codeCols = 65
)

var extensions = []string{".cpy", ".CPY", ".cbl", ".cob", ".CBL", ".COB", ""}

type Expander struct {
	Dirs  []string
	Diags []diagnostic.Diagnostic
}

type replacement struct {
	from, to string
	pseudo   bool
}

type copyStmt struct {
	name  string
	repl  []replacement
	lines int
	err   string
}

func (e *Expander) Expand(src string) string {
	return strings.Join(e.expand(strings.Split(src, "\n"), 0, 0), "\n")
}

func (e *Expander) expand(lines []string, depth, baseLine int) []string {
	var out []string
	for i := 0; i < len(lines); {
		st, ok := readCopy(lines, i)
		if !ok {
			out = append(out, lines[i])
			i++
			continue
		}
		line := baseLine + i + 1
		blanks := func() {
			for k := 0; k < st.lines; k++ {
				out = append(out, "")
			}
		}
		consumed := st.lines
		switch {
		case st.err != "":
			e.errorf(line, "COPY inválido: %s", st.err)
			blanks()
		case depth >= maxDepth:
			e.errorf(line, "COPY '%s' excede a profundidade máxima de inclusão (%d)", st.name, maxDepth)
			blanks()
		default:
			path := e.find(st.name)
			if path == "" {
				e.errorf(line, "copybook '%s' não encontrado (COPY)", st.name)
				blanks()
				break
			}
			data, err := os.ReadFile(path)
			if err != nil {
				e.errorf(line, "erro ao ler copybook '%s': %v", st.name, err)
				blanks()
				break
			}
			body := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			if len(st.repl) > 0 {
				body = e.replace(body, st, line)
			}
			out = append(out, e.expand(body, depth+1, baseLine+i)...)
		}
		i += consumed
	}
	return out
}

func (e *Expander) errorf(line int, format string, args ...any) {
	e.Diags = append(e.Diags, diagnostic.Diagnostic{
		Severity: diagnostic.Error,
		Line:     line,
		Col:      8,
		Message:  fmt.Sprintf(format, args...),
	})
}

func (e *Expander) find(name string) string {
	for _, dir := range e.Dirs {
		for _, variant := range []string{name, strings.ToUpper(name), strings.ToLower(name)} {
			for _, ext := range extensions {
				path := filepath.Join(dir, variant+ext)
				if st, err := os.Stat(path); err == nil && !st.IsDir() {
					return path
				}
			}
		}
	}
	return ""
}

func codeArea(raw string) (text string, comment bool) {
	r := []rune(strings.TrimRight(raw, "\r"))
	if len(r) < 8 {
		if strings.TrimSpace(string(r)) == "" {
			return "", true
		}
		return "", false
	}
	switch r[6] {
	case '*', '/':
		return "", true
	}
	end := len(r)
	if end > 72 {
		end = 72
	}
	return string(r[7:end]), false
}

func startsCopy(raw string) bool {
	line := strings.TrimRight(raw, "\r")
	if len(line) < 8 {
		return false
	}
	switch line[6] {
	case '*', '/', 'D', 'd', '-':
		return false
	}
	end := len(line)
	if end > 72 {
		end = 72
	}
	fields := strings.Fields(line[7:end])
	return len(fields) >= 2 && strings.EqualFold(fields[0], "COPY")
}

func readCopy(lines []string, i int) (*copyStmt, bool) {
	if !startsCopy(lines[i]) {
		return nil, false
	}
	stmt, _ := codeArea(lines[i])
	n := 1
	for {
		toks, ended, _ := splitCopy(stmt)
		if ended || i+n >= len(lines) {
			break
		}
		j := i + n
		for j < len(lines) {
			if _, skip := codeArea(lines[j]); !skip {
				break
			}
			j++
		}
		if j >= len(lines) {
			break
		}
		next, _ := codeArea(lines[j])
		first := ""
		if f := strings.Fields(next); len(f) > 0 {
			first = strings.ToUpper(f[0])
		}
		if !hasWord(toks, "REPLACING") && first != "REPLACING" && first != "OF" && first != "IN" {
			break
		}
		stmt += " " + next
		n = j + 1 - i
	}

	toks, _, serr := splitCopy(stmt)
	st := &copyStmt{lines: n}
	if serr != "" {
		st.err = serr
		st.name = firstName(toks)
		return st, true
	}
	st.name = firstName(toks)
	st.repl, st.err = parseReplacing(toks)
	return st, true
}

func firstName(toks []token) string {
	if len(toks) >= 2 {
		return toks[1].text
	}
	return ""
}

func hasWord(toks []token, w string) bool {
	for _, t := range toks {
		if !t.pseudo && strings.EqualFold(t.text, w) {
			return true
		}
	}
	return false
}

type token struct {
	text   string
	pseudo bool
}

func isSpace(r rune) bool { return r == ' ' || r == '\t' }

func splitCopy(s string) (toks []token, ended bool, err string) {
	r := []rune(s)
	for i := 0; i < len(r); {
		if isSpace(r[i]) {
			i++
			continue
		}
		if r[i] == '=' && i+1 < len(r) && r[i+1] == '=' {
			j := i + 2
			for j+1 < len(r) && !(r[j] == '=' && r[j+1] == '=') {
				j++
			}
			if j+1 >= len(r) {
				return toks, false, "pseudo-texto sem fechamento (==)"
			}
			toks = append(toks, token{text: strings.TrimSpace(string(r[i+2 : j])), pseudo: true})
			i = j + 2
			if i < len(r) && r[i] == '.' && (i+1 == len(r) || isSpace(r[i+1])) {
				return toks, true, ""
			}
			continue
		}
		j := i
		for j < len(r) && !isSpace(r[j]) {
			j++
		}
		word := string(r[i:j])
		i = j
		if strings.HasSuffix(word, ".") {
			word = strings.TrimSuffix(word, ".")
			if word != "" {
				toks = append(toks, token{text: word})
			}
			return toks, true, ""
		}
		toks = append(toks, token{text: word})
	}
	return toks, false, ""
}

func parseReplacing(toks []token) ([]replacement, string) {
	i := 2
	if i+1 < len(toks) && !toks[i].pseudo && (strings.EqualFold(toks[i].text, "OF") || strings.EqualFold(toks[i].text, "IN")) {
		i += 2
	}
	if i >= len(toks) || toks[i].pseudo || !strings.EqualFold(toks[i].text, "REPLACING") {
		return nil, ""
	}
	i++
	if i >= len(toks) {
		return nil, "REPLACING sem operandos"
	}
	var out []replacement
	for i < len(toks) {
		a := toks[i]
		if msg := checkOperand(a); msg != "" {
			return nil, msg
		}
		i++
		if i >= len(toks) || toks[i].pseudo || !strings.EqualFold(toks[i].text, "BY") {
			return nil, fmt.Sprintf("esperado BY depois de '%s' em REPLACING", a.text)
		}
		i++
		if i >= len(toks) {
			return nil, "esperado operando depois de BY em REPLACING"
		}
		b := toks[i]
		if msg := checkOperand(b); msg != "" {
			return nil, msg
		}
		i++
		if a.text == "" {
			return nil, "pseudo-texto vazio como origem em REPLACING"
		}
		out = append(out, replacement{from: a.text, to: b.text, pseudo: a.pseudo})
	}
	return out, ""
}

func checkOperand(t token) string {
	if t.pseudo {
		return ""
	}
	up := strings.ToUpper(t.text)
	if strings.HasPrefix(t.text, "\"") || strings.HasPrefix(t.text, "'") || up == "LEADING" || up == "TRAILING" {
		return fmt.Sprintf("operando '%s' não suportado em REPLACING (use ==pseudo-texto== ou uma palavra)", t.text)
	}
	return ""
}

func (e *Expander) replace(body []string, st *copyStmt, line int) []string {
	out := make([]string, len(body))
	for k, raw := range body {
		r := []rune(strings.TrimRight(raw, "\r"))
		if len(r) < 8 || r[6] == '*' || r[6] == '/' {
			out[k] = raw
			continue
		}
		end := len(r)
		if end > 72 {
			end = 72
		}
		code := replaceText(string(r[7:end]), st.repl)
		tail := string(r[end:])
		if n := len([]rune(code)); n > codeCols {
			if trimmed := strings.TrimRight(code, " "); len([]rune(trimmed)) <= codeCols {
				code = trimmed
			} else {
				e.errorf(line, "linha %d do copybook '%s' passa da coluna 72 depois de REPLACING", k+1, st.name)
				out[k] = raw
				continue
			}
		}
		if tail != "" {
			code += strings.Repeat(" ", codeCols-len([]rune(code)))
		}
		out[k] = string(r[:7]) + code + tail
	}
	return out
}

func isWordChar(c rune) bool {
	return c == '-' || c == '_' || (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func replaceText(s string, repl []replacement) string {
	r := []rune(s)
	var b strings.Builder
	var quote rune
	for i := 0; i < len(r); {
		c := r[i]
		if quote != 0 {
			b.WriteRune(c)
			if c == quote {
				quote = 0
			}
			i++
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			b.WriteRune(c)
			i++
			continue
		}
		matched := false
		for _, rp := range repl {
			n := len([]rune(rp.from))
			if i+n > len(r) || !strings.EqualFold(string(r[i:i+n]), rp.from) {
				continue
			}
			if !rp.pseudo && ((i > 0 && isWordChar(r[i-1])) || (i+n < len(r) && isWordChar(r[i+n]))) {
				continue
			}
			b.WriteString(rp.to)
			i += n
			matched = true
			break
		}
		if !matched {
			b.WriteRune(c)
			i++
		}
	}
	return b.String()
}
