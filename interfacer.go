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
	"fmt"
	"github.com/SimFG/interfacer/scanner"
	"github.com/SimFG/interfacer/tool"
	"github.com/SimFG/interfacer/writer"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type SubModule struct {
	ProjectDir    string   `yaml:"project_dir"`
	ProjectModule string   `yaml:"project_module"`
	ExcludeDirs   []string `yaml:"exclude_dirs,flow"`
}

type Config struct {
	WritePaths          []string `yaml:"write_paths,flow"`
	ExcludeDirs         []string `yaml:"exclude_dirs,flow"`
	ProjectDir          string   `yaml:"project_dir"`
	ProjectModule       string   `yaml:"project_module"`
	InterfaceFullName   string   `yaml:"interface_full_name"`
	NewMethod           string   `yaml:"new_method"`
	ReturnDefaultValues string   `yaml:"return_default_values"`
	EnableRecord        bool     `yaml:"enable_record"`
	EnableDebug         bool     `yaml:"enable_debug"`
}

var (
	interfacer = &cobra.Command{
		Use:   "interfacer",
		Short: "Implement a method of the interface anywhere",
		Run:   implement,
	}

	yamlFile            string
	projectDir          string
	projectModule       string
	interfaceFullName   string
	newMethod           string
	returnDefaultValues string
	writePaths          = make(map[string]string)
	config              = &Config{}
)

func init() {
	interfacer.Flags().StringVar(&yamlFile, "yaml-file", "interfacer.yaml", "full project dir")
	interfacer.Flags().StringVar(&projectDir, "project-dir", config.ProjectDir, "full project dir")
	interfacer.Flags().StringVar(&projectModule, "project-module", config.ProjectModule, "project module")
	interfacer.Flags().StringVar(&interfaceFullName, "interface", config.InterfaceFullName, "interface full name, like: go.uber.org/zap/zapcore.Core")
	interfacer.Flags().StringVar(&newMethod, "method", config.NewMethod, "the method declaration")
	interfacer.Flags().StringVar(&returnDefaultValues, "returns", config.ReturnDefaultValues, "the return value of the method, like: nil,nil")

	tool.Info("cmd params", zap.String("yaml-file", yamlFile), zap.String("project_dir", projectDir), zap.String("project_module", projectModule),
		zap.String("interface_full_name", interfaceFullName), zap.String("method", newMethod),
		zap.String("return_default_value", returnDefaultValues), zap.Any("config", config))
}

func readYaml() {
	defer func() {
		if e := recover(); e != nil {
			tool.Info("panic readYaml", zap.Any("err", e))
		}
	}()

	_, err := os.Stat(yamlFile)
	tool.HandleErrorWithMsg(err, "fail to stat "+yamlFile)

	f, err := os.Open(yamlFile)
	tool.HandleErrorWithMsg(err, "fail to open "+yamlFile)

	err = yaml.NewDecoder(f).Decode(config)
	tool.HandleErrorWithMsg(err, "fail to decode "+yamlFile)

	if projectDir == "" {
		projectDir = config.ProjectDir
	}
	if projectModule == "" {
		projectModule = config.ProjectModule
	}
	if interfaceFullName == "" {
		interfaceFullName = config.InterfaceFullName
	}
	if newMethod == "" {
		newMethod = config.NewMethod
	}
	if returnDefaultValues == "" {
		returnDefaultValues = config.ReturnDefaultValues
	}
}

func implement(cmd *cobra.Command, args []string) {
	readYaml()
	lo.ForEach[string](config.WritePaths, func(item string, index int) {
		pathInfo := strings.Split(item, ",")
		writePaths[pathInfo[0]] = pathInfo[1]
	})
	config.ExcludeDirs = append(config.ExcludeDirs, []string{".idea", ".git", "vendor", ".github"}...)
	tool.EnableRecord(config.EnableRecord)
	tool.EnableDebug(config.EnableDebug)

	if projectDir == "" || projectModule == "" || interfaceFullName == "" || newMethod == "" {
		tool.HandleErrorWithMsg(errors.New("invalid param"), "The params should be filled")
	}

	s := scanner.New(projectModule, projectDir)
	tool.Timer("Interfacer", func() {
		s.Start(projectDir, config.ExcludeDirs)
		s.Print()

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

		interfaceFileName := interfaceInfo.FilePaths()[0]
		writer.WriteFileForLine(interfaceFileName, []writer.Writer{writer.GetInterfaceWrite2(interfaceFileName, interfaceName, "\t"+newMethod)})
		lo.ForEach[*scanner.StructInfo](interfaceInfo.GetImplements(), func(item *scanner.StructInfo, index int) {
			writePath := item.FilePaths()[0]
			if p, ok := writePaths[item.Name()]; ok {
				writePath = p
			}
			receiverName, receiverType := item.MethodReceiver()
			writer.WriteFile(writePath, []writer.Writer{writer.GetFuncWriter(receiverName, receiverType, funcName, paramNames, paramTypes, returnTypes, returnDefaults)})
		})
	})
}

func main() {
	if err := interfacer.Execute(); err != nil {
		if interfacer.SilenceErrors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(-1)
	}
}
