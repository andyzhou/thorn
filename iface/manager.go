package iface

/*
 * interface of manager
 */

type IManager interface {
	Close()
	GetRooms() int32
	CloseRoom(id uint64) bool
	GetRoom(id uint64) IRoom
	AddRoom(room IRoom) bool
}