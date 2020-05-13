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
	log "github.com/Sirupsen/logrus"
	"github.com/disintegration/letteravatar"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/matrix-org/dugong"
	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/matrix-static/mxclient"
	"github.com/matrix-org/matrix-static/sanitizer"
	"github.com/matrix-org/matrix-static/templates"
	"github.com/matrix-org/matrix-static/utils"
	"github.com/matrix-org/matrix-static/workers"
	"github.com/t3chguy/go-gin-prometheus"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const PublicRoomsPageSize = 20
const RoomTimelineSize = 30
const RoomMembersPageSize = 20

type configVars struct {
	ConfigFile string
	NumWorkers int

	PublicServePrefix       string
	EnablePrometheusMetrics bool
	EnablePprof             bool

	LastAccessDiscardDuration time.Duration
	KeepAtLeastNRooms         int

	LogDir string
}

func main() {
	// startup checks
	if stat, err := os.Stat("./assets"); os.IsNotExist(err) || !stat.IsDir() {
		log.WithError(err).Error("./assets/ directory is not accessible")
		return
	}

	config := configVars{}

	flag.StringVar(&config.ConfigFile, "config-file", "./config.json", "The path to the desired config file.")
	flag.IntVar(&config.NumWorkers, "num-workers", 32, "Number of Worker goroutines to start.")

	flag.StringVar(&config.PublicServePrefix, "public-serve-prefix", "/", "Prefix for publicly accessible routes.")
	flag.BoolVar(&config.EnablePrometheusMetrics, "enable-prometheus-metrics", false, "Whether or not to enable the /metrics endpoint.")
	flag.BoolVar(&config.EnablePprof, "enable-pprof", false, "Whether or not to enable the /debug/pprof endpoints.")
	flag.StringVar(&config.LogDir, "logger-directory", "", "Where to write the info, warn and error logs to.")

	flag.DurationVar(&config.LastAccessDiscardDuration, "cache-ttl", 30*time.Minute, "")
	flag.IntVar(&config.KeepAtLeastNRooms, "cache-min-rooms", 10, "")

	flag.Parse()

	if config.LogDir != "" {
		log.AddHook(dugong.NewFSHook(
			filepath.Join(config.LogDir, "info.log"),
			filepath.Join(config.LogDir, "warn.log"),
			filepath.Join(config.LogDir, "error.log"),
			&log.TextFormatter{
				TimestampFormat:  "2006-01-02 15:04:05.000000",
				DisableColors:    true,
				DisableTimestamp: false,
				DisableSorting:   false,
			}, &dugong.DailyRotationSchedule{GZip: false},
		))
	}

	log.Infof("Matrix-Static (%+v)", config)

	client, err := mxclient.NewClient(config.ConfigFile)
	if err != nil {
		log.WithError(err).Error("Unable to start new Client")
		return
	}

	worldReadableRooms := client.NewWorldReadableRooms()
	pool := workers.NewWorkers(uint32(config.NumWorkers), client)
	sanitizerFn := sanitizer.InitSanitizer()

	router := gin.New()
	router.RedirectTrailingSlash = false

	if config.EnablePprof {
		pprof.Register(router, nil)
	}

	// This is temporary until generated server-side in Synapse as suggested by riot-web issues.
	avatarRouter := router.Group(config.PublicServePrefix)
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

		if err != nil {
			log.WithError(err).Error("Failed to write Image Buffer out.")
		}
	}))

	publicRouter := router.Group(config.PublicServePrefix)
	publicRouter.Use(gin.Logger(), gin.Recovery())

	if config.EnablePrometheusMetrics {
		ginProm := ginprometheus.NewPrometheus("http")
		publicRouter.Use(ginProm.HandlerFunc())
		router.GET(ginProm.MetricsPath, ginprometheus.PrometheusHandler())
	}

	publicRouter.Static("/img", "./assets/img")
	publicRouter.Static("/css", "./assets/css")
	publicRouter.StaticFile("/robots.txt", "./assets/robots.txt")

	publicRouter.GET("/", func(c *gin.Context) {
		page := utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1)
		templates.WritePageTemplate(c.Writer, &templates.RoomsPage{
			Rooms:    worldReadableRooms.GetPage(page, PublicRoomsPageSize),
			PageSize: PublicRoomsPageSize,
			Page:     page,
		})
	})

	roomAliasCache := persistence.NewInMemoryStore(time.Hour)
	publicRouter.GET("/alias/:roomAlias", cache.CachePage(roomAliasCache, time.Hour, func(c *gin.Context) {
		roomAlias := c.Param("roomAlias")
		resp, err := client.GetRoomDirectoryAlias(roomAlias)

		// TODO better error page
		if err != nil || resp.RoomID == "" {
			templates.WritePageTemplate(c.Writer, &templates.ErrorPage{
				ErrType: "Unable to resolve Room Alias.",
				Error:   err,
			})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, "/room/"+resp.RoomID+"/")
	}))

	roomRouter := publicRouter.Group("/room/:roomID/")
	{
		const permalinkOffset = 10

		roomRouter.GET("/$:eventID", func(c *gin.Context) {
			eventID := c.Param("eventID")
			roomID := c.Param("roomID")

			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/room/%s/?anchor=$%s&offset=-%d&highlight=$%s", roomID, eventID, permalinkOffset, eventID))
		})

		// Load room worker into request object so that we can do any clean up etc here
		roomRouter.Use(func(c *gin.Context) {
			roomID := c.Param("roomID")

			if roomID[0] != '!' {
				templates.WritePageTemplate(c.Writer, &templates.ErrorPage{
					ErrType: "Unable to Load Room.",
					Details: "Room ID must start with a '!'",
				})
				c.Abort()
				return
			}

			worker := pool.GetWorkerForRoomID(roomID)

			worker.Queue <- &workers.RoomInitialSyncJob{RoomID: roomID}
			resp := (<-worker.Output).(*workers.RoomInitialSyncResp)

			if resp.Err != nil {
				defer c.Abort()

				if respErr, ok := mxclient.UnwrapRespError(resp.Err); ok {
					templates.WritePageTemplate(c.Writer, &templates.ErrorPage{
						ErrType: "Unable to Join Room.",
						Details: mxclient.TextForRespError(respErr),
					})
					return
				}

				if err, ok := resp.Err.(gomatrix.HTTPError); ok {
					templates.WritePageTemplate(c.Writer, &templates.ErrorPage{
						ErrType: "Cannot Load Room.",
						Details: err.Message,
					})
					return
				}

				templates.WritePageTemplate(c.Writer, &templates.ErrorPage{
					ErrType: "Cannot Load Room. Internal Server Error.",
					Error:   resp.Err,
				})

				return
			}

			c.Set("RoomWorker", worker)
			c.Next()
		})

		roomRouter.GET("/", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(workers.Worker)
			offset := utils.StrToIntDefault(c.DefaultQuery("offset", "0"), 0)
			eventID := c.Query("anchor")

			worker.Queue <- workers.Job(workers.RoomEventsJob{
				RoomID:   c.Param("roomID"),
				Anchor:   eventID,
				Offset:   offset,
				PageSize: RoomTimelineSize,
			})

			jobResult := (<-worker.Output).(workers.RoomEventsResp)
			if jobResult.Err != nil {
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
			highlight := c.Query("highlight")

			templates.WritePageTemplate(c.Writer, &templates.RoomChatPage{
				RoomInfo:      jobResult.RoomInfo,
				MemberMap:     jobResult.MemberMap,
				Events:        events,
				PageSize:      RoomTimelineSize,
				CurrentOffset: offset,
				Anchor:        eventID,

				AtTopEnd:    jobResult.AtTopEnd,
				AtBottomEnd: jobResult.AtBottomEnd,

				Sanitizer:    sanitizerFn,
				MediaBaseURL: client.MediaBaseURL,
				Highlight:    highlight,
			})
		})

		const RoomServersPageSize = 30

		roomRouter.GET("/servers", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(workers.Worker)
			worker.Queue <- workers.RoomServersJob{
				RoomID:   c.Param("roomID"),
				Page:     utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1),
				PageSize: RoomServersPageSize,
			}

			jobResult := templates.RoomServersPage((<-worker.Output).(workers.RoomServersResp))
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
			worker := c.MustGet("RoomWorker").(workers.Worker)
			worker.Queue <- workers.RoomAliasesJob{
				RoomID:   c.Param("roomID"),
				Page:     utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1),
				PageSize: RoomAliasesPageSize,
			}

			jobResult := templates.RoomAliasesPage((<-worker.Output).(workers.RoomAliasesResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})

		roomRouter.GET("/members", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(workers.Worker)
			worker.Queue <- workers.RoomMembersJob{
				RoomID:   c.Param("roomID"),
				Page:     utils.StrToIntDefault(c.DefaultQuery("page", "1"), 1),
				PageSize: RoomMembersPageSize,
			}

			jobResult := templates.RoomMembersPage((<-worker.Output).(workers.RoomMembersResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})

		roomRouter.GET("/members/:mxid", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(workers.Worker)
			worker.Queue <- workers.RoomMemberInfoJob{
				RoomID: c.Param("roomID"),
				Mxid:   c.Param("mxid"),
			}

			//c.AbortWithStatus(http.StatusNotFound)

			jobResult := templates.RoomMemberInfoPage((<-worker.Output).(workers.RoomMemberInfoResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})

		roomRouter.GET("/power_levels", func(c *gin.Context) {
			worker := c.MustGet("RoomWorker").(workers.Worker)
			worker.Queue <- workers.RoomPowerLevelsJob{RoomID: c.Param("roomID")}

			jobResult := templates.RoomPowerLevelsPage((<-worker.Output).(workers.RoomPowerLevelsResp))
			templates.WritePageTemplate(c.Writer, &jobResult)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	go startForwardPaginator(config, pool)
	go startPublicRoomListTimer(worldReadableRooms)
	log.Info("Listening on port " + port)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
		Addr:         ":" + port,
	}

	log.Fatal(srv.ListenAndServe())
}

const LoadPublicRoomsPeriod = time.Hour

func startPublicRoomListTimer(worldReadableRooms *mxclient.WorldReadableRooms) {
	t := time.NewTicker(LoadPublicRoomsPeriod)
	for {
		<-t.C
		log.Info("Reloading public room list")
		worldReadableRooms.Update()
	}
}

const LazyForwardPaginateRooms = 2 * time.Minute

func startForwardPaginator(config configVars, pool *workers.Workers) {
	//t := time.NewTicker(LazyForwardPaginateRooms)
	wg := sync.WaitGroup{}
	for {
		//<-t.C
		time.Sleep(LazyForwardPaginateRooms)
		wg.Add(int(pool.NumWorkers))
		log.Info("Forward paginating all loaded rooms")
		pool.JobForAllWorkers(workers.RoomForwardPaginateJob{
			Wg:      &wg,
			TTL:     config.LastAccessDiscardDuration,
			KeepMin: config.KeepAtLeastNRooms,
		})
		wg.Wait()
	}
}
