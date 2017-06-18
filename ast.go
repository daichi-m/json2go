package json2go

import (
	"bytes"
	"go/ast"
	"go/token"
	"sort"
)

func astMakeDecls(rootNodes []*node) []ast.Decl {
	var decls []ast.Decl

	for _, rootNode := range rootNodes {
		name := attrName(rootNode.key)
		if name == "" {
			continue
		}
		decls = append(decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(name),
					Type: astTypeFromNode(rootNode),
				},
			},
		})
	}

	return decls
}

func astTypeFromNode(n *node) ast.Expr {
	var resultType ast.Expr
	var pointable bool

	switch n.t.id {
	case nodeTypeBool:
		resultType = ast.NewIdent("bool")
		pointable = true
	case nodeTypeInt:
		resultType = ast.NewIdent("int")
		pointable = true
	case nodeTypeFloat:
		resultType = ast.NewIdent("float64")
		pointable = true
	case nodeTypeString:
		resultType = ast.NewIdent("string")
		pointable = true

	case nodeTypeArrayUnknown:
		resultType = &ast.ArrayType{
			Elt: &ast.InterfaceType{
				Methods: &ast.FieldList{
					Opening: 1,
					Closing: 2,
				},
			},
		}
	case nodeTypeArrayBool:
		resultType = &ast.ArrayType{
			Elt: ast.NewIdent("bool"),
		}
	case nodeTypeArrayInt:
		resultType = &ast.ArrayType{
			Elt: ast.NewIdent("int"),
		}
	case nodeTypeArrayFloat:
		resultType = &ast.ArrayType{
			Elt: ast.NewIdent("float64"),
		}
	case nodeTypeArrayString:
		resultType = &ast.ArrayType{
			Elt: ast.NewIdent("string"),
		}
	case nodeTypeArrayInterface:
		resultType = &ast.ArrayType{
			Elt: &ast.InterfaceType{
				Methods: &ast.FieldList{
					Opening: 1,
					Closing: 2,
				},
			},
		}

	case nodeTypeArrayObject:
		resultType = &ast.ArrayType{
			Elt: astStructTypeFromNode(n),
		}

	case nodeTypeObject:
		resultType = astStructTypeFromNode(n)
		pointable = true

	case nodeTypeExternalNode:
		extName := n.externalTypeID
		if extName == "" {
			extName = attrName(n.key)
		}
		resultType = ast.NewIdent(extName)
		pointable = true

	default:
		resultType = &ast.InterfaceType{
			Methods: &ast.FieldList{
				Opening: token.Pos(1),
				Closing: token.Pos(2),
			},
		}
	}

	if pointable && !n.required {
		resultType = &ast.StarExpr{
			X: resultType,
		}
	}
	return resultType
}

func astStructTypeFromNode(n *node) *ast.StructType {
	typeDesc := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{},
		},
	}

	// sort children by name
	type nodeWithName struct {
		key  string
		node *node
	}
	var sortedChildren []nodeWithName
	for _, child := range n.children {
		sortedChildren = append(sortedChildren, nodeWithName{
			key:  child.key,
			node: child,
		})
	}
	sort.Slice(sortedChildren, func(i, j int) bool {
		return sortedChildren[i].key < sortedChildren[j].key
	})

	for _, child := range sortedChildren {
		childName := attrName(child.key)
		if childName == "" {
			continue
		}

		typeDesc.Fields.List = append(typeDesc.Fields.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(childName)},
			Type:  astTypeFromNode(child.node),
			Tag:   astJSONTag(child.key, !child.node.required),
		})
	}

	return typeDesc
}

func astJSONTag(key string, omitempty bool) *ast.BasicLit {
	var buf bytes.Buffer
	buf.WriteString("`json:\"")
	buf.WriteString(key)
	if omitempty {
		buf.WriteString(",omitempty")
	}
	buf.WriteString("\"`")

	return &ast.BasicLit{
		Value: buf.String(),
	}
}
