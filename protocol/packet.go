package protocol

import (
	"encoding/binary"
	"github.com/andyzhou/thorn/iface"
	"github.com/golang/protobuf/proto"
	"log"
)

/*
 * packet data face, implement of IPacket
 * packet for server send to client
 */

/*
s->c
|--totalDataLen(uint16)--|--msgIDLen(uint8)--|--------------data--------------|
|-------------2----------|---------1---------|---------(totalDataLen-2-1)-----|
*/

//inter macro define
const (
	DataLen       = 2
	MessageIdLen  = 1
	MinPacketLen  = DataLen + MessageIdLen
	MaxPacketLen  = (2 << 8) * DataLen
	PacketMaxSize = 4096 //4KB
)

//data info
type Packet struct {
	id   uint8 //message id
	data []byte
}

type PlayerPacket struct {
	id     uint64        //player id
	packet iface.IPacket //original packet
}

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
		id uint8,
		data interface{},
	) *Packet {
	//self init
	p := &Packet{
		id:id,
	}

	//process data
	switch v := data.(type) {
	case []byte:
		{
			p.data = v
		}
	case proto.Message:
		orgData, err := proto.Marshal(v)
		if err == nil {
			p.data = orgData
		}else{
			log.Println("NewPacketWithPara failed, err:", err)
			return nil
		}
	case nil:
		{
			//do nothing
		}
	default:
		log.Println("NewPacketWithPara error type:", id)
		return nil
	}

	return p
}

//pack data
func (f *Packet) Pack() []byte {
	//basic check
	if f.id < 0 {
		return nil
	}

	//init data buff
	//dataBuff := bytes.NewBuffer(nil)
	dataBuff := make([]byte, MinPacketLen, MinPacketLen)

	//write length
	dataLen := len(f.data)
	binary.BigEndian.PutUint16(dataBuff, uint16(dataLen))

	//write message id
	dataBuff[DataLen] = f.id

	//write data
	dataBuff = append(dataBuff, f.data...)

	//return dataBuff.Bytes()
	return dataBuff
}

func (f *Packet) UnmarshalPB(msg proto.Message) error {
	return proto.Unmarshal(f.data, msg)
}

//get
func (f *Packet) GetData() []byte {
	return f.data
}

func (f *Packet) GetMessageId() uint8 {
	return f.id
}

//set
func (f *Packet) SetMessageId(messageId uint8) {
	f.id = messageId
}

func (f *Packet) SetData(data []byte) {
	f.data = data
}