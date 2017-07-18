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
	"net/url"
	"path"
	"regexp"
	"strconv"
)

// mxcRegex allows splitting an mxc into a serverName and mediaId
// FindStringSubmatch of which results in [_, serverName, mediaId] if valid
// and [] if invalid mxc is provided.
// Examples:
// "mxc://foo/bar" => ["mxc://foo/bar", "foo", "bar"]
// "mxc://bar/foo#auto" => ["mxc://bar/foo#auto", "bar", "foo"]
// "invalidMxc://whatever" => [] (Invalid MXC Caught)
var mxcRegex = regexp.MustCompile(`mxc://(.+?)/(.+?)(?:#.+)?$`)

type MXCURL struct {
	string
	homeserverURL string
}

func NewMXCURL(url string, baseUrl string) *MXCURL {
	return &MXCURL{url, baseUrl}
}

func (m *MXCURL) IsValid() bool {
	ok, _, _ := m.split()
	return ok
}

func (m *MXCURL) split() (ok bool, serverName string, mediaId string) {
	mxc := m.string
	matches := mxcRegex.FindStringSubmatch(mxc)

	ok = true
	if len(matches) != 3 {
		return false, "", ""
	}

	serverName = matches[1]
	mediaId = matches[2]
	return
}

func (m *MXCURL) mapMxcUrl(kind string) *url.URL {
	ok, serverName, mediaId := m.split()
	if !ok {
		return nil
	}

	hsURL, _ := url.Parse(m.homeserverURL)
	parts := []string{hsURL.Path}
	parts = append(parts, "_matrix", "media", "r0", kind, serverName, mediaId)
	hsURL.Path = path.Join(parts...)
	return hsURL
}

func (m *MXCURL) ToThumbURL(width, height int, method string) string {
	mediaUrl := m.mapMxcUrl("thumbnail")

	if mediaUrl == nil {
		return ""
	}

	q := mediaUrl.Query()
	q.Set("width", strconv.Itoa(width))
	q.Set("height", strconv.Itoa(height))
	q.Set("method", method)

	mediaUrl.RawQuery = q.Encode()

	return mediaUrl.String()
}

func (m *MXCURL) ToURL() string {
	return m.mapMxcUrl("download").String()
}
