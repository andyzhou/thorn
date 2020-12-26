package room

import (
	"fmt"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/pb"
	"github.com/andyzhou/thorn/protocol"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

/*
 * game face, implement of IGame
 */

//game state
const (
	GameCreated = iota
	GameReady
	Gaming
	GameOver
	GameStop
)

//inter macro define
const (
	MaxReadyTime          int64  = 20            //max prepare time, force close if no any login
	MaxGameFrame          uint32 = 30*60*3 + 100 //max frames per game round
	BroadcastOffsetFrames        = 3             //cast per frames
	kMaxFrameDataPerMsg          = 60            //max message packet per frame
	kBadNetworkThreshold         = 2             //max time for no heart beat
)


//face info
type Game struct {
	id uint64
	startTime int64
	randSeed int32
	state int
	gl iface.IGameListener //original game listener
	logic iface.ILockStep
	players map[uint64]iface.IPlayer //player map
	frameCount uint32
	result map[uint64]uint64
	dirty bool
}

//construct
func NewGame(
			roomId uint64,
			players []uint64,
			randSeed int32,
			gl iface.IGameListener,
		) *Game {
	//self init
	this := &Game{
		id:roomId,
		randSeed:randSeed,
		gl:gl,
		logic:NewLockStep(),
		state:GameCreated,
		players:make(map[uint64]iface.IPlayer),
		result:make(map[uint64]uint64),
	}
	//init players
	for idx, v := range players {
		this.players[v] = NewPlayer(v, int32(idx + 1))
	}
	return this
}

//join game
func (f *Game) JoinGame(playerId uint64, conn iface.IConn) bool {
	//basic check
	if playerId <= 0 || conn == nil {
		return false
	}

	//check player
	p, ok := f.players[playerId]
	if !ok {
		return false
	}

	//init message
	msg := &pb.S2C_ConnectMsg{
		ErrorCode:pb.ERRORCODE_ERR_Ok.Enum(),
	}

	//check status
	if f.state != GameReady && f.state != Gaming {
		log.Printf("[game(%d)] player[%d] game is over\n", f.id, playerId)
		//reset msg
		msg.ErrorCode = pb.ERRORCODE_ERR_RoomState.Enum()

		//notify client
		p.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_Connect), msg))
		return false
	}

	//check conn
	if p.GetConn() != nil {
		//TODO 这里有多线程操作的危险 如果调 p.client.Close() 会把现有刚进来的玩家提调
		p.GetConn().SetExtraData(nil)
		log.Printf("[game(%d)] player[%d] replace\n", f.id, playerId)
	}

	//sync conn
	p.Connect(conn)

	//send message
	p.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_Connect), msg))

	//call cb of game listener
	f.gl.OnJoinGame(f.id, playerId)

	return true
}

//leave game
func (f *Game) LeaveGame(playerId uint64) bool {
	//basic check
	if playerId <= 0 {
		return false
	}

	//check player
	p, ok := f.players[playerId]
	if !ok {
		return false
	}

	//clean up
	p.CleanUp()

	//call cb of game listener
	f.gl.OnLeaveGame(f.id, playerId)

	return true
}

