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

var mxcRegex = regexp.MustCompile(`mxc://(.+)/(.+)(?:#.+)?`)

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
		return time.Unix(int64(timestamp), 0).Format(time.RFC822)
	},
	"plus": func(a, b int) int {
		return a + b
	},
	"minus": func(a, b int) int {
		return a - b
	},
}).ParseGlob("templates/*.html"))
