package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

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
	gen := generator.New()
	ast.PackageExports(pkg)
	gen.WasmWrapperPkg(pkg)

	outDir := filepath.Join(filepath.Dir(srcPath), filepath.Base(srcPath) + "_wasm")
	os.MkdirAll(outDir, os.ModePerm)
	for srcPath, fNode := range pkg.Files {
		srcSegments := strings.SplitN(filepath.Base(srcPath), ".", 2)
		outPath := filepath.Join(outDir, srcSegments[0] + "_wasm." + srcSegments[1])
		file, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Error creating file %v: %v", outPath, err)
			return
		}
		defer file.Close()

		format.Node(file, fset, fNode)
	}
}