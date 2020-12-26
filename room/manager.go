package room

import (
	"github.com/andyzhou/thorn/iface"
	"sync"
	"sync/atomic"
)

/*
 * manager face, implement of IManager
 */

//face info
type Manager struct {
	roomCount int32
	rooms *sync.Map
}

//construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		rooms:new(sync.Map),
		roomCount:0,
	}
	return this
}

//close
func (f *Manager) Close() {
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
	atomic.AddInt32(&f.roomCount, -1)
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

