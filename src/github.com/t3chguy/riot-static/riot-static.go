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
	"strconv"
	"sync"
)

type RespInitialSync struct {
	AccountData []gomatrix.Event `json:"account_data"`

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

type PublicRooms struct {
	sync.RWMutex
	NumRooms int
	List     []gomatrix.PublicRoomsChunk
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	panic(err)
}

type TemplateRooms struct {
	Rooms    []gomatrix.PublicRoomsChunk
	NumRooms int
	Page     int
}

func paginate(x []gomatrix.PublicRoomsChunk, page int, size int) []gomatrix.PublicRoomsChunk {
	skip := (page - 1) * size

	if skip > len(x) {
		skip = len(x)
	}

	end := skip + size
	if end > len(x) {
		end = len(x)
	}

	return x[skip:end]
}

func GetPublicRoomsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var page int
	query := r.URL.Query()
	if query["page"] != nil {
		page, _ = strconv.Atoi(query["page"][0])
	}

	if page <= 0 {
		page = 1
	}

	pageSize := 20

	publicRooms.RWMutex.RLock()
	numRooms := publicRooms.NumRooms
	someRooms := paginate(publicRooms.List, page, pageSize)
	publicRooms.RWMutex.RUnlock()

	templateRooms := TemplateRooms{someRooms, numRooms, page}

	err := tpl.ExecuteTemplate(w, "rooms.html", templateRooms)

	if err != nil {
		ErrorHandler(w, r, err)
	}

}

func GetPublicRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)

	urlPath := cli.BuildURLWithQuery([]string{"rooms", "!" + vars["roomId"], "initialSync"}, map[string]string{"limit": "64"})
	//urlPath := cli.BuildURL("publicRooms", vars["roomId"], "initialSync")
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

var publicRooms = new(PublicRooms)
var rooms = map[string]struct {
	sync.RWMutex
	InitialSync RespInitialSync
	Alias       RespGetRoomAlias
}{}

func LoadPublicRooms() {
	publicRooms.RWMutex.Lock()
	fmt.Println("Loading public publicRooms")
	resp, err := cli.PublicRooms(0, "", "")

	if err == nil {
		b := resp.Chunk[:0]

		// filter on actually WorldReadable publicRooms
		for _, x := range resp.Chunk {
			if x.WorldReadable {
				b = append(b, x)
			}
		}

		// copy them so we don't encounter slice hell
		publicRooms.NumRooms = len(b)
		publicRooms.List = make([]gomatrix.PublicRoomsChunk, publicRooms.NumRooms)
		copy(publicRooms.List, b)
	}
	publicRooms.RWMutex.Unlock()

	if err != nil {
		panic(err)
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
		"plus": func(a, b int) int {
			return a + b
		},
		"minus": func(a, b int) int {
			return a - b
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
			register.HomeServer = config.HomeServer
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

	go LoadPublicRooms()

	r := mux.NewRouter()

	r.HandleFunc("/", GetPublicRoomsList)
	r.HandleFunc("/!{roomId}", GetPublicRoom)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Fatal(http.ListenAndServe(":"+port, r))
}
