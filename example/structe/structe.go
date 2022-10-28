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

package structe

import "github.com/SimFG/interfacer/example/structe/base"

type Structe struct {
}

func (s Structe) S1() {

}

type Common struct {
}

func (s *Common) S2(i string) string {
	return ""
}

type StructePoint struct {
	Common
	*Structe
	base.Gen
	*base.Basic
	att string
}

func (s *StructePoint) S2() {

}

func (s StructePoint) S3() {

}
