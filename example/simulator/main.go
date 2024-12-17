package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/pb"
	"github.com/andyzhou/thorn/protocol"
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
	SecretKey = "testRoom"
)

func main()  {
	var (
		m any = nil
	)
	wg := new(sync.WaitGroup)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Println("panic happened, err:", err)
			wg.Done()
		}
	}()

	wg.Add(1)

	key := pbkdf2.Key([]byte(Password), []byte(Salt), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	//create batch players
	playerIds := []uint64{
		1,
		2,
	}
	for _, playerId := range playerIds {
		createPlayer(playerId, &block)
	}
	wg.Wait()
}

//create one player
func createPlayer(playerId uint64, block *kcp.BlockCrypt) {
	//dial server
	session, err := kcp.DialWithOptions(UdpServerAddr, *block, 10, 3)
	if err != nil {
		log.Println("connect server failed, err:", err)
		return
	}
	log.Println("connect server success")
	go runMainProcess(session, playerId)
}

//read packet
func readPacket()  {

}

//write packet
func writePacket(sess *kcp.UDPSession, packet iface.IPacket) bool {
	if sess == nil || packet == nil {
		return false
	}
	_, err := sess.Write(packet.Pack())
	if err != nil {
		log.Println("writePacket failed, err:", err)
		return false
	}
	log.Println("write packet success")
	return true
}

//gen connect room packet
func genConnRoomPacket(roomId, playerId uint64, token string) *protocol.Packet {
	msg := &pb.C2S_ConnectMsg{
		BattleID:roomId,
		PlayerID:playerId,
		Token:token,
	}
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Connect), msg)
	return packet
}

//gen join room packet
func genJoinRoomPacket() *protocol.Packet {
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_JoinRoom), nil)
	return packet
}

//gen room progress packet
func genRoomProgressPacket(progress int32) *protocol.Packet {
	msg := &pb.C2S_ProgressMsg {
		Pro:progress,
	}
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Progress), msg)
	return packet
}

//gen heart beat packet
func genHeartBeatPacket() *protocol.Packet {
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Heartbeat), nil)
	return packet
}

//main process
func runMainProcess(sess *kcp.UDPSession, playerId uint64) {
	fmt.Println("connect server success")

	//set up
	roomId := uint64(1)
	token := SecretKey
	progress := int32(1)
	//maxProgress := 10

	defer func() {
		if sess != nil {
			sess.Close()
		}
	}()

	//send connect packet
	packet := genConnRoomPacket(roomId, playerId, token)
	if packet != nil {
		writePacket(sess, packet)
	}

	//send join room packet
	packet = genJoinRoomPacket()
	if packet != nil {
		writePacket(sess, packet)
	}

	//loop
	for {
		//gen heart beat packet
		packet = genHeartBeatPacket()
		if packet != nil {
			writePacket(sess, packet)
		}

		//gen progress packet
		packet = genRoomProgressPacket(progress)
		if packet != nil {
			writePacket(sess, packet)
		}
		time.Sleep(time.Second/10)
		progress++
		//if maxProgress > 0 {
		//	break
		//}
	}
}
