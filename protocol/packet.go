package protocol

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"log"
)

/*
 * packet data face, implement of IPacket
 */

/*
s->c
|--totalDataLen(uint16)--|--msgIDLen(uint8)--|--------data-----|
|-------------2----------|---------1---------|--(totalDataLen-2-1)--|
*/

//inter macro define
const (
	DataLen      = 2
	MessageIDLen = 1


	MinPacketLen = DataLen + MessageIDLen
	MaxPacketLen = (2 << 8) * DataLen
	MaxMessageID = (2 << 8) * MessageIDLen
)

//data info
type Packet struct {
	id uint8
	data []byte
}

//construct
func NewPacket(
			id uint8,
			msg interface{},
		) *Packet {
	this := &Packet{
		id:id,
		data:make([]byte, 0),
	}

	switch v := msg.(type) {
	case []byte:
		this.data = v
	case proto.Message:
		if mashData, err := proto.Marshal(v); err == nil {
			this.data = mashData
		}else{
			log.Printf("[NewPacket] proto marshal msg: %d error: %v\n",
				id, err)
			return nil
		}
	default:
		log.Printf("[NewPacket] error msg type msg: %d\n", id)
		return nil
	}

	return this
}

func (f *Packet) GetMessageId() uint8 {
	return f.id
}

func (f *Packet) GetData() []byte {
	return f.data
}

func (f *Packet) Serialize() []byte {
	buff := make([]byte, MinPacketLen, MinPacketLen)
	dataLen := len(f.data)
	binary.BigEndian.PutUint16(buff, uint16(dataLen))
	buff[dataLen] = f.id
	buff = append(buff, f.data...)
	return buff
}

func (f *Packet) UnmarshalPB(msg proto.Message) error {
	return proto.Unmarshal(f.data, msg)
}
