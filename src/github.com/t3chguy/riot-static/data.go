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

// Notes
// Reads should be done on clones
// Treat Rooms as if Immutable
// ONLY MUTATE DURING A FULL LOCK

package main

import (
	"sync"
)

type DataStore struct {
	sync.RWMutex
	roomList []*Room
	roomMap  map[string]*Room
}

//func (data *DataStore) GetRoomPointer(roomID string) *Room {
//	data.RLock()
//	defer data.RUnlock()
//	return data.roomMap[roomID]
//}

func (data *DataStore) GetRoom(roomID string) (room Room, exists bool) {
	data.RLock()
	roomPointer := data.roomMap[roomID]
	data.RUnlock()

	if roomPointer == nil {
		return Room{}, false
	}
	exists = true

	data.RLock()
	expired := roomPointer.Cached.CheckExpired()
	data.RUnlock()

	if expired {
		data.Lock()
		defer data.Unlock()
		// Get the *REAL* room to update self
		if !roomPointer.FetchAndSelfUpdate() {
			return Room{}, false
		}
	}

	return *roomPointer, true
}

func (data *DataStore) GetRoomList(start int, end int) []*Room {
	data.RLock()
	defer data.RUnlock()
	length := data.GetNumRooms()

	if end == 0 {
		return data.roomList[min(start, length):]
	}

	return data.roomList[min(start, length):min(end, length)]
}

func (data *DataStore) SetRoomList(roomList []*Room) {
	length := len(roomList)
	newRoomList := make([]*Room, length, length)
	newRoomMap := make(map[string]*Room)

	copy(newRoomList, roomList)
	for _, room := range roomList {
		newRoomMap[room.RoomID] = room
	}

	data.Lock()
	defer data.Unlock()

	data.roomList = newRoomList
	data.roomMap = newRoomMap
}

func (data *DataStore) GetNumRooms() int {
	data.RLock()
	defer data.RUnlock()
	return len(data.roomList)
}
