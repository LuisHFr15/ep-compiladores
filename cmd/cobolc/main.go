package main

import (
	"fmt"
	"os"

	"ep-compiladores/internal/lexer"
	"ep-compiladores/internal/parser"
	"ep-compiladores/internal/source"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "uso: cobolc <arquivo.cbl>")
		os.Exit(1)
	}

	content, err := source.Read(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}

	tokens := lexer.New(string(content)).Tokenize()
	program, err := parser.New(tokens).ParseProgram()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("PROGRAM-ID: %s\n", program.Identification.ProgramID)
}
