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

package matrixClient

import (
	"fmt"
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/utils"
	"strings"
	"sync"
	"time"
)

type PowerLevel int

func (powerLevel PowerLevel) String() string {
	switch int(powerLevel) {
	case 100:
		return "Admin"
	case 50:
		return "Moderator"
	case 0:
		return "User"
	case -1:
		return "Muted"
	default:
		return "Custom"
	}
}

func (powerLevel PowerLevel) ToInt() int {
	return int(powerLevel)
}

type Room struct {
	client *Client // each room has a client that is responsible for its state being up to date

	sync.RWMutex

	ID string // IMMUTABLE

	backPaginationToken    string
	forwardPaginationToken string
	lastForwardPagination  time.Time

	eventList []gomatrix.Event
	//eventMap        map[string]*gomatrix.Event
	latestRoomState RoomState // IMMUTABLE

	hasInitialSynced bool
}

func (r *Room) concatBackpagination(oldEvents []gomatrix.Event, newToken *string) {
	r.Lock() //deadlock
	defer r.Unlock()

	//for _, event := range oldEvents {
	//	r.eventMap[event.ID] = &event
	//}
	r.eventList = append(r.eventList, oldEvents...)

	if newToken != nil {
		r.backPaginationToken = *newToken
	}

}

func (r *Room) GetTokens() (string, string) {
	r.RLock()
	defer r.RUnlock()
	return r.backPaginationToken, r.forwardPaginationToken
}

func (r *Room) GetMemberList(start int, end int) []*MemberInfo {
	length := r.GetNumMembers()

	if end == 0 {
		return r.latestRoomState.memberList[utils.Min(start, length):]
	}

	return r.latestRoomState.memberList[utils.Min(start, length):utils.Min(end, length)]
}

func (r *Room) GetMemberIgnore(mxid string) (memberInfo MemberInfo) {
	memberInfo, _ = r.GetMember(mxid)
	return
}

func (r *Room) GetMember(mxid string) (memberInfo MemberInfo, exists bool) {
	if memberInfoPointer := r.latestRoomState.memberMap[mxid]; memberInfoPointer != nil {
		return *memberInfoPointer, true
	}
	return MemberInfo{}, false
}

func (r *Room) GetNumMembers() int {
	return len(r.latestRoomState.memberList)
}

func (r *Room) findEventIndex(anchor string, backpaginate bool) (int, bool) {
	r.RLock()
	for index, event := range r.eventList {
		if event.ID == anchor {
			defer r.RUnlock()
			return index, true
		}
	}
	r.RUnlock()

	if backpaginate {
		if numNew := r.client.backpaginateRoom(r, 100); numNew > 0 {
			return r.findEventIndex(anchor, false)
		}
	}
	return 0, false
}

const overcompensatePaginationQuantity = 5

func (r *Room) getBackwardEventRange(index, number int) []gomatrix.Event {
	r.RLock()
	length := len(r.eventList)
	r.RUnlock()

	var numNew int
	if delta := index + number + overcompensatePaginationQuantity; delta >= length {
		numNew = r.client.backpaginateRoom(r, -delta)
	}
	length += numNew

	eventList := r.eventList[index:utils.Min(index+number+1, length)]

	fmt.Println(eventList)

	return eventList
}

func (r *Room) getForwardEventRange(index, number int) []gomatrix.Event {
	r.RLock()
	defer r.RUnlock()
	length := len(r.eventList)

	oldestIndex := utils.FixRange(0, index+number, length)
	latestIndex := utils.Min(index, oldestIndex)
	return r.eventList[latestIndex:oldestIndex]
}

