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
	"errors"
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/riot-static/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
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
type Client struct{ *gomatrix.Client }

func (m *Client) RoomInitialSync(roomID string, limit int) (resp *RespInitialSync, err error) {
	urlPath := m.BuildURLWithQuery([]string{"rooms", roomID, "initialSync"}, map[string]string{
		"limit": strconv.Itoa(limit),
	})
	_, err = m.MakeRequest("GET", urlPath, nil, &resp)
	return
}

const minimumPagination = 64

// TODO split into runs of max 999 recursively otherwise we get capped.
func (m *Client) backpaginateRoom(room *Room, amount int) (int, error) {
	amount = utils.Max(amount, minimumPagination)
	backPaginationToken, _ := room.GetTokens()
	resp, err := m.Messages(room.ID, backPaginationToken, "", 'b', amount)

	if err != nil {
		return -1, err
	}

	room.concatBackpagination(resp.Chunk, resp.End)
	return len(resp.Chunk), nil
}

func (m *Client) forwardpaginateRoom(room *Room, amount int) (int, error) {
	amount = utils.Max(amount, minimumPagination)

	_, forwardPaginationToken := room.GetTokens()
	resp, err := m.Messages(room.ID, forwardPaginationToken, "", 'f', amount)

	if err != nil {
		return -1, err
	}

	// I would have thought to use resp.Start here but NOPE
	room.concatForwardPagination(resp.Chunk, resp.End)
	return len(resp.Chunk), nil
}

func newClient(homeserverURL, userID, accessToken string) (*Client, error) {
	cli, err := gomatrix.NewClient(homeserverURL, userID, accessToken)
	cli.Client = &http.Client{
		Timeout: 5 * time.Second,
	}
	return &Client{cli}, err
}

func NewGuest(configPath string, homeserverURL string) error {
	m, err := newClient(homeserverURL, "", "")
	if err != nil {
		return err
	}

	register, inter, err := m.RegisterGuest(&gomatrix.ReqRegister{})

	if err != nil {
		return err
	}
	if inter != nil || register == nil {
		return errors.New("Error encountered during guest registration")
	}

	// TODO consider SRV Query instead.
	register.HomeServer = homeserverURL

	configJson, err := json.Marshal(register)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, configJson, 0600)
}

func NewClient(configPath string) (*Client, error) {
	var config *gomatrix.RespRegister

	if _, err := os.Stat(configPath); err != nil {
		return nil, errors.New("Config file not found and Guest Registration not permitted by lack of command line flag (--create-guest-account)")
	}

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &config)

	if config.HomeServer == "" {
		return nil, errors.New("No user configuration found and Guest Registration not permitted by lack of command line flag (--create-guest-account)")
	}

	return newClient(config.HomeServer, config.UserID, config.AccessToken)
}
