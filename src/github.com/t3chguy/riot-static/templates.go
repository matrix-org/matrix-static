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
	"github.com/t3chguy/riot-static/matrix-client"
	"golang.org/x/net/html"
	"html/template"
	"strings"
	"time"
)

type MemberEventContent struct {
	Membership  string               `json:"membership,omitempty"`
	AvatarURL   matrix_client.MXCURL `json:"avatar_url,omitempty"`
	DisplayName string               `json:"displayname,omitempty"`
}

func InitTemplates(client *matrix_client.Client) *template.Template {
	return template.Must(template.New("main").Funcs(template.FuncMap{
		"time": func(timestamp int) string {
			return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).Format("2 Jan 2006 15:04:05")
		},
		"plus": func(a, b int) int {
			return a + b
		},
		"minus": func(a, b int) int {
			return a - b
		},
		"URL": func(str string) template.URL {
			return template.URL(str)
		},
		"HTML": func(str string) template.HTML {
			return template.HTML(str)
		},
		"MXCtoThumbUrl": func(mxc matrix_client.MXCURL) template.URL {
			return template.URL(client.MXCToThumbUrl(mxc))
		},
		"MXCtoUrl": func(mxc matrix_client.MXCURL) template.URL {
			return template.URL(client.MXCToUrl(mxc))
		},
		"mRoomMember": func(room *matrix_client.Room, event *gomatrix.Event) interface{} {
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
			senderPretty := room.GetMemberIgnore(event.Sender).GetName()
			target := *event.StateKey
			targetPretty := room.GetMemberIgnore(*event.StateKey).GetName()

			switch content.Membership {
			case "invite":
				return senderPretty + " invited " + targetPretty + "."
			case "ban":
				var reasonString string
				if reason, ok := event.Content["reason"].(string); ok {
					reasonString = " (" + reason + ")"
				}
				return senderPretty + " banned " + targetPretty + reasonString + "."
			case "join":
				if event.PrevContent != nil && prevContent.Membership == "join" {
					if prevContent.DisplayName == "" && content.DisplayName != "" {
						return senderPretty + " set their display name to " + content.DisplayName + "."
					} else if prevContent.DisplayName != "" && content.DisplayName == "" {
						return senderPretty + " removed their display name " + prevContent.DisplayName + "."
					} else if prevContent.DisplayName != content.DisplayName {
						return senderPretty + " changed their display name from " + prevContent.DisplayName + " to " + content.DisplayName + "."
					} else if prevContent.AvatarURL == "" && content.AvatarURL != "" {
						return senderPretty + " set a profile picture."
					} else if prevContent.AvatarURL != "" && content.AvatarURL == "" {
						return senderPretty + " removed their profile picture."
					} else if prevContent.AvatarURL != content.AvatarURL {
						return senderPretty + " changed their profile picture."
					} else {
						return ""
					}
				} else {
					return targetPretty + " joined the room."
				}
			case "leave":
				if sender == target {
					if prevContent.Membership == "invite" {
						return targetPretty + " rejected invite."
					} else {
						return targetPretty + " left the room."
					}
				} else if prevContent.Membership == "ban" {
					return senderPretty + " unbanned " + targetPretty + "."
				} else if prevContent.Membership == "leave" {
					return senderPretty + " kicked " + targetPretty + "."
				} else if prevContent.Membership == "invite" {
					return senderPretty + " withdrew " + targetPretty + "'s invite."
				} else {
					return targetPretty + " left the r"
				}
			}

			return effect
		},
		"mRoomMessage": func(event *gomatrix.Event) interface{} {
			switch event.Content["msgtype"] {
			//case "m.image":
			//case "m.file":
			//case "m.location":
			//case "m.video":
			//case "m.audio":
			case "m.notice", "m.emote", "m.text":
				// These use the default HTML capable renderer
				fallthrough
			default:
				if event.Content["format"] == "org.matrix.custom.html" {
					p := bluemonday.NewPolicy()

					p.AllowElements("font", "del", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "p", "a", "ul", "ol", "nl", "li", "b", "i", "u", "strong", "em", "strike", "code", "hr", "br", "div", "table", "thead", "caption", "tbody", "tr", "th", "td", "pre", "span")

					p.AllowAttrs("color", "data-mx-bg-color", "data-mx-color").OnElements("font")
					p.AllowAttrs("data-mx-bg-color", "data-mx-color").OnElements("span")
					p.AllowAttrs("href", "name", "targetPretty", "rel").OnElements("a")

					//p.AllowAttrs("src").OnElements("img")
					//p.AllowAttrs("start").OnElements("ol")

					p.AllowURLSchemes("http", "https", "ftp", "mailto")
					p.AddTargetBlankToFullyQualifiedLinks(true)
					p.AddSpaceWhenStrippingTag(true)

					reader := strings.NewReader(event.Content["formatted_body"].(string))
					root, err := html.Parse(reader)

					if err != nil {
						return event.Content["body"]
					}

					var b bytes.Buffer
					html.Render(&b, root.FirstChild.LastChild)

					sanitized := p.SanitizeBytes(b.Bytes())

					return template.HTML(sanitized)
				}
				return event.Content["body"]
			}
		},
	}).ParseGlob("templates/*.html"))
}
