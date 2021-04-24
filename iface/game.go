package iface

/*
 * interface of game
 */

type IGameListener interface {
	OnJoinGame(conn IConn, roomId, playerId uint64)
	OnStartGame(roomId uint64)
	OnLeaveGame(roomId, playerId uint64)
	OneGameOver(roomId uint64)
}

type IGame interface {
	Close()
	GetResult() map[uint64]uint64
	Tick(int64) bool
	ProcessMessage(playerId uint64, packet IPacket) bool
	JoinGame(playerId uint64, conn IConn) bool
	LeaveGame(playerId uint64) bool
}
