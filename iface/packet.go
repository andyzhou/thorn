package iface

import "github.com/golang/protobuf/proto"

/*
 * interface of packet
 */

type IPacket interface {
	GetMessageId() uint8
	GetData() []byte
	Serialize() []byte
	UnmarshalPB(msg proto.Message) error
}