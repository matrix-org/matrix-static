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

package matrix_client

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

func (data *RoomStore) LoadRoom(roomID string) (room *Room, ok bool) {
	if room = data.GetRoom(roomID); room != nil {
		ok = room.LazyInitialSync()
	}
	return
}

func (data *RoomStore) GetRoom(roomID string) *Room {
	data.RLock()
	defer data.RUnlock()
	return data.roomMap[roomID]
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
