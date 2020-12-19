package room

import (
	"github.com/andyzhou/thorn/iface"
	"time"
)

/*
 * player face, implement of IPlayer
 */

//face info
type Player struct {
	id uint64
	idx uint64
	isReady bool
	isOnline bool
	lastHeartBeatTime int64
	sendFrameCount uint32
	client iface.IConn
}

//construct
func NewPlayer(id, idx uint64) *Player {
	//self init
	this := &Player{
		id:id,
		idx:idx,
	}
	return this
}

func (f *Player) CleanUp() {
	if f.client != nil {
		f.client.Close()
	}
	f.client = nil
	f.isOnline = false
	f.isReady = false
}

func (f *Player) GetConn() iface.IConn {
	return f.client
}

func (f *Player) Connect(conn iface.IConn) {
	f.client = conn
	f.isOnline = true
	f.isReady = false
	f.lastHeartBeatTime = time.Now().Unix()
}

func (f *Player) IsOnline() bool {
	return f.client != nil && f.isOnline
}

func (f *Player) RefreshHeartbeatTime() {
	f.lastHeartBeatTime = time.Now().Unix()
}

func (f *Player) GetLastHeartbeatTime() int64 {
	return f.lastHeartBeatTime
}

func (f *Player) SetSendFrameCount(c uint32) {
	f.sendFrameCount = c
}

func (f *Player) GetSendFrameCount() uint32 {
	return f.sendFrameCount
}

func (f *Player) SendMessage(packet iface.IPacket) {
	if packet == nil || !f.IsOnline() {
		return
	}
	if nil != f.client.AsyncWritePacket(packet, 0) {
		f.client.Close()
	}
}
