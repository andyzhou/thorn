package network

import (
	"github.com/andyzhou/thorn/iface"
	"net"
	"sync/atomic"
	"time"
)

/*
 * conn face, implement of IConn
 */

//face info
type Conn struct {
	conn net.Conn //raw connection
	callback iface.IConnCallBack //cb interface
	extraData interface{}
	closeFlag int32
	packetSendChan chan iface.IPacket
	packetReceiveChan chan iface.IPacket
	closeChan chan bool
}

//construct
func NewConn(conn net.Conn) *Conn {
	//self init
	this := &Conn{
		conn:conn,
		closeChan:make(chan bool, 1),
	}
	return this
}

//close
func (f *Conn) Close() {
}

//check is closed
func (f *Conn) IsClosed() bool {
	return atomic.LoadInt32(&f.closeFlag) == 1
}

//get extra data
func (f *Conn) GetExtraData() interface{} {
	return f.extraData
}

//set extra data
func (f *Conn) SetExtraData(data interface{}) bool {
	if data == nil {
		return false
	}
	f.extraData = data
	return true
}

//get raw connect
func (f *Conn) GetRawConn() net.Conn {
	return f.conn
}

//set call back
func (f *Conn) SetCallBack(cb iface.IConnCallBack)  {
	f.callback = cb
}

//async send packet
func (f *Conn) AsyncWritePacket(
					packet iface.IPacket,
					duration time.Duration,
				) error {
	return nil
}