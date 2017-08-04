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

package mxclient

import "github.com/matrix-org/gomatrix"

// Keeping here in case it becomes used again.
//func ConcatEventsSlices(slices ...[]gomatrix.Event) []gomatrix.Event {
//	var totalLen int
//	for _, s := range slices {
//		totalLen += len(s)
//	}
//	tmp := make([]gomatrix.Event, totalLen)
//	var i int
//	for _, s := range slices {
//		i += copy(tmp[i:], s)
//	}
//	return tmp
//}

// ReverseEventsCopy returns a copy of the input slice with all elements in reverse order.
func ReverseEventsCopy(events []gomatrix.Event) []gomatrix.Event {
	var newEvents []gomatrix.Event
	for i := len(events) - 1; i >= 0; i-- {
		newEvents = append(newEvents, events[i])
	}
	return newEvents
}

// ShouldHideEvent returns a bool the event should be ignored in the timeline view, mimicking riot-web
func ShouldHideEvent(ev gomatrix.Event) bool {
	// m.room.create ?

	// we want to hide all unknowns +:
	// m.room.redaction
	// m.room.aliases
	// m.room.canonical_alias

	if ev.Type == "m.room.join_rules" ||
		ev.Type == "m.room.member" ||
		ev.Type == "m.room.power_levels" ||
		ev.Type == "m.room.message" ||
		ev.Type == "m.room.name" ||
		ev.Type == "m.room.topic" ||
		ev.Type == "m.room.avatar" {
		return false
	}

	return true
}
