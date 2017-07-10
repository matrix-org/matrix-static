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
	"github.com/t3chguy/riot-static/utils"
	"io/ioutil"
	"os"
	"strconv"
)

// This is a Truncated RespInitialSync as we only need SOME information from it.
type RespInitialSync struct {
	//AccountData []gomatrix.Event `json:"account_data"`

	Messages gomatrix.RespMessages `json:"messages"`
	//Membership string                 `json:"membership"`
	State []gomatrix.Event `json:"state"`
	//RoomID     string                 `json:"room_id"`
	//Receipts   []*gomatrix.Event      `json:"receipts"`
	//Presence   []*PresenceEvent       `json:"presence"`
}

// Our Client extension adds some methods and ties in a RoomStore (Ordered Map)
type Client struct {
	*gomatrix.Client
	*RoomStore
}

func (m *Client) RoomInitialSync(roomID string, limit int) (resp *RespInitialSync, err error) {
	urlPath := m.BuildURLWithQuery([]string{"rooms", roomID, "initialSync"}, map[string]string{
		"limit": strconv.Itoa(limit),
	})
	_, err = m.MakeRequest("GET", urlPath, nil, &resp)
	return
}

const minimumPagination = 64

// TODO split into runs of max 999 recursively otherwise we get capped.
func (m *Client) backpaginateRoom(room *Room, amount int) int {
	amount = utils.Max(amount, minimumPagination)
	backPaginationToken, _ := room.GetTokens()
	resp, err := m.Messages(room.ID, backPaginationToken, "", 'b', amount)

	if err != nil {
		return 0
	}

	room.concatBackpagination(resp.Chunk, resp.End)
	return len(resp.Chunk)
}

func (m *Client) forwardpaginateRoom(room *Room, amount int) int {
	amount = utils.Max(amount, minimumPagination)

	room.requestLock.Lock()
	defer room.requestLock.Unlock()

	_, forwardPaginationToken := room.GetTokens()
	resp, err := m.Messages(room.ID, forwardPaginationToken, "", 'f', amount)

	if err != nil {
		return 0
	}

	// I would have thought to use resp.Start here but NOPE
	room.concatForwardPagination(resp.Chunk, resp.End)
	return len(resp.Chunk)
}

func NewClient() *Client {
	var config *gomatrix.RespRegister
	if _, err := os.Stat("./config.json"); err == nil {
		file, e := ioutil.ReadFile("./config.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		json.Unmarshal(file, &config)
	}

	if config == nil {
		config = new(gomatrix.RespRegister)
	}

	if config.HomeServer == "" {
		config.HomeServer = "https://matrix.org"
	}

	cli, _ := gomatrix.NewClient(config.HomeServer, "", "")

	if config.AccessToken == "" || config.UserID == "" {
		register, inter, err := cli.RegisterGuest(&gomatrix.ReqRegister{})

		if err != nil || inter != nil || register == nil {
			fmt.Println("Error encountered during guest registration")
			os.Exit(1)
		}

		register.HomeServer = config.HomeServer
		config = register

		configJson, _ := json.Marshal(config)
		err = ioutil.WriteFile("./config.json", configJson, 0600)
		if err != nil {
			fmt.Println(err)
		}
	}

	cli.SetCredentials(config.UserID, config.AccessToken)

	return &Client{cli, NewRoomStore()}
}
