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
	"encoding/json"
	"github.com/matrix-org/gomatrix"
	"sync"
)

type PowerLevels struct {
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

type RoomState struct {
	sync.RWMutex

	creator        string
	topic          string
	Name           string
	CanonicalAlias string
	AvatarURL      MXCURL

	Aliases          []string
	NumJoinedMembers int

	powerLevels PowerLevels
	memberList  []*MemberInfo
	memberMap   map[string]*MemberInfo // include leave?
}

func (rs *RoomState) UpdateOnEvent(event *gomatrix.Event) {
	rs.Lock()
	defer rs.Unlock()

	stateKey := *event.StateKey

	switch event.Type {
	case "m.room.aliases": // We do not (yet) care about m.room.aliases
	case "m.room.canonical_alias":
		if alias, ok := event.Content["alias"].(string); ok {
			rs.CanonicalAlias = alias
		}
	case "m.room.create":
		if creator, ok := event.Content["creator"].(string); ok {
			rs.creator = creator
		}
	case "m.room.join_rules": // We do not (yet) care about m.room.join_rules
	case "m.room.member":
		var currentMemberState *MemberInfo
		if currentMemberState = rs.memberMap[stateKey]; currentMemberState == nil {
			newMemberInfo := NewMemberInfo(stateKey)
			currentMemberState = newMemberInfo
			rs.memberMap[stateKey] = newMemberInfo
		}

		if membership, ok := event.Content["membership"].(string); ok {
			currentMemberState.Membership = membership
		}
		if avatarUrl, ok := event.Content["avatar_url"].(string); ok {
			currentMemberState.AvatarURL = MXCURL(avatarUrl)
		}
		if displayName, ok := event.Content["displayname"].(string); ok {
			currentMemberState.DisplayName = displayName
		}
	case "m.room.power_levels":
		switch event.Type {
		case "m.room.power_levels":
			// ez convert to powerLevels
			if data, err := json.Marshal(event.Content); err == nil {
				var powerLevels PowerLevels
				err = json.Unmarshal(data, &powerLevels)
				if err == nil {
					rs.powerLevels = powerLevels
				}
			}
		}
	}
}

func (rs *RoomState) CalculateMemberList() []*MemberInfo {
	rs.RLock()
	memberList := make([]*MemberInfo, 0, len(rs.memberMap))
	for _, member := range rs.memberMap {
		if member.Membership == "join" {
			memberList = append(memberList, member)
		}
	}
	rs.RUnlock()

	return memberList
}

func (rs *RoomState) Topic() string {
	rs.RLock()
	defer rs.RUnlock()
	return rs.topic
}
