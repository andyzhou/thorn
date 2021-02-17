package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/pb"
	"github.com/andyzhou/thorn/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"sync"
	"time"
)

/*
 * client simulator
 */


//inter macro define
const (
	UdpServerAddr = "127.0.0.1:6100"
	Password = "test"
	Salt = "abc"
)

func main()  {
	wg := new(sync.WaitGroup)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic happened, err:", err)
			wg.Done()
		}
	}()

	wg.Add(1)

	key := pbkdf2.Key([]byte(Password), []byte(Salt), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	//dial server
	client, err := kcp.DialWithOptions(UdpServerAddr, block, 10, 3)
	if err != nil {
		log.Println("connect server failed, err:", err)
		return
	}
	log.Println("connect server success")
	go runMainProcess(client)
	wg.Wait()
}

//read packet
func readPacket()  {

}

//write packet
func writePacket(sess *kcp.UDPSession, packet iface.IPacket) bool {
	if sess == nil || packet == nil {
		return false
	}
	_, err := sess.Write(packet.Serialize())
	if err != nil {
		log.Println("writePacket failed, err:", err)
		return false
	}
	log.Println("write packet success")
	return true
}

//gen connect room packet
func genConnRoomPacket(roomId, playerId uint64) *protocol.Packet {
	msg := &pb.C2S_ConnectMsg{
		BattleID:proto.Uint64(roomId),
		PlayerID:proto.Uint64(playerId),
	}
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Connect), msg)
	return packet
}

//main process
func runMainProcess(sess *kcp.UDPSession) {
	fmt.Println("connect server success")

	roomId := uint64(1)
	playerId := uint64(1)
	//status := 0
	//
	////get packet
	//packet := genConnRoomPacket(roomId, playerId)
	//byteData := packet.Serialize()
	//if byteData != nil {
	//	log.Println(byteData)
	//}
	//
	//return
	for {
		//send login packet
		packet := genConnRoomPacket(roomId, playerId)
		if packet != nil {
			writePacket(sess, packet)
		}
		time.Sleep(time.Second)
	}
}