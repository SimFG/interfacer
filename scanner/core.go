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
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var logger = zap.L().With(zap.String("package", "scanner"))

type PostParser interface {
	Post(structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo)
}

type WrapperFunc struct {
	CurrentName, InnerName string
	PostFunc               PostParserFunc
}

func (w *WrapperFunc) Post(structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo) {
	w.PostFunc(w.CurrentName, w.InnerName, structs, interfaces)
}

type PostParserFunc func(currentName string, innerName string, structs map[string]*StructInfo, interfaces map[string]*InterfaceInfo)

type Scanner struct {
	structs         map[string]*StructInfo
	interfaces      map[string]*InterfaceInfo
	packageStr      string
	rootPath        string
	postParserFuncs []PostParser
}

func New(p string, r string) *Scanner {
	return &Scanner{
		structs:    make(map[string]*StructInfo),
		interfaces: make(map[string]*InterfaceInfo),
		packageStr: p,
		rootPath:   r,
	}
}

func (s *Scanner) Start(dir string, excludeDir []string) error {
	excludeDirMap := tool.ToMap(excludeDir)
	if err := tool.FileWalk(dir, true, func(absPath string, fileInfo os.FileInfo) bool {
		if _, ok := excludeDirMap[fileInfo.Name()]; ok {
			return false
		}
		s.parseDir(absPath)
		return true
	}); err != nil {
		return err
	}

	lo.ForEach[PostParser](s.postParserFuncs, func(item PostParser, index int) {
		item.Post(s.structs, s.interfaces)
	})

	// TODO more go routine
	for _, structInfo := range s.structs {
		structInfo.Tokens()
	}

	for _, interfaceInfo := range s.interfaces {
		interfaceInfo.Tokens()
	}

	for _, structInfo := range s.structs {
		for _, interfaceInfo := range s.interfaces {
			if structInfo.HasImplementInterface(interfaceInfo) {
				interfaceInfo.structs = append(interfaceInfo.structs, structInfo)
			}
		}
	}

	//s.structs["github.com/milvus-io/milvus/internal/proxy.queryTask"].Print()
	return nil
}

func (s *Scanner) parseDir(dir string) error {
	fs := token.NewFileSet()
	result, err := parser.ParseDir(fs, dir, nil, 0)
	if err != nil {
		logger.Error("fail to parse dir", zap.Error(err))
		return err
	}
	for _, r := range result {
		lastSep := strings.LastIndex(dir, tool.FileSep)
		lastWord := dir[lastSep+1:]
		//curPackage := strings.Replace(dir[:lastSep], s.rootPath, s.packageStr, 1)
		//curPackage += tool.FileSep + r.Name
		curPackage := strings.Replace(dir, s.rootPath, s.packageStr, 1)
		if lastWord != r.Name {
			logger.Warn("package name is uncommon", zap.String("dir", dir), zap.String("package", r.Name))
		}
		//log.Println("package name:", r.Name, dir, curPackage)
		p := NewPackageParser(s, curPackage, dir, r)
		p.Parse()
	}
	return nil
}

func (s *Scanner) Print() {
	for _, info := range s.structs {
		info.Print()
	}

	for _, info := range s.interfaces {
		info.Print()
	}
}

func (s *Scanner) GetInterface(name string) *InterfaceInfo {
	interfaceInfo := s.interfaces[name]
	return interfaceInfo
}
