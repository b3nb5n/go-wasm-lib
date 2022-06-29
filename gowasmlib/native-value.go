package gowasmlib

import (
	"go/ast"
)

func primitiveNativeValue(jsValue ast.Expr, nativeType *ast.Ident) ast.Expr {
	var method, typeCast string

	switch typeStr := nativeType.String(); typeStr {
	case "bool":
		method = "Bool"
	case "string":
		method = "String"
	case "int", "int8", "int16", "int32", "rune", "int64",
		"uint", "uint8", "byte", "uint16", "uint32", "uint64", "uintptr":
		method = "Int"
		if typeStr != "int" {
			typeCast = typeStr
		}
	case "float32", "float64", "complex64", "complex128":
		method = "Float"
		if typeStr != "float64" {
			typeCast = typeStr
		}
	default:
		panic("unknown native type")
	}

	expr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   jsValue,
			Sel: &ast.Ident{Name: method},
		},
	}

	if typeCast != "" {
		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: typeCast},
			Args: []ast.Expr{expr},
		}
	}

	return expr
}

func NativeValue(jsValue ast.Expr, nativeType ast.Expr) ast.Expr {
	switch nativeType := nativeType.(type) {
	case *ast.Ident:
		return primitiveNativeValue(jsValue, nativeType)
	default:
		panic("unknown type expression")
	}
}
