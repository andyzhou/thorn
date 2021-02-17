package network

import (
	"crypto/sha1"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/protocol"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"time"
)

/*
 * kcp server face
 */

//face info
type KcpServer struct {
	address string //like ':10086'
	password string
	salt string
	cb iface.IRoomCallback
	router iface.IConnCallBack
	protocol iface.IProtocol
	config iface.IConfig
	listener *kcp.Listener
	manager iface.IManager
	needQuit bool
}

//construct
func NewKcpServer(
			address,
			password,
			salt string,
		) *KcpServer {
	//init manager
	manager := NewManager()

	//self init
	this := &KcpServer{
		address:address,
		password:password,
		salt:salt,
		protocol:protocol.NewProtocol(),
		listener:new(kcp.Listener),
		manager:manager,
		router:NewRouter(manager),
	}

	//inter init
	this.interInit()

	//spawn main process
	go this.runMainProcess()

	return this
}

//stop
func (f *KcpServer) Quit() {
	f.needQuit = true
	f.listener.Close()
	f.listener = nil
}

//get router
func (f *KcpServer) GetRouter() iface.IConnCallBack {
	return f.router
}

//get manager
func (f *KcpServer) GetManager() iface.IManager {
	return f.manager
}

//get protocol
func (f *KcpServer) GetProtocol() iface.IProtocol {
	return f.protocol
}

//get config
func (f *KcpServer) GetConfig() iface.IConfig {
	return f.config
}

//set room callback
func (f *KcpServer) SetCallback(cb iface.IRoomCallback) bool {
	if cb == nil {
		return false
	}
	f.cb = cb
	return true
}

//set config
func (f *KcpServer) SetConfig(config iface.IConfig) bool {
	if config == nil {
		return false
	}
	f.config = config
	return true
}

//////////////////
//private func
//////////////////

//run main process
func (f *KcpServer) runMainProcess() {
	if f.listener == nil {
		return
	}
	log.Println("KcpServer:runMainProcess wait connect..")

	//loop
	for {
		if f.needQuit {
			break
		}
		//accept new connect
		sess, err := f.listener.AcceptKCP()
		if err != nil {
			log.Println("Server accept failed, err:", err)
			continue
		}
		log.Println("conn sess:", sess.RemoteAddr())

		//set upd mode
		f.setUdpMode(sess)

		//simple testing
		//go f.handleEcho(sess)

		//new udp connect
		conn := NewConn(sess, f)
		if f.cb != nil {
			conn.SetCallBack(f.cb)
		}
		conn.Do()
	}
}

//test read
func (f *KcpServer) handleEcho(conn *kcp.UDPSession) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("handleEcho, read buf:", string(buf))

		n, err = conn.Write(buf[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}

//set udp mode
func (f *KcpServer) setUdpMode(session *kcp.UDPSession) bool {
	if session == nil {
		return false
	}
	session.SetNoDelay(1, 10, 2, 1)
	session.SetStreamMode(true)
	session.SetWindowSize(4096, 4096)
	session.SetReadBuffer(4 * 1024 * 1024)
	session.SetWriteBuffer(4 * 1024 * 1024)
	session.SetACKNoDelay(true)
	return true
}

//inter init
func (f *KcpServer) interInit() {
	//init AES key
	key := pbkdf2.Key([]byte(f.password), []byte(f.salt), 1024, 32, sha1.New)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		panic(err)
		return
	}

	//init kcp listener, udp protocol
	f.listener, err = kcp.ListenWithOptions(f.address, block, 10, 3)
	if err != nil {
		panic(err)
		return
	}
	log.Println("KcpServer:init, listen on ", f.address)

	//init chan limit
	packetChanLimit := uint32(1024)
	timeOut := time.Second * 5

	//init default config
	f.config = protocol.NewConfig(
			packetChanLimit,
			packetChanLimit,
			timeOut,
			timeOut,
		)
}