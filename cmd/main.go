package main

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/baldwin-dev-co/go-wasm-lib/generator"
	cli "github.com/jawher/mow.cli"
	"github.com/radovskyb/watcher"
)

type opts struct {
	srcPath string
	build bool
	binName string
	watch bool
}

func main() {
	app := cli.App("gowasm", "Generate wasm libraries from go source")
	app.Spec = "SRC [-e] [-b BIN] [-w]"

	var (
		// cmd options
		srcPath = app.StringArg("SRC", ".", "A path to the directory containing the source package")
		build       = app.BoolOpt("b build", false, "Build a wasm binary after code generation")
		binName       = app.StringArg("BIN", "", "The name of the built wasm binary (relative to src)")
		watch       = app.BoolOpt("w watch", false, "Regenerate when a source file is changed")

		// generator options
		exportWrappers = app.BoolOpt("e export", false, "Export wasm wrappers")

	)
	
	app.Action = func() {
		err := execute(
			&opts{
				srcPath: *srcPath,
				build: *build,
				binName: *binName,
				watch: *watch,
			},
			&generator.Config{
				ExportWrappers: *exportWrappers,
			},
		)
		if err != nil {
			fmt.Println(err)
			cli.Exit(1)
		}
	}

	app.Run(os.Args)
}

func execute(cliOpts *opts, genConfig *generator.Config) error {
	err := gowasm(cliOpts.srcPath, genConfig)
	if err != nil {
		return err
	}

	if cliOpts.build {
		err := build(cliOpts.srcPath, cliOpts.binName)
		if err != nil {
			return err
		}
	}

	if cliOpts.watch {
		err := watch(cliOpts, genConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func gowasm(srcPath string, genConfig *generator.Config) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, srcPath, nil, 0)
	if err != nil {
		return fmt.Errorf("Error parsing dir: %v", err)
	}

	var pkgName string
	for name := range pkgs {
		pkgName = name
		break
	}

	pkg := pkgs[pkgName]
	wrapperFile, err := generator.GenerateWrapperFile(pkg, genConfig)
	if err != nil {
		return fmt.Errorf("Error generating go wasm wrappers: %v", err)
	}

	outPath := filepath.Join(srcPath, "wasm-wrappers.go")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating wrapper file: %v\n", err)
		cli.Exit(1)
	}
	defer outFile.Close()

	err = format.Node(outFile, fset, wrapperFile)
	if err != nil {
		return fmt.Errorf("Error formatting wrapper file: %v", err)
	}

	return nil
}

func build(srcPath, binName string) error {
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(srcPath, binName), srcPath)
	buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	err := buildCmd.Run()
	if err != nil {
		return fmt.Errorf("Error building wasm binary: %v", err)
	}

	return nil
}

func watch(cliOpts *opts, genConfig *generator.Config) error {
	w := watcher.New()
	w.FilterOps(watcher.Write)
	cliOpts.watch = false

	go func() {
		for {
			select {
			case event := <-w.Event:	
				if filepath.Base(event.Path) == "wasm-wrappers.go" {
					continue
				}

				err := execute(cliOpts, genConfig)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Successfully generated wasm lib")
				}

			case err := <-w.Error:
				fmt.Printf("watcher error: %v\n", err)
			case <-w.Closed:
				break
			}
		}
	}()

	err := w.Add(cliOpts.srcPath)
	if err != nil {
		return fmt.Errorf("Error adding %s to file watcher: %v", cliOpts.srcPath, err)
	}

	err = w.Start(time.Millisecond * 100)
	if err != nil {
		return fmt.Errorf("Error starting file watcher: %v", err)
	}

	return nil
}
