package protocol

import "github.com/andyzhou/thorn/iface"

/*
 * message data face, implement of IMessage
 */

//data info
type Message struct {
	id uint64 //player id
	packet iface.IPacket
}

//construct
func NewMessage(
			id uint64,
			packet iface.IPacket,
		) *Message {
	//self init
	this := &Message{
		id:id,
		packet:packet,
	}
	return this
}

func (f *Message) GetId() uint64 {
	return f.id
}

func (f *Message) GetPacket() iface.IPacket {
	return f.packet
}