package iface

/*
 * interface of packet
 */

type IPacket interface {
	GetMessageId() uint8
	GetData() []byte
	Serialize() []byte
}