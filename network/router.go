package network

import (
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
	manager iface.IManager
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

func (f *Router) OnConnect(conn iface.IConn) bool {
	log.Println("Router:OnConnect")
	atomic.AddUint64(&f.totalConn, 1)
	return true
}


//room message process
func (f *Router) OnMessage(
					conn iface.IConn,
					packet iface.IPacket,
				) bool {
	//get message id
	messageId := pb.ID(packet.GetMessageId())
	log.Println("Router:OnMessage, messageId:", messageId)

	switch messageId {
	case pb.ID_MSG_Connect://connect
		{
			//unpack message
			msg := &pb.C2S_ConnectMsg{}
			if err := packet.UnmarshalPB(msg); nil != err {
				log.Printf("[router] msg.UnmarshalPB error=[%s]\n", err.Error())
				return false
			}

			//get relate id
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
				return false
			}

			//check room status
			if room.IsOver() {
				ret.ErrorCode = pb.ERROR_CODE_ERR_RoomState
				f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
				log.Printf("[router] room is over player=[%d] room==[%d] token=[%s]\n",
							playerId, roomId, token)
				return false
			}

			//check player
			if !room.HasPlayer(playerId) {
				ret.ErrorCode = pb.ERROR_CODE_ERR_NoPlayer
				f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
				log.Printf("[router] !room.HasPlayer(playerID) player=[%d] room==[%d] token=[%s]\n",
							playerId, roomId, token)
				return false
			}

			//verify token
			if !room.VerifyToken(token) {
				ret.ErrorCode = pb.ERROR_CODE_ERR_Token
				f.writePacket(conn, uint8(pb.ID_MSG_Connect), ret)
				log.Printf("[router] verifyToken failed player=[%d] room==[%d] token=[%s]\n",
							playerId, roomId, token)
				return false
			}

			//put extra data
			conn.SetExtraData(playerId)

			//callback of room
			bRet := room.OnConnect(conn)
			return bRet
		}

	case pb.ID_MSG_Heartbeat://heart beat
		{
			f.writePacket(conn, uint8(pb.ID_MSG_Heartbeat), nil)
			return true
		}

	case pb.ID_MSG_END://end
		{
			f.writePacket(conn, uint8(pb.ID_MSG_END), packet.GetData())
		}
	default:
		return false
	}

	return false
}

func (f *Router) OnClose(conn iface.IConn) {
	val := atomic.LoadUint64(&f.totalConn) - 1
	atomic.StoreUint64(&f.totalConn, val)
}

//////////////////
//private func
//////////////////

//async write packet
func (f *Router) writePacket(conn iface.IConn, msgId uint8, data interface{}) bool {
	if conn == nil {
		return false
	}
	conn.AsyncWritePacket(
		protocol.NewPacketWithPara(msgId, data),
		time.Microsecond,
	)
	return true
}