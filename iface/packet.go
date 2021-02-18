package iface

import "github.com/golang/protobuf/proto"

/*
 * interface of packet
 */

type IPacket interface {
	Pack() []byte
	UnmarshalPB(msg proto.Message) error

	//get
	GetMessageId() uint8
	GetData() []byte

	//set
	SetMessageId(uint8)
	SetData([]byte)
}

type IPlayerPacket interface {
	//get
	GetId() uint64
	GetPacket() IPacket

	//set
	SetId(id uint64)
	SetPacket(packet IPacket)
}