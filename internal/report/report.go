package report

import (
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/lexer"
	"ep-compiladores/internal/token"
)

func Render(path string, stream *lexer.TokenStream, program *ast.Program, parseErr error) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Arquivo: %s\n\n", path)

	fmt.Fprintf(&b, "=== Fluxo de tokens (%d) ===\n", len(stream.Tokens))
	fmt.Fprintf(&b, "%-6s %-6s %-5s %-16s %s\n", "#", "LINHA", "COL", "TIPO", "LEXEMA")
	for i, t := range stream.Tokens {
		fmt.Fprintf(&b, "%-6d %-6d %-5d %-16s %s\n", i+1, t.Line, t.Col, category(t.Type), t.Literal)
	}

	fmt.Fprintf(&b, "\n=== Erros léxicos (%d) ===\n", len(stream.Diagnostics))
	if len(stream.Diagnostics) == 0 {
		b.WriteString("nenhum\n")
	}
	for _, d := range stream.Diagnostics {
		b.WriteString(d.String() + "\n")
	}

	b.WriteString("\n=== Identificação (§3.1) ===\n")
	if parseErr != nil {
		b.WriteString(parseErr.Error() + "\n")
	} else {
		fmt.Fprintf(&b, "PROGRAM-ID: %s\n", program.Identification.ProgramID)
	}

	b.WriteString("\nResultado: ")
	switch {
	case stream.HasErrors():
		fmt.Fprintf(&b, "FALHA (%d erros léxicos)\n", len(stream.Diagnostics))
	case parseErr != nil:
		b.WriteString("FALHA (erro sintático)\n")
	default:
		b.WriteString("SUCESSO\n")
	}

	return b.String()
}

func category(t token.Type) string {
	if token.IsKeyword(t) {
		return "KEYWORD"
	}
	return string(t)
}
