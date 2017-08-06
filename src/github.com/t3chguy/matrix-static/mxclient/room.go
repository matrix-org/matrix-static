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

package mxclient

import (
	"errors"
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/matrix-static/utils"
)

type RoomInfo struct {
	RoomID          string
	Name            string
	CanonicalAlias  string
	Topic           string
	AvatarURL       MXCURL
	NumMemberEvents int
	NumMembers      int
	NumServers      int
}

type Room struct {
	// each room has a client that is responsible for its state being up to date
	client *Client

	ID string

	backPaginationToken    string
	forwardPaginationToken string

	// eventList[0] is the latest event we know
	eventList []gomatrix.Event
	//eventMap        map[string]*gomatrix.Event
	latestRoomState RoomState

	HasReachedHistoricEndOfTimeline bool
}

// ForwardPaginateRoom queries the API for any events newer than the latest one currently in the timeline and appends them.
func (r *Room) ForwardPaginateRoom() {
	r.client.forwardpaginateRoom(r, 0)
}

func (r *Room) concatBackpagination(oldEvents []gomatrix.Event, newToken string) {
	for _, event := range oldEvents {
		if ShouldHideEvent(event) {
			continue
		}
		//if event.Type == "m.room.redaction" {
		// The server has already handled these for us
		// so just consume them to prevent them blanking on timeline
		//continue
		//}

		r.eventList = append(r.eventList, event)
	}
	r.backPaginationToken = newToken
	r.latestRoomState.RecalculateMemberListAndServers()
}

func (r *Room) concatForwardPagination(newEvents []gomatrix.Event, newToken string) {
	for _, event := range newEvents {
		// TODO Handle redaction and skip adding to TL
		//if event.Type == "m.room.redaction" {
		// Might want an Event Map->*Event so we can skip an O(n) task
		//}

		if ShouldHideEvent(event) {
			continue
		}

		r.latestRoomState.UpdateOnEvent(&event, false)
		r.eventList = append([]gomatrix.Event{event}, r.eventList...)
	}
	r.forwardPaginationToken = newToken
	r.latestRoomState.RecalculateMemberListAndServers()
}

func (r *Room) findEventIndex(anchor string, backpaginate bool) (int, bool) {
	for index, event := range r.eventList {
		if event.ID == anchor {
			return index, true
		}
	}

	if backpaginate {
		if numNew, _ := r.client.backpaginateRoom(r, 100); numNew > 0 {
			return r.findEventIndex(anchor, false)
		}
	}
	return 0, false
}

// overcompenesatePaginationBy, number to try and keep as a buffer at the end of our in-memory timeline so we don't
// backpaginate on every single call.
const overcompensateBackpaginationBy = 32

func (r *Room) backpaginateIfNeeded(anchorIndex, offset, number int) {
	if r.HasReachedHistoricEndOfTimeline {
		return
	}

	// delta is the number of events we should have, to comfortably handle this request, if we do not have this many
	// then ask the mxclient to backpaginate this room by at least delta-length events.
	// TODO if numNew = 0, we are at end of TL as we know it, mark this room as such.
	length := len(r.eventList)
	if delta := anchorIndex + offset + number + overcompensateBackpaginationBy; delta >= length {
		// if no error encountered and zero events then we are likely at the last historical event.
		if numNew, err := r.client.backpaginateRoom(r, delta-length); err == nil {
			if numNew == 0 {
				r.HasReachedHistoricEndOfTimeline = true
			}
		}
	}
}

func (r *Room) getBackwardEventRange(anchorIndex, offset, number int) []gomatrix.Event {
	r.backpaginateIfNeeded(anchorIndex, offset, number)

	length := len(r.eventList)
	startIndex := utils.Min(anchorIndex+offset, length)
	return r.eventList[startIndex:utils.Min(startIndex+number, length)]
}

func (r *Room) getForwardEventRange(index, offset, number int) []gomatrix.Event {
	topIndex := utils.Bound(0, index+number-offset, len(r.eventList))
	return r.eventList[utils.Max(topIndex-number, 0):topIndex]
}

// GetState returns an instance of RoomState believed to represent the current state of the room.
func (r *Room) GetState() RoomState {
	return r.latestRoomState
}

// GetEventPage returns a paginated slice of events, as well as whether this slice rests at either/both ends of the timeline.
func (r *Room) GetEventPage(anchor string, offset int, pageSize int) (events []gomatrix.Event, atTopEnd, atBottomEnd bool, err error) {
	var anchorIndex int
	if anchor != "" {
		if index, found := r.findEventIndex(anchor, false); found {
			anchorIndex = index
		} else {
			err = errors.New("Could not find event")
			return
		}
	}

	if offset >= 0 {
		events = r.getBackwardEventRange(anchorIndex, offset, pageSize)
	} else {
		events = r.getForwardEventRange(anchorIndex, -offset, pageSize)
	}

	// Consider ourselves at end if the ID matches the respective end of the stored event list.
	numEvents, totalNumEvents := len(events), len(r.eventList)
	if numEvents > 0 {
		atTopEnd = events[numEvents-1].ID == r.eventList[totalNumEvents-1].ID
		atBottomEnd = events[0].ID == r.eventList[0].ID
	}
	return
}

const RoomInitialSyncLimit = 256

// NewRoom fetches :roomId/initialSync for a room and instantiates a room to represent it.
func (m *Client) NewRoom(roomID string) (*Room, error) {
	resp, err := m.RoomInitialSync(roomID, RoomInitialSyncLimit)

	if err != nil {
		return nil, err
	}

	// filter out m.room.redactions and reverse ordering at once.
	var filteredEventList []gomatrix.Event
	for _, event := range resp.Messages.Chunk {
		if ShouldHideEvent(event) {
			continue
		}

		filteredEventList = append([]gomatrix.Event{event}, filteredEventList...)
	}

	newRoom := &Room{
		client: m,
		ID:     roomID,
		forwardPaginationToken: resp.Messages.End,
		backPaginationToken:    resp.Messages.Start,
		eventList:              filteredEventList,
		latestRoomState:        *NewRoomState(m),
	}

	for _, event := range resp.State {
		newRoom.latestRoomState.UpdateOnEvent(&event, true)
	}

	newRoom.latestRoomState.RecalculateMemberListAndServers()

	return newRoom, nil
}

// RoomInfo summates basic currentState parameters
func (r *Room) RoomInfo() RoomInfo {
	return RoomInfo{
		r.ID,
		r.latestRoomState.CalculateName(),
		r.latestRoomState.canonicalAlias,
		r.latestRoomState.Topic,
		r.latestRoomState.AvatarURL,
		r.latestRoomState.GetNumMemberEvents(),
		r.latestRoomState.NumMembers(),
		len(r.latestRoomState.Servers()),
	}
}
