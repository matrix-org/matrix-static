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

type PowerLevel int

// TODO don't bother with this and have a map similar to react-sdk "Roles.js"

// String is the Stringer implementation for PowerLevel
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

// Int allows a quick denature of PowerLevel to an int
func (powerLevel PowerLevel) Int() int {
	return int(powerLevel)
}

// implements sort.Interface
type MemberList []*MemberInfo

func (ml MemberList) Len() int { return len(ml) }
func (ml MemberList) Less(i, j int) bool {
	a, b := ml[i], ml[j]
	plA, plB := a.PowerLevel.Int(), b.PowerLevel.Int()
	if plA == plB {
		// Secondary sort is Low->High Lexicographically on GetName()
		return a.GetName() < b.GetName()
	}

	// Primary Sort is High->Low on PowerLevel
	return plA > plB
}
func (ml MemberList) Swap(i, j int) { ml[i], ml[j] = ml[j], ml[i] }

type MemberInfo struct {
	MXID        string
	Membership  string
	DisplayName string
	AvatarURL   MXCURL
	PowerLevel  PowerLevel
}

// NewMemberInfo returns a new MemberInfo with defaults (membership=leave) applied.
func NewMemberInfo(mxid string) *MemberInfo {
	return &MemberInfo{
		MXID:       mxid,
		Membership: "leave",
	}
}

// GetName returns either the user's DisplayName, or if empty, their MXID.
// TODO make this disambiguate users if their DisplayName is not unique,
// implementation tips say to make a map of DisplayNames (or Set?) -> MXID?
func (memberInfo MemberInfo) GetName() string {
	if memberInfo.DisplayName != "" {
		return memberInfo.DisplayName
	} else {
		return memberInfo.MXID
	}
}
