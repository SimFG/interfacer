/*
 * // Copyright 2022 The SimFG Authors
 * //
 * // Licensed under the Apache License, Version 2.0 (the "License");
 * // you may not use this file except in compliance with the License.
 * // You may obtain a copy of the License at
 * //
 * //     http://www.apache.org/licenses/LICENSE-2.0
 * //
 * // Unless required by applicable law or agreed to in writing, software
 * // distributed under the License is distributed on an "AS IS" BASIS,
 * // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * // See the License for the specific language governing permissions and
 * // limitations under the License.
 */

package writer

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func WriteFile(fileName string, writers []Writer) {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	for _, writer := range writers {
		writer.Write(fileNode)
	}

	err = format.Node(&buf, fset, fileNode)
	// TODO handle like:
	//type Component struct {
	//}
	//
	//func (Component) Dummy(){
	//}
	if err != nil {
		fmt.Println("node filename:", fileName, "err:", err)
		return
	}
	err = ioutil.WriteFile(fileName, buf.Bytes(), 0)
	if err != nil {
		fmt.Println("node err:", err)
		return
	}
	//fmt.Println(string(buf.Bytes()))
}

type Writer interface {
	Write(fileNode *ast.File)
}

type WriteFunc func(fileNode *ast.File)

func (w WriteFunc) Write(fileNode *ast.File) {
	w(fileNode)
}

func GetImportWriter(alia string, importValue string) Writer {
	return WriteFunc(func(fileNode *ast.File) {
		var ident *ast.Ident
		var importSpec *ast.GenDecl
		if alia != "" {
			ident = &ast.Ident{Name: alia}
		}
		for _, decl := range fileNode.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok || d.Tok != token.IMPORT {
				continue
			}
			d.Specs = append(d.Specs, &ast.ImportSpec{
				Name: ident,
				Path: &ast.BasicLit{Value: "\"" + importValue + "\""},
			})
			return
		}
		if importSpec == nil {
			importSpec = &ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					&ast.ImportSpec{
						Name: ident,
						Path: &ast.BasicLit{Value: "\"" + importValue + "\""},
					},
				},
			}
			fileNode.Decls = append(fileNode.Decls, importSpec)
		}
	})
}

func GetIdent(i string) ast.Expr {
	isStar := false
	if i[0] == '*' {
		isStar = true
		i = i[1:]
	}
	pointIndex := strings.Index(i, ".")
	selectorName := ""
	if pointIndex > 0 {
		selectorName = i[0:pointIndex]
		i = i[pointIndex+1:]
	}
	var result ast.Expr
	result = &ast.Ident{Name: i}
	if selectorName != "" {
		result = &ast.SelectorExpr{
			Sel: &ast.Ident{Name: i},
			X:   &ast.Ident{Name: selectorName},
		}
	}
	if isStar {
		result = &ast.StarExpr{X: result}
	}
	return result
}

func GetFuncWriter(receiverName string, receiverType string, funcName string, paramNames []string, paramTypes []string, returnTypes []string, returnDefaultValues []string) Writer {
	return WriteFunc(func(fileNode *ast.File) {
		paramFieldList := &ast.FieldList{}
		if len(paramNames) > 0 {
			lo.ForEach[string](paramNames, func(item string, index int) {
				paramFieldList.List = append(paramFieldList.List, &ast.Field{
					Names: []*ast.Ident{{Name: item}},
					Type:  GetIdent(paramTypes[index]),
				})
			})
		}
		returnFieldList := &ast.FieldList{}
		if len(returnTypes) > 0 {
			lo.ForEach[string](returnTypes, func(item string, index int) {
				returnFieldList.List = append(returnFieldList.List, &ast.Field{
					Type: GetIdent(item),
				})
			})
		}

		funcDecl := &ast.FuncDecl{
			Name: &ast.Ident{Name: funcName},
			Type: &ast.FuncType{
				Params:  paramFieldList,
				Results: returnFieldList,
			},
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: receiverName}},
						Type:  GetIdent(receiverType),
					},
				},
			},
		}

		if len(returnDefaultValues) > 0 {
			funcDecl.Body = &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{Name: strings.Join(returnDefaultValues, ", ")},
						},
					},
				},
			}
		}

		fileNode.Decls = append(fileNode.Decls, funcDecl)
	})
}

func GetInterfaceWrite(interfaceName string, funcName string, paramNames []string, paramTypes []string, returnTypes []string) Writer {
	return WriteFunc(func(fileNode *ast.File) {
		var (
			ok            bool
			typeSpec      *ast.TypeSpec
			interfaceType *ast.InterfaceType
			methodField   *ast.Field
		)
		ast.Inspect(fileNode, func(x ast.Node) bool {
			if typeSpec, ok = x.(*ast.TypeSpec); !ok {
				return true
			}
			if interfaceType, ok = typeSpec.Type.(*ast.InterfaceType); !ok {
				return true
			}
			typeName := typeSpec.Name.Name
			if typeName != interfaceName {
				return true
			}

			paramFieldList := &ast.FieldList{}
			if len(paramNames) > 0 {
				lo.ForEach[string](paramNames, func(item string, index int) {
					paramFieldList.List = append(paramFieldList.List, &ast.Field{
						Names: []*ast.Ident{{Name: item}},
						Type:  GetIdent(paramTypes[index]),
					})
				})
			}
			returnFieldList := &ast.FieldList{}
			if len(returnTypes) > 0 {
				lo.ForEach[string](returnTypes, func(item string, index int) {
					returnFieldList.List = append(returnFieldList.List, &ast.Field{
						Type: GetIdent(item),
					})
				})
			}
			methodField = &ast.Field{
				Names: []*ast.Ident{{Name: funcName}},
				Type: &ast.FuncType{
					Params:  paramFieldList,
					Results: returnFieldList,
				},
			}
			interfaceType.Methods.List = append(interfaceType.Methods.List, methodField)
			return true
		})
	})
}
