package workers

import (
	"bytes"
	"github.com/matrix-org/matrix-static/mxclient"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"
)

var count = 0

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func makeWorker(rooms map[string]*mxclient.Room) *Worker {
	worker := &Worker{
		ID:     count,
		client: nil,
		Queue:  make(chan Job),
		Output: make(chan JobResp),
		rooms:  rooms,
	}
	go worker.Start()
	count += 1
	return worker
}

func TestRoomForwardPaginateJob_Work(t *testing.T) {
	cli, _ := mxclient.NewRawClient("", "", "", "")
	cli.Client.Client.Transport = RoundTripFunc(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	room1 := &mxclient.Room{Client: cli, ID: "room1", LastAccess: time.Now().Add(-10 * time.Minute)}
	room2 := &mxclient.Room{Client: cli, ID: "room2", LastAccess: time.Now().Add(-11 * time.Minute)}
	room3 := &mxclient.Room{Client: cli, ID: "room3", LastAccess: time.Now().Add(-12 * time.Minute)}
	room4 := &mxclient.Room{Client: cli, ID: "room4", LastAccess: time.Now().Add(-13 * time.Minute)}

	tests := []struct {
		name   string
		fields RoomForwardPaginateJob
		rooms  map[string]*mxclient.Room
		exp    map[string]*mxclient.Room
	}{
		{
			"should not remove any if less rooms than KeepMin",
			RoomForwardPaginateJob{
				&sync.WaitGroup{},
				100 * time.Minute,
				10,
			},
			map[string]*mxclient.Room{
				"room1": room1,
				"room2": room2,
				"room3": room3,
				"room4": room4,
			},
			map[string]*mxclient.Room{
				"room1": room1,
				"room2": room2,
				"room3": room3,
				"room4": room4,
			},
		}, {
			"should not remove any if less rooms than KeepMin even if exceed TTL",
			RoomForwardPaginateJob{
				&sync.WaitGroup{},
				10 * time.Minute,
				10,
			},
			map[string]*mxclient.Room{
				"room1": room1,
				"room2": room2,
				"room3": room3,
				"room4": room4,
			},
			map[string]*mxclient.Room{
				"room1": room1,
				"room2": room2,
				"room3": room3,
				"room4": room4,
			},
		}, {
			"should remove any exceeding TTL but keeping KeepMin",
			RoomForwardPaginateJob{
				&sync.WaitGroup{},
				10 * time.Minute,
				3,
			},
			map[string]*mxclient.Room{
				"room1": room1,
				"room2": room2,
				"room3": room3,
				"room4": room4,
			},
			map[string]*mxclient.Room{
				"room1": room1,
				"room2": room2,
				"room3": room3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.Wg.Add(1)
			job := RoomForwardPaginateJob{
				Wg:      tt.fields.Wg,
				TTL:     tt.fields.TTL,
				KeepMin: tt.fields.KeepMin,
			}
			w := makeWorker(tt.rooms)
			job.Work(w)
			job.Wg.Wait()

			if !reflect.DeepEqual(w.rooms, tt.exp) {
				t.Error("Rooms mismatch expectation", w.rooms, tt.exp)
			}
		})
	}
}
