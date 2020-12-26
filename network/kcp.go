package network

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/protocol"
	"github.com/xtaci/kcp-go"
	"log"
	"net"
	"time"
)

/*
 * kcp server face
 */

//face info
type KcpServer struct {
	address string //like ':10086'
	config iface.IConfig
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
	server.Start(f.listener)
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
	log.Printf("KcpServer:init, listen on %s\n", f.address)

	//init chan limit
	packetChanLimit := uint32(1024)
	timeOut := time.Second * 5

	//init config
	f.config = protocol.NewConfig(
			packetChanLimit,
			packetChanLimit,
			timeOut,
			timeOut,
		)
}