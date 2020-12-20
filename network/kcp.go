package network

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/xtaci/kcp-go"
	"net"
	"time"
)

/*
 * kcp server face
 */

//face info
type KcpServer struct {
	address string //like ':10086'
	config *Config
	listener net.Listener
}

//construct
func NewKcpServer(
			address string,
		) *KcpServer {
	//self init
	this := &KcpServer{
		address:address,
	}

	//inter init
	this.interInit()
	return this
}

//stop
func (f *KcpServer) Quit() {
	f.listener.Close()
	f.listener = nil
}

//start
func (f *KcpServer) Start(
			cb iface.IConnCallBack,
			protocol iface.IProtocol,
		) {
	//init server
	server := NewServer(f.config, cb, protocol)

	//spawn main process
	go server.Start(f.listener)
}

//////////////////
//private func
//////////////////

//inter init
func (f *KcpServer) interInit() {
	//init kcp listener
	listener, err := kcp.Listen(f.address)
	if err != nil {
		panic(err)
	}
	f.listener = listener

	//init config
	f.config = &Config{
		PacketReceiveChanLimit: 1024,
		PacketSendChanLimit:    1024,
		ConnReadTimeout:        time.Second * 5,
		ConnWriteTimeout:       time.Second * 5,
	}
}