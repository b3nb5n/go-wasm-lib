package gowasmlib

import (
	"go/ast"
	"go/token"
	"strconv"
)

func WasmCall(fn *ast.FuncDecl) (*ast.CallExpr, []ast.Stmt) {
	var i int
	args := make([]ast.Expr, fn.Type.Params.NumFields())
	resolvers := make([]ast.Stmt, 0)

	for _, param := range fn.Type.Params.List {
		for _, name := range param.Names {
			arg, resolver := ResolveValue(
				name,
				&ast.IndexExpr{
					X: &ast.Ident{Name: "args"},
					Index: &ast.BasicLit{
						Kind:  token.INT,
						Value: strconv.Itoa(i),
					},
				},
				param.Type,
				nil,
			)

			args[i] = arg
			if resolver != nil {
				resolvers = append(resolvers, resolver...)
			}

			i++
		}
	}

	call := &ast.CallExpr{
		Fun: &ast.Ident{
			Name: fn.Name.Name,
			Obj:  fn.Name.Obj,
		},
		Args: args,
	}

	return call, resolvers
}

func WrapFunc(fn *ast.FuncDecl) *ast.FuncDecl {
	nativeCall, argResolvers := WasmCall(fn)

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
			List: append(argResolvers, &ast.ExprStmt{
				X: nativeCall,
			}),
		},
	}
}
