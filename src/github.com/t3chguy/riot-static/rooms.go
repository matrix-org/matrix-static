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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/matrix-org/gomatrix"
	"github.com/matryer/resync"
	"net/url"
	"path"
	"regexp"
	"sync"
)

// mxcRegex allows splitting an mxc into a serverName and mediaId
// FindStringSubmatch of which results in [_, serverName, mediaId] if valid
// and [] if invalid mxc is provided.
// Examples:
// "mxc://foo/bar" => ["mxc://foo/bar", "foo", "bar"]
// "mxc://bar/foo#auto" => ["mxc://bar/foo#auto", "bar", "foo"]
// "invalidMxc://whatever" => [] (Invalid MXC Caught)
var mxcRegex = regexp.MustCompile(`mxc://(.+?)/(.+?)(?:#.+)?$`)

type MxcUrl string

func (mxcUrl MxcUrl) ToThumbUrl() string {
	mxc := string(mxcUrl)
	matches := mxcRegex.FindStringSubmatch(mxc)

	if len(matches) != 3 {
		return ""
	}

	serverName := matches[1]
	mediaId := matches[2]

	hsURL, _ := url.Parse(cli.HomeserverURL.String())
	parts := []string{hsURL.Path}
	parts = append(parts, "_matrix", "media", "r0", "thumbnail", serverName, mediaId)
	hsURL.Path = path.Join(parts...)

	q := hsURL.Query()
	q.Set("width", "50")
	q.Set("height", "50")
	q.Set("method", "crop")

	hsURL.RawQuery = q.Encode()

	return hsURL.String()
}
func (mxcUrl MxcUrl) ToUrl() string {
	mxc := string(mxcUrl)
	matches := mxcRegex.FindStringSubmatch(mxc)

	if len(matches) != 3 {
		return ""
	}

	serverName := matches[1]
	mediaId := matches[2]
	hsURL, _ := url.Parse(cli.HomeserverURL.String())
	parts := []string{hsURL.Path}
	parts = append(parts, "_matrix", "media", "r0", "download", serverName, mediaId)
	hsURL.Path = path.Join(parts...)

	q := hsURL.Query()
	q.Set("width", "50")
	q.Set("height", "50")
	q.Set("method", "crop")

	hsURL.RawQuery = q.Encode()

	return hsURL.String()
}

type PowerLevel int

func (powerLevel PowerLevel) String() string {
	switch int(powerLevel) {
	case 100:
		return "Admin"
	case 50:
		return "Moderator"
	case 0:
		return "User"
	case -1:
		return "Muted"
	default:
		return "Custom"
	}
}

func (powerLevel PowerLevel) ToInt() int {
	return int(powerLevel)
}

type MemberInfo struct {
	MXID        string
	Membership  string
	DisplayName string
	AvatarURL   MxcUrl
	PowerLevel
}

func (memberInfo *MemberInfo) GetName() string {
	if memberInfo.DisplayName != "" {
		return memberInfo.DisplayName
	} else {
		return memberInfo.MXID
	}
}

type Room struct {
	resync.Once
	RoomID  string
	Servers []string

	InitialSync *RespInitialSync
	MemberInfo  map[string]*MemberInfo

	PowerLevels *PowerLevelsEvent

	CanonicalAlias   string
	Name             string
	WorldReadable    bool
	Topic            string
	NumJoinedMembers int
	AvatarUrl        MxcUrl
	GuestCanJoin     bool
	Aliases          []string
}

func (room *Room) GetName() string {
	if room.Name != "" {
		return room.Name
	} else if room.CanonicalAlias != "" {
		return room.CanonicalAlias
	} else {
		return room.RoomID
	}
}

