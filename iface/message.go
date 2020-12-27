package iface

import "github.com/golang/protobuf/proto"

/*
 * interface of message
 */

type IMessage interface {
	//unpack pb
	UnmarshalPB(msg proto.Message) error

	//get
	GetId() uint32
	GetPlayerId() uint64
	GetLen() uint32
	GetData() []byte

	//set
	SetId(id uint32)
	SetPlayerId(id uint64)
	SetLen(len uint32)
	SetData(data []byte)
}
