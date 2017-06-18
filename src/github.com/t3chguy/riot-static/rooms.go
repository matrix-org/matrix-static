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

func (room *Room) fetchInitialSync(wg *sync.WaitGroup) {
	defer wg.Done()

	urlPath := cli.BuildURL("rooms", room.RoomID, "initialSync")
	var resp RespInitialSync
	_, err := cli.MakeRequest("GET", urlPath, nil, &resp)

	if err == nil {
		data.Lock()
		data.Rooms[room.RoomID].InitialSync = resp
		data.Unlock()
	}
}

func (room *Room) fetch() {
	var wg sync.WaitGroup

	wg.Add(1)

	go room.fetchInitialSync(&wg)

	wg.Wait()
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
