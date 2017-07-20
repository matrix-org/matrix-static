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

import (
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/riot-static/utils"
	"sync"
)

type WorldReadableRooms struct {
	mxclient   *Client
	roomsMutex sync.RWMutex
	rooms      []gomatrix.PublicRoomsChunk
}

// processRoomDirectory replaces AvatarUrl from mxc to its https counterpart and filters on WorldReadable rooms.
func processRoomDirectory(homeserverBaseUrl string, roomList []gomatrix.PublicRoomsChunk) (filteredRooms []gomatrix.PublicRoomsChunk) {
	for _, room := range roomList {
		if !room.WorldReadable {
			continue
		}

		// Hack to get a "Primary Alias" to match Room Directory of riot-web
		if room.CanonicalAlias == "" && len(room.Aliases) > 0 {
			room.CanonicalAlias = room.Aliases[0]
		}

		room.AvatarUrl = NewMXCURL(room.AvatarUrl, homeserverBaseUrl).ToThumbURL(60, 60, "crop")

		// Append world readable room to the filtered list.
		filteredRooms = append(filteredRooms, room)
	}
	return
}

func (m *Client) NewWorldReadableRooms() *WorldReadableRooms {
	worldReadableRooms := &WorldReadableRooms{mxclient: m}
	if err := worldReadableRooms.Update(); err != nil {
		panic(err)
	}

	return worldReadableRooms
}

func (r *WorldReadableRooms) Update() error {
	resp, err := r.mxclient.PublicRooms(0, "", "")
	if err != nil {
		return err
	}
	filteredRooms := processRoomDirectory(r.mxclient.HomeserverURL.String(), resp.Chunk)

	r.roomsMutex.Lock()
	defer r.roomsMutex.Unlock()

	r.rooms = filteredRooms
	return nil
}

// For future when we support filtering the public room directory (LOCALLY)
//func (r *WorldReadableRooms) GetFilteredPage(page, pageSize int, query string) []gomatrix.PublicRoomsChunk {
//	r.roomsMutex.RLock()
//	defer r.roomsMutex.RUnlock()
//	return nil
//}

func (r *WorldReadableRooms) GetPage(page, pageSize int) []gomatrix.PublicRoomsChunk {
	r.roomsMutex.RLock()
	defer r.roomsMutex.RUnlock()
	start, end := utils.CalcPaginationStartEnd(page, pageSize, len(r.rooms))
	return r.rooms[start:end]
}

func (r *WorldReadableRooms) GetAll() []gomatrix.PublicRoomsChunk {
	r.roomsMutex.RLock()
	defer r.roomsMutex.RUnlock()
	return r.rooms
}
