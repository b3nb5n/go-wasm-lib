package gowasmlib

import (
	"go/ast"
)

func WrapFunc(fn *ast.FuncDecl) *ast.FuncDecl {
	wrapper := &ast.FuncDecl{
		Doc:  fn.Doc,
		Recv: fn.Recv,
		Name: &ast.Ident{Name: fn.Name.Name + "Wasm"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "Value"},
						},
						Names: []*ast.Ident{
							{Name: "this"},
						},
					},
					{
						Type: &ast.ArrayType{
							Elt: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "js"},
								Sel: &ast.Ident{Name: "Value"},
							},
						},
						Names: []*ast.Ident{
							{Name: "args"},
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.Ident{Name: "any"}},
				},
			},
		},
		Body: fn.Body,
	}

	return wrapper
}
