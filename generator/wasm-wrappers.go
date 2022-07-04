package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// returns a file containing wasm wrappers for each of the top-level function declarations in the pkg
// and a wasmMain function that exposes each of the exported functions to js
func GenerateWrapperFile(pkg *ast.Package, config *Config) (*ast.File, error) {
	if pkg == nil {
		return nil, fmt.Errorf("Pkg can't be nil")
	}

	gen := newGenerator(pkg, config)
	funcSignatures := make(map[string]*ast.FuncType)
	funcWrappers := make([]ast.Decl, 0)
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if !fn.Name.IsExported() || strings.HasSuffix(fn.Name.Name, "Wasm") {
					continue
				}

				funcSignatures[fn.Name.Name] = fn.Type
				wrapper, err := gen.wasmWrapperFunc(fn)
				if err != nil {
					return nil, fmt.Errorf("Error wrapping function \"%s\": %v", fn.Name.Name, err)
				}

				funcWrappers = append(funcWrappers, wrapper)
			}
		}
	}

	wrapperFile := &ast.File{
		Name:  &ast.Ident{Name: pkg.Name},
		Decls: append(funcWrappers, gen.wasmMainFunc(funcSignatures)),
	}

	astutil.AddImport(token.NewFileSet(), wrapperFile, "syscall/js")
	return wrapperFile, nil
}

// returns a wrapper function that:
// transforms dynamic js args into the given static function signature,
// calls the given function with the resolved arguments,
// and returns its results (nil if the given function has no return value)
// 
// wasm wrapper signature:
// 	func exampleWasm(this js.Value, args []js.Value) any { ...
func (gen *generator) wasmWrapperFunc(fn *ast.FuncDecl) (*ast.FuncDecl, error) {
	args, argResolvers, err := gen.resolveFuncArgs(fn.Type.Params)
	if err != nil {
		return nil, err
	}

	funcCall := &ast.CallExpr{
		Fun: &ast.Ident{
			Name: fn.Name.Name,
			Obj:  fn.Name.Obj,
		},
		Args: args,
	}

	var returnStmt *ast.ReturnStmt
	if fn.Type.Results.NumFields() == 0 {
		argResolvers = append(argResolvers, &ast.ExprStmt{X: funcCall})
		returnStmt = &ast.ReturnStmt{
			Results: []ast.Expr{&ast.Ident{Name: "nil"}},
		}
	} else {
		returnStmt = &ast.ReturnStmt{
			Results: []ast.Expr{funcCall},
		}
	}

	return &ast.FuncDecl{
		Name: &ast.Ident{Name: gen.wrapperName(fn.Name.Name)},
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
			List: append(argResolvers, returnStmt),
		},
	}, nil
}

// returns an new function called "wasmMain" that exposes each of the given functions to js
func (gen *generator) wasmMainFunc(funcs map[string]*ast.FuncType) *ast.FuncDecl {
	var i int
	jsGlobalDecls := make([]ast.Stmt, len(funcs))
	for name := range funcs {
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
						Value: "\"" + name + "\"",
					},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "FuncOf"},
						},
						Args: []ast.Expr{&ast.Ident{Name: gen.wrapperName(name)}},
					},
				},
			},
		}

		i++
	}

	return &ast.FuncDecl{
		Name: &ast.Ident{Name: "mainWasm"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: jsGlobalDecls,
		},
	}
}
