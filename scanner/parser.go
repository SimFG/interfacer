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

package scanner

import (
	"github.com/SimFG/interfacer/tool"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go/ast"
	"strings"
)

type PackageParser struct {
	scanner            *Scanner
	curPack            string
	curDir             string
	astPack            *ast.Package
	structs            []string
	interfaces         []string
	innerStructPost    map[string][]string
	innerInterfacePost map[string][]string
}

func NewPackageParser(s *Scanner, curPack string, curDir string, astPack *ast.Package) *PackageParser {
	tool.Info("NewPackageParser", zap.String("cur_pack", curPack), zap.String("cur_dir", curDir))
	return &PackageParser{
		scanner:            s,
		curPack:            curPack,
		curDir:             curDir,
		astPack:            astPack,
		innerStructPost:    make(map[string][]string),
		innerInterfacePost: make(map[string][]string),
	}
}

func (p *PackageParser) Parse() {
	for name, file := range p.astPack.Files {
		p.ParseFile(name, file)
	}

	p.handInner()
}

func (p *PackageParser) handInner() {
	tool.Info("handInner innerStructPost")
	for s, i := range p.innerStructPost {
		tool.Info(s, zap.Strings("inners", i))
		lo.ForEach[string](i, func(item string, _ int) {
			fullName := p.curPack + "." + item
			if lo.Contains[string](p.structs, item) {
				p.scanner.postParserFuncs = append(p.scanner.postParserFuncs, &WrapperFunc{
					CurrentName: s,
					InnerName:   fullName,
					PostFunc: func(currentName string, innerName string, structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo) {
						if structs[currentName] != nil {
							if structs[innerName] != nil {
								structs[currentName].innerStruct = append(structs[currentName].innerStruct, structs[innerName])
							}
						}
					},
				})
			}

			if lo.Contains[string](p.interfaces, item) {
				p.scanner.postParserFuncs = append(p.scanner.postParserFuncs, &WrapperFunc{
					CurrentName: s,
					InnerName:   fullName,
					PostFunc: func(currentName string, innerName string, structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo) {
						if structs[currentName] != nil {
							if interfaces[innerName] != nil {
								structs[currentName].innerInterface = append(structs[currentName].innerInterface, interfaces[innerName])
							}
						}
					},
				})
			}
		})
	}

	tool.Info("handInner innerInterfacePost")
	for s, i := range p.innerInterfacePost {
		tool.Info(s, zap.Strings("inners", i))
		lo.ForEach[string](i, func(item string, _ int) {
			fullName := p.curPack + "." + item
			if lo.Contains[string](p.interfaces, item) {
				p.scanner.postParserFuncs = append(p.scanner.postParserFuncs, &WrapperFunc{
					CurrentName: s,
					InnerName:   fullName,
					PostFunc: func(currentName string, innerName string, structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo) {
						if interfaces[currentName] != nil {
							if interfaces[innerName] != nil {
								interfaces[currentName].innerInterface = append(interfaces[currentName].innerInterface, interfaces[innerName])
							}
						}
					},
				})
			}
		})
	}
}

// GetValueFromType Get value from the `*ast.Ident`/`*ast.SelectorExpr`/`*ast.StarExpr`
// TODO handle the map / slice / func
func (p *PackageParser) GetValueFromType(e ast.Expr) string {
	tool.Info("GetValueFromType", zap.String("type", tool.TypeString(e)))
	switch e.(type) {
	case *ast.SelectorExpr:
		selectExpr := e.(*ast.SelectorExpr)
		// if the inner interface is "http.File"
		// selectExpr.Sel.Name -> File
		// selectExpr.X.(*ast.Ident).Name -> http
		return selectExpr.X.(*ast.Ident).Name + "." + selectExpr.Sel.Name
	case *ast.Ident:
		return e.(*ast.Ident).Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StarExpr:
		return "*" + p.GetValueFromType(e.(*ast.StarExpr).X)
	default:
		tool.Info("WARN GetValueFromType")
		return ""
	}

}

func (p *PackageParser) HandleFieldListForInterface(fields *ast.FieldList, f func(value string, namesLen int)) {
	if fields == nil {
		return
	}
	for _, param := range fields.List {
		tool.Info("HandleFieldListForInterface param", zap.Any("names", param.Names))
		value := p.GetValueFromType(param.Type)
		times := len(param.Names)
		if times == 0 {
			times = 1
		}
		f(value, times)
	}
}

