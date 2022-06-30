package gowasmlib

import (
	"go/ast"
	"go/token"
)

// returns an expression that casts jsValue to the given native type
// if the value cant be directly cast, a runtime resolver is also returned.
func ResolveValue(
	name *ast.Ident,
	jsValue ast.Expr,
	nativeType ast.Expr,
	dst ast.Expr,
) (ast.Expr, []ast.Stmt) {
	switch nativeType := nativeType.(type) {
	case *ast.Ident:
		return resolveIdent(name, jsValue, nativeType, dst)
	case *ast.StarExpr:
		return resolvePointer(name, jsValue, nativeType, dst)
	case *ast.ArrayType:
		return resolveArray(name, jsValue, nativeType, dst)
	case *ast.StructType:
		return resolveStruct(name, jsValue, nativeType, dst)
	default:
		panic("unknown native type")
	}
}

func resolveIdent(
	name *ast.Ident,
	jsValue ast.Expr,
	nativeType *ast.Ident,
	dst ast.Expr,
) (expr ast.Expr, resolver []ast.Stmt) {
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
	case "float32", "float64":
		method = "Float"
		if typeStr != "float64" {
			typeCast = typeStr
		}
	default:
		panic("unknown type identifier")
	}

	expr = &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   jsValue,
			Sel: &ast.Ident{Name: method},
		},
	}

	if typeCast != "" {
		expr = &ast.CallExpr{
			Fun:  &ast.Ident{Name: typeCast},
			Args: []ast.Expr{expr},
		}
	}

	if dst != nil {
		resolver = append(resolver, &ast.AssignStmt{
			Lhs: []ast.Expr{dst},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{expr},
		})

		expr = dst
	}

	return expr, resolver
}

func resolvePointer(
	name *ast.Ident,
	jsValue ast.Expr,
	nativeType *ast.StarExpr,
	dst ast.Expr,
) (_ ast.Expr, resolver []ast.Stmt) {
	if dst == nil {
		dst = name
		resolver = append(resolver, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{name},
						Type:  nativeType,
					},
				},
			},
		})
	}

	_, eltResolver := ResolveValue(
		&ast.Ident{Name: name.Name + "Elt"},
		jsValue,
		nativeType.X,
		dst,
	)

	return dst, append(
		resolver,
		&ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "jsType"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   jsValue,
							Sel: &ast.Ident{Name: "Type"},
						},
					},
				},
			},
			Cond: &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "jsType"},
					Op: token.NEQ,
					Y: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "js"},
						Sel: &ast.Ident{Name: "TypeUndefined"},
					},
				},
				Op: token.LOR,
				Y: &ast.BinaryExpr{
					X:  &ast.Ident{Name: "jsType"},
					Op: token.NEQ,
					Y: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "js"},
						Sel: &ast.Ident{Name: "TypeNull"},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: eltResolver,
			},
		},
	)
}

func resolveArray(
	name *ast.Ident,
	jsValue ast.Expr,
	nativeType *ast.ArrayType,
	dst ast.Expr,
) (_ ast.Expr, resolver []ast.Stmt) {
	lenExpr := nativeType.Len
	if lenExpr == nil { // if the native type represents a slice
		// create a variable to hold the runtime length
		lenExpr = &ast.Ident{Name: name.Name + "Len"}

		// resolve the runtime length
		resolver = append(resolver, &ast.AssignStmt{
			Lhs: []ast.Expr{lenExpr},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   jsValue,
						Sel: &ast.Ident{Name: "Length"},
					},
				},
			},
		})
	}

	if dst == nil {
		if nativeType.Len == nil {
			// declare a new slice using make and add it to the resolver
			resolver = append(resolver, &ast.AssignStmt{
				Lhs: []ast.Expr{name},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun:  &ast.Ident{Name: "make"},
						Args: []ast.Expr{nativeType, lenExpr},
					},
				},
			})
		} else {
			// declare a new array and add it to the resolver
			resolver = append(resolver, &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{name},
							Type:  nativeType,
						},
					},
				},
			})
		}

		// set dst to the newly declared destination
		dst = name
	}

	idxIdent := &ast.Ident{Name: name.Name + "Idx"}
	_, eltResolver := ResolveValue(
		&ast.Ident{Name: name.Name + "Elt"},
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   jsValue,
				Sel: &ast.Ident{Name: "Index"},
			},
			Args: []ast.Expr{idxIdent},
		},
		nativeType.Elt,
		&ast.IndexExpr{X: dst, Index: idxIdent},
	)

	return dst, append(
		resolver,
		&ast.ForStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{idxIdent},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.INT,
						Value: "0",
					},
				},
			},
			Cond: &ast.BinaryExpr{
				X:  idxIdent,
				Op: token.LSS,
				Y:  lenExpr,
			},
			Post: &ast.IncDecStmt{
				X:   idxIdent,
				Tok: token.INC,
			},
			Body: &ast.BlockStmt{
				List: eltResolver,
			},
		},
	)
}

func resolveStruct(
	name *ast.Ident,
	jsValue ast.Expr,
	nativeType *ast.StructType,
	dst ast.Expr,
) (_ ast.Expr, resolver []ast.Stmt) {
	if dst == nil {
		resolver = append(resolver, &ast.AssignStmt{
			Lhs: []ast.Expr{name},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CompositeLit{Type: nativeType},
			},
		})

		dst = name
	}

	for _, field := range nativeType.Fields.List {
		for _, fieldName := range field.Names {
			_, fieldResolver := ResolveValue(
				&ast.Ident{Name: name.Name + fieldName.Name},
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   jsValue,
						Sel: &ast.Ident{Name: "Get"},
					},
					Args: []ast.Expr{fieldName},
				},
				field.Type,
				&ast.SelectorExpr{
					X:   dst,
					Sel: fieldName,
				},
			)

			resolver = append(resolver, fieldResolver...)
		}
	}

	return dst, resolver
}
