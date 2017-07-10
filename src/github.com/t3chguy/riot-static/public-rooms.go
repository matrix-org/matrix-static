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
	"fmt"
	"github.com/t3chguy/riot-static/mxclient"
)

func LoadPublicRooms(client *mxclient.Client, first bool) {
	fmt.Println("Loading publicRooms")
	resp, err := client.PublicRooms(0, "", "")

	if err != nil {
		// Only panic if first one fails, after that we only have outdated data (less important)
		if first {
			panic(err)
		} else {
			fmt.Println(err)
		}
	}

	var worldReadableRooms []*mxclient.Room

	// filter on actually WorldReadable publicRooms
	for _, x := range resp.Chunk {
		if !x.WorldReadable {
			continue
		}

		var room *mxclient.Room
		if existingRoom := client.GetRoom(x.RoomId); existingRoom != nil {
			room = existingRoom
		} else {
			room = client.NewRoom(x)
		}

		// Append world readable r to the filtered list.
		worldReadableRooms = append(worldReadableRooms, room)
	}
	client.SetRoomList(worldReadableRooms)
}
