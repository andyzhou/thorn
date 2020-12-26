package network

import (
	"errors"
	"github.com/andyzhou/thorn/iface"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * conn face, implement of IConn
 */

// Error type
var (
	ErrConnClosing   = errors.New("use of closed network connection")
	ErrWriteBlocking = errors.New("write packet was blocking")
	ErrReadBlocking  = errors.New("read packet was blocking")
)

//face info
type Conn struct {
	server iface.IServer
	conn net.Conn //raw connection
	callback iface.IConnCallBack //cb interface
	extraData interface{}
	closeOnce sync.Once
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
	f.closeOnce.Do(func() {
		atomic.StoreInt32(&f.closeFlag, 1)
		close(f.closeChan)
		close(f.packetSendChan)
		close(f.packetReceiveChan)
		f.conn.Close()
		f.callback.OnClose(f)
	})
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
					timeout time.Duration,
				) error {
	if f.IsClosed() {
		return ErrConnClosing
	}

	defer func() {
		if err := recover(); err != nil {
			err = ErrConnClosing
		}
	}()

	if timeout == 0 {
		select {
		case f.packetSendChan <- packet:
			return nil
		default:
			return ErrWriteBlocking
		}
	}else{
		select {
		case f.packetSendChan <- packet:
			return nil
		case <- f.closeChan:
			return ErrConnClosing
		case <- time.After(timeout):
			return ErrWriteBlocking
		}
	}

	return nil
}

///////////////
//private func
///////////////

//write loop
func (f *Conn) writeLoop() {
	defer func() {
		recover()
		f.Close()
	}()

	//loop
	for {
		select {
		case <- f.closeChan:
			return
		case p, ok := <- f.packetSendChan:
			if ok {
				if f.IsClosed() {
					return
				}
				serverConf := f.server.GetConfig()
				f.conn.SetWriteDeadline(
							time.Now().Add(serverConf.GetConnWriteTimeout()),
						)
				if _, err := f.conn.Write(p.Serialize()); err != nil {
					return
				}
			}
		}
	}
}

//read loop
func (f *Conn) readLoop() {
	defer func() {
		recover()
		f.Close()
	}()

	//loop
	for {
		select {
		case <-f.closeChan:
			return
		}
		if f.IsClosed() {
			return
		}
		//read packet
		serverConf := f.server.GetConfig()
		f.conn.SetReadDeadline(
				time.Now().Add(serverConf.GetConnReadTimeout()),
			)
		p, err := f.server.GetProtocol().ReadPacket(f.conn)
		if err != nil {
			return
		}
		//send to receive chan
		f.packetReceiveChan <- p
	}
}

//handle loop
func (f *Conn) handleLoop() {
	defer func() {
		recover()
		f.Close()
	}()

	//loop
	for {
		select {
		case <- f.closeChan:
			return
		case p, ok := <- f.packetReceiveChan:
			if ok {
				if f.IsClosed() {
					return
				}
				//callback
				if !f.callback.OnMessage(f, p) {
					return
				}
			}
		}

	}
}

func (f *Conn) asyncDo(fun func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fun()
		wg.Done()
	}()
}