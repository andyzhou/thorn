package iface

/*
 * interface of room
 */

type IRoom interface {
	Stop()
	Start()
	GetId() uint64
	GetSecretKey() string
	GetTimeStamp() int64
	IsOver() bool
	HasPlayer(id uint64) bool
	VerifyToken(string) bool
	IGameListener
	IConnCallBack
}
