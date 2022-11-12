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

package tool

import (
	"fmt"
	"go.uber.org/zap"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

const FileSep = string(os.PathSeparator)

func StringsEqual(s, t []string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := range s {
		if s[i] != t[i] {
			return false
		}
	}
	return true
}

func Timer(name string, fn func()) {
	start := time.Now()
	fn()
	end := time.Now()
	Info("timer_"+name, zap.Time("start", start), zap.Time("end", end), zap.Int64("cost/ms", end.Sub(start).Milliseconds()))
}

// PathJoin path.Join()
func PathJoin(paths ...string) string {
	return strings.Join(paths, FileSep)
}

func FileNumInDir(dir string) int {
	k := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			k++
		}
		return nil
	})
	return k
}

// FileWalk fn, if the return value is false, it won't continue to walk
func FileWalk(dir string, onlyDir bool, fn func(absPath string, fileInfo os.FileInfo) bool) {
	Info("FileWalk", zap.String("dir", dir), zap.Bool("only_dir", onlyDir))

	rootInfo, err := os.Stat(dir)
	if err != nil {
		Panic("os.Stat err", zap.Error(err))
	}
	if onlyDir && !rootInfo.IsDir() {
		return
	}
	if ok := fn(dir, rootInfo); !ok {
		return
	}
	if !rootInfo.IsDir() {
		return
	}
	file, err := os.Open(dir)
	if err != nil {
		Panic("os.Open err", zap.Error(err))
	}

	fileInfos, err := file.Readdir(0)
	file.Close()
	if err != nil {
		Panic("file.Readdir err", zap.Error(err))
	}
	for _, fileInfo := range fileInfos {
		Info("sub_file", zap.String("file_name", fileInfo.Name()))
		if onlyDir && !fileInfo.IsDir() {
			continue
		}
		FileWalk(PathJoin(file.Name(), fileInfo.Name()), onlyDir, fn)
	}
}

// ToMap More functions like it, visit: https://github.com/samber/lo
func ToMap[T comparable](a []T) map[T]struct{} {
	b := make(map[T]struct{}, len(a))
	for _, t := range a {
		b[t] = struct{}{}
	}
	return b
}

// ImportName get the import name from the import spec value
func ImportName(i string) string {
	lastSep := strings.LastIndex(i, FileSep)
	if lastSep == -1 {
		return i
	}
	return i[lastSep+1:]
}

func TypeString(i interface{}) string {
	if i == nil {
		return ""
	}
	return reflect.TypeOf(i).String()
}

// HandleError panic directly
func HandleError(e error) {
	if e != nil {
		Panic("Panic HandleError", zap.Error(e))
	}
}

func HandleErrorWithMsg(e error, msg ...string) {
	if e != nil {
		fmt.Println(e, msg)
		Panic("Panic HandleErrorWithMsg", zap.Strings("msg", msg), zap.Error(e))
	}
}

func PrintDetail(objectName string, i interface{}) {
	Info("object detail", zap.String("objectName", objectName), zap.Any("detail", i))
}
