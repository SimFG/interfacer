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
	"go.uber.org/zap"
	"go/ast"
)

// GetValueFromType Get value from the `*ast.Ident`/`*ast.SelectorExpr`/`*ast.StarExpr`
// TODO handle the map / slice / func
func GetValueFromType(e ast.Expr) string {
	Info("GetValueFromType", zap.String("type", TypeString(e)))
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
		return "*" + GetValueFromType(e.(*ast.StarExpr).X)
	default:
		Info("WARN GetValueFromType")
		return ""
	}
}
