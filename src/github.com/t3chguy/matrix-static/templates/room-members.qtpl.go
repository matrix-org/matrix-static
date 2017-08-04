// This file is automatically generated by qtc from "room-members.qtpl".
// See https://github.com/valyala/quicktemplate for details.

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:1
package templates

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:1
import "github.com/t3chguy/matrix-static/mxclient"

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:5
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:5
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:5
type RoomMembersPage struct {
	RoomInfo mxclient.RoomInfo
	Members  []mxclient.MemberInfo
	PageSize int
	Page     int
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:14
func (p *RoomMembersPage) streamprintMemberRow(qw422016 *qt422016.Writer, Member *mxclient.MemberInfo) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:14
	qw422016.N().S(`
    <tr>
        <td><a href="`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:16
	qw422016.E().S(p.BaseUrl())
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:16
	qw422016.E().S(Member.MXID)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:16
	qw422016.N().S(`">`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:16
	qw422016.E().S(Member.MXID)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:16
	qw422016.N().S(`</a></td>
        <td>
            `)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:18
	if Member.AvatarURL.IsValid() {
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:18
		qw422016.N().S(`
                <img class="memberListAvatar" src="`)
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:19
		qw422016.E().S(Member.AvatarURL.ToThumbURL(48, 48, "crop"))
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:19
		qw422016.N().S(`" />
            `)
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:20
	} else {
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:20
		qw422016.N().S(`
                <img class="memberListAvatar" src="./avatar/`)
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:21
		qw422016.N().U(Member.GetName())
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:21
		qw422016.N().S(`" />
            `)
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:22
	}
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:22
	qw422016.N().S(`
        </td>
        <td>`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:24
	qw422016.E().S(Member.DisplayName)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:24
	qw422016.N().S(`</td>
        <td>`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:25
	qw422016.E().S(Member.PowerLevel.String())
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:25
	qw422016.N().S(` (`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:25
	qw422016.N().D(Member.PowerLevel.Int())
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:25
	qw422016.N().S(`)</td>
    </tr>
`)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
func (p *RoomMembersPage) writeprintMemberRow(qq422016 qtio422016.Writer, Member *mxclient.MemberInfo) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	p.streamprintMemberRow(qw422016, Member)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	qt422016.ReleaseWriter(qw422016)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
func (p *RoomMembersPage) printMemberRow(Member *mxclient.MemberInfo) string {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	qb422016 := qt422016.AcquireByteBuffer()
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	p.writeprintMemberRow(qb422016, Member)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	qs422016 := string(qb422016.B)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	qt422016.ReleaseByteBuffer(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
	return qs422016
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:27
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:31
func (p *RoomMembersPage) StreamTitle(qw422016 *qt422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:31
	qw422016.N().S(`
    Public Room Members - `)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:32
	qw422016.E().S(p.RoomInfo.Name)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:32
	qw422016.N().S(` - Riot Static
`)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
func (p *RoomMembersPage) WriteTitle(qq422016 qtio422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	p.StreamTitle(qw422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	qt422016.ReleaseWriter(qw422016)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
func (p *RoomMembersPage) Title() string {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	qb422016 := qt422016.AcquireByteBuffer()
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	p.WriteTitle(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	qs422016 := string(qb422016.B)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	qt422016.ReleaseByteBuffer(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
	return qs422016
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:33
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:35
func (p *RoomMembersPage) StreamHead(qw422016 *qt422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:35
	qw422016.N().S(`
`)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
func (p *RoomMembersPage) WriteHead(qq422016 qtio422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	p.StreamHead(qw422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	qt422016.ReleaseWriter(qw422016)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
func (p *RoomMembersPage) Head() string {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	qb422016 := qt422016.AcquireByteBuffer()
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	p.WriteHead(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	qs422016 := string(qb422016.B)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	qt422016.ReleaseByteBuffer(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
	return qs422016
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:36
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:38
func (p *RoomMembersPage) StreamHeader(qw422016 *qt422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:38
	qw422016.N().S(`
    `)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:39
	StreamPrintRoomHeader(qw422016, p.RoomInfo)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:39
	qw422016.N().S(`
`)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
func (p *RoomMembersPage) WriteHeader(qq422016 qtio422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	p.StreamHeader(qw422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	qt422016.ReleaseWriter(qw422016)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
func (p *RoomMembersPage) Header() string {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	qb422016 := qt422016.AcquireByteBuffer()
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	p.WriteHeader(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	qs422016 := string(qb422016.B)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	qt422016.ReleaseByteBuffer(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
	return qs422016
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:40
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:43
func (p *RoomMembersPage) StreamBody(qw422016 *qt422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:43
	qw422016.N().S(`<div>`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:45
	qw422016.N().D(p.RoomInfo.NumMemberEvents)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:45
	qw422016.N().S(` `)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:45
	qw422016.N().S(`users have interacted with this room.</div>`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:47
	StreamPaginatorCurPage(qw422016, p)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:47
	qw422016.N().S(`<table><thead><tr><td>MXID</td><td>Avatar</td><td>Display Name</td><td>Power Level</td></tr></thead><tbody>`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:59
	for _, Member := range p.Members {
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:60
		p.streamprintMemberRow(qw422016, &Member)
		//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:61
	}
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:61
	qw422016.N().S(`</tbody></table>`)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:65
	StreamPaginatorFooter(qw422016, p)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
func (p *RoomMembersPage) WriteBody(qq422016 qtio422016.Writer) {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	p.StreamBody(qw422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	qt422016.ReleaseWriter(qw422016)
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
func (p *RoomMembersPage) Body() string {
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	qb422016 := qt422016.AcquireByteBuffer()
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	p.WriteBody(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	qs422016 := string(qb422016.B)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	qt422016.ReleaseByteBuffer(qb422016)
	//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
	return qs422016
//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:67
}

//line src\github.com\t3chguy\matrix-static\templates\room-members.qtpl:74
func (p *RoomMembersPage) CurPage() int {
	return p.Page
}
func (p *RoomMembersPage) HasNextPage() bool {
	return len(p.Members) == p.PageSize
}
func (p *RoomMembersPage) BaseUrl() string {
	return RoomBaseUrl(p.RoomInfo.RoomID) + "/members/"
}