package iface

/*
 * interface of player
 */

type IPlayer interface {
	CleanUp()
	GetConn() IConn
	Connect(conn IConn)
	IsOnline() bool
	RefreshHeartbeatTime()
	GetLastHeartbeatTime() int64
	SetSendFrameCount(c uint32)
	GetSendFrameCount() uint32
	SendMessage(packet IPacket)
}