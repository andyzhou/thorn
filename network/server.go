package network

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/xtaci/kcp-go"
	"log"
	"net"
	"sync"
)

/*
 * server face, implement of IServer
 */

//face info
type Server struct {
	config iface.IConfig
	protocol iface.IProtocol
	cb iface.IConnCallBack
	listener  net.Listener
	closeChan chan bool
	closeOnce sync.Once
	wg *sync.WaitGroup
}

//construct
func NewServer(
			config iface.IConfig,
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
	f.closeChan <- true
}

//start
//accept new connection
func (f *Server) Start(listener net.Listener) {
	//sync listener
	f.listener = listener

	//spawn main process
	go f.runMainProcess()
}

//get protocol
func (f *Server) GetProtocol() iface.IProtocol {
	return f.protocol
}

//get config
func (f *Server) GetConfig() iface.IConfig {
	return f.config
}

//////////////////
//private func
//////////////////

//run main process
func (f *Server) runMainProcess() {
	//defer
	defer func() {
		close(f.closeChan)
		f.listener.Close()
	}()

	log.Println("Server start running...")

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