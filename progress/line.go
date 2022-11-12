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

package progress

import "fmt"

type Line struct {
}

func (l *Line) Print(format string, a ...any) {
	fmt.Printf("\r")
	fmt.Printf(format, a...)
}

type LineGroup struct {
	lines []*Line
}

func (lg *LineGroup) SetLineNum(num int) {
	for i := 0; i < num; i++ {
		lg.lines = append(lg.lines, &Line{})
	}
}

func (lg *LineGroup) Print(f func(i int, l *Line)) {
	lg.hideCursor()
	for i, line := range lg.lines {
		f(i, line)
		fmt.Printf("\n")
	}
	lg.lineMoveUp(len(lg.lines))
}

func (lg *LineGroup) End(f func(i int, l *Line)) {
	for i, line := range lg.lines {
		f(i, line)
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
	lg.showCursor()
}

// lineMoveUp ANSI escape code ANSI转义序列
func (lg *LineGroup) lineMoveUp(n int) {
	fmt.Printf("\033[%dA", n)
}

func (lg *LineGroup) lineMoveDown(n int) {
	fmt.Printf("\033[%dB", n)
}

func (lg *LineGroup) hideCursor() {
	fmt.Printf("\033[?25l")
}

func (lg *LineGroup) showCursor() {
	fmt.Printf("\033[?25h")
}