// Event represents a single Matrix event.
type StateEvent struct {
	StateKey    string                 `json:"state_key,omitempty"` // The state key for the event. Only present on State Events.
	Sender      string                 `json:"sender"`              // The user ID of the sender of the event
	Type        string                 `json:"type"`                // The event type
	Timestamp   int                    `json:"origin_server_ts"`    // The unix timestamp when this message was sent by the origin server
	ID          string                 `json:"event_id"`            // The unique ID of this event
	RoomID      string                 `json:"room_id"`             // The room the event was sent to. May be nil (e.g. for presence)
	Content     map[string]interface{} `json:"content"`             // The JSON content of the event.
	Membership  string                 `json:"membership"`
	PrevContent map[string]interface{} `json:"prev_content,omitempty"`
}

type RespInitialSync struct {
	AccountData []gomatrix.Event `json:"account_data"`

	Messages   *gomatrix.RespMessages `json:"messages"`
	Membership string                 `json:"membership"`
	State      []*StateEvent          `json:"state"`
	RoomID     string                 `json:"room_id"`
	Receipts   []*gomatrix.Event      `json:"receipts"`
}

type PowerLevelsEvent struct {
	Ban           int            `json:"ban"`
	Events        map[string]int `json:"events"`
	EventsDefault int            `json:"events_default"`
	Invite        int            `json:"invite"`
	Kick          int            `json:"kick"`
	Redact        int            `json:"redact"`
	StateDefault  int            `json:"state_default"`
	Users         map[string]int `json:"users"`
	UsersDefault  int            `json:"users_default"`
}

func (room *Room) fetchInitialSync(wg *sync.WaitGroup) {
	defer wg.Done()

	urlPath := cli.BuildURLWithQuery([]string{"rooms", room.RoomID, "initialSync"}, map[string]string{
		"limit": "20",
	})
	//urlPath := cli.BuildURL("rooms", room.RoomID, "initialSync")
	fmt.Println(urlPath)
	var resp RespInitialSync
	_, err := cli.MakeRequest("GET", urlPath, nil, &resp)

	if err == nil {
		memberInfo := make(map[string]*MemberInfo)
		var powerLevelsEvent *PowerLevelsEvent

		for _, stateEvent := range resp.State {
			switch stateEvent.Type {
			case "m.room.member":
				if memberInfo[stateEvent.StateKey] == nil {
					memberInfo[stateEvent.StateKey] = &MemberInfo{MXID: stateEvent.StateKey}
				}

				if avatarUrl := stateEvent.Content["avatar_url"]; avatarUrl != nil {
					memberInfo[stateEvent.StateKey].AvatarURL = MxcUrl(avatarUrl.(string))
				}
				if membership := stateEvent.Content["membership"]; memberInfo != nil {
					memberInfo[stateEvent.StateKey].Membership = membership.(string)
				}
				if displayname := stateEvent.Content["displayname"]; displayname != nil {
					memberInfo[stateEvent.StateKey].DisplayName = displayname.(string)
				}

				//fmt.Println(stateEvent.PrevContent)

			case "m.room.power_levels":
				data, _ := json.Marshal(stateEvent.Content)
				err := json.Unmarshal(data, &powerLevelsEvent)

				fmt.Println(err)

				if stateEvent.Content["users"] == nil {
					break
				}

				for mxid, powerLevel := range stateEvent.Content["users"].(map[string]interface{}) {
					if memberInfo[mxid] == nil {
						memberInfo[mxid] = &MemberInfo{MXID: mxid}
					}

					if powerLevel != nil {
						memberInfo[mxid].PowerLevel = PowerLevel(powerLevel.(float64))
					}
				}
			}
		}

		data.Lock()
		data.Rooms[room.RoomID].PowerLevels = powerLevelsEvent
		data.Rooms[room.RoomID].MemberInfo = memberInfo
		data.Rooms[room.RoomID].InitialSync = &resp
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
		AvatarUrl:        MxcUrl(publicRoomInfo.AvatarUrl),
		GuestCanJoin:     publicRoomInfo.GuestCanJoin,
		Aliases:          publicRoomInfo.Aliases,
	}

	return
}
