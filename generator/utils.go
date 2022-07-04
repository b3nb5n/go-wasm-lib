package generator

import (
	"fmt"
	"go/ast"
	"strings"
)

func (gen *generator) wrapperName(srcName string) string {
	name := srcName + "Wasm"
	if gen.config.ExportWrappers {
		return strings.ToUpper(string(name[0])) + name[1:]
	} else {
		return strings.ToLower(string(name[0])) + name[1:]
	}
}

func (gen *generator) getTypeAlias(name string) (ast.Expr, error) {
	if expr, ok := gen.typeAliases[name]; ok {
		return expr, nil
	}

	// This is just looking throug the current packages top level declarations
	// for a type alias with the given name
	for _, file := range gen.pkg.Files {
		for _, decl := range file.Decls {
			if gDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range gDecl.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == name {
						gen.typeAliases[name] = ts.Type
						return ts.Type, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("No type alias \"%s\" found in the current package", name)
}