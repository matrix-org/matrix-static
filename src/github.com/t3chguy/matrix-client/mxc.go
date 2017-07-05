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
	"github.com/matrix-org/gomatrix"
	"net/url"
	"path"
	"regexp"
)

// mxcRegex allows splitting an mxc into a serverName and mediaId
// FindStringSubmatch of which results in [_, serverName, mediaId] if valid
// and [] if invalid mxc is provided.
// Examples:
// "mxc://foo/bar" => ["mxc://foo/bar", "foo", "bar"]
// "mxc://bar/foo#auto" => ["mxc://bar/foo#auto", "bar", "foo"]
// "invalidMxc://whatever" => [] (Invalid MXC Caught)
var mxcRegex = regexp.MustCompile(`mxc://(.+?)/(.+?)(?:#.+)?$`)

type MXCURL string

func (m *MXCURL) split() (ok bool, serverName string, mediaId string) {
	mxc := string(*m)
	matches := mxcRegex.FindStringSubmatch(mxc)

	ok = true
	if len(matches) != 3 {
		return false, "", ""
	}

	serverName = matches[1]
	mediaId = matches[2]
	return
}

func (m *MXCURL) mapMxcUrl(cli *gomatrix.Client, kind string) string {
	ok, serverName, mediaId := m.split()
	if !ok {
		return ""
	}

	hsURL, _ := url.Parse(cli.HomeserverURL.String())
	parts := []string{hsURL.Path}
	parts = append(parts, "_matrix", "media", "r0", kind, serverName, mediaId)
	hsURL.Path = path.Join(parts...)

	q := hsURL.Query()
	q.Set("width", "50")
	q.Set("height", "50")
	q.Set("method", "crop")

	hsURL.RawQuery = q.Encode()

	return hsURL.String()
}

func (m *Client) MXCToThumbUrl(mxcurl MXCURL) string {
	return mxcurl.mapMxcUrl(m.Client, "thumbnail")
}

func (m *Client) MXCToUrl(mxcurl MXCURL) string {
	return mxcurl.mapMxcUrl(m.Client, "download")
}
