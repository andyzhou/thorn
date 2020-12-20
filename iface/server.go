package iface

import "net"

/*
 * interface of server
 */

type IServer interface {
	Stop()
	Start(listener net.Listener)
}