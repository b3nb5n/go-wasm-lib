package generator

import (
	"go/ast"

	"golang.org/x/tools/go/ast/astutil"
)

// transforms the package  into a wasm wrapper package in place
func (gen *Generator) WasmWrapperPkg(pkg *ast.Package) {
	if pkg == nil {
		pkg = gen.pkg
	} else {
		gen.pkg = pkg
	}

	pkgIdent := &ast.Ident{Name: pkg.Name + "wasm"}
	for _, file := range pkg.Files {
		file.Name = pkgIdent
		gen.WasmWrapperFile(file)
	}

	gen.wasmMain(pkg)
}

// transforms the file into a wasm wrapper file in place.
// each function declaration is wrapped, all other declarations are removed
// if the given file is nil, a new wrapper file will be returned
func (gen *Generator) WasmWrapperFile(file *ast.File) *ast.File {
	if file == nil {
		file = &ast.File{}
	}
	if file.Decls == nil {
		file.Decls = make([]ast.Decl, 0)
	}

	declWrappers := make([]ast.Decl, 0)
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			gen.WasmWrapperFunc(fn)
			declWrappers = append(declWrappers, fn)
		}
	}

	file.Decls = declWrappers
	astutil.AddImport(gen.fset, file, "syscall/js")
	
	return file
}

// transforms the function declaration in place into a wasm wrapper around a call to the given function
func (gen *Generator) WasmWrapperFunc(fn *ast.FuncDecl) {
	args, argsResolver := gen.resolveFuncArgs(fn.Type.Params)
	fn.Body = &ast.BlockStmt{
		List: append(argsResolver, &ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{
						Name: fn.Name.Name,
						Obj:  fn.Name.Obj,
					},
					Args: args,
				},
			},
		}),
	}

	fn.Name.Name += "Wasm"
	gen.funcs[fn.Name.Name] = fn.Type

	fn.Type = &ast.FuncType{
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
	}
}
