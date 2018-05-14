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
	log "github.com/Sirupsen/logrus"
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/matrix-static/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// This is a Truncated RespInitialSync as we only need SOME information from it.
type RespInitialSync struct {
	// AccountData []gomatrix.Event `json:"account_data"`

	Messages gomatrix.RespMessages `json:"messages"`
	// Membership string                 `json:"membership"`
	State []gomatrix.Event `json:"state"`
	// RoomID     string                 `json:"room_id"`
	// Receipts   []*gomatrix.Event      `json:"receipts"`
	// Presence   []*PresenceEvent       `json:"presence"`
}

// Our Client extension adds some methods
// and a field expanding an MXC URL
type Client struct {
	*gomatrix.Client
	MediaBaseURL string
}

// Register makes an HTTP request according to http://matrix.org/docs/spec/client_server/r0.2.0.html#get-matrix-client-r0-rooms-roomid-initialsync
func (m *Client) RoomInitialSync(roomID string, limit int) (resp *RespInitialSync, err error) {
	urlPath := m.BuildURLWithQuery([]string{"rooms", roomID, "initialSync"}, map[string]string{
		"limit": strconv.Itoa(limit),
	})
	_, err = m.MakeRequest("GET", urlPath, nil, &resp)
	return
}

type RespRoomDirectoryAlias struct {
	RoomID  string   `json:"room_id"`
	Servers []string `json:"servers"`
}

func (m *Client) GetRoomDirectoryAlias(roomAlias string) (resp *RespRoomDirectoryAlias, err error) {
	urlPath := m.BuildURL("directory", "room", roomAlias)
	_, err = m.MakeRequest("GET", urlPath, nil, &resp)
	return
}

const minimumPagination = 64

// TODO split into runs of max size recursively otherwise synapse may enforce its own limit (999?)
func (m *Client) backpaginateRoom(room *Room, amount int) (int, error) {
	loggerWithFields := log.WithField("roomID", room.ID).WithField("amount", amount)
	loggerWithFields.Info("Backpaginating Room")

	amount = utils.Max(amount, minimumPagination)
	resp, err := m.Messages(room.ID, room.backPaginationToken, "", 'b', amount)

	if err != nil {
		loggerWithFields.WithError(err).Error("Failed Backpaginating Room")
		return -1, err
	}

	room.concatBackpagination(resp.Chunk, resp.End)
	loggerWithFields.Info("Finished Backpaginating Room")
	return len(resp.Chunk), nil
}

func (m *Client) forwardpaginateRoom(room *Room, amount int) (int, error) {
	amount = utils.Max(amount, minimumPagination)
	resp, err := m.Messages(room.ID, room.forwardPaginationToken, "", 'f', amount)

	if err != nil {
		return -1, err
	}

	// I would have thought to use resp.Start here but NOPE
	room.concatForwardPagination(resp.Chunk, resp.End)
	return len(resp.Chunk), nil
}

// NewRawClient returns a wrapped client with http client timeouts applied.
func NewRawClient(homeserverURL, mediaBaseURL, userID, accessToken string) (*Client, error) {
	cli, err := gomatrix.NewClient(homeserverURL, userID, accessToken)
	cli.Client = &http.Client{
		Timeout: 30 * time.Second,
	}
	return &Client{cli, mediaBaseURL}, err
}

// The struct representing the json config file format.
type Config struct {
	AccessToken  string `json:"access_token"`
	DeviceID     string `json:"device_id"`
	HomeServer   string `json:"home_server"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	MediaBaseUrl string `json:"media_base_url"`
}

// NewClient returns a Client configured by the config file found at configPath or an error if encountered.
func NewClient(configPath string) (*Client, error) {
	var config Config

	if _, err := os.Stat(configPath); err != nil {
		return nil, errors.New("config file not found")
	}

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &config)

	if config.HomeServer == "" {
		return nil, errors.New("no user configuration found")
	}

	if config.MediaBaseUrl == "" {
		config.MediaBaseUrl = config.HomeServer
	}

	return NewRawClient(config.HomeServer, config.MediaBaseUrl, config.UserID, config.AccessToken)
}
