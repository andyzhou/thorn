package room

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/pb"
	"github.com/andyzhou/thorn/protocol"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
	"log"
	"sync/atomic"
	"time"
)

/*
 * room router face, implement of IConnCallBack
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

	switch messageId {
	case pb.ID_MSG_Connect://connect
		{
			//unpack message
			msg := &pb.C2S_ConnectMsg{}
			if err := packet.UnmarshalPB(msg); nil != err {
				fmt.Printf("[router] msg.UnmarshalPB error=[%s]\n", err.Error())
				return false
			}

			//get relate id
			playerId := msg.GetPlayerID()
			roomId := msg.GetBattleID()
			token := msg.GetToken()

			//ret message
			ret := &pb.S2C_ConnectMsg{
				ErrorCode:pb.ERRORCODE_ERR_Ok.Enum(),
			}

			//get room
			room := f.manager.GetRoom(roomId)
			if room == nil {
				ret.ErrorCode = pb.ERRORCODE_ERR_NoRoom.Enum()
				conn.AsyncWritePacket(protocol.NewPacket(uint8(pb.ID_MSG_Connect), ret), time.Millisecond)
				log.Printf("[router] no room player=[%d] room=[%d] token=[%s]\n",
							playerId, roomId, token)
				return false
			}

			//check room status
			if room.IsOver() {
				ret.ErrorCode = pb.ERRORCODE_ERR_RoomState.Enum()
				conn.AsyncWritePacket(protocol.NewPacket(uint8(pb.ID_MSG_Connect), ret), time.Millisecond)
				log.Printf("[router] room is over player=[%d] room==[%d] token=[%s]\n",
							playerId, roomId, token)
				return false
			}

			//check player
			if !room.HasPlayer(playerId) {
				ret.ErrorCode = pb.ERRORCODE_ERR_NoPlayer.Enum()
				conn.AsyncWritePacket(protocol.NewPacket(uint8(pb.ID_MSG_Connect), ret), time.Millisecond)
				log.Printf("[router] !room.HasPlayer(playerID) player=[%d] room==[%d] token=[%s]\n",
							playerId, roomId, token)
				return false
			}

			//verify token
			if !room.VerifyToken(token) {
				ret.ErrorCode = pb.ERRORCODE_ERR_Token.Enum()
				conn.AsyncWritePacket(protocol.NewPacket(uint8(pb.ID_MSG_Connect), ret), time.Millisecond)
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
			conn.AsyncWritePacket(
					protocol.NewPacket(uint8(pb.ID_MSG_Heartbeat), nil),
					time.Microsecond,
				)
			return true
		}

	case pb.ID_MSG_END://end
		{
			conn.AsyncWritePacket(
					protocol.NewPacket(uint8(pb.ID_MSG_END), packet.GetData()),
					time.Microsecond,
				)
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