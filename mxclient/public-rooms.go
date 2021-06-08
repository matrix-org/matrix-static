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
	"github.com/matrix-org/matrix-static/utils"
	"strings"
	"sync"
)

type WorldReadableRooms struct {
	mxclient   *Client
	roomsMutex sync.RWMutex
	rooms      []gomatrix.PublicRoom
}

// processRoomDirectory replaces AvatarUrl from mxc to its https counterpart and filters on WorldReadable rooms.
func processRoomDirectory(homeserverBaseUrl string, roomList []gomatrix.PublicRoom) (filteredRooms []gomatrix.PublicRoom) {
	for _, room := range roomList {
		if !room.WorldReadable {
			continue
		}

		// Hack to get a "Primary Alias" to match Room Directory of riot-web
		if room.CanonicalAlias == "" && len(room.Aliases) > 0 {
			room.CanonicalAlias = room.Aliases[0]
		}

		room.AvatarURL = NewMXCURL(room.AvatarURL, homeserverBaseUrl).ToThumbURL(60, 60, "crop")

		// Append world readable room to the filtered list.
		filteredRooms = append(filteredRooms, room)
	}
	return
}

// NewWorldReadableRooms instantiates a WorldReadableRooms Collection
func (m *Client) NewWorldReadableRooms() *WorldReadableRooms {
	worldReadableRooms := &WorldReadableRooms{mxclient: m}
	if err := worldReadableRooms.Update(); err != nil {
		panic(err)
	}

	return worldReadableRooms
}

// Update updates the state of the WorldReadableRooms Collection by doing an API Call.
func (r *WorldReadableRooms) Update() error {
	resp, err := r.mxclient.PublicRooms(0, "", "")
	if err != nil {
		return err
	}
	filteredRooms := processRoomDirectory(r.mxclient.MediaBaseURL, resp.Chunk)

	r.roomsMutex.Lock()
	defer r.roomsMutex.Unlock()

	r.rooms = filteredRooms
	return nil
}

// GetFilteredPage returns a filtered & paginated slice of the WorldReadableRooms Collection
func (r *WorldReadableRooms) GetFilteredPage(page, pageSize int, query string) []gomatrix.PublicRoom {
	r.roomsMutex.RLock()
	defer r.roomsMutex.RUnlock()

	lowerQuery := strings.ToLower(query)

	filteredRooms := make([]gomatrix.PublicRoom, 0, pageSize)
	for _, room := range r.rooms {
		if len(filteredRooms) > pageSize {
			break
		}

		if (lowerQuery[0] == '#' && strings.Contains(strings.ToLower(room.CanonicalAlias), lowerQuery)) ||
			strings.Contains(strings.ToLower(room.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(room.Topic), lowerQuery) {
			filteredRooms = append(filteredRooms, room)
		}
	}

	start, end := utils.CalcPaginationStartEnd(page, pageSize, len(filteredRooms))
	return filteredRooms[start:end]
}

// GetPage returns a paginated slice of the WorldReadableRooms Collection
func (r *WorldReadableRooms) GetPage(page, pageSize int) []gomatrix.PublicRoom {
	r.roomsMutex.RLock()
	defer r.roomsMutex.RUnlock()
	start, end := utils.CalcPaginationStartEnd(page, pageSize, len(r.rooms))
	return r.rooms[start:end]
}
