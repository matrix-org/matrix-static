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

type RespInitialSync struct {
	AccountData []gomatrix.Event `json:"account_data"`

	Messages   gomatrix.RespMessages `json:"messages"`
	Membership string                `json:"membership"`
	State      []gomatrix.Event      `json:"state"`
	RoomID     string                `json:"room_id"`
	Receipts   []gomatrix.Event      `json:"receipts"`
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

type RespGetRoomAlias struct {
	RoomID  string   `json:"room_id"`
	Servers []string `json:"servers"`
}

func (room *Room) fetchRoomAliasInfo(wg *sync.WaitGroup) {
	defer wg.Done()

	urlPath := cli.BuildURL("directory", "room", room.PublicRoomInfo.CanonicalAlias)
	var resp RespGetRoomAlias
	_, err := cli.MakeRequest("GET", urlPath, nil, &resp)

	if err == nil {
		data.Lock()
		data.Rooms[room.RoomID].Servers = resp.Servers
		data.Unlock()
	}
}

func (room *Room) fetch() {
	var wg sync.WaitGroup

	wg.Add(1)
	go room.fetchInitialSync(&wg)

	if room.PublicRoomInfo.CanonicalAlias != "" {
		wg.Add(1)
		go room.fetchRoomAliasInfo(&wg)
	}

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
