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

package mxclient

import (
	"encoding/json"
	"fmt"
	"github.com/matrix-org/gomatrix"
	"sort"
	"strings"
)

type PowerLevels struct {
	Ban           PowerLevel            `json:"ban"`
	Events        map[string]PowerLevel `json:"events"`
	EventsDefault PowerLevel            `json:"events_default"`
	Invite        PowerLevel            `json:"invite"`
	Kick          PowerLevel            `json:"kick"`
	Redact        PowerLevel            `json:"redact"`
	StateDefault  PowerLevel            `json:"state_default"`
	Users         map[string]PowerLevel `json:"users"`
	UsersDefault  PowerLevel            `json:"users_default"`
}

type RoomState struct {
	client *Client

	Creator        string
	Topic          string
	Name           string
	canonicalAlias string
	AvatarURL      MXCURL
	Aliases        []string

	PowerLevels PowerLevels
	memberList  []*MemberInfo
	MemberMap   map[string]*MemberInfo
}

func NewRoomState(client *Client) *RoomState {
	return &RoomState{
		client:    client,
		MemberMap: make(map[string]*MemberInfo),
	}
}

func (rs RoomState) NumMembers() int {
	return len(rs.memberList)
}

func (rs *RoomState) GetNumMemberEvents() int {
	return len(rs.MemberMap)
}

func (rs *RoomState) UpdateOnEvent(event *gomatrix.Event, usePrevContent bool) {
	if event.StateKey == nil {
		fmt.Println("Debug Event", event)
		return
	}

	stateKey := *event.StateKey

	switch event.Type {
	case "m.room.aliases": // We do not (yet) care about m.room.aliases
	case "m.room.canonical_alias":
		if alias, ok := event.Content["alias"].(string); ok {
			rs.canonicalAlias = alias
		}
	case "m.room.create":
		if creator, ok := event.Content["creator"].(string); ok {
			rs.Creator = creator
		}
	case "m.room.join_rules": // We do not (yet) care about m.room.join_rules
	case "m.room.member":
		var currentMemberState *MemberInfo
		if currentMemberState = rs.MemberMap[stateKey]; currentMemberState == nil {
			newMemberInfo := NewMemberInfo(stateKey)
			currentMemberState = newMemberInfo
			rs.MemberMap[stateKey] = newMemberInfo
		}

		if usePrevContent {
			if membership, ok := event.PrevContent["membership"].(string); ok {
				currentMemberState.Membership = membership
			}
			if avatarUrl, ok := event.PrevContent["avatar_url"].(string); ok {
				currentMemberState.AvatarURL = *NewMXCURL(avatarUrl, rs.client.HomeserverURL.String())
			}
			if displayName, ok := event.PrevContent["displayname"].(string); ok {
				currentMemberState.DisplayName = displayName
			}
		}

		if membership, ok := event.Content["membership"].(string); ok {
			currentMemberState.Membership = membership
		}
		if avatarUrl, ok := event.Content["avatar_url"].(string); ok {
			currentMemberState.AvatarURL = *NewMXCURL(avatarUrl, rs.client.HomeserverURL.String())
		}
		if displayName, ok := event.Content["displayname"].(string); ok {
			currentMemberState.DisplayName = displayName
		}

		rs.memberList = rs.CalculateMemberList()
	case "m.room.power_levels":
		// ez convert to powerLevels
		if data, err := json.Marshal(event.Content); err == nil {
			var powerLevels PowerLevels
			err = json.Unmarshal(data, &powerLevels)
			if err == nil {
				rs.PowerLevels = powerLevels
			}
		}

	case "m.room.name":
		if name, ok := event.Content["name"].(string); ok {
			rs.Name = name
		}
	case "m.room.topic":
		if topic, ok := event.Content["topic"].(string); ok {
			rs.Topic = topic
		}
	case "m.room.avatar":
		if url, ok := event.Content["url"].(string); ok {
			rs.AvatarURL = *NewMXCURL(url, rs.client.HomeserverURL.String())
		}
	}
}

func (rs *RoomState) CalculateMemberList() []*MemberInfo {
	memberList := make([]*MemberInfo, 0, len(rs.MemberMap))
	for _, member := range rs.MemberMap {
		if member.Membership == "join" {
			memberList = append(memberList, member)
		}
	}

	return memberList
}

func (rs RoomState) Members() []*MemberInfo {
	return rs.memberList
}

func (rs RoomState) Servers() StringIntPairList {
	serverMap := make(map[string]int)
	for _, member := range rs.CalculateMemberList() {
		if mxidSplit := strings.SplitN(member.MXID, ":", 2); len(mxidSplit) == 2 {
			serverMap[mxidSplit[1]]++
		}
	}

	serverList := make(StringIntPairList, 0, len(serverMap))
	for server, num := range serverMap {
		serverList = append(serverList, StringIntPair{server, num})
	}

	sort.Sort(sort.Reverse(serverList))
	return serverList
}

// Partial implementation of http://matrix.org/docs/spec/client_server/r0.2.0.html#calculating-the-display-name-for-a-room
// Does not handle based on members if there is no Name/Alias (yet)
// Falls back to first alias. TODO find edge case rooms for which this is needed.
func (rs RoomState) CalculateName() string {
	if rs.Name != "" {
		return rs.Name
	}
	if rs.canonicalAlias != "" {
		return rs.canonicalAlias
	}
	if len(rs.Aliases) > 0 {
		return rs.Aliases[0]
	}

	return "Empty Room"
}
