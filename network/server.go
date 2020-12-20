package network

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/xtaci/kcp-go"
	"net"
	"sync"
)

/*
 * server face, implement of IServer
 */

//face info
type Server struct {
	config *Config
	protocol iface.IProtocol
	cb iface.IConnCallBack
	listener  net.Listener
	closeChan chan bool
	closeOnce sync.Once
	wg *sync.WaitGroup
}

//construct
func NewServer(
			config *Config,
			cb iface.IConnCallBack,
			protocol iface.IProtocol,
		) *Server {
	//self init
	this := &Server{
		config:config,
		cb:cb,
		protocol:protocol,
		closeChan:make(chan bool, 1),
		wg:new(sync.WaitGroup),
	}
	return this
}

//stop
func (f *Server) Stop() {
	f.closeOnce.Do(func() {
			close(f.closeChan)
			f.listener.Close()
		})
	f.wg.Wait()
}

//start
func (f *Server) Start(listener net.Listener) {
	f.listener = listener
	f.wg.Add(1)
	defer func() {
		f.wg.Done()
	}()

	//loop
	for {
		//receive close chan
		select {
		case <- f.closeChan:
			return
		}

		//accept new connect
		conn, err := f.listener.Accept()
		if err != nil {
			continue
		}

		//set udp mode
		f.setUdpMode(conn)

		//create new client connect
		f.wg.Add(1)
		go func() {
			conn := NewConn(conn, f)
			conn.Do()
			f.wg.Done()
		}()
	}
}

//get protocol
func (f *Server) GetProtocol() iface.IProtocol {
	return f.protocol
}

//get config
func (f *Server) GetConfig() *Config {
	return f.config
}

//////////////////
//private func
//////////////////

//set udp mode
func (f *Server) setUdpMode(conn net.Conn) {
	kcpConn := conn.(*kcp.UDPSession)
	kcpConn.SetNoDelay(1, 10, 2, 1)
	kcpConn.SetStreamMode(true)
	kcpConn.SetWindowSize(4096, 4096)
	kcpConn.SetReadBuffer(4 * 1024 * 1024)
	kcpConn.SetWriteBuffer(4 * 1024 * 1024)
	kcpConn.SetACKNoDelay(true)
}