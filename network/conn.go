package network

import (
	"github.com/andyzhou/thorn/iface"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * conn face, implement of IConn
 */

//face info
type Conn struct {
	server iface.IServer
	conn net.Conn //raw connection
	callback iface.IConnCallBack //cb interface
	extraData interface{}
	closeFlag int32
	packetSendChan chan iface.IPacket
	packetReceiveChan chan iface.IPacket
	closeChan chan bool
	wg *sync.WaitGroup
}

//construct
func NewConn(conn net.Conn, server iface.IServer) *Conn {
	//self init
	this := &Conn{
		server:server,
		conn:conn,
		closeChan:make(chan bool, 1),
		wg:new(sync.WaitGroup),
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

//do it
func (f *Conn) Do() {
	if !f.callback.OnConnect(f) {
		return
	}
	f.asyncDo(f.handleLoop, f.wg)
	f.asyncDo(f.readLoop, f.wg)
	f.asyncDo(f.writeLoop, f.wg)
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

///////////////
//private func
///////////////

//write loop
func (f *Conn) writeLoop() {

}

//read loop
func (f *Conn) readLoop() {

}

//handle loop
func (f *Conn) handleLoop() {

}

func (f *Conn) asyncDo(fun func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fun()
		wg.Done()
	}()
}