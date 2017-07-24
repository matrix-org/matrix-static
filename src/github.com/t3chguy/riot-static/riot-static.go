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
	"github.com/t3chguy/riot-static/mxclient"
	"github.com/t3chguy/riot-static/sanitizer"
	"github.com/t3chguy/riot-static/templates"
	"github.com/t3chguy/riot-static/utils"
	"net/http"
	"os"
	"time"
)

// TODO Cache memberList+serverList until it changes

const PublicRoomsPageSize = 20
const RoomTimelineSize = 30
const RoomMembersPageSize = 20

const NumWorkers uint32 = 32

func main() {
	client := mxclient.NewClient()
	worldReadableRooms := client.NewWorldReadableRooms()

	workers := NewWorkers(NumWorkers, client)
	sanitizerFn := sanitizer.InitSanitizer()

	router := gin.Default()
	router.Static("/img", "./assets/img")
	router.Static("/css", "./assets/css")

	router.GET("/", func(c *gin.Context) {
		page := utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1)
		templates.WritePageTemplate(c.Writer, &templates.RoomsPage{
			Rooms:    worldReadableRooms.GetPage(page, PublicRoomsPageSize),
			PageSize: PublicRoomsPageSize,
			Page:     page,
		})
	})

	roomRouter := router.Group("/room/:roomID/")
	{
		// Load room worker into request object so that we can do any clean up etc here
		roomRouter.Use(func(c *gin.Context) {
			roomID := c.Param("roomID")
			worker := workers.GetWorkerForRoomID(roomID)

			worker.Queue <- &RoomInitialSyncJob{roomID}
			resp := (<-worker.Output).(*RoomInitialSyncResp)

			if resp.err != nil {
				c.String(http.StatusNotFound, "Room Not Found")
				c.Abort()
				return
			}

			c.Set("RoomWorker", worker)
			c.Next()

			//	c.HTML(http.StatusInternalServerError, "room_error.html", gin.H{
			//		"Error": "Failed to load room.",
			//		"Room":  room,
			//	})
		})

		roomRouter.GET("/", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			offset := utils.StrToIntDefault(c.DefaultQuery("offset", "0"), 0)
			eventID := c.DefaultQuery("anchor", "")

			worker.Queue <- Job(RoomEventsJob{
				c.Param("roomID"),
				eventID,
				offset,
				RoomTimelineSize,
			})

			jobResult := (<-worker.Output).(RoomEventsResp)
			if jobResult.err != nil {
				templates.WritePageTemplate(c.Writer, &templates.RoomErrorPage{
					Error:    "Some error has occurred",
					RoomInfo: jobResult.RoomInfo,
				})
				return
			}

			numEvents := len(jobResult.Events)

			if eventID == "" && numEvents > 0 {
				eventID = jobResult.Events[0].ID
			}

			events := mxclient.ReverseEventsCopy(jobResult.Events)

			var reachedRoomCreate bool
			if numEvents > 0 {
				reachedRoomCreate = events[0].Type == "m.room.create" && *events[0].StateKey == ""
			}

			templates.WritePageTemplate(c.Writer, &templates.RoomChatPage{
				RoomInfo:          jobResult.RoomInfo,
				MemberMap:         jobResult.MemberMap,
				Events:            events,
				PageSize:          RoomTimelineSize,
				ReachedRoomCreate: reachedRoomCreate,
				CurrentOffset:     offset,
				Anchor:            eventID,

				Sanitizer:         sanitizerFn,
				HomeserverBaseURL: client.HomeserverURL.String(),
			})
		})

		const RoomServersPageSize = 30

		roomRouter.GET("/servers", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			page := utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1)

			worker.Queue <- RoomServersJob{
				c.Param("roomID"),
				page,
				RoomServersPageSize,
			}

			jobResult := templates.RoomServersPage((<-worker.Output).(RoomServersResp))
			templates.WritePageTemplate(c.Writer, &jobResult)

			/*
				templates.WritePageTemplate(c.Writer, &worker.RoomServers(RoomServersJob{
					c.Param("roomID"),
					page,
					RoomServersPageSize,
				}))
			*/
		})

		roomRouter.GET("/members", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			page := utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1)

			worker.Queue <- RoomMembersJob{
				c.Param("roomID"),
				page,
				RoomMembersPageSize,
			}

			jobResult := templates.RoomMembersPage((<-worker.Output).(RoomMembersResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})

		roomRouter.GET("/members/:mxid", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			worker.Queue <- RoomMemberInfoJob{
				c.Param("roomID"),
				c.Param("mxid"),
			}

			//c.AbortWithStatus(http.StatusNotFound)

			jobResult := templates.RoomMemberInfoPage((<-worker.Output).(RoomMemberInfoResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})

		roomRouter.GET("/power_levels", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			worker.Queue <- RoomPowerLevelsJob{c.Param("roomID")}

			jobResult := templates.RoomPowerLevelsPage((<-worker.Output).(RoomPowerLevelsResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	go startForwardPaginator(workers)
	go startPublicRoomListTimer(worldReadableRooms)
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

func startPublicRoomListTimer(worldReadableRooms *mxclient.WorldReadableRooms) {
	t := time.NewTicker(LoadPublicRoomsPeriod)
	for {
		<-t.C
		worldReadableRooms.Update()
	}
}

const LazyForwardPaginateRooms = time.Minute

func startForwardPaginator(workers *Workers) {
	t := time.NewTicker(LazyForwardPaginateRooms)
	for {
		<-t.C
		workers.JobForAllWorkers(RoomForwardPaginateJob{})
	}
}
