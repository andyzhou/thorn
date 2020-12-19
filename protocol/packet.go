package protocol

import "encoding/binary"

/*
 * packet face, implement of IPacket
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

//face info
type Packet struct {
	id uint8
	data []byte
}

//construct
func NewPacket() *Packet {
	this := &Packet{
		data:make([]byte, 0),
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
