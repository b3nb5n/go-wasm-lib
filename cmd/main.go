package main

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/baldwin-dev-co/go-wasm-lib/generator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing package directory")
		return
	}

	srcPath := os.Args[1]
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, srcPath, nil, 0)
	if (err != nil) {
		fmt.Printf("Error parsing dir: %v\n", err)
		return
	}

	var pkgName string
	for name := range pkgs {
		pkgName = name
		break
	}

	pkg := pkgs[pkgName]
	wrapperFile := generator.GenerateWrapperFile(pkg)
	outPath := filepath.Join(srcPath, "wasm-wrappers.go")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating wrapper file: %v\n", err)
		return
	}
	defer outFile.Close()

	err = format.Node(outFile, fset, wrapperFile)
	if err != nil {
		fmt.Printf("Error formatting wrapper file: %v\n", err)
		return
	}
}