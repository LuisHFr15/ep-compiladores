package ast

type Pos struct {
	Line int
	Col  int
}

func (p Pos) Position() Pos { return p }

type Name struct {
	Pos
	Text string
}

type Program struct {
	Identification *IdentificationDivision
	Environment    *EnvironmentDivision
	Data           *DataDivision
	Procedure      *ProcedureDivision
}

type IdentificationDivision struct {
	Pos
	ProgramID string
}

type EnvironmentDivision struct {
	DecimalComma bool
	Files        []*Select
}

type Select struct {
	Pos
	Name         string
	Assign       string
	AssignIsName bool
	Organization string
	Access       string
	RecordKey    *Name
	AltKeys      []*Name
	Status       *Name
}

type DataDivision struct {
	Files   []*FileDesc
	Working []*DataItem
	Linkage []*DataItem
}

type FileDesc struct {
	Pos
	Name    string
	Records []*DataItem
}

type ValueKind int

const (
	ValueNumber ValueKind = iota
	ValueString
	ValueFigurative
)

type Value struct {
	Kind ValueKind
	Text string
}

type DataItem struct {
	Pos
	Level    int
	Name     string
	Pic      string
	Usage    string
	Value    *Value
	Occurs   int
	Indexes  []string
	Children []*DataItem
}

type ProcedureDivision struct {
	Pos
	Using  []string
	Labels []*Label
}

type Label struct {
	Pos
	Name    string
	Section bool
	Body    []Stmt
}
