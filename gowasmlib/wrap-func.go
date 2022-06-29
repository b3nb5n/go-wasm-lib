package gowasmlib

import (
	"go/ast"
	"go/token"
	"strconv"
)

func WasmCall(fn *ast.FuncDecl) *ast.CallExpr {
	var i int
	args := make([]ast.Expr, fn.Type.Params.NumFields())
	for _, param := range fn.Type.Params.List {
		for range param.Names {
			args[i] = NativeValue(
				&ast.IndexExpr{
					X: &ast.Ident{Name: "args"},
					Index: &ast.BasicLit{
						Kind: token.INT,
						Value: strconv.Itoa(i),
					},
				},
				param.Type,
			)

			i++
		}
	}

	return &ast.CallExpr{
		Fun: &ast.Ident{
			Name: fn.Name.Name,
			Obj:  fn.Name.Obj,
		},
		Args: args,
	}
}

func WrapFunc(fn *ast.FuncDecl) *ast.FuncDecl {
	return &ast.FuncDecl{
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
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{X: WasmCall(fn)},
			},
		},
	}
}
