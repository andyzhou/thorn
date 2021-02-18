package protocol

import (
	"encoding/binary"
	"errors"
	"github.com/andyzhou/thorn/iface"
	"io"
)

/*
 * protocol face, implement of IProtocol
 * - read and un-packet data from client side
 */

//face info
type Protocol struct {
}

//construct
func NewProtocol() *Protocol {
	//self init
	this := &Protocol{
	}
	return this
}

//read packet
func (f *Protocol) ReadPacket(reader io.Reader) (iface.IPacket, error) {
	//init header
	buff := make([]byte, MinPacketLen, MinPacketLen)

	//read header
	_, err := io.ReadFull(reader, buff)
	if err != nil {
		return nil, err
	}

	//unpack header
	//get data length
	dataLen := binary.BigEndian.Uint16(buff)
	if dataLen > MaxPacketLen {
		return nil, errors.New("data too max")
	}

	//message id
	p := &Packet{
		id:buff[DataLen],
	}

	//data
	if dataLen > 0 {
		p.data = make([]byte, dataLen, dataLen)
		_, err = io.ReadFull(reader, p.data)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}
