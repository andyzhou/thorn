package main

import (
	"github.com/andyzhou/thorn/iface"
	"log"
)

/*
 * call back for room, implement of IRoomCallback
 */

//face info
type RoomCallBack struct {
}

//construct
func NewRoomCallBack()  *RoomCallBack {
	//self init
	this := &RoomCallBack{
	}
	return this
}

//implement of IConnCallBack
func (f *RoomCallBack) OnJoinGame(roomId, playerId uint64) {
	log.Println("RoomCallBack:OnJoinGame")
}

func (f *RoomCallBack)  OnStartGame(roomId uint64) {
	log.Println("RoomCallBack:OnStartGame")
}

func (f *RoomCallBack)  OnLeaveGame(roomId, playerId uint64) {
	log.Println("RoomCallBack:OnLeaveGame")
}

func (f *RoomCallBack)  OneGameOver(roomId uint64) {
	log.Println("RoomCallBack:OneGameOver")
}


//implement of IGameListener
//cb for connected
func (f *RoomCallBack) OnConnect(conn iface.IConn) bool {
	log.Println("RoomCallBack:OnConnect")
	return true
}

//cb for received packet
func (f *RoomCallBack) OnMessage(conn iface.IConn, packet iface.IPacket) bool {
	log.Println("RoomCallBack:OnMessage")
	return true
}

//cb for closed conn
func (f *RoomCallBack) OnClose(conn iface.IConn) {
	log.Println("RoomCallBack:OnClose")
}
