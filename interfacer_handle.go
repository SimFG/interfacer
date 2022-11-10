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

package main

import (
	"errors"
	"github.com/SimFG/interfacer/scanner"
	"github.com/SimFG/interfacer/tool"
	"github.com/SimFG/interfacer/writer"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"strings"
)

func WriteMethod(s *scanner.Scanner, interfaceFullName string, newMethod string, returnDefaultValues string, skipInterface bool) {
	interfaceInfo := s.GetInterface(interfaceFullName)
	if interfaceInfo == nil {
		tool.HandleErrorWithMsg(errors.New("not found the interface"), "interface name:", interfaceFullName)
	}

	interfaceName := interfaceFullName[strings.LastIndex(interfaceFullName, ".")+1:]
	i := strings.Index(newMethod, "(")
	j := strings.Index(newMethod, ")")
	x := strings.LastIndex(newMethod, "(")
	y := strings.LastIndex(newMethod, ")")
	tool.Info("new method split", zap.Ints("splits", []int{i, j, x, y}))

	var (
		funcName       string
		paramNames     []string
		paramTypes     []string
		returnTypes    []string
		returnDefaults []string
	)

	funcName = newMethod[:i]
	returnDefaults = strings.Split(returnDefaultValues, ",")

	lo.ForEach[string](strings.Split(newMethod[i+1:j], ","), func(item string, index int) {
		item = strings.TrimSpace(item)
		if item == "" {
			return
		}
		paramInfo := strings.Split(item, " ")
		paramNames = append(paramNames, paramInfo[0])
		paramTypes = append(paramTypes, paramInfo[1])
	})
	if i == x && j == y {
		returnType := strings.TrimSpace(newMethod[j+1:])
		if returnType != "" {
			returnTypes = append(returnTypes, strings.TrimSpace(newMethod[j+1:]))
		}
	} else {
		lo.ForEach[string](strings.Split(newMethod[x+1:y], ","), func(item string, index int) {
			item = strings.TrimSpace(item)
			if item == "" {
				return
			}
			returnTypes = append(returnTypes, item)
		})
	}
	tool.Info("method signature", zap.String("func_name", funcName),
		zap.Strings("param_names", paramNames), zap.Strings("param_types", paramTypes),
		zap.Strings("return_types", returnTypes), zap.Strings("return_defaults", returnDefaults))

	if !skipInterface {
		interfaceFileName := interfaceInfo.FilePaths()[0]
		writer.WriteFileForLine(interfaceFileName, []writer.Writer{writer.GetInterfaceWrite2(interfaceFileName, interfaceName, "\t"+newMethod)})
	}
	lo.ForEach[*scanner.StructInfo](interfaceInfo.GetImplements(), func(item *scanner.StructInfo, index int) {
		writePath := item.FilePaths()[0]
		if p, ok := writePaths[item.Name()]; ok {
			writePath = p
		}
		receiverName, receiverType := item.MethodReceiver()
		writer.WriteFile(writePath, []writer.Writer{writer.GetFuncWriter(receiverName, receiverType, funcName, paramNames, paramTypes, returnTypes, returnDefaults)})
	})
}