func (r *Room) GetEvents(anchor string, amount int, towardsHistory bool) (events []gomatrix.Event, nextAnchor string) {
	// normal (towards history): after=X - returns X+N including X
	// return (towards present): before=X - returnx X-n not including X

	// 0 is the LATEST Event
	// Len-1 is the OLDEST event

	if amount <= 0 {
		return []gomatrix.Event{}, anchor
	}

	var eventIndex int

	if anchor != "" {
		if index, found := r.findEventIndex(anchor, true); found {
			eventIndex = index
		} else {
			return []gomatrix.Event{}, anchor
		}
	}

	if towardsHistory {
		eventsO := r.getBackwardEventRange(eventIndex, amount)

		var nextHistorical gomatrix.Event
		lastEventIndex := len(eventsO) - 1

		if lastEventIndex >= amount {
			eventsO, nextHistorical = eventsO[:lastEventIndex], eventsO[lastEventIndex]
		} else {
			nextHistorical = eventsO[lastEventIndex]
		}

		return eventsO, nextHistorical.ID
	} else {
		return []gomatrix.Event{}, anchor
	}
}

const RoomInitialSyncLimit = 100

func (m *Client) NewRoom(publicRoomInfo gomatrix.PublicRoomsChunk) *Room {
	newRoom := &Room{
		client: m,
		ID:     publicRoomInfo.RoomId,
		latestRoomState: RoomState{
			Name:             publicRoomInfo.Name,
			topic:            publicRoomInfo.Topic,
			AvatarURL:        MXCURL(publicRoomInfo.AvatarUrl),
			CanonicalAlias:   publicRoomInfo.CanonicalAlias,
			Aliases:          publicRoomInfo.Aliases,
			memberMap:        make(map[string]*MemberInfo),
			NumJoinedMembers: publicRoomInfo.NumJoinedMembers,
		},
		lastForwardPagination: time.Now(),
		//eventMap: make(map[string]*gomatrix.Event),
		//eventList: make([]*gomatrix.Event, 0, RoomInitialSyncLimit),
	}

	//newRoom.latestRoomState.CalculateMemberList()
	return newRoom
}

func (r *Room) LazyInitialSync() {
	r.RLock()
	hasInitialSynced := r.hasInitialSynced
	r.RUnlock()

	if hasInitialSynced {
		return
	}

	resp, err := r.client.RoomInitialSync(r.ID, RoomInitialSyncLimit)

	if err != nil {
		panic(err)
	}

	r.Lock()
	defer r.Unlock()

	for _, event := range resp.State {
		r.latestRoomState.UpdateOnEvent(&event)
	}

	r.backPaginationToken = resp.Messages.Start
	r.forwardPaginationToken = resp.Messages.End

	utils.ReverseEvents(resp.Messages.Chunk)
	r.eventList = resp.Messages.Chunk
	r.hasInitialSynced = true
}

// Partial implementation of http://matrix.org/docs/spec/client_server/r0.2.0.html#calculating-the-display-name-for-a-room
// falling back to room ID instead of "Empty Room" (though this should not be possible with guest-world-readble-rooms)
func (r *Room) GetName() string {
	r.RLock()
	defer r.RUnlock()

	if r.latestRoomState.Name != "" {
		return r.latestRoomState.Name
	}
	if r.latestRoomState.CanonicalAlias != "" {
		return r.latestRoomState.CanonicalAlias
	}
	if len(r.latestRoomState.Aliases) > 0 {
		return r.latestRoomState.Aliases[0]
	}

	return r.ID
}

func (r *Room) NumJoinedMembers() int {
	r.RLock()
	defer r.RUnlock()
	return r.latestRoomState.NumJoinedMembers
}

func (r *Room) CanonicalAlias() string {
	r.RLock()
	defer r.RUnlock()

	if r.latestRoomState.CanonicalAlias != "" {
		return r.latestRoomState.CanonicalAlias
	}
	if len(r.latestRoomState.Aliases) > 0 {
		return r.latestRoomState.Aliases[0]
	}
	return ""
}
func (r *Room) AvatarUrl() MXCURL {
	r.RLock()
	defer r.RUnlock()
	return r.latestRoomState.AvatarURL
}
func (r *Room) Topic() string {
	r.RLock()
	defer r.RUnlock()
	return r.latestRoomState.topic
}

func (r *Room) GetServers() map[string]int {
	r.RLock()
	defer r.RUnlock()

	serverMap := make(map[string]int)
	for _, member := range r.latestRoomState.memberList {
		if mxidSplit := strings.SplitN(member.MXID, ":", 2); len(mxidSplit) == 2 {
			serverMap[mxidSplit[1]]++
		}
	}
	return serverMap
}
