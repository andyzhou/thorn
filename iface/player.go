package iface

/*
 * interface of player
 */

type IPlayer interface {
	CleanUp()
	GetId() uint64
	GetIdx() int32
	GetConn() IConn
	Connect(conn IConn)
	IsOnline() bool
	IsReady() bool
	SetReady()
	GetProgress() int32
	SetProgress(int32)
	RefreshHeartbeatTime()
	GetLastHeartbeatTime() int64
	SetSendFrameCount(c uint32)
	GetSendFrameCount() uint32
	SendMessage(packet IPacket)
}