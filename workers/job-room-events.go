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

package workers

import (
	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/matrix-static/mxclient"
)

type RoomEventsResp struct {
	Events      []gomatrix.Event
	RoomInfo    mxclient.RoomInfo
	MemberMap   map[string]mxclient.MemberInfo
	AtTopEnd    bool
	AtBottomEnd bool
	Err         error
}

type RoomEventsJob struct {
	RoomID   string
	Anchor   string
	Offset   int
	PageSize int
}

func (job RoomEventsJob) Work(w *Worker) {
	room := w.rooms[job.RoomID]
	events, atTopEnd, atBottomEnd, err := room.GetEventPage(job.Anchor, job.Offset, job.PageSize)

	membersMap := make(map[string]mxclient.MemberInfo)
	for mxid, member := range room.GetState().MemberMap {
		membersMap[mxid] = *member
	}

	w.Output <- RoomEventsResp{
		events,
		room.RoomInfo(),
		membersMap,
		atTopEnd,
		atBottomEnd,
		err,
	}
	room.Access()
}
