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
	"github.com/samber/lo"
	"go.uber.org/zap"
	"os"
	"strings"
)

type ConfigChecker struct {
}

// CheckWritePaths the path should be "xxxx,xxxx"
func (c ConfigChecker) CheckWritePaths(paths []string) {
	lo.ForEach[string](paths, func(item string, index int) {
		if strings.Index(item, ",") < 0 {
			Panic("invalid write path", zap.String("path", item))
		}
	})
}

// CheckProjectDir check whether the dir param is existed and valid
func (c ConfigChecker) CheckProjectDir(dir string) {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		Panic("invalid project dir", zap.String("dir", dir))
	}
	if !fileInfo.IsDir() {
		Panic("the dir param isn't a dir", zap.String("dir", dir))
	}
}

// CheckModuleName the module shouldn't be empty
func (c ConfigChecker) CheckModuleName(module string) {
	if module == "" {
		Panic("the module param shouldn't be empty")
	}
}

// CheckInterface check the method and the return default values accord the method, if the interface name isn't empty
func (c ConfigChecker) CheckInterface(interfaceName string, method string, returnDefaultValues string) {
	if interfaceName == "" {
		return
	}

	method = strings.TrimSpace(method)
	i := strings.Index(method, "(")
	j := strings.Index(method, ")")
	x := strings.LastIndex(method, "(")
	y := strings.LastIndex(method, ")")

	if i < 0 || j < 0 || i > j {
		Panic("the method should contain the '(' and ')' orderly",
			zap.String("method", method))
	}

	if i == 0 {
		Panic("the method should contain a method name",
			zap.String("method", method))
	}

	if x > y {
		Panic("the method should contain the '(' and ')' orderly for the more return values",
			zap.String("method", method))
	}

	params := strings.TrimSpace(method[i+1 : j])
	if params != "" {
		lo.ForEach[string](strings.Split(params, ","), func(item string, index int) {
			if item == "" {
				Panic("the method name is invalid ",
					zap.String("method", method))
			}
		})
	}

	returns := strings.TrimSpace(method[j+1:])
	returnParams := 0
	if returns != "" {
		if strings.Contains(returns, ",") && x < j {
			Panic("the method has more return values that should be included by the bracket, '()'",
				zap.String("method", method))
		}
		returnParams = len(strings.Split(returns, ","))
	}

	if returnDefaultValues != "" {
		if len(strings.Split(returnDefaultValues, ",")) != returnParams {
			Panic("the number of the return default values should be equal the number of the method return values",
				zap.String("method", method), zap.String("default_values", returnDefaultValues))
		}
	}
}
