package iface

import "github.com/golang/protobuf/proto"

/*
 * interface of packet
 */

type IPacket interface {
	UnPackHead(data []byte) (IMessage, error)
	Pack() []byte
	UnmarshalPB(msg proto.Message) error
	GetId() uint32
	GetHeadLen() uint32
	GetData() []byte
}

type IPlayerPacket interface {
	//get
	GetId() uint64
	GetPacket() IPacket

	//set
	SetId(id uint64)
	SetPacket(packet IPacket)
}