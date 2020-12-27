package thorn

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/network"
	"github.com/andyzhou/thorn/room"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

/*
 * server api face
 */

//face info
type Server struct {
	address string
	password string
	salt string
	cb iface.IRoomCallback //callback for api client
	kcp iface.IKcpServer
	wg *sync.WaitGroup
	wgVal int32
}

//construct, step-1
func NewServer(
			address,
			password,
			salt string,
		) *Server {
	//self init
	this := &Server{
		address:  address,
		password: password,
		salt:     salt,
		wg:       new(sync.WaitGroup),
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
	f.syncGroupDone()
}

//start, step-1
func (f *Server) Start() {
	if f.wgVal > 0 {
		return
	}
	f.wg.Add(1)
	atomic.AddInt32(&f.wgVal, 1)
	f.wg.Wait()
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
	bRet := f.kcp.GetManager().AddRoom(room)

	return bRet
}

//start room, step-3
func (f *Server) StartRoom(roomId uint64) bool {
	//basic check
	if roomId <= 0 {
		return false
	}

	//get room
	room := f.kcp.GetManager().GetRoom(roomId)
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
	if f.kcp != nil {
		//set call back
		f.kcp.SetCallback(cb)
	}
	f.cb = cb
	return true
}

//set config
func (f *Server) SetConfig(config iface.IConfig) bool {
	return f.kcp.SetConfig(config)
}

///////////////
//private func
///////////////

//sync group done
func (f *Server) syncGroupDone() {
	if f.wgVal <= 0 {
		return
	}
	atomic.AddInt32(&f.wgVal, -1)
	f.wg.Done()
}

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
					f.syncGroupDone()
					return
				}
			}
		}
	}()

	//init kcp server
	f.kcp = network.NewKcpServer(f.address, f.password, f.salt)

	//set wait group value
	atomic.StoreInt32(&f.wgVal, 0)
}