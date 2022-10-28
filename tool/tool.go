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
	"log"
	"os"
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
	log.Println(name, "start:", start, "end:", end, "cost:", end.Second()-start.Second())
}

// PathJoin path.Join()
func PathJoin(paths ...string) string {
	return strings.Join(paths, FileSep)
}

// FileWalk fn, if the return value is false, it won't continue to walk
func FileWalk(dir string, onlyDir bool, fn func(absPath string, fileInfo os.FileInfo) bool) error {
	rootInfo, err := os.Stat(dir)
	if err != nil {
		log.Println("os.Stat err", dir, err)
		return err
	}
	if onlyDir && !rootInfo.IsDir() {
		return nil
	}
	if ok := fn(dir, rootInfo); !ok {
		return nil
	}
	if !rootInfo.IsDir() {
		return nil
	}
	file, err := os.Open(dir)
	if err != nil {
		log.Println("os.Open err", err)
	}

	//log.Println("fileName", file.Name())
	fileInfos, err := file.Readdir(0)
	file.Close()
	if err != nil {
		log.Println("file.Readdir err", err)
		return err
	}
	for _, fileInfo := range fileInfos {
		if onlyDir && !fileInfo.IsDir() {
			continue
		}
		if err = FileWalk(PathJoin(file.Name(), fileInfo.Name()), onlyDir, fn); err != nil {
			return err
		}
	}
	return nil
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

func PrintDetail(detail string, i interface{}) {
	log.Println(detail+" - PrintDetail", fmt.Sprintf("%#v", i))
}
