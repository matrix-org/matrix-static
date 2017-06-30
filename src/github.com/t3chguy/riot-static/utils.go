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

package main

import "strconv"

func calcPaginationPage(pageString string, size int) (page int, skip int, end int) {
	var err error
	if page, err = strconv.Atoi(pageString); err != nil {
		page = 1
	}

	skip = (page - 1) * size
	end = skip + size
	return
}

// min returns the minimum value of two ints
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
