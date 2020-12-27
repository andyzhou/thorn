package protocol

import "github.com/golang/protobuf/proto"

/*
 * message data face, implement of IMessage
 */

//data info
type Message struct {
	id uint32 //message id
	playerId uint64
	len uint32
	data []byte
}

//construct
func NewMessage() *Message {
	//self init
	this := &Message{
		data:make([]byte, 0),
	}
	return this
}

//get
func (f *Message) GetId() uint32 {
	return f.id
}

func (f *Message) GetPlayerId() uint64 {
	return f.playerId
}

func (f *Message) GetLen() uint32 {
	return f.len
}

func (f *Message) GetData() []byte {
	return f.data
}

//set
func (f *Message) SetId(id uint32) {
	f.id = id
}

func (f *Message) SetPlayerId(id uint64) {
	f.playerId = id
}

func (f *Message) SetLen(len uint32) {
	f.len = len
}

func (f *Message) SetData(data []byte) {
	f.data = data
	f.len = uint32(len(data))
}

func (f *Message) UnmarshalPB(msg proto.Message) error {
	return proto.Unmarshal(f.data, msg)
}
