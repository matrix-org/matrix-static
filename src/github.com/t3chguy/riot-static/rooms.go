package main

import (
	"github.com/matrix-org/gomatrix"
	"sync"
)

type Room struct {
	sync.Once
	RoomID         string
	Servers        []string
	InitialSync    RespInitialSync
	PublicRoomInfo gomatrix.PublicRoomsChunk
}

func (room *Room) fetch() {
	urlPath := cli.BuildURL("rooms", room.RoomID, "initialSync")
	var resp RespInitialSync
	_, err := cli.MakeRequest("GET", urlPath, nil, &resp)

	if err == nil {
		data.RWMutex.Lock()
		data.Rooms[room.RoomID].InitialSync = resp
		data.RWMutex.Unlock()
	}
}

func (room *Room) Fetch() {
	room.Once.Do(room.fetch)
}

func NewRoom(roomId string, publicRoomInfo gomatrix.PublicRoomsChunk) (room *Room) {

	room = &Room{
		RoomID:         roomId,
		PublicRoomInfo: publicRoomInfo,
	}

	return
}
