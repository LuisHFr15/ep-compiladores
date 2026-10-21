package diagnostic

import "fmt"

type Severity string

const (
	Info    Severity = "I"
	Warning Severity = "W"
	Error   Severity = "E"
	Severe  Severity = "S"
)

type Diagnostic struct {
	Severity Severity
	Line     int
	Col      int
	Message  string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("[%s] linha %d, col %d: %s", d.Severity, d.Line, d.Col, d.Message)
}