func (p *PackageParser) HandleFuncType(funcType *ast.FuncType, methodInfo *MethodInfo) {
	p.HandleFieldListForInterface(funcType.Params, func(value string, namesLen int) {
		tool.IfF(value != "", func() {
			tool.Times(namesLen, func(index int) {
				methodInfo.params = append(methodInfo.params, value)
			})
		})
	})
	p.HandleFieldListForInterface(funcType.Results, func(value string, namesLen int) {
		tool.IfF(value != "", func() {
			tool.Times(namesLen, func(index int) {
				methodInfo.returns = append(methodInfo.returns, value)
			})
		})
	})
}

func (p *PackageParser) ParseFile(fileFullPath string, astFile *ast.File) {
	tool.Info("PackageParser ParseFile", zap.String("file_full_path", fileFullPath), zap.Any("ast_file_name", astFile.Name))

	var (
		importList      = make(map[string]string)
		structList      = make(map[string]*StructInfo)
		interfaceList   = make(map[string]*InterfaceInfo)
		funcList        = make(map[string][]*MethodInfo) // the method maybe use the struct which isn't scanned
		innerInterfaces = make(map[string][]string)
		innerStructs    = make(map[string][]string)
	)

	ast.Inspect(astFile, func(x ast.Node) bool {
		switch x.(type) {
		case *ast.ImportSpec:
			importSpec := x.(*ast.ImportSpec)
			v := strings.Trim(importSpec.Path.Value, "\"")
			if importSpec.Name != nil {
				importList[importSpec.Name.Name] = v
			} else {
				importList[tool.ImportName(v)] = v
			}
		case *ast.TypeSpec:
			typeSpec := x.(*ast.TypeSpec)
			typeName := typeSpec.Name.Name
			baseInfo := &BaseInfo{name: p.curPack + "." + typeName, packageName: p.curPack, filePaths: []string{fileFullPath}}
			switch typeSpec.Type.(type) {
			case *ast.StructType:
				structType := typeSpec.Type.(*ast.StructType)
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						value := p.GetValueFromType(field.Type)
						innerStructs[typeName] = append(innerStructs[typeName], value)
					}
				}
				structList[typeName] = &StructInfo{BaseInfo: baseInfo, methods: make(map[string]*MethodInfo)}
			case *ast.InterfaceType:
				interfaceList[typeName] = &InterfaceInfo{BaseInfo: baseInfo}
				interfaceType := typeSpec.Type.(*ast.InterfaceType)
				if interfaceType.Methods != nil {
					info := interfaceList[typeName]
					for _, filed := range interfaceType.Methods.List {
						switch filed.Type.(type) {
						case *ast.FuncType: // TODO support generics type
							funcName := filed.Names[0].Name
							funcType := filed.Type.(*ast.FuncType)
							methodInfo := &MethodInfo{name: funcName}
							p.HandleFuncType(funcType, methodInfo)

							info.methods = append(info.methods, methodInfo)
						default:
							// get inner interface
							value := p.GetValueFromType(filed.Type)
							if value == "" {
								tool.Info("default field type spec type", zap.String("type", tool.TypeString(filed.Type)))
							} else {
								innerInterfaces[typeName] = append(innerInterfaces[typeName], value)
							}
						}

					}
				}
			default:
				tool.Info("default type spec type", zap.String("type", tool.TypeString(typeSpec.Type)))
			}
		case *ast.FuncDecl:
			funcDecl := x.(*ast.FuncDecl)
			tool.IfF(funcDecl.Recv != nil, func() {
				funcName := funcDecl.Name.Name
				methodInfo := &MethodInfo{name: funcName}
				structName := p.GetValueFromType(funcDecl.Recv.List[0].Type)
				if len(structName) == 0 {
					//tool.PrintDetail("FuncDecl-receiver", funcDecl.Recv.List[0].Type)
					// log.Println(funcName, "-file", fileFullPath)
				} else {
					if len(funcDecl.Recv.List[0].Names) != 0 {
						methodInfo.receiverName = funcDecl.Recv.List[0].Names[0].Name
					} else {
						//fmt.Println("Test Fubang", "structName", structName, "file", fileFullPath)
						return
					}
					methodInfo.receiverType = structName
					tool.IfF(structName[0] == '*', func() {
						methodInfo.isPointReceiver = true
						structName = structName[1:]
					})
					p.HandleFuncType(funcDecl.Type, methodInfo)
					//log.Println("funcDecl", fmt.Sprintf("%#v", methodInfo))
					funcList[structName] = append(funcList[structName], methodInfo)
				}
			})
		default:
			if x != nil {
				tool.Info("default node type", zap.String("type", tool.TypeString(x)))
			}
		}
		return true
	})

	tool.Info("import list", zap.Any("value", importList))
	tool.Info("struct list", zap.Any("value", structList))
	tool.Info("interface list", zap.Any("value", interfaceList))
	tool.Info("func list", zap.Any("value", funcList))
	tool.Info("inner interface list", zap.Any("value", innerInterfaces))
	tool.Info("inner struct list", zap.Any("value", innerStructs))

	// convert the simple fileFullPath to full fileFullPath
	handleName := func(name string) string {
		if _, ok := structList[name]; ok {
			return p.curPack + "." + name
		}
		if len(name) > 0 && name[0] == '*' {
			if _, ok := structList[name[1:]]; ok {
				return "*" + p.curPack + "." + name[1:]
			}
		}
		if _, ok := interfaceList[name]; ok {
			return p.curPack + "." + name
		}
		i := strings.Index(name, ".")

		if i > 0 {
			s := 0
			if name[0] == '*' {
				s = 1
			}
			spaceName := name[s:i]
			return lo.If[string](s == 1, "*").Else("") + importList[spaceName] + "." + name[i+1:]
		}
		return name
	}

	handMethod := func(item *MethodInfo) {
		item.params = lo.Map[string, string](item.params, func(item2 string, index int) string {
			return handleName(item2)
		})

		item.returns = lo.Map[string, string](item.returns, func(item2 string, index int) string {
			return handleName(item2)
		})
	}

	for _, info := range interfaceList {
		lo.ForEach[*MethodInfo](info.methods, func(item *MethodInfo, index int) {
			handMethod(item)
		})
		p.scanner.interfaces[info.name] = info
	}

	handleInnerName := func(name string) (string, bool) {
		if name[0] == '*' {
			name = name[1:]
		}
		pointIndex := strings.Index(name, ".")
		if pointIndex > 0 {
			return importList[name[0:pointIndex]] + name[pointIndex:], true
		}
		if structList[name] != nil || interfaceList[name] != nil {
			return p.curPack + "." + name, true
		}

		return name, false
	}

	for i, inners := range innerStructs {
		fullStructName := p.curPack + "." + i
		lo.ForEach[string](inners, func(item string, _ int) {
			v, ok := handleInnerName(item)
			if ok {
				p.scanner.postParserFuncs = append(p.scanner.postParserFuncs, &WrapperFunc{
					CurrentName: fullStructName,
					InnerName:   v,
					PostFunc: func(currentName string, innerName string, structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo) {
						if structs[currentName] != nil {
							if structs[innerName] != nil {
								structs[currentName].innerStruct = append(structs[currentName].innerStruct, structs[innerName])
							}
							if interfaces[innerName] != nil {
								structs[currentName].innerInterface = append(structs[currentName].innerInterface, interfaces[innerName])
							}
						}
					},
				})

			} else {
				p.innerStructPost[fullStructName] = append(p.innerStructPost[fullStructName], v)
			}
		})
	}

	for i, inners := range innerInterfaces {
		fullInterfaceName := p.curPack + "." + i
		lo.ForEach[string](inners, func(item string, _ int) {
			v, ok := handleInnerName(item)
			if ok {
				p.scanner.postParserFuncs = append(p.scanner.postParserFuncs, &WrapperFunc{
					CurrentName: fullInterfaceName,
					InnerName:   v,
					PostFunc: func(currentName string, innerName string, structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo) {
						if interfaces[currentName] != nil {
							if interfaces[innerName] != nil {
								interfaces[currentName].innerInterface = append(interfaces[currentName].innerInterface, interfaces[innerName])
							}
						}
					},
				})
			} else {
				p.innerInterfacePost[fullInterfaceName] = append(p.innerInterfacePost[fullInterfaceName], v)
			}
		})
	}

	for _, info := range structList {
		if _, ok := p.scanner.structs[info.name]; ok {
			curStructInfo := p.scanner.structs[info.name]
			curStructInfo.packageName = info.packageName
			curStructInfo.filePaths = append(curStructInfo.filePaths, info.filePaths...)
			continue
		}
		p.scanner.structs[info.name] = info
	}

	for i, funcs := range funcList {
		fullName := handleName(i)
		// Methods and structs are in different files, but the corresponding methods are scanned first
		if i == fullName || i == "*"+fullName {
			fullName = p.curPack + "." + fullName
		}
		structInfo := p.scanner.structs[fullName]
		if structInfo == nil {
			structInfo = &StructInfo{BaseInfo: &BaseInfo{name: fullName, filePaths: []string{fileFullPath}}, methods: make(map[string]*MethodInfo)}
			p.scanner.structs[structInfo.name] = structInfo
		}
		lo.ForEach[*MethodInfo](funcs, func(item *MethodInfo, _ int) {
			handMethod(item)
			structInfo.methods[item.name] = item
		})
	}

	p.structs = append(p.structs, lo.Keys[string, *StructInfo](structList)...)
	p.interfaces = append(p.interfaces, lo.Keys[string, *InterfaceInfo](interfaceList)...)
}
