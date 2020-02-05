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
	log "github.com/Sirupsen/logrus"
	"github.com/matrix-org/matrix-static/mxclient"
	"sort"
	"sync"
	"time"
)

// This Job has no Resp.

type RoomForwardPaginateJob struct {
	Wg      *sync.WaitGroup
	TTL     time.Duration
	KeepMin int
}

func (job RoomForwardPaginateJob) Work(w *Worker) {
	numRoomsBefore := len(w.rooms)

	// discard old rooms, ignoring the first N
	if numRoomsBefore > job.KeepMin {
		rooms := make([]*mxclient.Room, 0, numRoomsBefore)

		// create a slice of rooms ordered by LastAccess descending
		for _, room := range w.rooms {
			rooms = append(rooms, room)
		}
		sort.Slice(rooms, func(i, j int) bool {
			return rooms[i].LastAccess.After(rooms[j].LastAccess)
		})

		for _, room := range rooms[job.KeepMin:] {
			if room.LastAccess.Before(time.Now().Add(-job.TTL)) {
				delete(w.rooms, room.ID)
			}
		}
		numRoomsAfter := len(w.rooms)
		log.WithField("worker", w.ID).WithField("numRooms", numRoomsAfter).Infof("Removed %d rooms", numRoomsBefore-numRoomsAfter)
	}

	for _, room := range w.rooms {
		room.ForwardPaginateRoom()
	}
	job.Wg.Done()
}
