// Copyright 2017 Michael Telatynski <7t3cghuy@gmail.com>
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

package main

import (
	"fmt"
	"github.com/matrix-org/gomatrix"
	"sync"
)

type Room struct {
	sync.Once
	RoomID  string
	Servers []string

	InitialSync RespInitialSync

	CanonicalAlias   string
	Name             string
	WorldReadable    bool
	Topic            string
	NumJoinedMembers int
	AvatarUrl        string
	GuestCanJoin     bool
	Aliases          []string
}

// Event represents a single Matrix event.
type StateEvent struct {
	StateKey   *string                `json:"state_key,omitempty"` // The state key for the event. Only present on State Events.
	Sender     string                 `json:"sender"`              // The user ID of the sender of the event
	Type       string                 `json:"type"`                // The event type
	Timestamp  int                    `json:"origin_server_ts"`    // The unix timestamp when this message was sent by the origin server
	ID         string                 `json:"event_id"`            // The unique ID of this event
	RoomID     string                 `json:"room_id"`             // The room the event was sent to. May be nil (e.g. for presence)
	Content    map[string]interface{} `json:"content"`             // The JSON content of the event.
	Membership string                 `json:"membership"`
}

type RespInitialSync struct {
	AccountData []gomatrix.Event `json:"account_data"`

	Messages   gomatrix.RespMessages `json:"messages"`
	Membership string                `json:"membership"`
	State      []StateEvent          `json:"state"`
	RoomID     string                `json:"room_id"`
	Receipts   []gomatrix.Event      `json:"receipts"`
}

func (room *Room) fetchInitialSync(wg *sync.WaitGroup) {
	defer wg.Done()

	urlPath := cli.BuildURL("rooms", room.RoomID, "initialSync")
	fmt.Println(urlPath)
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

	urlPath := cli.BuildURL("directory", "room", room.CanonicalAlias)
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

	if room.CanonicalAlias != "" {
		wg.Add(1)
		go room.fetchRoomAliasInfo(&wg)
	}

	wg.Wait()
}

func (room *Room) Fetch() {
	room.Once.Do(room.fetch)
}

func NewRoom(publicRoomInfo gomatrix.PublicRoomsChunk) (room *Room) {

	room = &Room{
		RoomID:           publicRoomInfo.RoomId,
		CanonicalAlias:   publicRoomInfo.CanonicalAlias,
		Name:             publicRoomInfo.Name,
		WorldReadable:    publicRoomInfo.WorldReadable,
		Topic:            publicRoomInfo.Topic,
		NumJoinedMembers: publicRoomInfo.NumJoinedMembers,
		AvatarUrl:        publicRoomInfo.AvatarUrl,
		GuestCanJoin:     publicRoomInfo.GuestCanJoin,
		Aliases:          publicRoomInfo.Aliases,
	}

	return
}
