package ast

type Expr interface {
	Position() Pos
	exprNode()
}

type NumLit struct {
	Pos
	Text string
}

type StrLit struct {
	Pos
	Value string
}

type Figurative struct {
	Pos
	Kind string
}

type Ref struct {
	Pos
	Name  string
	Quals []string
	Subs  []Expr
}

type FuncCall struct {
	Pos
	Name string
	Args []Expr
}

type Binary struct {
	Pos
	Op   string
	L, R Expr
}

type Neg struct {
	Pos
	X Expr
}

func (NumLit) exprNode()     {}
func (StrLit) exprNode()     {}
func (Figurative) exprNode() {}
func (Ref) exprNode()        {}
func (FuncCall) exprNode()   {}
func (Binary) exprNode()     {}
func (Neg) exprNode()        {}

type Cond interface {
	Position() Pos
	condNode()
}

type Rel struct {
	Pos
	L, R Expr
	Op   string
}

type And struct {
	Pos
	L, R Cond
}

type Or struct {
	Pos
	L, R Cond
}

type Not struct {
	Pos
	X Cond
}

type ClassCond struct {
	Pos
	X     Expr
	Class string
	Neg   bool
}

func (Rel) condNode()       {}
func (And) condNode()       {}
func (Or) condNode()        {}
func (Not) condNode()       {}
func (ClassCond) condNode() {}
