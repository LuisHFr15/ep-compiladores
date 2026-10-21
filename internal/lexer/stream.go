package lexer

import (
	"ep-compiladores/internal/diagnostic"
	"ep-compiladores/internal/token"
)

type TokenStream struct {
	Tokens      []token.Token
	Diagnostics []diagnostic.Diagnostic
}

func (s *TokenStream) HasErrors() bool {
	for _, d := range s.Diagnostics {
		if d.Severity == diagnostic.Error || d.Severity == diagnostic.Severe {
			return true
		}
	}
	return false
}
