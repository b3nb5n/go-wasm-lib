package generator

import "go/ast"

// used to generate go code
type Generator struct {
	pkg *ast.Package
	funcs map[string]*ast.FuncType
	aliasResolvers map[string]*ast.Ident
}

