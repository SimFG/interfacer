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
	"fmt"
	"github.com/SimFG/interfacer/tool"
	"github.com/samber/lo"
	"log"
	"sort"
	"strings"
)

type BaseInfo struct {
	filePaths   []string
	packageName string
	name        string
	tokens      []string
}

func (b *BaseInfo) FilePaths() []string {
	return b.filePaths
}

func (b *BaseInfo) Name() string {
	return b.name
}

type StructInfo struct {
	*BaseInfo
	// map is easy to check whether it has implemented the interface
	methods map[string]*MethodInfo
	// TODO
	innerStruct    []*StructInfo
	innerInterface []*InterfaceInfo
}

func (s *StructInfo) CheckMethod(m *MethodInfo) bool {
	sm, ok := s.methods[m.name]
	if ok && sm.Equal(m) {
		return true
	}

	for _, inner := range s.innerStruct {
		if inner.CheckMethod(m) {
			return true
		}
	}

	for _, inner := range s.innerInterface {
		if inner.CheckMethod(m) {
			return true
		}
	}

	return false
}

func (s *StructInfo) Tokens() {
	s.tokens = s.innerToken()
	sort.Strings(s.tokens)
}

func (s *StructInfo) innerToken() []string {
	var tokens []string
	for _, info := range s.methods {
		tokens = append(tokens, info.token())
	}
	lo.ForEach[*InterfaceInfo](s.innerInterface, func(item *InterfaceInfo, index int) {
		tokens = append(tokens, item.innerToken()...)
	})
	lo.ForEach[*StructInfo](s.innerStruct, func(item *StructInfo, index int) {
		tokens = append(tokens, item.innerToken()...)
	})
	return tokens
}

// TOOD stack overflow
func (s *StructInfo) HasImplementInterface(i *InterfaceInfo) bool {
	var x, y int

	for x < len(s.tokens) && y < len(i.tokens) {
		if s.tokens[x] == i.tokens[y] {
			x++
			y++
		} else if s.tokens[x] < i.tokens[y] {
			x++
		} else {
			return false
		}
	}

	return y == len(i.tokens)

	//for _, m := range i.methods {
	//	if !s.CheckMethod(m) {
	//		return false
	//	}
	//}
	//
	//for _, innerInterface := range i.innerInterface {
	//	if !s.HasImplementInterface(innerInterface) {
	//		return false
	//	}
	//}
	//return true
}

func (s *StructInfo) MethodReceiver() (receiverName string, receiverType string) {
	for _, info := range s.methods {
		if info.receiverName != "" {
			return info.receiverName, info.receiverType
		}
	}
	return strings.ToLower(s.name), "*" + s.name
}

func (s *StructInfo) Print() {
	log.Println("struct:", s.name, "package:", s.packageName)
	log.Println("filePath:", s.filePaths)
	log.Println("methods:")
	for _, method := range s.methods {
		method.Print()
	}
	log.Println("innerStructs:")
	for _, inner := range s.innerStruct {
		log.Println(inner.name)
	}
	log.Println("innerInterface:")
	for _, inner := range s.innerInterface {
		log.Println(inner.name)
	}
	log.Println("--------------------")
}

type InterfaceInfo struct {
	*BaseInfo
	innerInterface []*InterfaceInfo
	methods        []*MethodInfo
	structs        []*StructInfo // implement the interfaceinnerInterface
}

func (i *InterfaceInfo) CheckMethod(m *MethodInfo) bool {
	for _, mi := range i.methods {
		if mi.Equal(m) {
			return true
		}
	}

	for _, inner := range i.innerInterface {
		if inner.CheckMethod(m) {
			return true
		}
	}

	return false
}

func (i *InterfaceInfo) GetImplements() []*StructInfo {
	return i.structs
}

func (i *InterfaceInfo) Tokens() {
	i.tokens = i.innerToken()
	sort.Strings(i.tokens)
}

func (i *InterfaceInfo) innerToken() []string {
	var tokens []string
	lo.ForEach[*MethodInfo](i.methods, func(item *MethodInfo, index int) {
		tokens = append(tokens, item.token())
	})
	lo.ForEach[*InterfaceInfo](i.innerInterface, func(item *InterfaceInfo, index int) {
		tokens = append(tokens, item.innerToken()...)
	})
	return tokens
}

func (i *InterfaceInfo) Print() {
	log.Println("interface:", i.name, "package:", i.packageName)
	log.Println("filePath:", i.filePaths)
	log.Println("implements:")
	for _, structInfo := range i.structs {
		log.Println(structInfo.name, "path:", structInfo.filePaths)
	}
	log.Println("methods:")
	for _, method := range i.methods {
		method.Print()
	}
	log.Println("inner:")
	for _, inner := range i.innerInterface {
		log.Println(inner.name)
	}
	log.Println("--------------------")
}

type MethodInfo struct {
	name            string
	isPointReceiver bool
	receiverType    string
	receiverName    string
	types           []string
	params          []string
	returns         []string
}

func (m *MethodInfo) token() string {
	return fmt.Sprintf("%s%v%v", m.name, m.params, m.returns)
}

func (m *MethodInfo) Equal(t *MethodInfo) bool {
	return m.name == t.name &&
		tool.StringsEqual(m.params, t.params) &&
		tool.StringsEqual(m.returns, t.returns)
}

func (m *MethodInfo) Print() {
	log.Println(m.name, "params:", m.params, "returns:", m.returns)
}
