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
	"github.com/SimFG/interfacer/progress"
	"github.com/SimFG/interfacer/tool"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"sync"
	"time"
)

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
	enableImplement bool
	postParserFuncs []PostParser

	fileSum    int
	currentNum int
	startTime  time.Time
	done       chan struct{}
	lg         *progress.LineGroup
}

func New(p string, r string) *Scanner {
	tool.Info("Scanner New", zap.String("package", p), zap.String("path", r))
	return &Scanner{
		structs:         make(map[string]*StructInfo),
		interfaces:      make(map[string]*InterfaceInfo),
		packageStr:      p,
		rootPath:        r,
		enableImplement: true,
		lg:              &progress.LineGroup{},
		done:            make(chan struct{}),
	}
}

func (s *Scanner) DisableImplementRelation() {
	s.enableImplement = false
}

func (s *Scanner) SubModule(sub *Scanner, fullInterfaceName string, method string) {
	tool.Info("sub module", zap.String("full_interface_name", fullInterfaceName), zap.String("method", method))
	interfaceInfo, ok := sub.interfaces[fullInterfaceName]
	if !ok {
		tool.Panic("not found the interface name in the sub module")
	}
	if _, ok = s.interfaces[fullInterfaceName]; ok {
		tool.Panic("found the interface name in the root module")
	}
	interfaceInfo.ExcludeTokens([]string{method})
	for _, structInfo := range s.structs {
		if structInfo.HasImplementInterface(interfaceInfo) {
			interfaceInfo.structs = append(interfaceInfo.structs, structInfo)
		}
	}
	s.interfaces[fullInterfaceName] = interfaceInfo
}

func (s *Scanner) GetLineFunc(current int) func(i int, l *progress.Line) {
	return func(i int, l *progress.Line) {
		if i == 0 {
			l.Print("cost: %dms", time.Now().Sub(s.startTime).Milliseconds())
		} else if i == 1 {
			l.Print("progress: %d/%d", current, s.fileSum)
		}
	}
}

func (s *Scanner) Start(dir string, excludeDir []string) {
	tool.Info("Scanner Start", zap.String("dir", dir), zap.Strings("exclude_dir", excludeDir))
	excludeDirMap := tool.ToMap(excludeDir)

	fmt.Println("start to scan the dir:", dir)
	s.lg.SetLineNum(2)
	defer func() {
		s.lg.End(s.GetLineFunc(s.fileSum))
	}()
	s.fileSum = tool.FileNumInDir(dir)
	s.startTime = time.Now()
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-time.After(500 * time.Millisecond):
				s.lg.Print(s.GetLineFunc(s.currentNum))
			}
		}
	}()

	tool.FileWalk(dir, true, func(absPath string, fileInfo os.FileInfo) bool {
		tool.Info("File Walk inner", zap.String("abs_path", absPath))
		if _, ok := excludeDirMap[fileInfo.Name()]; ok {
			s.currentNum += tool.FileNumInDir(absPath)
			return false
		}
		s.parseDir(absPath)
		return true
	})
	close(s.done)

	lo.ForEach[PostParser](s.postParserFuncs, func(item PostParser, _ int) {
		item.Post(s.structs, s.interfaces)
	})

	// TODO more go routine
	w := sync.WaitGroup{}
	w.Add(1)
	go func() {
		defer w.Done()
		for _, structInfo := range s.structs {
			structInfo.Tokens()
		}
	}()

	w.Add(1)
	go func() {
		defer w.Done()
		for _, interfaceInfo := range s.interfaces {
			interfaceInfo.Tokens()
		}
	}()
	w.Wait()

	if !s.enableImplement {
		return
	}

	for _, structInfo := range s.structs {
		for _, interfaceInfo := range s.interfaces {
			if structInfo.HasImplementInterface(interfaceInfo) {
				interfaceInfo.structs = append(interfaceInfo.structs, structInfo)
			}
		}
	}
}

func (s *Scanner) parseDir(dir string) error {
	tool.Info("Scanner parseDir", zap.String("dir", dir))

	fs := token.NewFileSet()
	result, err := parser.ParseDir(fs, dir, nil, 0)
	tool.HandleErrorWithMsg(err, "fail to parse dir:", dir)

	for _, r := range result {
		lastSep := strings.LastIndex(dir, tool.FileSep)
		lastWord := dir[lastSep+1:]
		curPackage := strings.Replace(dir, s.rootPath, s.packageStr, 1)
		if lastWord != r.Name {
			tool.Info("WARN package name is uncommon", zap.String("dir", dir), zap.String("package", r.Name))
		}
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
	if interfaceInfo != nil {
		interfaceInfo.Print()
	}
	return interfaceInfo
}
