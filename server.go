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
	address string //host:port
	password string
	salt string
	cb iface.IConnCallBack //callback for api client
	kcp iface.IKcpServer
	wg *sync.WaitGroup
	wgVal int32
}

//construct, step-1
//address format: ip/domain:port
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

//start, step-2
func (f *Server) Start() {
	if f.wgVal > 0 {
		return
	}
	f.wg.Add(1)
	atomic.AddInt32(&f.wgVal, 1)
	f.wg.Wait()
}


//register cb for api client, step-3
//client should implement this callback
func (f *Server) SetCallback(cb iface.IConnCallBack) bool {
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


//create room, step-4
func (f *Server) CreateRoom(
			roomId uint64,
			players []uint64,
			randSeed int32,
			secretKey string,
		) iface.IRoom {
	//basic check
	if roomId <= 0 || players == nil {
		return nil
	}

	//try check room
	roomObj := f.GetRoom(roomId)
	if roomObj != nil {
		return roomObj
	}

	//init room
	roomObj = room.NewRoom(roomId, players, randSeed, secretKey)

	//add into manager
	f.kcp.GetManager().AddRoom(roomObj)

	return roomObj
}

//get room
func (f *Server) GetRoom(roomId uint64) iface.IRoom {
	//basic check
	if roomId <= 0 {
		return nil
	}

	//get room
	room := f.kcp.GetManager().GetRoom(roomId)
	return room
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