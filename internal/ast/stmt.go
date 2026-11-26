package ast

type Stmt interface {
	Position() Pos
	stmtNode()
}

type Branch struct {
	On  []Stmt
	Not []Stmt
}

type Display struct {
	Pos
	Items       []Expr
	NoAdvancing bool
}

type Accept struct {
	Pos
	Target *Ref
}

type Move struct {
	Pos
	Src     Expr
	Targets []*Ref
}

type Add struct {
	Pos
	Operands []Expr
	To       []Expr
	Giving   []*Ref
	Size     *Branch
}

type Subtract struct {
	Pos
	Operands []Expr
	From     []Expr
	Giving   []*Ref
	Size     *Branch
}

type Multiply struct {
	Pos
	Operand Expr
	By      []Expr
	Giving  []*Ref
	Size    *Branch
}

type Divide struct {
	Pos
	A, B      Expr
	Into      bool
	Giving    []*Ref
	Remainder *Ref
	Size      *Branch
}

type Compute struct {
	Pos
	Targets []*Ref
	Expr    Expr
	Size    *Branch
}

type If struct {
	Pos
	Cond Cond
	Then []Stmt
	Else []Stmt
}

type WhenAlt struct {
	Op    string
	Value Expr
	Class string
	Neg   bool
}

type When struct {
	Pos
	Alts []WhenAlt
	Body []Stmt
}

type Evaluate struct {
	Pos
	Subject  Expr
	Whens    []*When
	Other    []Stmt
	HasOther bool
}

type Varying struct {
	Var      *Ref
	From, By Expr
}

type Perform struct {
	Pos
	Target  *Name
	Thru    *Name
	Varying *Varying
	Until   Cond
	Body    []Stmt
}

type GoTo struct {
	Pos
	Target *Name
}

type StopRun struct{ Pos }

type Goback struct{ Pos }

type Exit struct{ Pos }

type Continue struct{ Pos }

type Set struct {
	Pos
	Targets []*Ref
	Value   Expr
}

type SearchWhen struct {
	Pos
	Cond Cond
	Body []Stmt
}

type Search struct {
	Pos
	Table    *Ref
	AtEnd    []Stmt
	HasAtEnd bool
	Whens    []*SearchWhen
}

type Region struct {
	After  Expr
	Before Expr
}

type InspectTally struct {
	Counter *Ref
	Kind    string
	Pattern Expr
	Region  Region
}

type InspectReplace struct {
	Kind   string
	From   Expr
	To     Expr
	Region Region
}

type InspectConvert struct {
	From, To Expr
	Region   Region
}

type Inspect struct {
	Pos
	Target    *Ref
	Func      *FuncCall
	Tallying  []InspectTally
	Replacing []InspectReplace
	Convert   *InspectConvert
}

type InitReplace struct {
	Category string
	Value    Expr
}

type Initialize struct {
	Pos
	Targets   []*Ref
	Replacing []InitReplace
}

type OpenItem struct {
	Mode string
	File *Name
}

type Open struct {
	Pos
	Items []OpenItem
}

type Close struct {
	Pos
	Files []*Name
}

type Read struct {
	Pos
	File    *Name
	Into    *Ref
	Key     *Ref
	AtEnd   *Branch
	Invalid *Branch
}

type Write struct {
	Pos
	Record  *Name
	From    Expr
	Invalid *Branch
}

type Start struct {
	Pos
	File    *Name
	Op      string
	Key     *Ref
	Invalid *Branch
}

func (Display) stmtNode()    {}
func (Accept) stmtNode()     {}
func (Move) stmtNode()       {}
func (Add) stmtNode()        {}
func (Subtract) stmtNode()   {}
func (Multiply) stmtNode()   {}
func (Divide) stmtNode()     {}
func (Compute) stmtNode()    {}
func (If) stmtNode()         {}
func (Evaluate) stmtNode()   {}
func (Perform) stmtNode()    {}
func (GoTo) stmtNode()       {}
func (StopRun) stmtNode()    {}
func (Goback) stmtNode()     {}
func (Exit) stmtNode()       {}
func (Continue) stmtNode()   {}
func (Set) stmtNode()        {}
func (Search) stmtNode()     {}
func (Inspect) stmtNode()    {}
func (Initialize) stmtNode() {}
func (Open) stmtNode()       {}
func (Close) stmtNode()      {}
func (Read) stmtNode()       {}
func (Write) stmtNode()      {}
func (Start) stmtNode()      {}
