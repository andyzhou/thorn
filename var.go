package thorn

import "net"

/*
 * shared variable or struct
 * most used for callback
 */

//server conf
type ServerConf struct {
	Host     string
	Port     int
	Password string
	Salt     string
}

//connect info
type VThornConn struct {
	conn     net.Conn
	roomId   uint64
	playerId uint64
}

