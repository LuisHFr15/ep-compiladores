package token

type Type string

const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"

	IDENT  Type = "IDENT"
	PERIOD Type = "PERIOD"

	IDENTIFICATION Type = "IDENTIFICATION"
	DIVISION       Type = "DIVISION"
	PROGRAM_ID     Type = "PROGRAM-ID"
)

type Token struct {
	Type    Type
	Literal string
	Line    int
	Col     int
}

var keywords = map[string]Type{
	"IDENTIFICATION": IDENTIFICATION,
	"DIVISION":       DIVISION,
	"PROGRAM-ID":     PROGRAM_ID,
}

func LookupWord(w string) Type {
	if t, ok := keywords[w]; ok {
		return t
	}
	return IDENT
}
