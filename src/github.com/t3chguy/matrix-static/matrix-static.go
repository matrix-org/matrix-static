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
	"bytes"
	"flag"
	"fmt"
	"github.com/disintegration/letteravatar"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/t3chguy/go-gin-prometheus"
	"github.com/t3chguy/matrix-static/mxclient"
	"github.com/t3chguy/matrix-static/sanitizer"
	"github.com/t3chguy/matrix-static/templates"
	"github.com/t3chguy/matrix-static/utils"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

// TODO Cache memberList+serverList until it changes

const PublicRoomsPageSize = 20
const RoomTimelineSize = 30
const RoomMembersPageSize = 20

func main() {
	configPath := flag.String("config-file", "./config.json", "The path to the desired config file.")
	numWorkers := flag.Int("num-workers", 32, "Number of Worker goroutines to start.")

	publicServePrefix := flag.String("public-serve-prefix", "/", "Prefix for publicly accessible routes.")
	enablePrometheusMetrics := flag.Bool("enable-prometheus-metrics", false, "Whether or not to enable the /metrics endpoint.")
	enablePprof := flag.Bool("enable-pprof", false, "Whether or not to enable the /debug/pprof endpoints.")

	flag.Parse()

	client, err := mxclient.NewClient(*configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	worldReadableRooms := client.NewWorldReadableRooms()
	workers := NewWorkers(uint32(*numWorkers), client)
	sanitizerFn := sanitizer.InitSanitizer()

	router := gin.New()
	router.RedirectTrailingSlash = false

	if *enablePprof {
		pprof.Register(router, nil)
	}

	// This is temporary until generated server-side in Synapse as suggested by riot-web issues.
	avatarRouter := router.Group(*publicServePrefix)
	avatarRouter.Use(gin.Recovery())
	generatedAvatarCache := persistence.NewInMemoryStore(time.Hour)
	avatarRouter.GET("/avatar/:identifier", cache.CachePage(generatedAvatarCache, time.Hour, func(c *gin.Context) {
		identifier := c.Param("identifier")
		if (identifier[0] == '#' || identifier[0] == '!' || identifier[0] == '@') && len(identifier) > 1 {
			identifier = identifier[1:]
		}

		avatarChar, _ := utf8.DecodeRuneInString(identifier)
		img, err := letteravatar.Draw(100, unicode.ToUpper(avatarChar), nil)

		if err != nil {
			c.Error(err)
			return
		}

		buffer := new(bytes.Buffer)
		err = png.Encode(buffer, img)

		if err != nil {
			c.Error(err)
			return
		}

		c.Writer.Header().Set("Content-Type", "image/png")
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		_, err = c.Writer.Write(buffer.Bytes())

		//if err != nil {
		//	panic(err)
		//}
	}))

	publicRouter := router.Group(*publicServePrefix)
	publicRouter.Use(gin.Logger(), gin.Recovery())

	if *enablePrometheusMetrics {
		ginProm := ginprometheus.NewPrometheus("http")
		publicRouter.Use(ginProm.HandlerFunc())
		router.GET(ginProm.MetricsPath, ginprometheus.PrometheusHandler())
	}

	publicRouter.Static("/img", "./assets/img")
	publicRouter.Static("/css", "./assets/css")

	publicRouter.GET("/", func(c *gin.Context) {
		page := utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1)
		templates.WritePageTemplate(c.Writer, &templates.RoomsPage{
			Rooms:    worldReadableRooms.GetPage(page, PublicRoomsPageSize),
			PageSize: PublicRoomsPageSize,
			Page:     page,
		})
	})

	roomRouter := publicRouter.Group("/room/:roomID/")
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

			templates.WritePageTemplate(c.Writer, &templates.RoomChatPage{
				RoomInfo:      jobResult.RoomInfo,
				MemberMap:     jobResult.MemberMap,
				Events:        events,
				PageSize:      RoomTimelineSize,
				CurrentOffset: offset,
				Anchor:        eventID,

				AtTopEnd:    jobResult.AtTopEnd,
				AtBottomEnd: jobResult.AtBottomEnd,

				Sanitizer:         sanitizerFn,
				HomeserverBaseURL: client.HomeserverURL.String(),
			})
		})

		const RoomServersPageSize = 30

		roomRouter.GET("/servers", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			worker.Queue <- RoomServersJob{
				c.Param("roomID"),
				utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1),
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

		const RoomAliasesPageSize = 10

		roomRouter.GET("/aliases", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			worker.Queue <- RoomAliasesJob{
				c.Param("roomID"),
				utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1),
				RoomAliasesPageSize,
			}

			jobResult := templates.RoomAliasesPage((<-worker.Output).(RoomAliasesResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})

		roomRouter.GET("/members", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(Worker)
			worker.Queue <- RoomMembersJob{
				c.Param("roomID"),
				utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1),
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
