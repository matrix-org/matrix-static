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

import "github.com/t3chguy/matrix-static/mxclient"

type RoomMemberNotFoundError struct {
	roomID string
	mxid   string
}

func (err *RoomMemberNotFoundError) Error() string {
	return "Member " + err.mxid + " not found in " + err.roomID + "."
}

type RoomMemberInfoResp struct {
	RoomInfo   mxclient.RoomInfo
	MemberInfo mxclient.MemberInfo
	Err        error
}

type RoomMemberInfoJob struct {
	roomID string
	mxid   string
}

func (job RoomMemberInfoJob) Work(w *Worker) {
	room := w.rooms[job.roomID]

	var err error
	var memberInfo mxclient.MemberInfo

	if member := room.GetState().MemberMap[job.mxid]; member == nil {
		err = &RoomMemberNotFoundError{
			job.roomID,
			job.mxid,
		}
	} else {
		memberInfo = *member
	}

	w.Output <- RoomMemberInfoResp{
		room.RoomInfo(),
		memberInfo,
		err,
	}
}
