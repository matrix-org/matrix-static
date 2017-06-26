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

/*
	@TODO Show more than just textual events (Images etc)
	@TODO Historical member info
	@TODO make Peeking more useful (paginate etc)
	@TODO investigate best practices for the spider to get the most out of it
	@TODO make cache invalidate self after N requests of M minutes
*/

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"sync"
)

func paginate(page int, size int, length int) (skip int, end int) {
	if skip = (page - 1) * size; skip > length {
		skip = length
	}
	if end = skip + size; end > length {
		end = length
	}
	return
}

func GetPublicRoomsList(c *gin.Context) {
	var page int
	var err error
	if page, err = strconv.Atoi(c.DefaultQuery("page", "1")); err != nil {
		page = 1
	}

	pageSize := 20

	data.RLock()
	skip, end := paginate(page, pageSize, data.NumRooms)
	c.HTML(http.StatusOK, "rooms.html", gin.H{
		"Rooms":    data.Ordered[skip:end],
		"NumRooms": data.NumRooms,
		"Page":     page,
	})
	data.RUnlock()
}

func GetPublicRoom(c *gin.Context) {
	roomId := c.Param("roomId")

	data.RLock()
	c.HTML(http.StatusOK, "room.html", gin.H{
		"Room": data.Rooms[roomId],
	})
	data.RUnlock()
}

func GetPublicRoomServers(c *gin.Context) {
	roomId := c.Param("roomId")

	data.RLock()
	c.HTML(http.StatusOK, "room_servers.html", gin.H{
		"Room": data.Rooms[roomId],
	})
	data.RUnlock()
}

func GetPublicRoomMembers(c *gin.Context) {
	roomId := c.Param("roomId")

	//var page int
	//var err error
	//if page, err = strconv.Atoi(c.DefaultQuery("page", "1")); err != nil {
	//	page = 1
	//}

	//pageSize := 20

	data.RLock()
	length := len(data.Rooms[roomId].MemberInfo)
	//skip, end := paginate(page, pageSize, length)
	c.HTML(http.StatusOK, "room_members.html", gin.H{
		"Room": data.Rooms[roomId],
		//"MemberInfo": data.Rooms[roomId].Members[skip:end],
		"NumMembers": length,
	})
	data.RUnlock()
}

func GetPublicRoomPowerLevels(c *gin.Context) {
	roomId := c.Param("roomId")

	data.RLock()
	c.HTML(http.StatusOK, "power_levels.html", gin.H{
		"PowerLevels": data.Rooms[roomId].PowerLevels,
	})
	data.RUnlock()
}

func GetPublicRoomMember(c *gin.Context) {
	roomId := c.Param("roomId")
	mxid := c.Param("mxid")

	data.RLock()
	if memberInfo := data.Rooms[roomId].MemberInfo[mxid]; memberInfo != nil {
		c.HTML(http.StatusOK, "member_info.html", gin.H{
			"RoomID":     roomId,
			"MXID":       mxid,
			"MemberInfo": memberInfo,
		})
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
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
	// @TODO: fix this.
	data.Lock()
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

		//data.Lock()
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

func main() {
	setupCli()

	router := gin.Default()
	router.SetHTMLTemplate(tpl)
	router.Static("/assets", "./assets")

	router.Use(func(c *gin.Context) {
		data.Once.Do(LoadPublicRooms)
	})

	router.GET("/", GetPublicRoomsList)

	roomRouter := router.Group("/room/")
	{
		roomRouter.Use(func(c *gin.Context) {
			roomId := c.Param("roomId")
			if data.Rooms[roomId] == nil {
				c.String(http.StatusNotFound, "Room Not Found")
				c.Abort()
			} else {
				data.Rooms[roomId].Fetch()
				c.Next()
			}
		})

		roomRouter.GET("/:roomId", GetPublicRoom)
		roomRouter.GET("/:roomId/servers", GetPublicRoomServers)
		roomRouter.GET("/:roomId/members", GetPublicRoomMembers)
		roomRouter.GET("/:roomId/members/:mxid", GetPublicRoomMember)
		roomRouter.GET("/:roomId/power_levels", GetPublicRoomPowerLevels)
	}

	router.GET("/clear", func(c *gin.Context) {
		data.Once = sync.Once{}
		for _, room := range data.Rooms {
			room.Once = sync.Once{}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router.Run(":" + port)
}
