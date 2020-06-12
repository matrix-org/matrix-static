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

import log "github.com/Sirupsen/logrus"

type RoomInitialSyncResp struct {
	Err error
}

type RoomInitialSyncJob struct {
	RoomID string
}

func (job RoomInitialSyncJob) Work(w *Worker) {
	resp := &RoomInitialSyncResp{}

	if _, exists := w.rooms[job.RoomID]; !exists {
		loggerWithFields := log.WithField("worker", w.ID).WithField("RoomID", job.RoomID)
		loggerWithFields.Info("Started Initial Syncing Room")
		if newRoom, err := w.client.NewRoom(job.RoomID); err == nil {
			loggerWithFields.Info("Finished Initial Syncing Room")
			w.rooms[job.RoomID] = newRoom
		} else {
			loggerWithFields.WithError(err).Error("Failed Initial Syncing Room")
			resp.Err = err
		}
	}

	w.Output <- resp
}
