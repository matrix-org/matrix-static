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
	"fmt"
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/utils"
	"io/ioutil"
	"os"
	"strconv"
)

type RespInitialSync struct {
	//AccountData []gomatrix.Event `json:"account_data"`

	Messages gomatrix.RespMessages `json:"messages"`
	//Membership string                 `json:"membership"`
	State []gomatrix.Event `json:"state"`
	//RoomID     string                 `json:"room_id"`
	//Receipts   []*gomatrix.Event      `json:"receipts"`
	//Presence   []*PresenceEvent       `json:"presence"`
}

type Client struct {
	*gomatrix.Client
	*RoomStore
}

// Implement Cache as a layer on top of this?
// Caching will cache r state historical calculations too
// rather than just the raw data
// memberInfo etc can just be stored as a list and calculated OD for the cache.

func (m *Client) RoomInitialSync(roomID string, limit int) (resp *RespInitialSync, err error) {
	urlPath := m.BuildURLWithQuery([]string{"rooms", roomID, "initialSync"}, map[string]string{
		"limit": strconv.Itoa(limit),
	})
	_, err = m.MakeRequest("GET", urlPath, nil, &resp)
	return
}

const minimumPagination = 64

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

var config *gomatrix.RespRegister

func NewClient() *Client {
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

	return &Client{cli, new(RoomStore)}
}
