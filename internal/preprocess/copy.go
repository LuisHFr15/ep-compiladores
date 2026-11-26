package preprocess

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ep-compiladores/internal/diagnostic"
)

const maxDepth = 10

var extensions = []string{".cpy", ".CPY", ".cbl", ".cob", ".CBL", ".COB", ""}

type Expander struct {
	Dirs  []string
	Diags []diagnostic.Diagnostic
}

func (e *Expander) Expand(src string) string {
	return strings.Join(e.expand(strings.Split(src, "\n"), 0, 0), "\n")
}

func (e *Expander) expand(lines []string, depth, baseLine int) []string {
	var out []string
	for i, raw := range lines {
		name, ok := copyName(raw)
		if !ok {
			out = append(out, raw)
			continue
		}
		line := baseLine + i + 1
		if depth >= maxDepth {
			e.errorf(line, "COPY '%s' excede a profundidade máxima de inclusão (%d)", name, maxDepth)
			out = append(out, "")
			continue
		}
		path := e.find(name)
		if path == "" {
			e.errorf(line, "copybook '%s' não encontrado (COPY)", name)
			out = append(out, "")
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			e.errorf(line, "erro ao ler copybook '%s': %v", name, err)
			out = append(out, "")
			continue
		}
		out = append(out, e.expand(strings.Split(strings.TrimRight(string(data), "\n"), "\n"), depth+1, baseLine+i)...)
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

func copyName(raw string) (string, bool) {
	line := strings.TrimRight(raw, "\r")
	if len(line) < 8 {
		return "", false
	}
	switch line[6] {
	case '*', '/', 'D', 'd', '-':
		return "", false
	}
	end := len(line)
	if end > 72 {
		end = 72
	}
	fields := strings.Fields(line[7:end])
	if len(fields) < 2 || !strings.EqualFold(fields[0], "COPY") {
		return "", false
	}
	return strings.TrimSuffix(fields[1], "."), true
}
