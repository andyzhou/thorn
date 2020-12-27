package protocol

import (
	"errors"
	"github.com/andyzhou/thorn/iface"
	"io"
	"log"
)

/*
 * protocol face, implement of IProtocol
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
	header := make([]byte, PacketHeadSize)

	//read header
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, err
	}

	//unpack header
	packet := NewPacket()
	message, err := packet.UnPackHead(header)
	if err != nil {
		return nil, err
	}

	//read real data and storage into message object
	if message.GetLen() <= 0 {
		return nil, errors.New("message len is zero")
	}

	data := make([]byte, message.GetLen())
	_, err = io.ReadFull(reader, data)
	if err != nil {
		log.Println("read data failed, err:", err.Error())
		return nil, err
	}

	//set packet message id
	packet.id = message.GetId()

	//set packet message
	packet.data = data

	return packet, nil
}
