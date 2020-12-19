package iface

/*
 * interface of manager
 */

type IManager interface {
	Close()
	GetRoom(id uint64) IRoom
	AddRoom(room IRoom) bool
}