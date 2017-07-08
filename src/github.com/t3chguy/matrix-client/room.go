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

package matrix_client

import (
	"fmt"
	"github.com/matrix-org/gomatrix"
	"github.com/t3chguy/utils"
	"sort"
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

type RoomEventErrorEnum int

const (
	RoomEventsCouldNotFindEvent RoomEventErrorEnum = iota
	RoomEventsUnknownError
	RoomEventsFine
)

type Room struct {
	client *Client // each room has a client that is responsible for its state being up to date

	// Active lock is called for the duration of a Read/Write on any field of the room.
	// This lock may be Locked/Unlocked several times in a request, so may be hit by inconsistent data,
	// for this reason requests which may move the anchor point of our eventList/latestRoomState Lock requestLock (TOO?)
	activeLock sync.RWMutex

	// Request lock is called for the duration of a request on the room object
	// To prevent the CRON forward pagination from changing the 0 point of a timeline (and mangling latestRoomState)
	requestLock sync.RWMutex

	ID string // IMMUTABLE

	backPaginationToken    string
	forwardPaginationToken string
	lastForwardPagination  time.Time

	eventList []gomatrix.Event
	//eventMap        map[string]*gomatrix.Event
	latestRoomState RoomState // IMMUTABLE

	hasInitialSynced bool
}

func (r *Room) concatBackpagination(oldEvents []gomatrix.Event, newToken string) {
	r.activeLock.Lock()
	defer r.activeLock.Unlock()

	fmt.Println("concatBackpagination", len(oldEvents), newToken)

	//for _, event := range oldEvents {
	//fmt.Println(event)
	//}
	r.eventList = append(r.eventList, oldEvents...)
	r.backPaginationToken = newToken
}

func (r *Room) concatForwardpagination(newEvents []gomatrix.Event, newToken string) {
	r.activeLock.Lock()
	defer r.activeLock.Unlock()

	//for _, event := range newEvents {
	//fmt.Println(event)
	//}
	//r.eventList = ...
	r.forwardPaginationToken = newToken
}

func (r *Room) GetTokens() (string, string) {
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()
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
	r.activeLock.RLock()
	for index, event := range r.eventList {
		if event.ID == anchor {
			defer r.activeLock.RUnlock()
			return index, true
		}
	}
	r.activeLock.RUnlock()

	if backpaginate {
		if numNew := r.client.backpaginateRoom(r, 100); numNew > 0 {
			return r.findEventIndex(anchor, false)
		}
	}
	return 0, false
}

const overcompensatePaginationQuantity = 32

func (r *Room) getBackwardEventRange(index, offset, number int) []gomatrix.Event {
	r.activeLock.RLock()
	length := len(r.eventList)
	r.activeLock.RUnlock()

	if delta := index + offset + number + overcompensatePaginationQuantity; delta >= length {
		length += r.client.backpaginateRoom(r, delta-length)
	}
	index = utils.Min(index+offset, length)

	r.activeLock.RLock()
	defer r.activeLock.RUnlock()

	return r.eventList[index:utils.Min(index+number, length)]
}

func (r *Room) getForwardEventRange(index, offset, number int) []gomatrix.Event {
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()

	length := len(r.eventList)
	topIndex := utils.FixRange(0, index+number-offset, length)

	return r.eventList[utils.Max(topIndex-number, 0):topIndex]
}

//func (r *Room) GetEvents(anchor string, amount int, towardsHistory bool) ([]gomatrix.Event, RoomEventErrorEnum) {
//	 normal (towards history): after=X - returns X+N including X
//	 return (towards present): before=X - returnx X-n not including X
//
// 0 is the LATEST Event
// Len-1 is the OLDEST event
//
//if amount <= 0 {
//	return []gomatrix.Event{}, RoomEventsUnknownError
//}

//var eventIndex int
//
//if anchor != "" {
//	if index, found := r.findEventIndex(anchor, true); found {
//		eventIndex = index
//	} else {
//		return []gomatrix.Event{}, RoomEventsCouldNotFindEvent
//	}
//}
//
//if towardsHistory {
//	return r.getBackwardEventRange(eventIndex, amount), RoomEventsFine
//} else {
//	return r.getForwardEventRange(eventIndex, amount), RoomEventsFine
//}
//}

func (r *Room) GetEventPage(anchor string, offset int, pageSize int) (events []gomatrix.Event, state RoomEventErrorEnum) {
	var anchorIndex int

	if anchor != "" {
		if index, found := r.findEventIndex(anchor, false); found {
			anchorIndex = index
		} else {
			return []gomatrix.Event{}, RoomEventsCouldNotFindEvent
		}
	}

	if offset >= 0 { // backwards
		return r.getBackwardEventRange(anchorIndex, offset, pageSize), RoomEventsFine
	} else { // forwards
		return r.getForwardEventRange(anchorIndex, -offset, pageSize), RoomEventsFine
	}

	return
}

const RoomInitialSyncLimit = 256

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

func (r *Room) LazyInitialSync() bool {
	r.activeLock.RLock()
	hasInitialSynced := r.hasInitialSynced
	r.activeLock.RUnlock()

	if hasInitialSynced {
		return true
	}

	resp, err := r.client.RoomInitialSync(r.ID, RoomInitialSyncLimit)

	if err != nil {
		fmt.Println(err)
		return false
	}

	r.activeLock.Lock()
	defer r.activeLock.Unlock()

	for _, event := range resp.State {
		r.latestRoomState.UpdateOnEvent(&event)
	}

	r.backPaginationToken = resp.Messages.Start
	r.forwardPaginationToken = resp.Messages.End

	utils.ReverseEvents(resp.Messages.Chunk)
	r.eventList = resp.Messages.Chunk
	r.hasInitialSynced = true
	return true
}

// Partial implementation of http://matrix.org/docs/spec/client_server/r0.2.0.html#calculating-the-display-name-for-a-room
// falling back to room ID instead of "Empty Room" (though this should not be possible with guest-world-readble-rooms)
func (r *Room) GetName() string {
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()

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
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()
	return r.latestRoomState.NumJoinedMembers
}

func (r *Room) CanonicalAlias() string {
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()

	if r.latestRoomState.CanonicalAlias != "" {
		return r.latestRoomState.CanonicalAlias
	}
	if len(r.latestRoomState.Aliases) > 0 {
		return r.latestRoomState.Aliases[0]
	}
	return ""
}
func (r *Room) AvatarUrl() MXCURL {
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()
	return r.latestRoomState.AvatarURL
}
func (r *Room) Topic() string {
	r.activeLock.RLock()
	defer r.activeLock.RUnlock()
	return r.latestRoomState.topic
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (r *Room) GetServers() PairList {
	serverMap := make(map[string]int)
	for _, member := range r.latestRoomState.CalculateMemberList() {
		if mxidSplit := strings.SplitN(member.MXID, ":", 2); len(mxidSplit) == 2 {
			serverMap[mxidSplit[1]]++
		}
	}

	serverList := make(PairList, 0, len(serverMap))
	for server, num := range serverMap {
		serverList = append(serverList, Pair{server, num})
	}

	sort.Sort(sort.Reverse(serverList))
	return serverList
}

func (r *Room) GetMembers() []*MemberInfo {
	return r.latestRoomState.CalculateMemberList()
}

func (r *Room) GetNumMemberEvents() int {
	r.latestRoomState.RLock()
	defer r.latestRoomState.RUnlock()
	return len(r.latestRoomState.memberMap)
}
