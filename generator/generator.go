package generator

import (
	"go/ast"
	"go/token"
)

// used to generate go code
type Generator struct {
	fset *token.FileSet
	pkg *ast.Package
	funcs map[string]*ast.FuncType
	aliasResolvers map[string]*ast.Ident
}

func New(fset *token.FileSet) *Generator {
	return &Generator{
		fset: fset,
		funcs: make(map[string]*ast.FuncType),
		aliasResolvers: make(map[string]*ast.Ident),
	}
}