//process message
func (f *Game) ProcessMessage(playerId uint64, packet iface.IPacket) bool {
	//basic check
	if playerId <= 0 || packet == nil {
		return false
	}

	//check player
	player, ok := f.players[playerId]
	if !ok {
		return false
	}

	log.Printf("[game(%d)] processMsg player[%d] msg=[%d]\n",
				f.id, player.GetId(), packet.GetMessageId())

	//get message id
	messageId := pb.ID(packet.GetMessageId())

	//do relate opt by message id
	switch messageId {
	case pb.ID_MSG_JoinRoom://join room
		{
			msg := &pb.S2C_JoinRoomMsg{
				Roomseatid:proto.Int32(player.GetIdx()),
				RandomSeed:proto.Int32(f.randSeed),
			}
			//loop players
			for _, v := range f.players {
				if player.GetId() == v.GetId() {
					continue
				}
				msg.Others = append(msg.Others, v.GetId())
				msg.Pros = append(msg.Pros, v.GetProgress())
			}

			//notify player
			player.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_JoinRoom), msg))
		}
	case pb.ID_MSG_Progress://progress
		{
			if f.state > GameReady {
				break
			}

			//unzip packet
			msg := &pb.C2S_ProgressMsg{}
			if err := packet.UnmarshalPB(msg); err != nil {
				log.Printf("[game(%d)] processMsg player[%d] msg=[%d] UnmarshalPB error:[%s]\n",
						f.id, player.GetId(), packet.GetMessageId(), err.Error())
				return false
			}
			//set player progress
			player.SetProgress(msg.GetPro())
			packet := protocol.NewPacket(uint8(pb.ID_MSG_Progress),
						&pb.S2C_ProgressMsg{
							Id:proto.Uint64(player.GetId()),
							Pro:msg.Pro,
						})
			//broadcast
			f.broadcastExclude(packet, player.GetId())
		}

	case pb.ID_MSG_Heartbeat://heart beat
		{
			player.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_Heartbeat), nil))
			player.RefreshHeartbeatTime()
		}

	case pb.ID_MSG_Ready://ready
		{
			if f.state == GameReady {
				f.doReady(player)
			}else if f.state == Gaming {
				log.Printf("[game(%d)] doReconnect [%d]\n", f.id, player.GetId())
				f.doReady(player)
				f.doReconnect(player)
			}else{
				log.Printf("[game(%d)] ID_MSG_Ready player[%d] state error:[%d]\n",
					f.id, player.GetId(), f.state)
			}
		}

	case pb.ID_MSG_Input://input
		{
			msg := &pb.C2S_InputMsg{}
			if err := packet.UnmarshalPB(msg); nil != err {
				fmt.Printf("[game(%d)] processMsg player[%d] msg=[%d] UnmarshalPB error:[%s]\n",
							f.id, player.GetId(), packet.GetMessageId(), err.Error())
				return false
			}
			//push input
			if !f.pushInput(player, msg) {
				log.Printf("[game(%d)] processMsg player[%d] msg=[%d] pushInput failed\n",
							f.id, player.GetId(), packet.GetMessageId())
				break
			}

			//force broadcast frame
			f.dirty = true
		}

	case pb.ID_MSG_Result://result
		{
			msg := &pb.C2S_ResultMsg{}
			if err := packet.UnmarshalPB(msg); nil != err {
				log.Printf("[game(%d)] processMsg player[%d] msg=[%d] UnmarshalPB error:[%s]\n",
							f.id, player.GetId(), packet.GetMessageId(), err.Error())
				return false
			}

			//set result
			f.result[player.GetId()] = msg.GetWinnerID()
			log.Printf("[game(%d)] ID_MSG_Result player[%d] winner=[%d]\n",
						f.id, player.GetId(), msg.GetWinnerID())
			player.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_Result), nil))
		}
	default:
		{
			log.Printf("[game(%d)] processMsg unknown message id[%d]\n", int32(messageId))
		}
	}

	return true
}

//tick main logic
func (f *Game) Tick(now int64) bool {
	//do relate opt by game state
	switch f.state {
	case GameReady:
		{
			delta := now - f.startTime
			if delta < MaxReadyTime {
				if f.checkReady() {
					//start
					f.doStart()
					f.state = Gaming
				}
			}else{
				if f.getOnlinePlayerCount() > 0 {
					//up to ready time, if player online, force start
					f.doStart()
					f.state = Gaming
					fmt.Printf("[game(%d)] force start game because ready state is timeout\n", f.id)
				}else{
					//all not join game, force finished
					f.state = GameOver
					log.Printf("[game(%d)] game over!! nobody ready\n", f.id)
				}
			}
			return true
		}
	case Gaming:
		{
			if f.checkOver() {
				f.state = GameOver
				log.Printf("[game(%d)] game over successfully!!\n", f.id)
				return true
			}

			if f.isTimeOut() {
				f.state = GameOver
				log.Printf("[game(%d)] game timeout\n", f.id)
				return true
			}

			//other logic
			f.logic.Tick()
			f.broadcastFrameData()
			return true
		}
	case GameOver:
		{
			f.doGameOver()
			f.state = GameStop
			log.Printf("[game(%d)] do game over\n", f.id)
			return true
		}
	case GameStop:
		{
			return false
		}
	}
	return false
}

//get result
func (f *Game) GetResult() map[uint64]uint64 {
	return f.result
}

//close game
func (f *Game) Close() {
	packet := protocol.NewPacket(uint8(pb.ID_MSG_Close), nil)
	f.broadcast(packet)
}

//clean up
func (f *Game) CleanUp() {
	for _, v := range f.players {
		v.CleanUp()
	}
	f.players = make(map[uint64]iface.IPlayer)
}

////////////////
//private func
////////////////

//do ready
func (f *Game) doReady(p iface.IPlayer) {
	if p.IsReady() {
		return
	}

	//set status
	p.SetReady()

	//init message
	packet := protocol.NewPacket(uint8(pb.ID_MSG_Ready), nil)
	p.SendMessage(packet)
}

//check game is ready
func (f *Game) checkReady() bool {
	for _, v := range f.players {
		if !v.IsReady() {
			return false
		}
	}
	return true
}

