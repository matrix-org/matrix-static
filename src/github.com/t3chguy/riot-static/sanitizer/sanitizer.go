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

package sanitizer

import (
	"bytes"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"strings"
)

type Sanitizer struct {
	*bluemonday.Policy
}

func (s *Sanitizer) Sanitize(str string) (sanitizedStr string, ok bool) {
	reader := strings.NewReader(str)
	root, err := html.Parse(reader)

	if err != nil {
		return "", false
	}

	var b bytes.Buffer
	html.Render(&b, root.FirstChild.LastChild)

	return string(s.SanitizeBytes(b.Bytes())), true
}

func InitSanitizer() *Sanitizer {
	p := bluemonday.NewPolicy()

	p.AllowElements("font", "del", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "p", "a", "ul", "ol", "nl", "li", "b", "i", "u", "strong", "em", "strike", "code", "hr", "br", "div", "table", "thead", "caption", "tbody", "tr", "th", "td", "pre", "span")

	p.AllowAttrs("color", "data-mx-bg-color", "data-mx-color").OnElements("font")
	p.AllowAttrs("data-mx-bg-color", "data-mx-color").OnElements("span")
	p.AllowAttrs("href", "name", "targetPretty", "rel").OnElements("a")

	p.AllowURLSchemes("http", "https", "ftp", "mailto")
	p.AddTargetBlankToFullyQualifiedLinks(true)
	p.AddSpaceWhenStrippingTag(true)

	return &Sanitizer{p}
}
