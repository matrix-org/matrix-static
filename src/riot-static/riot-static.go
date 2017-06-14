package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/matrix-org/gomatrix"
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

func main() {
	funcMap := template.FuncMap{
		"mxcToUrl": func(mxc string) string {
			if !strings.HasPrefix(mxc, "mxc://") {
				return ""
			}
			mxc = strings.TrimPrefix(mxc, "mxc://")
			split := strings.SplitN(mxc, "/", 2)

			return cli.BuildBaseURL("_matrix", "media", "r0", "thumbnail", split[0], split[1], "?width=50&height=50&method=crop")

			//return cli.BuildURL("download", split[0], split[1])
		},
	}

	tpl = template.Must(template.New("main").Funcs(funcMap).ParseGlob("templates/*.html"))

	cli, _ = gomatrix.NewClient("https://matrix.org", "", "")

	r := mux.NewRouter()

	r.HandleFunc("/", GetPublicRoomsList)
	r.HandleFunc("/!{roomId}", GetPublicRoom)

	log.Fatal(http.ListenAndServe(":8000", r))
}