//game start
func (f *Game) doStart() {
	//init for game start
	f.frameCount = 0
	f.logic.Reset()

	//init players
	for _, v := range f.players {
		v.SetReady()
		v.SetProgress(100)
	}

	//init message
	f.startTime = time.Now().Unix()

	msg := &pb.S2C_StartMsg{
		TimeStamp:proto.Int64(f.startTime),
	}
	packet := protocol.NewPacket(uint8(pb.ID_MSG_Start), msg)

	//broad cast to all
	f.broadcast(packet)

	//callback for game start
	f.gl.OnStartGame(f.id)
}

//game is over
func (f *Game) doGameOver() {
	f.gl.OneGameOver(f.id)
}

//push client input
func (f *Game) pushInput(p iface.IPlayer, msg *pb.C2S_InputMsg) bool {
	input := &pb.InputData{
		Id:		proto.Uint64(p.GetId()),
		Sid:	proto.Int32(msg.GetSid()),
		X:		proto.Int32(msg.GetX()),
		Y:		proto.Int32(msg.GetY()),
		Roomseatid:proto.Int32(p.GetIdx()),
	}
	f.logic.PushCommand(input)
	return true
}

//client reconnect
func (f *Game) doReconnect(p iface.IPlayer) bool {
	//init message
	msg := &pb.S2C_StartMsg{
		TimeStamp:proto.Int64(f.startTime),
	}
	packet := protocol.NewPacket(uint8(pb.ID_MSG_Start), msg)

	//send message
	p.SendMessage(packet)

	//process frame data
	framesCount := f.frameCount
	i := uint32(0)
	c := 0
	frameMsg := &pb.S2C_FrameMsg{}

	//loop
	for ; i <= framesCount; i++ {
		frameData := f.logic.GetFrame(i)
		if frameData == nil && i != (framesCount - 1) {
			continue
		}

		fd := &pb.FrameData{
			FrameID:proto.Uint32(i),
		}
		if frameData != nil {
			fd.Input = frameData.GetData()
		}
		frameMsg.Frames = append(frameMsg.Frames, fd)
		c++

		if c >= kMaxFrameDataPerMsg || i == (framesCount - 1) {
			p.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_Frame), frameMsg))
			c = 0
			frameMsg = &pb.S2C_FrameMsg{}
		}
	}

	//set send frame count
	p.SetSendFrameCount(f.frameCount)

	return true
}

//broad cast frame data
func (f *Game) broadcastFrameData() {
	//get frame count
	frameCount := f.logic.GetFrameCount()

	//check frame
	if !f.dirty && (frameCount - f.frameCount < BroadcastOffsetFrames) {
		return
	}

	now := time.Now().Unix()

	for _, p := range f.players {
		//check online
		if !p.IsOnline() {
			continue
		}

		//check status
		if !p.IsReady() {
			continue
		}

		//check network
		if now - p.GetLastHeartbeatTime() >= kBadNetworkThreshold {
			continue
		}

		//check player last frame
		i := p.GetSendFrameCount()
		c := 0
		msg := &pb.S2C_FrameMsg{}

		for ; i < frameCount; i++ {
			frameData := f.logic.GetFrame(i)
			if frameData == nil && i != (frameCount - 1) {
				continue
			}

			//init frame data
			fd := &pb.FrameData{
				FrameID:proto.Uint32(i),
			}
			if frameData != nil {
				fd.Input = frameData.GetData()
			}
			msg.Frames = append(msg.Frames, fd)
			c++

			//if last frame or up to max frame, send them
			if i == (frameCount - 1) || c >= kMaxFrameDataPerMsg {
				p.SendMessage(protocol.NewPacket(uint8(pb.ID_MSG_Frame), msg))
				c = 0
				msg = &pb.S2C_FrameMsg{}
			}
		}

		//set frame count
		p.SetSendFrameCount(frameCount)
	}
}

//broad cast
func (f *Game) broadcast(packet iface.IPacket) {
	for _, v := range f.players {
		v.SendMessage(packet)
	}
}

//broad cast exclude
func (f *Game) broadcastExclude(msg iface.IPacket, id uint64) {
	for _, v := range f.players {
		if v.GetId() == id {
			continue
		}
		v.SendMessage(msg)
	}
}

//get one player
func (f *Game) getPlayer(id uint64) iface.IPlayer {
	v, ok := f.players[id]
	if !ok {
		return nil
	}
	return v
}

//get player count
func (f *Game) getPlayerCount() int {
	return len(f.players)
}

//get online player count
func (f *Game) getOnlinePlayerCount() int {
	num := 0
	for _, v := range f.players {
		if v.IsOnline() {
			num++
		}
	}
	return num
}

//check over
func (f *Game) checkOver() bool {
	//if some one online and get result, will not over
	for _, v := range f.players {
		if !v.IsOnline() {
			continue
		}
		_, ok := f.result[v.GetId()]
		if !ok {
			return false
		}
	}
	return true
}

//is time out
func (f *Game) isTimeOut() bool {
	return f.logic.GetFrameCount() > MaxGameFrame
}