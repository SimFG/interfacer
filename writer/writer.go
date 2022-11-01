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
	"bufio"
	"bytes"
	"github.com/SimFG/interfacer/tool"
	"github.com/samber/lo"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func WriteFile(fileName string, writers []Writer) {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	tool.HandleError(err)
	for _, writer := range writers {
		writer.Write(fset, fileNode)
	}

	err = format.Node(&buf, fset, fileNode)
	tool.HandleErrorWithMsg(err, "format node filename:", fileName)
	// TODO handle like:
	//type Component struct {
	//}
	//
	//func (Component) Dummy(){
	//}

	err = ioutil.WriteFile(fileName, buf.Bytes(), 0)
	tool.HandleErrorWithMsg(err, "write node filename:", fileName)
}

func WriteFileForLine(fileName string, writers []Writer) {
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	tool.HandleError(err)
	for _, writer := range writers {
		writer.Write(fset, fileNode)
	}
}

type Writer interface {
	Write(fset *token.FileSet, fileNode *ast.File)
}

type WriteFunc func(fset *token.FileSet, fileNode *ast.File)

func (w WriteFunc) Write(fset *token.FileSet, fileNode *ast.File) {
	w(fset, fileNode)
}

func GetImportWriter(alia string, importValue string) Writer {
	return WriteFunc(func(fset *token.FileSet, fileNode *ast.File) {
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
	return WriteFunc(func(fset *token.FileSet, fileNode *ast.File) {
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
	return WriteFunc(func(fset *token.FileSet, fileNode *ast.File) {
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

// GetInterfaceWrite2 dismiss the influence of the comment
func GetInterfaceWrite2(fileName string, interfaceName string, method string) Writer {
	return WriteFunc(func(fset *token.FileSet, fileNode *ast.File) {
		var (
			ok       bool
			typeSpec *ast.TypeSpec
		)
		ast.Inspect(fileNode, func(x ast.Node) bool {
			if typeSpec, ok = x.(*ast.TypeSpec); !ok {
				return true
			}
			if _, ok = typeSpec.Type.(*ast.InterfaceType); !ok {
				return true
			}
			typeName := typeSpec.Name.Name
			if typeName != interfaceName {
				return true
			}
			pos := fset.Position(x.End())
			FileInsertContent(fileName, pos.Line-1, method)
			return false
		})
	})

}

func FileInsertContent(fileName string, line int, content string) {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0)
	tool.HandleErrorWithMsg(err, "File open failed!")

	reader := bufio.NewReader(file)

	tempFile, err := os.OpenFile(fileName+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0766)
	tool.HandleErrorWithMsg(err, "Temp create failed!")

	writer := bufio.NewWriter(tempFile)
	_ = writer.Flush()
	for i := 0; i < line; i++ {
		l, err := reader.ReadString('\n')
		tool.HandleErrorWithMsg(err, "File raed failed!")

		_, _ = writer.WriteString(l)
		_ = writer.Flush()
	}
	_, _ = tempFile.WriteString(content + "\n")
	for {
		l, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		tool.HandleErrorWithMsg(err, "File raed failed!")
		_, _ = writer.WriteString(l)
	}
	_ = writer.Flush()

	file.Close()
	tempFile.Close()
	err = os.Rename(fileName+".tmp", fileName)
	tool.HandleErrorWithMsg(err, "Rename file raed failed!")
}
