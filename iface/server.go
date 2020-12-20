package iface

import (
	"github.com/andyzhou/thorn/network"
	"net"
)

/*
 * interface of server
 */

type IServer interface {
	Stop()
	Start(listener net.Listener)
	GetProtocol() IProtocol
	GetConfig() *network.Config
}