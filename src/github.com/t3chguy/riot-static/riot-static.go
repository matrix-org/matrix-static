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
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

const PublicRoomsPageSize = 20
const RoomMembersPageSize = 20

func LoadPublicRooms(first bool) {
	fmt.Println("Loading publicRooms")
	resp, err := cli.PublicRooms(0, "", "")

	if err != nil {
		// Only panic if first one fails, after that we only have outdated data (less important)
		if first {
			panic(err)
		} else {
			fmt.Println(err)
		}
	}

	// Preallocate the maximum capacity possibly needed (if all rooms were world readable)
	worldReadableRooms := make([]*Room, 0, len(resp.Chunk))

	// filter on actually WorldReadable publicRooms
	for _, x := range resp.Chunk {
		if x.WorldReadable {
			room := NewRoom(x)

			if existingRoom, exists := data.GetRoom(x.RoomId); exists {
				room.Cached = existingRoom.Cached
				// Copy existing Cache
			}

			// Append world readable room to the filtered list.
			worldReadableRooms = append(worldReadableRooms, room)
		}
	}
	data.SetRoomList(worldReadableRooms)
}

var data = new(DataStore)

func main() {
	setupCli()

	router := gin.Default()
	router.SetHTMLTemplate(tpl)
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		page, skip, end := calcPaginationPage(c.DefaultQuery("page", "1"), PublicRoomsPageSize)
		c.HTML(http.StatusOK, "rooms.html", gin.H{
			"Rooms": data.GetRoomList(skip, end),
			"Page":  page,
		})
	})

	roomRouter := router.Group("/room/")
	{
		roomRouter.Use(func(c *gin.Context) {
			roomId := c.Param("roomId")

			if room, exists := data.GetRoom(roomId); exists {
				// Start of debug code
				//if _, exists := c.GetQuery("clear"); exists {
				//	room.Once.Reset()
				//}
				// End of debug code

				c.Set("Room", &room)
				c.Next()
			} else {
				c.String(http.StatusNotFound, "Room Not Found")
				c.Abort()
			}
		})

		roomRouter.GET("/:roomId/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "chat")
		})

		roomRouter.GET("/:roomId/chat", func(c *gin.Context) {
			c.HTML(http.StatusOK, "room.html", gin.H{
				"Room": c.MustGet("Room").(*Room),
			})
		})

		roomRouter.GET("/:roomId/servers", func(c *gin.Context) {
			c.HTML(http.StatusOK, "room_servers.html", gin.H{
				"Room": c.MustGet("Room").(*Room),
			})
		})

		roomRouter.GET("/:roomId/members", func(c *gin.Context) {
			page, skip, end := calcPaginationPage(c.DefaultQuery("page", "1"), RoomMembersPageSize)
			room := *c.MustGet("Room").(*Room)

			c.HTML(http.StatusOK, "room_members.html", gin.H{
				"Room":       room,
				"MemberInfo": room.GetMemberList(skip, end),
				"NumMembers": room.GetNumMembers(),
				"Page":       page,
			})
		})

		roomRouter.GET("/:roomId/members/:mxid", func(c *gin.Context) {
			room := c.MustGet("Room").(*Room)
			mxid := c.Param("mxid")

			if memberInfo, exists := room.GetMember(mxid); exists {
				c.HTML(http.StatusOK, "member_info.html", gin.H{
					"MemberInfo": memberInfo,
					"Room":       room,
				})
			} else {
				c.AbortWithStatus(http.StatusNotFound)
			}
		})

		roomRouter.GET("/:roomId/power_levels", func(c *gin.Context) {
			c.HTML(http.StatusOK, "power_levels.html", gin.H{
				"Room": c.MustGet("Room").(*Room),
			})
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	go runCron()
	fmt.Println("Listening on port " + port)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
		Addr:         ":" + port,
	}

	panic(srv.ListenAndServe())
}

const LoadPublicRoomsPeriod = time.Hour

func runCron() {
	LoadPublicRooms(true)
	t := time.NewTicker(LoadPublicRoomsPeriod)
	for {
		<-t.C
		LoadPublicRooms(false)
	}
}
