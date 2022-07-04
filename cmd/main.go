package main

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/baldwin-dev-co/go-wasm-lib/generator"
	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("gowasm", "Generate wasm libraries from go source")
	app.Spec = "SRC [-e [--all]] [-b OUT] [-w]"

	var (
		// cmd options
		srcPath = app.StringArg("SRC", ".", "A path to the directory containing the source package")
		_       = app.BoolOpt("b build", false, "Build a wasm binary after code generation")
		_       = app.StringArg("OUT", "build.wasm", "The name of the built wasm binary (relative to src)")
		_       = app.BoolOpt("w watch", false, "Regenerate when a source file is changed")

		// generator options
		exportWrappers = app.BoolOpt("e export", false, "Export wasm wrappers")
	)

	app.Action = func() {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, *srcPath, nil, 0)
		if err != nil {
			fmt.Printf("Error parsing dir: %v\n", err)
			cli.Exit(1)
		}

		var pkgName string
		for name := range pkgs {
			pkgName = name
			break
		}

		config := generator.Config{
			ExportWrappers: *exportWrappers,
		}

		pkg := pkgs[pkgName]
		wrapperFile, err := generator.GenerateWrapperFile(pkg, &config)
		if err != nil {
			fmt.Printf("Error generating go wasm wrappers: %v\n", err)
			cli.Exit(1)
		}

		outPath := filepath.Join(*srcPath, "wasm-wrappers.go")
		outFile, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Error creating wrapper file: %v\n", err)
			cli.Exit(1)
		}
		defer outFile.Close()

		err = format.Node(outFile, fset, wrapperFile)
		if err != nil {
			fmt.Printf("Error formatting wrapper file: %v\n", err)
			cli.Exit(1)
		}
	}

	app.Run(os.Args)
}
