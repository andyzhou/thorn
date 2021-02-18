package iface

import (
	"net"
	"time"
)

/*
 * interface of udp connect
 */

//callback for connect
type IConnCallBack interface {
	OnConnect(conn IConn) bool //cb for connected
	OnMessage(conn IConn, packet IPacket) bool //cb for received packet
	OnClose(conn IConn) //cb for closed conn
}

type IConn interface {
	Close()
	IsClosed() bool
	Do()
	AsyncWritePacket(packet IPacket, duration time.Duration) error
	GetActiveTime() int64
	GetRawConn() net.Conn
	GetExtraData() interface{}
	SetExtraData(data interface{}) bool
	SetCallBack(cb IConnCallBack)
}
