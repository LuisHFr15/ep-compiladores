package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ep-compiladores/internal/codegen/mips"
	"ep-compiladores/internal/lexer"
	"ep-compiladores/internal/parser"
	"ep-compiladores/internal/preprocess"
	"ep-compiladores/internal/report"
	"ep-compiladores/internal/semantic"
	"ep-compiladores/internal/source"
)

type dirList []string

func (d *dirList) String() string     { return strings.Join(*d, ",") }
func (d *dirList) Set(v string) error { *d = append(*d, v); return nil }

func main() {
	output := flag.String("o", "", "arquivo .txt para salvar o relatório (tokens e erros)")
	asmOut := flag.String("asm", "", "arquivo .asm de saída (padrão: <nome>.asm)")
	debug := flag.Bool("debug", false, "inclui as linhas de depuração (coluna 7 = D)")
	tokens := flag.Bool("tokens", false, "imprime o fluxo de tokens no terminal")
	var includes dirList
	flag.Var(&includes, "I", "pasta de busca de copybooks (pode repetir)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: cobolc [-debug] [-tokens] [-I dir] [-o relatorio.txt] [-asm saida.asm] <arquivo.cob>")
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	path := flag.Arg(0)

	content, err := source.Read(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}

	exp := &preprocess.Expander{Dirs: append([]string{filepath.Dir(path)}, includes...)}
	text := exp.Expand(string(content))

	res := &report.Result{Path: path, Copy: exp.Diags}
	res.Stream = lexer.New(text, *debug).Tokenize()

	var info *semantic.Info
	if len(res.Copy) == 0 && !res.Stream.HasErrors() {
		res.Program, res.Syntax = parser.New(res.Stream.Tokens).ParseProgram()
		if len(res.Syntax) == 0 {
			info = semantic.Analyze(res.Program)
			res.Semantic = info.Diags
		}
	}

	full := report.Render(res, true)
	if *tokens {
		fmt.Print(full)
	} else {
		fmt.Print(report.Render(res, false))
	}

	if *output != "" {
		if err := write(*output, full); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao gravar saída:", err)
			os.Exit(1)
		}
	}

	if res.Failed() {
		os.Exit(1)
	}

	asm, err := mips.Generate(info)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	target := *asmOut
	if target == "" {
		base := filepath.Base(path)
		target = strings.TrimSuffix(base, filepath.Ext(base)) + ".asm"
	}
	if err := write(target, asm); err != nil {
		fmt.Fprintln(os.Stderr, "erro ao gravar assembly:", err)
		os.Exit(1)
	}
}

func write(path, data string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
