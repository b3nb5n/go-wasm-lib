package generator

import (
	"fmt"
	"go/ast"
	"go/token"
)

// adds a main function to the package that exposes the package exports to javascript
func (gen *Generator) wasmEntry(pkg *ast.Package) {
	jsGlobalDecls := make([]ast.Stmt, len(gen.funcs))

	var i int
	for name := range gen.funcs {
		jsGlobalDecls[i] = &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "Global"},
						},
					},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind: token.STRING,
						Value: fmt.Sprint("\"", name, "\""),
					},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "FuncOf"},
						},
						Args: []ast.Expr{&ast.Ident{Name: name}},
					},
				},
			},
		}

		i++
	}

	pkg.Files["wasm_entry.go"] = &ast.File{
		Name: &ast.Ident{Name: pkg.Name},
		Imports: []*ast.ImportSpec{
			{ Path: &ast.BasicLit{Kind: token.STRING, Value: "\"syscall/js\""} },
		},
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: &ast.Ident{Name: "main"},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
				},
				Body: &ast.BlockStmt{
					List: append(
						jsGlobalDecls,
						&ast.ForStmt{Body: &ast.BlockStmt{}},
					),
				},
			},
		},
	}
}