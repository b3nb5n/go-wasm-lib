package generator

import "go/ast"

// used to generate go code
type Generator struct {
	pkg *ast.Package
	funcs map[string]*ast.FuncType
	aliasResolvers map[string]*ast.Ident
}

func New() *Generator {
	return &Generator{
		funcs: make(map[string]*ast.FuncType),
		aliasResolvers: make(map[string]*ast.Ident),
	}
}

