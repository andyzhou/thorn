package network

import (
	"errors"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/pb"
	"github.com/andyzhou/thorn/protocol"
	"log"
	"sync/atomic"
	"time"
)

/*
 * room router face, implement of IConnCallBack
 * - router for udp protocol
 */

//face info
type Router struct {
	manager   iface.IManager //reference
	totalConn uint64
}

//construct
func NewRouter(manager iface.IManager) *Router {
	//self init
	this := &Router{
		manager:manager,
	}
	return this
}

//cb for connected
func (f *Router) OnConnect(conn iface.IConn) bool {
	if conn == nil {
		return false
	}
	atomic.AddUint64(&f.totalConn, 1)
	return true
}

//room message process
func (f *Router) OnMessage(
		conn iface.IConn,
		packet iface.IPacket,
	) bool {
	var (
		err error
	)

	//get message id
	messageId := pb.ID(packet.GetMessageId())

	//do some opt by message id
	switch messageId {
	case pb.ID_MSG_Connect://connect
		{
			//process connect message
			err = f.processConnMessage(conn, packet)
		}

	case pb.ID_MSG_Heartbeat://heart beat
		{
			err = f.writePacket(conn, uint8(pb.ID_MSG_Heartbeat), nil)
		}

	case pb.ID_MSG_END://end
		{
			err = f.writePacket(conn, uint8(pb.ID_MSG_END), packet.GetData())
		}
	}

	return err == nil
}

//cb for connect closed
func (f *Router) OnClose(conn iface.IConn) {
	val := atomic.LoadUint64(&f.totalConn) - 1
	atomic.StoreUint64(&f.totalConn, val)
}

//////////////////
//private func
//////////////////

//process connect message
func (f *Router) processConnMessage(
		conn iface.IConn,
		packet iface.IPacket,
	) error {
	//check
	if conn == nil || packet == nil {
		return errors.New("invalid parameter")
	}

	//unpack connect message
	msg := &pb.C2S_ConnectMsg{}
	if err := packet.UnmarshalPB(msg); nil != err {
		log.Printf("[router] msg.UnmarshalPB error=[%s]\n", err.Error())
		return err
	}

	//get key data
	playerId := msg.GetPlayerID()
	roomId := msg.GetBattleID()
	token := msg.GetToken()

	//ret message
	ret := &pb.S2C_ConnectMsg{
		ErrorCode:pb.ERROR_CODE_ERR_Ok,
	}

	//get room
	room := f.manager.GetRoom(roomId)
	if room == nil {
		ret.ErrorCode = pb.ERROR_CODE_ERR_NoRoom
		f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
		log.Printf("[router] no room player=[%d] room=[%d] token=[%s]\n",
					playerId, roomId, token)
		return errors.New("can't get room by id")
	}

	//check room status
	if room.IsOver() {
		ret.ErrorCode = pb.ERROR_CODE_ERR_RoomState
		f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
		log.Printf("[router] room is over player=[%d] room==[%d] token=[%s]\n",
					playerId, roomId, token)
		return errors.New("room is over")
	}

	//check player
	if !room.HasPlayer(playerId) {
		ret.ErrorCode = pb.ERROR_CODE_ERR_NoPlayer
		f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
		log.Printf("[router] !room.HasPlayer(playerID) player=[%d] room==[%d] token=[%s]\n",
					playerId, roomId, token)
		return errors.New("room no such player id")
	}

	//verify token
	if !room.VerifyToken(token) {
		ret.ErrorCode = pb.ERROR_CODE_ERR_Token
		f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
		log.Printf("[router] verifyToken failed player=[%d] room==[%d] token=[%s]\n",
					playerId, roomId, token)
		return errors.New("invalid room token")
	}

	//put extra data
	conn.SetExtraData(playerId)

	//call cb of room
	room.OnConnect(conn)
	return nil
}

//async write packet
func (f *Router) writePacket(
		conn iface.IConn,
		msgId uint8,
		data interface{},
	) error {
	if conn == nil {
		return errors.New("invalid parameter")
	}
	err := conn.AsyncWritePacket(
		protocol.NewPacketWithPara(msgId, data),
		time.Microsecond,
	)
	return err
}