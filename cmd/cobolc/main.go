package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ep-compiladores/internal/lexer"
	"ep-compiladores/internal/parser"
	"ep-compiladores/internal/report"
	"ep-compiladores/internal/source"
)

func main() {
	output := flag.String("o", "", "arquivo .txt para salvar o fluxo de tokens")
	debug := flag.Bool("debug", false, "inclui as linhas de depuração (coluna 7 = D)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: cobolc [-debug] [-o saida.txt] <arquivo.cob>")
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

	stream := lexer.New(string(content), *debug).Tokenize()
	program, parseErr := parser.New(stream.Tokens).ParseProgram()
	text := report.Render(path, stream, program, parseErr)

	fmt.Print(text)

	if *output != "" {
		err := os.MkdirAll(filepath.Dir(*output), 0o755)
		if err == nil {
			err = os.WriteFile(*output, []byte(text), 0o644)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro ao gravar saída:", err)
			os.Exit(1)
		}
	}

	if stream.HasErrors() || parseErr != nil {
		os.Exit(1)
	}
}
