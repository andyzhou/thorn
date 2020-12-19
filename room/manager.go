package room

import (
	"github.com/andyzhou/thorn/iface"
	"sync"
)

/*
 * manager face, implement of IManager
 */

//face info
type Manager struct {
	rooms *sync.Map
}

//construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		rooms:new(sync.Map),
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

//add room
func (f *Manager) AddRoom(room iface.IRoom) bool {
	//basic check
	if room == nil {
		return false
	}
	//sync into map
	f.rooms.Store(room.GetId(), room)
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