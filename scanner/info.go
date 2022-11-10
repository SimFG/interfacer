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
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
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
	methods        map[string]*MethodInfo
	innerStruct    []*StructInfo
	innerInterface []*InterfaceInfo
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

func (s *StructInfo) HasImplementInterface(i *InterfaceInfo) bool {
	var x, y int

	for x < len(s.tokens) && y < len(i.tokens) {
		if s.tokens[x] == i.tokens[y] {
			x++
			y++
		} else if s.tokens[x] < i.tokens[y] {
			x++
		} else if slices.Contains(i.excludeTokens, i.tokens[y]) {
			y++
		} else {
			return false
		}
	}

	return y == len(i.tokens)
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
	tool.Record(zap.String("struct", s.name), zap.String("package", s.packageName),
		zap.Strings("file_paths", s.filePaths))
	innerStructNames := lo.Map[*StructInfo, string](s.innerStruct, func(item *StructInfo, _ int) string {
		return item.name
	})
	tool.Record(zap.Strings("inner_structs", innerStructNames))

	innerInterfaceNames := lo.Map[*InterfaceInfo, string](s.innerInterface, func(item *InterfaceInfo, _ int) string {
		return item.name
	})
	tool.Record(zap.Strings("inner_interfaces", innerInterfaceNames))
	tool.Record(zap.String("methods", ""))
	for _, method := range s.methods {
		method.Print()
	}
}

type InterfaceInfo struct {
	*BaseInfo
	innerInterface []*InterfaceInfo
	methods        []*MethodInfo
	structs        []*StructInfo // implement the interface
	excludeTokens  []string
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

func (i *InterfaceInfo) ExcludeTokens(methods []string) {
	lo.ForEach[string](methods, func(item string, index int) {
		if token := i.innerExcludeToken(item); token != "" {
			i.excludeTokens = append(i.excludeTokens, token)
		}
	})
}

func (i *InterfaceInfo) innerExcludeToken(method string) string {
	var token string
	for _, item := range i.methods {
		if item.name == method {
			token = item.token()
			return token
		}
	}
	for _, item := range i.innerInterface {
		if token = item.innerExcludeToken(method); token != "" {
			return token
		}
	}
	return ""
}

func (i *InterfaceInfo) Print() {
	tool.Record(zap.String("interface", i.name), zap.String("package_name", i.packageName),
		zap.Strings("file_paths", i.filePaths))
	tool.Record(zap.String("implements", ""))
	for _, structInfo := range i.structs {
		tool.Record(zap.String("struct_name", structInfo.name), zap.Strings("file_path", structInfo.filePaths))
	}
	innerInterfaceNames := lo.Map[*InterfaceInfo, string](i.innerInterface, func(item *InterfaceInfo, _ int) string {
		return item.name
	})
	tool.Record(zap.Strings("inner_interfaces", innerInterfaceNames))
	tool.Record(zap.String("methods", ""))
	for _, method := range i.methods {
		method.Print()
	}
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
	tool.Record(zap.String("name", m.name), zap.Strings("params", m.params), zap.Strings("returns", m.returns))
}
