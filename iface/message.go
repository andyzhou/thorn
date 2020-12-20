package iface

/*
 * interface of message
 */

type IMessage interface {
	GetId() uint64
	GetPacket() IPacket
}
