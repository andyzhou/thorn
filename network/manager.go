package network

import (
	"github.com/andyzhou/thorn/define"
	"github.com/andyzhou/thorn/iface"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * manager face, implement of IManager
 * - dynamic room manager
 */

//face info
type Manager struct {
	roomCount int32
	rooms *sync.Map //roomId -> IRoom
	closeChan chan bool
}

//construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		rooms:new(sync.Map),
		roomCount:0,
		closeChan:make(chan bool, 1),
	}
	//spawn main process
	go this.runMainProcess()
	return this
}

//close
func (f *Manager) Close() {
	var (
		m any = nil
	)
	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("Manager:Close panic, err:", err)
		}
	}()

	f.closeChan <- true
	if f.rooms == nil {
		return
	}
	sf := func(k, v interface{}) bool {
		room, ok := v.(iface.IRoom)
		if !ok {
			return false
		}
		room.Stop()
		return true
	}
	f.rooms.Range(sf)
}

//get rooms
func (f *Manager) GetRooms() int32 {
	return f.roomCount
}

//close room
func (f *Manager) CloseRoom(id uint64) bool {
	//basic check
	if id <= 0 || f.rooms == nil {
		return false
	}
	f.rooms.Delete(id)
	if f.roomCount > 0 {
		atomic.AddInt32(&f.roomCount, -1)
	}
	return true
}

//get room
func (f *Manager) GetRoom(id uint64) iface.IRoom {
	//basic check
	if id <= 0 || f.rooms == nil {
		return nil
	}
	//check room
	v, ok := f.rooms.Load(id)
	if !ok {
		return nil
	}
	room, ok := v.(iface.IRoom)
	if !ok {
		return nil
	}
	return room
}

//add room
func (f *Manager) AddRoom(room iface.IRoom) bool {
	//basic check
	if room == nil {
		return false
	}
	//sync into map
	f.rooms.Store(room.GetId(), room)
	atomic.AddInt32(&f.roomCount, 1)
	return true
}

//////////////
//private func
//////////////

//run main process
func (f *Manager) runMainProcess() {
	var (
		timer = time.NewTicker(time.Second * define.RoomCheckRate)
	)

	//defer
	defer func() {
		//clean up
		timer.Stop()
		close(f.closeChan)
	}()

	//loop
	for {
		select {
		case <- timer.C:
			{
				//clean up rooms
				f.cleanUpRooms()
			}
		case <- f.closeChan:
			return
		}
	}
}

//clean up closed room
func (f *Manager) cleanUpRooms() {
	if f.roomCount <= 0 {
		return
	}
	sf := func(k, v interface{}) bool {
		room, ok := v.(iface.IRoom)
		if ok && room != nil {
			if room.IsOver() {
				//clean up
				f.rooms.Delete(k)
				if f.roomCount > 0 {
					atomic.AddInt32(&f.roomCount, -1)
				}
			}
		}
		return true
	}
	f.rooms.Range(sf)
}
