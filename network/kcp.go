package network

import (
	"crypto/sha1"
	"github.com/andyzhou/thorn/define"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/protocol"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"sync"
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
	cb iface.IConnCallBack
	router iface.IConnCallBack
	protocol iface.IProtocol
	config iface.IConfig
	listener *kcp.Listener
	manager iface.IManager
	needQuit bool
	sync.RWMutex
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
	f.manager.Close()
	f.Lock()
	defer f.Unlock()
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
func (f *KcpServer) SetCallback(cb iface.IConnCallBack) bool {
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

		//set upd mode
		f.setUdpMode(sess)

		//new udp connect
		conn := NewConn(sess, f)
		if f.cb != nil {
			conn.SetCallBack(f.cb)
		}
		conn.Do()
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
	key := pbkdf2.Key(
					[]byte(f.password),
					[]byte(f.salt),
					1024,
					32,
					sha1.New,
				)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		panic(err)
		return
	}

	//init kcp listener, udp protocol
	f.listener, err = kcp.ListenWithOptions(
							f.address,
							block,
							10,
							3,
						)
	if err != nil {
		panic(err)
		return
	}

	//init chan limit
	packetChanLimit := uint32(define.DefaultChanSize)
	timeOut := time.Second * define.DefaultTimeOut

	//init default config
	f.config = protocol.NewConfig(
			packetChanLimit,
			packetChanLimit,
			timeOut,
			timeOut,
		)
}