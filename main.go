package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
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
	gen := generator.New()
	ast.PackageExports(pkg)
	gen.WasmWrapperPkg(pkg)

	var buf bytes.Buffer
	for _, file := range pkg.Files {
		format.Node(&buf, fset, file)
	}

	fmt.Println(buf.String())
}