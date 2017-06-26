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
	"bytes"
	"encoding/json"
	"github.com/matrix-org/gomatrix"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"html/template"
	"log"
	"regexp"
	"strings"
	"time"
)

func unpack3Values(val []string) (string, string, string) { return val[0], val[1], val[2] }

var mxcRegex = regexp.MustCompile(`mxc://(.+?)/(.+?)(?:#.+)?$`)

type MemberEventContent struct {
	Membership  string `json:"membership,omitempty"`
	AvatarURL   MxcUrl `json:"avatar_url,omitempty"`
	DisplayName string `json:"displayname,omitempty"`
}

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
	"mRoomMember": func(event *gomatrix.Event) interface{} {
		// join -> join = avatar/display name
		// join -> quit = kick/leave
		// * -> join = join
		// * -> invite = invite
		var content, prevContent *MemberEventContent

		dataContent, _ := json.Marshal(event.Content)
		if err := json.Unmarshal(dataContent, &content); err != nil {
			content = &MemberEventContent{}
		}

		var effect string

		dataPrevContent, _ := json.Marshal(event.PrevContent)
		if err := json.Unmarshal(dataPrevContent, &prevContent); err != nil {
			prevContent = &MemberEventContent{}
		}

		sender := event.Sender
		target := *event.StateKey

		switch content.Membership {
		case "invite":
			return sender + " invited " + target + "."
		case "ban":
			return sender + " banned " + target + "(" + event.Content["reason"].(string) + ")."
		case "join":
			if event.PrevContent != nil && prevContent.Membership == "join" {
				if prevContent.DisplayName == "" && content.DisplayName != "" {
					return sender + " set their display name to " + content.DisplayName + "."
				} else if prevContent.DisplayName != "" && content.DisplayName == "" {
					return sender + " removed their display name " + prevContent.DisplayName + "."
				} else if prevContent.DisplayName != content.DisplayName {
					return sender + " changed their display name from " + prevContent.DisplayName + " to " + content.DisplayName + "."
				} else if prevContent.AvatarURL == "" && content.AvatarURL != "" {
					return sender + " set a profile picture."
				} else if prevContent.AvatarURL != "" && content.AvatarURL == "" {
					return sender + " removed their profile picture."
				} else if prevContent.AvatarURL != content.AvatarURL {
					return sender + " changed their profile picture."
				} else {
					return ""
				}
			} else {
				return target + " joined the room."
			}
		case "leave":
			if sender == target {
				if prevContent.Membership == "invite" {
					return target + " rejected invite."
				} else {
					return target + " left the room."
				}
			} else if prevContent.Membership == "ban" {
				return sender + " unbanned " + target + "."
			} else if prevContent.Membership == "leave" {
				return sender + " kicked " + target + "."
			} else if prevContent.Membership == "invite" {
				return sender + " withdrew " + target + "'s invite."
			} else {
				return target + " left the room"
			}
		}

		return effect
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

				p.AllowAttrs("color", "data-mx-bg-color", "data-mx-color").OnElements("font")
				p.AllowAttrs("data-mx-bg-color", "data-mx-color").OnElements("span")
				p.AllowAttrs("href", "name", "target", "rel").OnElements("a")

				p.AllowAttrs("src").OnElements("img")
				p.AllowAttrs("start").OnElements("ol")

				p.AllowURLSchemes("http", "https", "ftp", "mailto")
				p.AddTargetBlankToFullyQualifiedLinks(true)
				p.AddSpaceWhenStrippingTag(true)

				reader := strings.NewReader(event.Content["formatted_body"].(string))
				root, err := html.Parse(reader)

				if err != nil {
					log.Fatal(err)
				}

				var b bytes.Buffer
				html.Render(&b, root.FirstChild.LastChild)
				partiallySanitized := b.String()

				sanitized := p.Sanitize(partiallySanitized)

				return template.HTML(sanitized)
			}
			return event.Content["body"]
		}
	},
}).ParseGlob("templates/*.html"))
