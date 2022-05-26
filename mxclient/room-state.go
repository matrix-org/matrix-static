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
	roomVersion    string
	AvatarURL      MXCURL
	aliasMap       map[string][]string
	Aliases        RoomAliases

	PowerLevels PowerLevels
	serverList  []ServerUserCount
	memberList  []*MemberInfo
	MemberMap   map[string]*MemberInfo
}

// NewRoomState creates a RoomState with defaults applied.
func NewRoomState(client *Client) *RoomState {
	return &RoomState{
		client:    client,
		MemberMap: make(map[string]*MemberInfo),
		aliasMap:  make(map[string][]string),
	}
}

// NumMembers returns the number of members with membership=join
func (rs RoomState) NumMembers() int {
	return len(rs.memberList)
}

// GetNumMemberEvents returns the total number of member events found in room state (i.e total number of unique users)
func (rs *RoomState) GetNumMemberEvents() int {
	return len(rs.MemberMap)
}

// UpdateOnEvent iterates the Room State based on the event observed.
func (rs *RoomState) UpdateOnEvent(event *gomatrix.Event, usePrevContent bool) {
	if event.StateKey == nil {
		return
	}

	stateKey := *event.StateKey

	switch event.Type {
	case "m.room.aliases":
		if aliases, ok := event.Content["aliases"].([]interface{}); ok {
			numAliases := len(aliases)
			if numAliases == 0 {
				break
			}

			processedAliases := make([]string, 0, numAliases)
			for _, alias := range aliases {
				if processedAlias, ok := alias.(string); ok {
					processedAliases = append(processedAliases, processedAlias)
				}
			}
			rs.aliasMap[stateKey] = processedAliases
		}
	case "m.room.canonical_alias":
		if alias, ok := event.Content["alias"].(string); ok {
			rs.canonicalAlias = alias
		}
	case "m.room.create":
		if creator, ok := event.Content["creator"].(string); ok {
			rs.Creator = creator
		}
		if roomVer, ok := event.Content["room_version"].(string); ok {
			rs.roomVersion = roomVer
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
				currentMemberState.AvatarURL = *NewMXCURL(avatarUrl, rs.client.MediaBaseURL)
			}
			if displayName, ok := event.PrevContent["displayname"].(string); ok {
				currentMemberState.DisplayName = displayName
			}
		}

		if membership, ok := event.Content["membership"].(string); ok {
			currentMemberState.Membership = membership
		}
		if avatarUrl, ok := event.Content["avatar_url"].(string); ok {
			currentMemberState.AvatarURL = *NewMXCURL(avatarUrl, rs.client.MediaBaseURL)
		}
		if displayName, ok := event.Content["displayname"].(string); ok {
			currentMemberState.DisplayName = displayName
		}
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
			rs.AvatarURL = *NewMXCURL(url, rs.client.MediaBaseURL)
		}
	}
}

// RecalculateMemberListAndServers does member list calculation, sorting and server calculations.
// ideally called at the end of concatenating so that its done as infrequently as possible
// whilst still never causing outdated information.
func (rs *RoomState) RecalculateMemberListAndServers() {
	for mxid, powerlevel := range rs.PowerLevels.Users {
		if _, ok := rs.MemberMap[mxid]; ok {
			rs.MemberMap[mxid].PowerLevel = PowerLevel(powerlevel)
		}
	}

	// Filter list of members with Membership=join
	memberList := make(MemberList, 0)
	for _, member := range rs.MemberMap {
		if member.Membership == "join" {
			memberList = append(memberList, member)
		}
	}

	// Create a map of servers by splitting memberList MXID's and incrementing count on server key
	serverMap := make(map[string]int)
	for _, member := range memberList {
		if mxidSplit := strings.SplitN(member.MXID, ":", 2); len(mxidSplit) == 2 {
			serverMap[mxidSplit[1]]++
		}
	}

	// Fit the server map into an sortable slice
	serverList := make(ServerUserCounts, 0, len(serverMap))
	for server, num := range serverMap {
		serverList = append(serverList, ServerUserCount{server, num})
	}

	// Sort and store both server and member lists
	sort.Sort(serverList)
	rs.serverList = serverList
	sort.Sort(memberList)
	rs.memberList = memberList

	// Recalculate Room Aliases
	roomAliases := make(RoomAliases, 0)
	for server, aliases := range rs.aliasMap {
		sort.Strings(aliases)
		roomAliases = append(roomAliases, RoomServerAliases{server, aliases})
	}

	// Sort and store the aliases list
	sort.Sort(roomAliases)
	rs.Aliases = roomAliases

}

// Members is an accessor for RoomState.memberList
func (rs RoomState) Members() []*MemberInfo {
	return rs.memberList
}

type ServerUserCount struct {
	ServerName string
	NumUsers   int
}

// implements sort.Interface
type ServerUserCounts []ServerUserCount

func (p ServerUserCounts) Len() int { return len(p) }
func (p ServerUserCounts) Less(i, j int) bool {
	a, b := p[i], p[j]
	if a.NumUsers == b.NumUsers {
		// Secondary Sort is Low->High Lexicographically on ServerName
		return a.ServerName < b.ServerName
	}
	// Primary Sort is High->Low on NumUsers
	return a.NumUsers > b.NumUsers
}
func (p ServerUserCounts) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type RoomServerAliases struct {
	ServerName string
	Aliases    []string
}

// implements sort.Interface
type RoomAliases []RoomServerAliases

func (p RoomAliases) Len() int { return len(p) }
func (p RoomAliases) Less(i, j int) bool {
	a, b := p[i], p[j]
	numA, numB := len(a.Aliases), len(b.Aliases)
	if numA == numB {
		// Secondary Sort is Low->High Lexicographically on ServerName
		return a.ServerName < b.ServerName
	}
	// Primary Sort is High->Low on len(Aliases)
	return numA > numB
}
func (p RoomAliases) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Servers iterates over the Member List (membership=join), splits each MXID and counts the number of each homeserver url.
func (rs RoomState) Servers() []ServerUserCount {
	return rs.serverList
}

// Partial implementation of http://matrix.org/docs/spec/client_server/r0.2.0.html#calculating-the-display-name-for-a-room
// Does not handle based on members if there is no Name/Alias (yet)
func (rs RoomState) CalculateName() string {
	if rs.Name != "" {
		return rs.Name
	}
	if rs.canonicalAlias != "" {
		return rs.canonicalAlias
	}
	return "Empty Room"
}
