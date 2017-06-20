package main

import (
	"html/template"
	"strings"

	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"sync"
	"time"
)

func paginate(x []*Room, page int, size int) []*Room {
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

func GetPublicRoomsList(c *gin.Context) {
	data.Once.Do(LoadPublicRooms)

	var page int
	var err error
	if page, err = strconv.Atoi(c.DefaultQuery("page", "1")); err != nil {
		page = 1
	}

	pageSize := 20

	data.RLock()
	c.HTML(http.StatusOK, "rooms.html", gin.H{
		"Rooms":    paginate(data.Ordered, page, pageSize),
		"NumRooms": data.NumRooms,
		"Page":     page,
	})
	data.RUnlock()
}

func GetPublicRoom(c *gin.Context) {
	roomId := c.Param("roomId")

	data.RLock()
	c.HTML(http.StatusOK, "room.html", gin.H{
		"Room": &data.Rooms[roomId],
	})
	data.RUnlock()
}

func GetPublicRoomServers(c *gin.Context) {
	roomId := c.Param("roomId")

	data.RLock()
	c.HTML(http.StatusOK, "room_servers.html", gin.H{
		"Room": &data.Rooms[roomId],
	})
	data.RUnlock()
}

func GetPublicRoomMembers(c *gin.Context) {
	roomId := c.Param("roomId")

	data.RLock()
	c.HTML(http.StatusOK, "room_members.html", gin.H{
		"Room": &data.Rooms[roomId],
	})
	data.RUnlock()
}

var data = struct {
	sync.Once
	sync.RWMutex
	NumRooms int
	Ordered  []*Room
	Rooms    map[string]*Room
}{}

func LoadPublicRooms() {
	fmt.Println("Loading public publicRooms")
	resp, err := cli.PublicRooms(0, "", "")

	if err == nil {
		b := []*Room{}
		c := map[string]*Room{}

		// filter on actually WorldReadable publicRooms
		for _, x := range resp.Chunk {
			if x.WorldReadable {
				var room *Room
				if data.Rooms[x.RoomId] != nil {
					room = data.Rooms[x.RoomId]
				} else {
					room = NewRoom(x)
				}
				b = append(b, room)
				c[x.RoomId] = room
			}
		}

		data.Lock()
		data.Rooms = c
		data.NumRooms = len(b)
		// copy order so we don't encounter slice hell
		data.Ordered = make([]*Room, data.NumRooms)
		copy(data.Ordered, b)

		data.Unlock()
	}

	if err != nil {
		panic(err)
	}
}

var tpl *template.Template = template.Must(template.New("main").Funcs(template.FuncMap{
	"mxcToUrl": func(mxc string) string {
		if !strings.HasPrefix(mxc, "mxc://") {
			return ""
		}

		_, serverName, mediaId := unpackThreeVals(mxcRegex.FindStringSubmatch(mxc))

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

func FetchRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomId := c.Param("roomId")
		data.Rooms[roomId].Fetch()
	}
}

func FailIfNoRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomId := c.Param("roomId")
		if data.Rooms[roomId] == nil {
			c.String(http.StatusNotFound, "Room Not Found")
			c.Abort()
		}
	}
}

var mxcRegex = regexp.MustCompile(`mxc://(.+)/(.+)(?:#.+)?`)

func unpackThreeVals(val []string) (string, string, string) { return val[0], val[1], val[2] }

func main() {
	setupCli()

	//go LoadPublicRooms()
	// Synchronous cache fill
	data.Once.Do(LoadPublicRooms)

	router := gin.Default()
	router.SetHTMLTemplate(tpl)

	router.GET("/", GetPublicRoomsList)

	roomRouter := router.Group("/room/")
	{
		roomRouter.Use(FailIfNoRoom())
		roomRouter.Use(FetchRoom())

		roomRouter.GET("/:roomId", GetPublicRoom)
		roomRouter.GET("/:roomId/servers", GetPublicRoomServers)
		roomRouter.GET("/:roomId/members", GetPublicRoomMembers)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router.Run(":" + port)
}
