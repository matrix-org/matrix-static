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
	"github.com/matrix-org/matrix-static/mxclient"
	"github.com/matrix-org/matrix-static/utils"
)

type RoomServersResp struct {
	RoomInfo mxclient.RoomInfo
	Servers  mxclient.ServerUserCounts
	PageSize int
	Page     int
}

type RoomServersJob struct {
	RoomID   string
	Page     int
	PageSize int
}

func (job RoomServersJob) Work(w *Worker) {
	room := w.rooms[job.RoomID]
	servers := room.GetState().Servers()

	start, end := utils.CalcPaginationStartEnd(job.Page, job.PageSize, len(servers))

	w.Output <- RoomServersResp{
		room.RoomInfo(),
		servers[start:end],
		job.PageSize,
		job.Page,
	}
	room.Access()
}
