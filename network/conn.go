package network

import (
	"github.com/andyzhou/thorn/define"
	"github.com/andyzhou/thorn/iface"
	"github.com/xtaci/kcp-go"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * conn face, implement of IConn
 * - original udp connect process
 * - read, write udp data
 */

//face info
type Conn struct {
	server            iface.IKcpServer    //reference
	conn              *kcp.UDPSession     //raw connection
	callback          iface.IConnCallBack //connect cb interface from outside
	extraData         interface{}
	activeTime        int64              //last active timestamp
	packetSendChan    chan iface.IPacket //send chan
	packetReceiveChan chan iface.IPacket //receive chan
	closeFlag         int32
	closeChan         chan bool
	closeOnce         sync.Once
	wg                sync.WaitGroup
}

//construct
func NewConn(
		sess *kcp.UDPSession,
		server iface.IKcpServer,
	) *Conn {
	//self init
	this := &Conn{
		conn:sess,
		server:server,
		activeTime:time.Now().Unix(),
		packetSendChan:make(chan iface.IPacket, define.ConnPacketChanSize),
		packetReceiveChan:make(chan iface.IPacket, define.ConnPacketChanSize),
		closeChan:make(chan bool, 1),
		wg:sync.WaitGroup{},
	}
	return this
}

//close
func (f *Conn) Close() {
	var (
		m any = nil
	)
	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("Conn:Close panic, err:", err)
		}
	}()

	//do some cleanup
	f.closeOnce.Do(func() {
		atomic.StoreInt32(&f.closeFlag, 1)
		close(f.packetSendChan)
		close(f.packetReceiveChan)
		close(f.closeChan)
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
	//check callback
	if f.callback != nil && !reflect.ValueOf(f.callback).IsNil() {
		if !f.callback.OnConnect(f) {
			return
		}
	}

	//check router
	router := f.server.GetRouter()
	if router != nil && !reflect.ValueOf(router).IsNil() {
		if !f.server.GetRouter().OnConnect(f) {
			return
		}
	}

	//async do three process
	f.asyncDo(f.handleLoop, &f.wg)
	f.asyncDo(f.readLoop, &f.wg)
	f.asyncDo(f.writeLoop, &f.wg)
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

//get last active time
func (f *Conn) GetActiveTime() int64 {
	return f.activeTime
}

//get raw connect
func (f *Conn) GetRawConn() net.Conn {
	return f.conn
}

//set connect call back
func (f *Conn) SetCallBack(cb iface.IConnCallBack)  {
	if cb == nil {
		return
	}
	f.callback = cb
}

//async send packet
func (f *Conn) AsyncWritePacket(packet iface.IPacket, timeout time.Duration) error {
	var (
		m any = nil
	)
	//basic check
	if packet == nil || f.IsClosed() {
		return define.ErrConnClosing
	}

	defer func() {
		if err := recover(); err != m {
			//err = define.ErrConnClosing
			log.Println("conn.AsyncWritePacket panic, err:", err)
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
	var (
		m any = nil
	)
	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("Conn:writeLoop panic, err:", err)
		}
		f.Close()
	}()

	//get server config
	//serverConf := f.server.GetConfig()
	//writeTimeOut := serverConf.GetConnWriteTimeout()

	//loop
	for {
		select {
		case <- f.closeChan:
			//log.Println("writeLoop close chan")
			return
		case p, ok := <- f.packetSendChan:
			if ok {
				if f.IsClosed() {
					return
				}
				//write packet
				//f.conn.SetWriteDeadline(time.Now().Add(writeTimeOut))
				_, err := f.conn.Write(p.Pack())
				if err != nil {
					log.Println("Conn:writeLoop, err:", err)
					return
				}
				//update active time
				f.activeTime = time.Now().Unix()
			}
		}
	}
}

//read loop
func (f *Conn) readLoop() {
	var (
		m any = nil
	)
	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("Conn:readLoop panic, err:", err)
		}
		f.Close()
	}()

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
		if err != nil {
			log.Println("Conn:readLoop, err:", err)
			continue
		}
		//send to receive chan
		f.packetReceiveChan <- message
	}
}

//handle loop
func (f *Conn) handleLoop() {
	var (
		m any = nil
	)
	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("Conn:handleLoop panic, err:", err)
		}
		f.Close()
	}()

	//loop
	for {
		select {
		case <- f.closeChan:
			log.Println("handleLoop close chan")
			return
		case p, ok := <- f.packetReceiveChan:
			if ok && &p != nil {
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
				f.activeTime = time.Now().Unix()
			}
		}
	}
}

//async do some func
func (f *Conn) asyncDo(fun func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fun()
		wg.Done()
	}()
}