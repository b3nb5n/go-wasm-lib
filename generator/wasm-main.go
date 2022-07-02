package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
)

// adds a main function that exposes the package exports to javascript
func (gen *Generator) wasmMain(pkg *ast.Package) {
	var mainFile *ast.File
	var dirPath string
	for path, file := range pkg.Files {
		dirPath = filepath.Dir(path)
		if filepath.Base(path) == "main.go" {
			mainFile = file
		}
	}

	if mainFile == nil {
		mainFile = gen.WasmWrapperFile(nil)
		pkg.Files[filepath.Join(dirPath, "main.go")] = mainFile
	}

	jsGlobalDecls := make([]ast.Stmt, len(gen.funcs))
	var i int
	for name := range gen.funcs {
		jsGlobalDecls[i] = &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "Global"},
						},
					},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprint("\"", name, "\""),
					},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "FuncOf"},
						},
						Args: []ast.Expr{&ast.Ident{Name: name}},
					},
				},
			},
		}

		i++
	}

	mainFile.Decls = append(mainFile.Decls, &ast.FuncDecl{
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
	})
}
