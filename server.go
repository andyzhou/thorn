package thorn

import (
	"errors"
	"fmt"
	"github.com/andyzhou/thorn/conf"
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
	conf    *ServerConf
	address string              //host:port
	cb      iface.IConnCallBack //callback for api client
	kcp     iface.IKcpServer
	wg      *sync.WaitGroup
	wgVal   int32
}

//construct, step-1
//address format: ip/domain:port
func NewServer(conf *ServerConf) *Server {
	//self init
	this := &Server{
		address: fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		conf:    conf,
		wg:      new(sync.WaitGroup),
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
		f.kcp = nil
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
	fmt.Printf("server listen on %v\n", f.conf.Port)
	f.wg.Wait()
}

//register cb for connect client, step-3
//client should implement this callback
func (f *Server) SetCallback(cb iface.IConnCallBack) error {
	if cb == nil {
		return errors.New("connect cb is nil")
	}
	if f.kcp != nil {
		//set call back
		f.kcp.SetCallback(cb)
	}
	f.cb = cb
	return nil
}

//create room, step-4
func (f *Server) CreateRoom(cfg *conf.RoomConf) (iface.IRoom, error) {
	//basic check
	if cfg == nil {
		return nil, errors.New("invalid parameter")
	}
	if cfg.RoomId <= 0 {
		return nil, errors.New("room id must exceed 0")
	}

	//try check room
	roomObj := f.GetRoom(cfg.RoomId)
	if roomObj != nil {
		return roomObj, nil
	}

	//init new room
	roomObj = room.NewRoom(cfg)

	//add into manager
	f.kcp.GetManager().AddRoom(roomObj)
	return roomObj, nil
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

//set kcp config
func (f *Server) SetConfig(config iface.IConfig) bool {
	return f.kcp.SetConfig(config)
}

///////////////
//private func
///////////////

//sync group done
func (f *Server) syncGroupDone() {
	if f.wgVal > 0 {
		atomic.AddInt32(&f.wgVal, -1)
	}
	if f.wg != nil {
		f.wg.Done()
	}
}

//signal catch
func (f *Server) signalCatch() {
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
}

//inter init
func (f *Server) interInit() {
	//signal catch
	f.signalCatch()

	//init kcp server
	f.kcp = network.NewKcpServer(f.address, f.conf.Password, f.conf.Salt)

	//set wait group value
	atomic.StoreInt32(&f.wgVal, 0)
}