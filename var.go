package thorn

import "net"

/*
 * shared variable or struct
 * most used for callback
 */

//connect info
type VThornConn struct {
	conn net.Conn
	roomId uint64
	playerId uint64
}

