package network

import (
	"github.com/andyzhou/thorn/define"
	"github.com/andyzhou/thorn/iface"
	"github.com/xtaci/kcp-go"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * conn face, implement of IConn
 * - original udp connect process
 * - read, write udp data
 */

//inter macro define
const (
	ConnPacketChanSize = 1024
)

//face info
type Conn struct {
	server iface.IKcpServer
	conn *kcp.UDPSession //raw connection
	callback iface.IConnCallBack //cb interface for out side
	extraData interface{}
	closeOnce sync.Once
	closeFlag int32
	packetSendChan chan iface.IPacket
	packetReceiveChan chan iface.IPacket
	closeChan chan bool
	wg *sync.WaitGroup
}

//construct
func NewConn(
				sess *kcp.UDPSession,
				server iface.IKcpServer,
			) *Conn {
	//self init
	this := &Conn{
		server:server,
		conn:sess,
		packetSendChan:make(chan iface.IPacket, ConnPacketChanSize),
		packetReceiveChan:make(chan iface.IPacket, ConnPacketChanSize),
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
	if f.callback != nil {
		f.callback.OnConnect(f)
	}
	if f.server.GetRouter() != nil {
		f.server.GetRouter().OnConnect(f)
	}

	//spawn three process
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
	//basic check
	if packet == nil || f.IsClosed() {
		return define.ErrConnClosing
	}

	defer func() {
		if err := recover(); err != nil {
			err = define.ErrConnClosing
		}
	}()

	if timeout == 0 {
		select {
		case f.packetSendChan <- packet:
			return nil
		default:
			return define.ErrWriteBlocking
		}
	}else{
		select {
		case f.packetSendChan <- packet:
			return nil
		case <- f.closeChan:
			return define.ErrConnClosing
		case <- time.After(timeout):
			return define.ErrWriteBlocking
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

	log.Println("Conn:writeLoop...")
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
				//serverConf := f.server.GetConfig()
				//f.conn.SetWriteDeadline(
				//			time.Now().Add(serverConf.GetConnWriteTimeout()),
				//		)
				_, err := f.conn.Write(p.Pack())
				log.Println("writeLoop, err:", err)
				if err != nil {
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

	log.Println("Conn:readLoop...")

	//get server config
	//serverConf := f.server.GetConfig()
	//readTimeOut := serverConf.GetConnReadTimeout()

	//loop
	for {
		if f.IsClosed() {
			return
		}
		//read packet
		//f.conn.SetReadDeadline(time.Now().Add(readTimeOut))
		message, err := f.server.GetProtocol().ReadPacket(f.conn)
		log.Println("readLoop, err:", err)
		if err != nil {
			continue
		}
		//send to receive chan
		f.packetReceiveChan <- message
	}
}

//handle loop
func (f *Conn) handleLoop() {
	defer func() {
		recover()
		f.Close()
	}()

	log.Println("Conn:handleLoop...")
	//loop
	for {
		log.Println("Conn:handleLoop..")
		select {
		case <- f.closeChan:
			return
		case p, ok := <- f.packetReceiveChan:
			if ok {
				if f.IsClosed() {
					return
				}
				//callback
				if f.server.GetRouter() != nil {
					f.server.GetRouter().OnMessage(f, p)
				}
				if f.callback != nil {
					f.callback.OnMessage(f, p)
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