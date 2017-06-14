package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/matrix-org/gomatrix"
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

type RespInitialSync struct {
	AccountData []gomatrix.Event `json:"account_data"`
	//Presence
	Messages   gomatrix.RespMessages `json:"messages"`
	Membership string                `json:"membership"`
	State      []gomatrix.Event      `json:"state"`
	RoomId     string                `json:"room_id"`
	Receipts   []gomatrix.Event      `json:"receipts"`
}

type RespGetRoomAlias struct {
	RoomID  string   `json:"room_id"`
	Servers []string `json:"servers"`
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	panic(err)
}

func GetPublicRoomsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var batch string

	query := r.URL.Query()
	if query["batch"] != nil {
		batch = query["batch"][0]
	}

	resp, err := cli.PublicRooms(30, batch, "")

	if err == nil {
		b := resp.Chunk[:0]
		for _, x := range resp.Chunk {
			if x.WorldReadable {
				b = append(b, x)
			}
		}
		resp.Chunk = b
		err = tpl.ExecuteTemplate(w, "rooms.html", resp)
	}

	if err != nil {
		ErrorHandler(w, r, err)
	}

}

func GetPublicRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	//log.Print(vars["roomId"])

	urlPath := cli.BuildURLWithQuery([]string{"rooms", "!" + vars["roomId"], "initialSync"}, map[string]string{"limit": "20"})
	//urlPath := cli.BuildURL("rooms", vars["roomId"], "initialSync")
	print(urlPath)
	var resp RespInitialSync
	_, err := cli.MakeRequest("GET", urlPath, nil, &resp)

	if err == nil {
		err = tpl.ExecuteTemplate(w, "room.html", resp)

	}

	if err != nil {
		ErrorHandler(w, r, err)
	}
}

var cli *gomatrix.Client
var tpl *template.Template
var config *gomatrix.RespRegister

func main() {
	funcMap := template.FuncMap{
		"mxcToUrl": func(mxc string) string {
			if !strings.HasPrefix(mxc, "mxc://") {
				return ""
			}
			mxc = strings.TrimPrefix(mxc, "mxc://")
			split := strings.SplitN(mxc, "/", 2)

			hsURL, _ := url.Parse(cli.HomeserverURL.String())
			parts := []string{hsURL.Path}
			parts = append(parts, "_matrix", "media", "r0", "thumbnail", split[0], split[1])
			hsURL.Path = path.Join(parts...)

			q := hsURL.Query()
			q.Set("width", "50")
			q.Set("height", "50")
			q.Set("method", "crop")

			hsURL.RawQuery = q.Encode()

			return hsURL.String()
		},
	}

	tpl = template.Must(template.New("main").Funcs(funcMap).ParseGlob("templates/*.html"))

	if _, err := os.Stat("./config.json"); err == nil {
		file, e := ioutil.ReadFile("./config.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		json.Unmarshal(file, &config)
	}

	if config == nil {
		config = new(gomatrix.RespRegister)
	}

	if config.HomeServer == "" {
		config.HomeServer = "https://matrix.org"
	}

	cli, _ = gomatrix.NewClient(config.HomeServer, "", "")

	if config.AccessToken == "" || config.UserID == "" {
		register, inter, err := cli.RegisterGuest(&gomatrix.ReqRegister{})

		if err == nil && inter == nil && register != nil {
			config = register
		} else {
			fmt.Println("Error encountered during guest registration")
			os.Exit(1)
		}

		configJson, _ := json.Marshal(config)
		err = ioutil.WriteFile("./config.json", configJson, 0600)
		if err != nil {
			fmt.Println(err)
		}
	}

	cli.SetCredentials(config.UserID, config.AccessToken)

	r := mux.NewRouter()

	r.HandleFunc("/", GetPublicRoomsList)
	r.HandleFunc("/!{roomId}", GetPublicRoom)

	log.Fatal(http.ListenAndServe(":8000", r))
}
