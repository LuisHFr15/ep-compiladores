package mips

import (
	"fmt"
	"strings"

	"ep-compiladores/internal/ast"
)

func sscanTmp(label string, n *int) (int, error) {
	if !strings.HasPrefix(label, "TN") {
		return 0, fmt.Errorf("not a tmp")
	}
	return fmt.Sscanf(label, "TN%d", n)
}

type fileGen struct{}

func (g *gen) fileInit()        {}
func (g *gen) fileData() string { return "" }

func (g *gen) search(s *ast.Search)    { g.errf(s.Pos, "SEARCH não implementado") }
func (g *gen) inspect(s *ast.Inspect)  { g.errf(s.Pos, "INSPECT não implementado") }
func (g *gen) open(s *ast.Open)        { g.errf(s.Pos, "OPEN não implementado") }
func (g *gen) closeFiles(s *ast.Close) { g.errf(s.Pos, "CLOSE não implementado") }
func (g *gen) read(s *ast.Read)        { g.errf(s.Pos, "READ não implementado") }
func (g *gen) write(s *ast.Write)      { g.errf(s.Pos, "WRITE não implementado") }
func (g *gen) start(s *ast.Start)      { g.errf(s.Pos, "START não implementado") }
