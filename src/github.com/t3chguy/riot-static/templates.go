// Copyright 2017 Michael Telatynski <7t3cghuy@gmail.com>
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
	"html/template"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

func unpack3Values(val []string) (string, string, string) { return val[0], val[1], val[2] }

var mxcRegex = regexp.MustCompile(`mxc://(.+?)/(.+?)(?:#.+)?$`)

var tpl *template.Template = template.Must(template.New("main").Funcs(template.FuncMap{
	"mxcToUrl": func(mxc string) string {
		if !strings.HasPrefix(mxc, "mxc://") {
			return ""
		}

		_, serverName, mediaId := unpack3Values(mxcRegex.FindStringSubmatch(mxc))

		hsURL, _ := url.Parse(cli.HomeserverURL.String())
		parts := []string{hsURL.Path}
		parts = append(parts, "_matrix", "media", "r0", "thumbnail", serverName, mediaId)
		hsURL.Path = path.Join(parts...)

		q := hsURL.Query()
		q.Set("width", "50")
		q.Set("height", "50")
		q.Set("method", "crop")

		hsURL.RawQuery = q.Encode()

		return hsURL.String()
	},
	"time": func(timestamp int) string {
		return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).Format("2 Jan 2006 15:04:05")
	},
	"plus": func(a, b int) int {
		return a + b
	},
	"minus": func(a, b int) int {
		return a - b
	},
	"fallback": func(args ...interface{}) string {
		for _, strO := range args {
			if strO != nil {
				if str := strO.(string); str != "" {
					return str
				}
			}
		}
		return ""
	},
}).ParseGlob("templates/*.html"))
