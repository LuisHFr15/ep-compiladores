package report

import (
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
	"ep-compiladores/internal/diagnostic"
	"ep-compiladores/internal/lexer"
	"ep-compiladores/internal/token"
)

type Result struct {
	Path     string
	Stream   *lexer.TokenStream
	Program  *ast.Program
	Copy     []diagnostic.Diagnostic
	Syntax   []diagnostic.Diagnostic
	Semantic []diagnostic.Diagnostic
}

func (r *Result) Failed() bool {
	return r.Stream.HasErrors() || len(r.Copy) > 0 || len(r.Syntax) > 0 || len(r.Semantic) > 0
}

func section(b *strings.Builder, title string, diags []diagnostic.Diagnostic) {
	fmt.Fprintf(b, "\n=== %s (%d) ===\n", title, len(diags))
	if len(diags) == 0 {
		b.WriteString("nenhum\n")
	}
	for _, d := range diags {
		b.WriteString(d.String() + "\n")
	}
}

func Render(r *Result, tokens bool) string {
	var b strings.Builder
	stream := r.Stream

	fmt.Fprintf(&b, "Arquivo: %s\n", r.Path)

	if tokens {
		fmt.Fprintf(&b, "\n=== Fluxo de tokens (%d) ===\n", len(stream.Tokens))
		fmt.Fprintf(&b, "%-6s %-6s %-5s %-16s %s\n", "#", "LINHA", "COL", "TIPO", "LEXEMA")
		for i, t := range stream.Tokens {
			fmt.Fprintf(&b, "%-6d %-6d %-5d %-16s %s\n", i+1, t.Line, t.Col, category(t.Type), t.Literal)
		}
	}

	section(&b, "Erros de COPY", r.Copy)
	section(&b, "Erros léxicos", stream.Diagnostics)
	section(&b, "Erros sintáticos", r.Syntax)
	section(&b, "Erros semânticos", r.Semantic)

	b.WriteString("\n=== Identificação (§3.1) ===\n")
	if r.Program != nil && r.Program.Identification != nil {
		fmt.Fprintf(&b, "PROGRAM-ID: %s\n", r.Program.Identification.ProgramID)
	} else {
		b.WriteString("não identificado\n")
	}

	b.WriteString("\nResultado: ")
	switch {
	case len(r.Copy) > 0:
		fmt.Fprintf(&b, "FALHA (%d erros de COPY)\n", len(r.Copy))
	case stream.HasErrors():
		fmt.Fprintf(&b, "FALHA (%d erros léxicos)\n", len(stream.Diagnostics))
	case len(r.Syntax) > 0:
		fmt.Fprintf(&b, "FALHA (%d erros sintáticos)\n", len(r.Syntax))
	case len(r.Semantic) > 0:
		fmt.Fprintf(&b, "FALHA (%d erros semânticos)\n", len(r.Semantic))
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
