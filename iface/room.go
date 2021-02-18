package iface

/*
 * interface of room
 */

type IRoom interface {
	Stop()
	GetId() uint64
	GetSecretKey() string
	IsOver() bool
	HasPlayer(id uint64) bool
	VerifyToken(string) bool
	IGameListener
	IConnCallBack
}

//call back for room
//api client should implement this
type IRoomCallback interface {
	IGameListener
	IConnCallBack
}