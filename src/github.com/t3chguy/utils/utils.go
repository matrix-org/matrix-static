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

import (
	"github.com/matrix-org/gomatrix"
	"strconv"
)

func CalcPaginationPage(pageString string, size int) (page int, skip int, end int) {
	var err error
	if page, err = strconv.Atoi(pageString); err != nil {
		page = 1
	}

	skip = (page - 1) * size
	end = skip + size
	return
}

func FixRange(min, val, max int) int {
	if val > max {
		return max
	}
	if val < min {
		return min
	}
	return val
}

// min returns the minimal value of N ints
func Min(nums ...int) int {
	curLowest := nums[0]
	for _, i := range nums {
		if i < curLowest {
			curLowest = i
		}
	}
	return curLowest
}

// max returns the maximal value of N ints
func Max(nums ...int) int {
	curHighest := nums[0]
	for _, i := range nums {
		if i > curHighest {
			curHighest = i
		}
	}
	return curHighest
}

func Concat(slices ...[]gomatrix.Event) []gomatrix.Event {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]gomatrix.Event, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}

func ReverseEventsCopy(events []gomatrix.Event) []gomatrix.Event {
	newEvents := make([]gomatrix.Event, len(events))
	for i := 0; i < len(events)/2; i++ {
		j := len(events) - i - 1
		newEvents[i], newEvents[j] = events[j], events[i]
	}
	return newEvents
}

func ReverseEvents(events []gomatrix.Event) {
	for i := 0; i < len(events)/2; i++ {
		j := len(events) - i - 1
		events[i], events[j] = events[j], events[i]
	}
}
