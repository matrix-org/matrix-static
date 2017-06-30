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
	"github.com/matrix-org/gomatrix"
)

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

func (memberInfo MemberInfo) GetName() string {
	if memberInfo.DisplayName != "" {
		return memberInfo.DisplayName
	} else {
		return memberInfo.MXID
	}
}

type Room struct {
	Cached

	RoomID  string
	Servers []string

	InitialSync *RespInitialSync

	memberList []*MemberInfo
	memberMap  map[string]*MemberInfo

	powerLevels *PowerLevelsEvent
	presenceMap map[string]*Presence

	CanonicalAlias   string
	Name             string
	WorldReadable    bool
	Topic            string
	NumJoinedMembers int
	AvatarUrl        MxcUrl
	GuestCanJoin     bool
	Aliases          []string
}

func (room Room) GetName() string {
	if room.Name != "" {
		return room.Name
	} else if room.CanonicalAlias != "" {
		return room.CanonicalAlias
	} else {
		return room.RoomID
	}
}

func (room Room) GetMemberList(start int, end int) []*MemberInfo {
	length := room.GetNumMembers()

	if end == 0 {
		return room.memberList[min(start, length):]
	}

	return room.memberList[min(start, length):min(end, length)]
}

func (room Room) GetMemberIgnore(mxid string) (memberInfo MemberInfo) {
	memberInfo, _ = room.GetMember(mxid)
	return
}

func (room Room) GetMember(mxid string) (memberInfo MemberInfo, exists bool) {
	if memberInfoPointer := room.memberMap[mxid]; memberInfoPointer != nil {
		return *memberInfoPointer, true
	}
	return MemberInfo{}, false
}

func (room Room) GetNumMembers() int {
	return len(room.memberList)
}

type Presence struct {
	CurrentlyActive bool   `json:"currently_active"`
	LastActiveAgo   int    `json:"last_active_ago"`
	Presence        string `json:"presence"`
	UserID          string `json:"user_id"`
}

type PresenceEvent struct {
	Content Presence `json:"content"`
}

type RespInitialSync struct {
	AccountData []gomatrix.Event `json:"account_data"`

	Messages   *gomatrix.RespMessages `json:"messages"`
	Membership string                 `json:"membership"`
	State      []*gomatrix.Event      `json:"state"`
	RoomID     string                 `json:"room_id"`
	Receipts   []*gomatrix.Event      `json:"receipts"`
	Presence   []*PresenceEvent       `json:"presence"`
}

type PowerLevelsEvent struct {
	Ban           int                   `json:"ban"`
	Events        map[string]int        `json:"events"`
	EventsDefault int                   `json:"events_default"`
	Invite        int                   `json:"invite"`
	Kick          int                   `json:"kick"`
	Redact        int                   `json:"redact"`
	StateDefault  int                   `json:"state_default"`
	Users         map[string]PowerLevel `json:"users"`
	UsersDefault  int                   `json:"users_default"`
}

func (room *Room) GetPresence(mxid string) Presence {
	return *room.presenceMap[mxid]
}

func (room *Room) setPresence(presenceEvents []*PresenceEvent) {
	presenceMap := make(map[string]*Presence)
	for _, presenceEvent := range presenceEvents {
		presenceMap[presenceEvent.Content.UserID] = &presenceEvent.Content
	}
	room.presenceMap = presenceMap
}

func (room *Room) FetchAndSelfUpdate() bool {
	initialSync, memberInfoMap, powerLevels, err := peekRoom(room.RoomID)
	if err != nil {
		return false
	}

	room.InitialSync = initialSync
	room.memberMap = memberInfoMap
	//room.memberList
	room.powerLevels = powerLevels

	return true
}

func NewRoom(publicRoomInfo gomatrix.PublicRoomsChunk) (room *Room) {

	room = &Room{
		RoomID:           publicRoomInfo.RoomId,
		CanonicalAlias:   publicRoomInfo.CanonicalAlias,
		Name:             publicRoomInfo.Name,
		Topic:            publicRoomInfo.Topic,
		NumJoinedMembers: publicRoomInfo.NumJoinedMembers,
		AvatarUrl:        MxcUrl(publicRoomInfo.AvatarUrl),
		Aliases:          publicRoomInfo.Aliases,
	}

	return
}
