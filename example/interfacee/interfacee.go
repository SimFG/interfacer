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

package interfacee

import (
	"net/http"
)

type S struct {
}

type OneMethod interface {
	One()
}

type TwoMethod interface {
	Two1()
	Two2()
}

type ParamMethod interface {
	Param1(a int)
	Param2(a int, b float32)
	Param3(a int) float32
	Param4(a int) (float32, float64)
	Param5(a int) (b string, c interface{})
}

type CombineInterface interface {
	OneMethod
	Combine1()
}

type MoreCombineInterface interface {
	OneMethod
	ParamMethod
	http.File
	MoreCombine(a int, b int64) (float32, float64)
}
