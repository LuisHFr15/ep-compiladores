package token

type Type string

const (
	EOF Type = "EOF"

	IDENT          Type = "IDENT"
	NUMBER         Type = "NUMBER"
	STRING         Type = "STRING"
	PICTURE_STRING Type = "PICTURE-STRING"
	OPERATOR       Type = "OPERATOR"

	PERIOD    Type = "PERIOD"
	COMMA     Type = "COMMA"
	SEMICOLON Type = "SEMICOLON"
	LPAREN    Type = "LPAREN"
	RPAREN    Type = "RPAREN"

	IDENTIFICATION Type = "IDENTIFICATION"
	ENVIRONMENT    Type = "ENVIRONMENT"
	DATA           Type = "DATA"
	PROCEDURE      Type = "PROCEDURE"
	DIVISION       Type = "DIVISION"
	SECTION        Type = "SECTION"
	DISPLAY        Type = "DISPLAY"
	ACCEPT         Type = "ACCEPT"
	PERFORM        Type = "PERFORM"
	STOP           Type = "STOP"
	RUN            Type = "RUN"
	IF             Type = "IF"
	ELSE           Type = "ELSE"
	MOVE           Type = "MOVE"
	PROGRAM_ID     Type = "PROGRAM-ID"
	PIC            Type = "PIC"
	PICTURE        Type = "PICTURE"
)

type Token struct {
	Type    Type
	Literal string
	Line    int
	Col     int
}

var keywords = map[string]Type{
	"IDENTIFICATION": IDENTIFICATION,
	"ENVIRONMENT":    ENVIRONMENT,
	"DATA":           DATA,
	"PROCEDURE":      PROCEDURE,
	"DIVISION":       DIVISION,
	"SECTION":        SECTION,
	"DISPLAY":        DISPLAY,
	"ACCEPT":         ACCEPT,
	"PERFORM":        PERFORM,
	"STOP":           STOP,
	"RUN":            RUN,
	"IF":             IF,
	"ELSE":           ELSE,
	"MOVE":           MOVE,
	"PROGRAM-ID":     PROGRAM_ID,
	"PIC":            PIC,
	"PICTURE":        PICTURE,
}

func LookupWord(w string) Type {
	if t, ok := keywords[w]; ok {
		return t
	}
	return IDENT
}

func IsKeyword(t Type) bool {
	for _, k := range keywords {
		if k == t {
			return true
		}
	}
	return false
}
