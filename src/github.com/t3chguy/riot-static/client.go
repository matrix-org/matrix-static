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
)

func /*(cli *client)*/ fetchRoomInitialSync(roomId string) (resp *RespInitialSync, err error) {
	urlPath := cli.BuildURLWithQuery([]string{"rooms", roomId, "initialSync"}, map[string]string{
		"limit": "20",
	})
	//urlPath := cli.BuildURL("rooms", room.RoomID, "initialSync")
	_, err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

func (initialSync *RespInitialSync) ReadState() (memberInfo map[string]*MemberInfo, powerLevels *PowerLevelsEvent, err error) {
	memberInfo = make(map[string]*MemberInfo)
	for _, stateEvent := range initialSync.State {
		stateKey := *stateEvent.StateKey
		switch stateEvent.Type {
		case "m.room.member":
			if memberInfo[stateKey] == nil {
				memberInfo[stateKey] = &MemberInfo{MXID: stateKey}
			}

			if avatarUrl, ok := stateEvent.Content["avatar_url"].(string); ok {
				memberInfo[stateKey].AvatarURL = MxcUrl(avatarUrl)
			}
			if membership, ok := stateEvent.Content["membership"].(string); ok {
				memberInfo[stateKey].Membership = membership
			}
			if displayname, ok := stateEvent.Content["displayname"].(string); ok {
				memberInfo[stateKey].DisplayName = displayname
			}

		case "m.room.power_levels":
			var data []byte
			if data, err = json.Marshal(stateEvent.Content); err != nil {
				return
			}
			if err = json.Unmarshal(data, &powerLevels); err != nil {
				return
			}
		}
	}
	return
}

func /*(cli *client)*/ peekRoom(roomId string) (initialSync *RespInitialSync, memberInfo map[string]*MemberInfo, powerLevels *PowerLevelsEvent, err error) {
	initialSync, err = fetchRoomInitialSync(roomId)
	if err != nil {
		return
	}
	memberInfo, powerLevels, err = initialSync.ReadState()
	return

	//room.powerLevels = powerLevelsEvent
	//room.memberMap = memberInfo
	//room.InitialSync = &resp
}
