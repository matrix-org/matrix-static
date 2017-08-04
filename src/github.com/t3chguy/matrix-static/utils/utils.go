// Copyright 2017 Michael Telatynski <7t3chguy@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import "strconv"

// StrToIntDefault converts str to its equivalent int, falling back to def if str does not represent an int.
func StrToIntDefault(str string, def int) (page int) {
	var err error
	if page, err = strconv.Atoi(str); err != nil {
		page = def
	}
	return
}

// CalcPaginationStartEnd calculates the slice offsets needed to perform pagination for desired page, pageSize and length
// if page=0 it will return slice offsets 0:length-1 for a "get all entries" page.
func CalcPaginationStartEnd(page, pageSize, length int) (start, end int) {
	if page == 0 {
		return 0, length - 1
	}

	start = Min((page-1)*pageSize, length)
	end = Min(start+pageSize, length)
	return
}

// Bound returns min if val<min, max if val>max, val else.
func Bound(min, val, max int) int {
	if val > max {
		return max
	}
	if val < min {
		return min
	}
	return val
}

// Min returns the minimal value of ints a and b
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximal value of ints a and b
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
