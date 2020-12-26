package thorn

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/network"
	"github.com/andyzhou/thorn/protocol"
	"github.com/andyzhou/thorn/room"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/*
 * server api face
 */

//face info
type Server struct {
	address string
	cb iface.IRoomCallback //callback for api client
	kcp iface.IKcpServer
	manager iface.IManager
}

//construct, step-1
func NewServer(address string) *Server {
	//self init
	this := &Server{
		address: address,
		manager: room.NewManager(),
	}
	//inter init
	this.interInit()
	return this
}

///////////////
//service api
///////////////

//stop
func (f *Server) Stop() {
	if f.kcp != nil {
		f.kcp.Quit()
	}
}

//start, step-1
func (f *Server) Start() {
	if f.kcp != nil {
		//start
		f.kcp.Start(room.NewRouter(f.manager), protocol.NewProtocol())
	}
}

//create room, step-2
func (f *Server) CreateRoom(
			roomId uint64,
			players []uint64,
			randSeed int32,
		) bool {
	//basic check
	if roomId <= 0 || players == nil {
		return false
	}

	//init room
	room := room.NewRoom(roomId, players, randSeed, f.cb)

	//add into manager
	bRet := f.manager.AddRoom(room)

	return bRet
}

//start room, step-3
func (f *Server) StartRoom(roomId uint64) bool {
	//basic check
	if roomId <= 0 {
		return false
	}

	//get room
	room := f.manager.GetRoom(roomId)
	if room == nil {
		return false
	}
	return true
}

//register cb for api client
//client should implement this callback
func (f *Server) SetCallback(cb iface.IRoomCallback) bool {
	if cb == nil {
		return false
	}
	f.cb = cb
	return true
}

///////////////
//private func
///////////////

//inter init
func (f *Server) interInit() {

	//init signal
	sig := make(chan os.Signal, 1)
	signal.Notify(
			sig,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
			os.Interrupt,
		)

	//watch signal
	go func() {
		for {
			select {
			case s, ok := <- sig:
				if ok {
					log.Printf("Get signal of %v\n", s.String())
					return
				}
			}
		}
	}()

	//init kcp server
	f.kcp = network.NewKcpServer(f.address)
}