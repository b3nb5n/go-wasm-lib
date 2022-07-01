package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/baldwin-dev-co/go-wasm-lib/generator"
)

func main() {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "test_src", nil, 0)
	if (err != nil) {
		fmt.Printf("Error parsing dir: %v", err)
		return
	}

	pkg := pkgs["test_src"]
	ast.PackageExports(pkg)
	generator.WasmWrapperPkg(pkg)

	var buf bytes.Buffer
	ast.Inspect(pkg, func(n ast.Node) bool {
		if decl, ok := n.(*ast.FuncDecl); ok {
			printer.Fprint(&buf, fset, decl)
		}

		return true
	})

	fmt.Println(buf.String())
}