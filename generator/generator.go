package generator

import "go/ast"

type generator struct {
	config *Config
	pkg *ast.Package
	typeAliases map[string]ast.Expr
	aliasResolvers map[string]*ast.FuncDecl
	funcSignatures map[string]*ast.FuncType
	funcWrappers map[string]*ast.FuncDecl
}

func newGenerator(pkg *ast.Package, config *Config) *generator {
	if config == nil {
		config = NewConfig()
	}

	return &generator{
		config: config,
		pkg: pkg,
		typeAliases: make(map[string]ast.Expr),
		aliasResolvers: make(map[string]*ast.FuncDecl),
		funcSignatures: make(map[string]*ast.FuncType),
		funcWrappers: make(map[string]*ast.FuncDecl),
	}
}

type Config struct {
	ExportWrappers bool
	AliasResolvers bool
}

func NewConfig() *Config {
	return &Config {
		AliasResolvers: true,
	}
}
