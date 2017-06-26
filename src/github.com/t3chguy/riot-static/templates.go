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
	"github.com/matrix-org/gomatrix"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"regexp"
	"time"
)

func unpack3Values(val []string) (string, string, string) { return val[0], val[1], val[2] }

var mxcRegex = regexp.MustCompile(`mxc://(.+?)/(.+?)(?:#.+)?$`)

var tpl *template.Template = template.Must(template.New("main").Funcs(template.FuncMap{
	"time": func(timestamp int) string {
		return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).Format("2 Jan 2006 15:04:05")
	},
	"plus": func(a, b int) int {
		return a + b
	},
	"minus": func(a, b int) int {
		return a - b
	},
	"HTML": func(str string) template.HTML {
		return template.HTML(str)
	},
	"mRoomMessage": func(event *gomatrix.Event) interface{} {
		switch event.Content["msgtype"] {
		case "m.notice":
			fallthrough
		case "m.emote":
			fallthrough
		case "m.text":
			fallthrough
		default:
			if event.Content["format"] == "org.matrix.custom.html" {
				//p := bluemonday.NewPolicy()
				p := bluemonday.UGCPolicy()

				p.AllowElements("font", "del", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "p", "a", "ul", "ol", "nl", "li", "b", "i", "u", "strong", "em", "strike", "code", "hr", "br", "div", "table", "thead", "caption", "tbody", "tr", "th", "td", "pre", "span")

				p.AllowAttrs("color", "data-mx-bg-color", "data-mx-color", "style").OnElements("font")
				p.AllowAttrs("data-mx-bg-color", "data-mx-color", "style").OnElements("span")
				p.AllowAttrs("href", "name", "target", "rel").OnElements("a")

				p.AllowAttrs("src").OnElements("img")
				p.AllowAttrs("start").OnElements("ol")

				p.AllowURLSchemes("http", "https", "ftp", "mailto")
				p.AddTargetBlankToFullyQualifiedLinks(true)
				p.AddSpaceWhenStrippingTag(true)

				return template.HTML(p.Sanitize(event.Content["formatted_body"].(string)))
			}
			return event.Content["body"]
		}
	},
}).ParseGlob("templates/*.html"))
