package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/andyzhou/thorn/iface"
	"github.com/golang/protobuf/proto"
)

/*
 * packet data face, implement of IPacket
 * packet for server send to client
 */

/*
s->c
|--totalDataLen(uint16)--|--msgIDLen(uint8)--|--------data-----|
|-------------4----------|---------4---------|--(totalDataLen-4)--|
*/

//inter macro define
//const (
//	DataLen      = 2
//	MessageIDLen = 1
//
//	MinPacketLen = DataLen + MessageIDLen
//	MaxPacketLen = (2 << 8) * DataLen
//	MaxMessageID = (2 << 8) * MessageIDLen
//)

const (
	PacketHeadSize = 8 //dataLen(4byte) + messageId(4byte)
	PacketMaxSize = 4096 //4KB
)

//data info
type Packet struct {
	id uint32 //message id
	len uint32
	data []byte
}

type PlayerPacket struct {
	id uint64 //player id
	packet iface.IPacket //original packet
}

//////////////////////////
//api for player packet
//////////////////////////

//construct
func NewPlayerPacket() *PlayerPacket {
	//self init
	this := &PlayerPacket{}
	return this
}

func (f *PlayerPacket) GetId() uint64 {
	return f.id
}

func (f *PlayerPacket) GetPacket() iface.IPacket {
	return f.packet
}

func (f *PlayerPacket) SetId(id uint64) {
	f.id = id
}

func (f *PlayerPacket) SetPacket(packet iface.IPacket) {
	f.packet = packet
}

//////////////////////////
//api for original packet
//////////////////////////

//construct
func NewPacket() *Packet {
	//self init
	this := &Packet{}
	return this
}

func NewPacketWithPara(
			id uint32,
			data interface{},
		) *Packet {
	var (
		orgData []byte
		isOk bool
		err error
	)

	//self init
	this := &Packet{
		id:id,
	}

	//process data
	orgData = make([]byte, 0)
	switch v := data.(type) {
	case []byte:
		{
			orgData, isOk = data.([]byte)
			if !isOk {
				return nil
			}
		}
	case proto.Message:
		orgData, err = proto.Marshal(v)
		if err != nil {
			return nil
		}
	default:
		return nil
	}

	//get length
	this.len = uint32(len(orgData))

	return this
}

//unpack header
func (f *Packet) UnPackHead(data []byte) (iface.IMessage, error)  {
	var (
		messageId uint32
		messageLen uint32
		err error
	)

	//basic check
	if data == nil || len(data) <= 0 {
		return nil, errors.New("invalid parameter")
	}

	//init data buff
	dataBuff := bytes.NewReader(data)

	//read length
	err = binary.Read(dataBuff, binary.LittleEndian, &messageLen)
	if err != nil {
		return nil, err
	}

	//read message id
	err = binary.Read(dataBuff, binary.LittleEndian, &messageId)
	if err != nil {
		return nil, err
	}

	//read data
	if messageLen > PacketMaxSize {
		tips := fmt.Sprintf("too large message data received, message length:%d", messageLen)
		return nil, errors.New(tips)
	}

	//init message data
	message := NewMessage()
	message.SetId(messageId)
	message.SetLen(messageLen)

	return message, nil
}

//pack data
func (f *Packet) Pack() []byte {
	var (
		err error
	)

	//basic check
	if f.id < 0 || f.data == nil {
		return nil
	}

	//init data buff
	dataBuff := bytes.NewBuffer(nil)

	//write length
	err = binary.Write(dataBuff, binary.BigEndian, f.len)
	if err != nil {
		return nil
	}

	//write message id
	err = binary.Write(dataBuff, binary.BigEndian, f.id)
	if err != nil {
		return nil
	}

	//write data
	err = binary.Write(dataBuff, binary.BigEndian, f.data)
	if err != nil {
		return nil
	}

	return dataBuff.Bytes()
}

func (f *Packet) UnmarshalPB(msg proto.Message) error {
	return proto.Unmarshal(f.data, msg)
}

func (f *Packet) GetData() []byte {
	return f.data
}

func (f *Packet) GetId() uint32 {
	return f.id
}

func (f *Packet) GetHeadLen() uint32 {
	return PacketHeadSize
}