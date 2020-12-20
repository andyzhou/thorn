package protocol

import (
	"encoding/binary"
	"errors"
	"github.com/andyzhou/thorn/iface"
	"io"
)

/*
 * protocol face, implement of IProtocol
 */

//face info
type Protocol struct {
}

//construct
func NewProcol() *Protocol {
	//self init
	this := &Protocol{
	}
	return this
}

//read packet
func (f *Protocol) ReadPacket(reader io.Reader) (iface.IPacket, error) {
	buff := make([]byte, MinPacketLen, MinPacketLen)
	_, err := io.ReadFull(reader, buff)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint16(buff)
	if dataLen > MaxPacketLen {
		return nil, errors.New("packet data too max")
	}

	//init packet
	packet := NewPacket()

	//set id
	packet.id = buff[dataLen]

	//set data
	if dataLen > 0 {
		packet.data = make([]byte, dataLen, dataLen)
		if _, err := io.ReadFull(reader, packet.data); err != nil {
			return nil, err
		}
	}
	return packet, nil
}
