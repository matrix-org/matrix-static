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
	"github.com/t3chguy/matrix-client"
	"github.com/t3chguy/utils"
	"net/http"
	"os"
	"time"
)

const PublicRoomsPageSize = 20
const RoomTimelineSize = 20
const RoomMembersPageSize = 20

func LoadPublicRooms(first bool) {
	fmt.Println("Loading publicRooms")
	resp, err := client.PublicRooms(0, "", "")

	if err != nil {
		// Only panic if first one fails, after that we only have outdated data (less important)
		if first {
			panic(err)
		} else {
			fmt.Println(err)
		}
	}

	// Preallocate the maximum capacity possibly needed (if all rooms were world readable)
	worldReadableRooms := make([]*matrix_client.Room, 0, len(resp.Chunk))

	// filter on actually WorldReadable publicRooms
	for _, x := range resp.Chunk {
		if !x.WorldReadable {
			continue
		}

		var room *matrix_client.Room
		if existingRoom := client.GetRoom(x.RoomId); existingRoom != nil {
			room = existingRoom
		} else {
			room = client.NewRoom(x)
		}

		// Append world readable r to the filtered list.
		worldReadableRooms = append(worldReadableRooms, room)
	}
	client.SetRoomList(worldReadableRooms)
}

var client *matrix_client.Client

func main() {
	client = matrix_client.NewClient()

	router := gin.Default()
	router.SetHTMLTemplate(tpl)
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		page, skip, end := utils.CalcPaginationPage(c.DefaultQuery("page", "1"), PublicRoomsPageSize)
		c.HTML(http.StatusOK, "rooms.html", gin.H{
			"Rooms": client.GetRoomList(skip, end),
			"Page":  page,
		})
	})

	roomRouter := router.Group("/room/")
	{
		// Load room into request object so that we can do any clean up etc here
		roomRouter.Use(func(c *gin.Context) {
			roomID := c.Param("roomID")

			if room := client.GetRoom(roomID); room != nil {
				if room.LazyInitialSync() {
					c.Set("Room", room)
					c.Next()
				} else {
					c.HTML(http.StatusInternalServerError, "room_error.html", gin.H{
						"Error": "Failed to load room.",
						"Room":  room,
					})
					c.Abort()
				}
			} else {
				c.String(http.StatusNotFound, "Room Not Found")
				c.Abort()
			}
		})

		roomRouter.GET("/:roomID/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "chat")
		})

		roomRouter.GET("/:roomID/chat", func(c *gin.Context) {
			room := c.MustGet("Room").(*matrix_client.Room)
			_, forward := c.GetQuery("forward")

			pageSize := RoomTimelineSize
			anchor := c.DefaultQuery("anchor", "")
			events, nextAnchor, eventsErr := room.GetEvents(anchor, pageSize, !forward)

			if eventsErr != matrix_client.RoomEventsFine {
				var errString string
				switch eventsErr {
				case matrix_client.RoomEventsCouldNotFindEvent:
					errString = "Given up while looking for given event."
				case matrix_client.RoomEventsUnknownError:
					errString = "Unknown error encountered."
				}
				c.HTML(http.StatusInternalServerError, "room_error.html", gin.H{
					"Error": errString,
					"Room":  room,
				})
				return // Bail early
			}

			var prevPage string
			if length := len(events); length > 0 {
				prevPage = events[length-1].ID
			}

			c.HTML(http.StatusOK, "room.html", gin.H{
				"Room":   room,
				"Events": utils.ReverseEventsCopy(events),
				//"PrevPage":  prevPage,
				//"NextPage":  nextPage,
				"PageSize": pageSize,
				//"NumBefore": numBefore,
				//"NumAfter":  numAfter,
				"PrevAnchor":    prevPage,
				"CurrentAnchor": anchor,
				"NextAnchor":    nextAnchor,
			})
		})

		roomRouter.GET("/:roomID/servers", func(c *gin.Context) {
			c.HTML(http.StatusOK, "room_servers.html", gin.H{
				"Room": c.MustGet("Room").(*matrix_client.Room),
			})
		})

		roomRouter.GET("/:roomID/members", func(c *gin.Context) {
			page, skip, end := utils.CalcPaginationPage(c.DefaultQuery("page", "1"), RoomMembersPageSize)
			room := c.MustGet("Room").(*matrix_client.Room)

			c.HTML(http.StatusOK, "room_members.html", gin.H{
				"Room":       room,
				"MemberInfo": room.GetMembers()[skip:end],
				"Page":       page,
			})
		})

		roomRouter.GET("/:roomID/members/:mxid", func(c *gin.Context) {
			room := c.MustGet("Room").(*matrix_client.Room)
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

		roomRouter.GET("/:roomID/power_levels", func(c *gin.Context) {
			c.HTML(http.StatusOK, "power_levels.html", gin.H{
				"Room": c.MustGet("Room").(*matrix_client.Room),
			})
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	LoadPublicRooms(true)
	go startPublicRoomListTimer()
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

func startPublicRoomListTimer() {
	t := time.NewTicker(LoadPublicRoomsPeriod)
	for {
		<-t.C
		LoadPublicRooms(false)
	}
}
