package main

import (
	"github.com/matrix-org/gomatrix"
	"sync"
)

type Room struct {
	sync.Once
	Servers        []string
	InitialSync    RespInitialSync
	PublicRoomInfo gomatrix.PublicRoomsChunk
}

func (room *Room) fetch() {
	data.RWMutex.Lock()

	data.RWMutex.Unlock()
}

func (room *Room) Fetch() {
	room.Once.Do(room.fetch)
}

func NewRoom(publicRoomInfo gomatrix.PublicRoomsChunk) (room *Room) {

	room = &Room{}
	room.PublicRoomInfo = publicRoomInfo

	return
}
