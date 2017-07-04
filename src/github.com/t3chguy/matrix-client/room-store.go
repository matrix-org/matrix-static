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

package matrixClient

import (
	"github.com/t3chguy/utils"
	"sync"
)

type RoomStore struct {
	sync.RWMutex // Protects just this level, Rooms are self-protecting.
	// The next two fields form an OrderedMap, roomList is the source of truth, map mirrors it with an r.ID => r mapping
	roomList []*Room
	roomMap  map[string]*Room
}

func (data *RoomStore) GetRoom(roomID string) *Room {
	data.RLock()
	room := data.roomMap[roomID]
	data.RUnlock()

	if room != nil {
		room.LazyInitialSync()
	}
	return room
}

func (data *RoomStore) GetRoomList(start int, end int) []*Room {
	data.RLock()
	defer data.RUnlock()
	length := len(data.roomList)

	if end == -1 {
		return data.roomList[utils.Min(start, length):]
	}

	return data.roomList[utils.Min(start, length):utils.Min(end, length)]
}

func (data *RoomStore) SetRoomList(roomList []*Room) {
	length := len(roomList)
	newRoomList := make([]*Room, length, length)
	newRoomMap := make(map[string]*Room)

	copy(newRoomList, roomList)
	for _, room := range roomList {
		newRoomMap[room.ID] = room
	}

	data.Lock()
	defer data.Unlock()

	data.roomList = newRoomList
	data.roomMap = newRoomMap
}

func (data *RoomStore) GetNumRooms() int {
	data.RLock()
	defer data.RUnlock()
	return len(data.roomList)
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		roomMap: make(map[string]*Room),
	}
}
