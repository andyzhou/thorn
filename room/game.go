package room

import (
	"fmt"
	"github.com/andyzhou/thorn/define"
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/pb"
	"github.com/andyzhou/thorn/protocol"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * game face, implement of IGame
 */

//face info
type Game struct {
	id uint64 //room id
	startTime int64
	randSeed int32
	state int
	gl iface.IGameListener //original game listener
	logic iface.ILockStep
	players sync.Map //player map, playerId -> IPlayer
	playerCount int32
	frameCount uint32
	result map[uint64]uint64
	dirty bool
	sync.RWMutex
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
		startTime:time.Now().Unix(),
		logic:NewLockStep(),
		players:sync.Map{},
		result:make(map[uint64]uint64),
	}
	//init players
	for idx, v := range players {
		player := NewPlayer(v, int32(idx + 1))
		this.players.Store(v, player)
	}
	return this
}

//player join game
func (f *Game) JoinGame(playerId uint64, conn iface.IConn) bool {
	//check
	if playerId <= 0 ||
		conn == nil ||
		reflect.ValueOf(conn).IsNil() {
		return false
	}

	//check player
	player := f.getPlayer(playerId)
	if player == nil {
		return false
	}

	//init message
	msg := &pb.S2C_ConnectMsg{
		ErrorCode:pb.ERROR_CODE_ERR_Ok,
	}

	//check status
	if f.state >= define.GameOver {
		log.Printf("[game(%d)] player[%d] game is over\n", f.id, playerId)
		//reset msg
		msg.ErrorCode = pb.ERROR_CODE_ERR_RoomState

		//notify client
		player.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_Connect), msg))
		return false
	}

	//check conn, if conn not nil, need reset
	if player.GetConn() != nil {
		//TODO, multi thread issue,
		//if call p.client.Close(),
		//will kick entry player
		player.GetConn().SetExtraData(nil)
		log.Printf("[game(%d)] player[%d] replace\n", f.id, playerId)
	}

	//sync conn
	player.Connect(conn)

	//send message to player
	player.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_Connect), msg))

	//call cb of game listener
	//this is the callback of room face
	f.gl.OnJoinGame(conn, f.id, playerId)
	atomic.AddInt32(&f.playerCount, 1)
	return true
}

//leave game
func (f *Game) LeaveGame(playerId uint64) bool {
	//basic check
	if playerId <= 0 {
		return false
	}

	//check player
	player := f.getPlayer(playerId)
	if player == nil {
		return false
	}

	//clean up
	player.CleanUp()

	//call cb of game listener
	//this is the callback of room face
	f.gl.OnLeaveGame(f.id, playerId)
	atomic.AddInt32(&f.playerCount, -1)
	return true
}

//process message
func (f *Game) ProcessMessage(playerId uint64, packet iface.IPacket) bool {
	//basic check
	if playerId <= 0 || packet == nil {
		return false
	}

	//check player
	player := f.getPlayer(playerId)
	if player == nil {
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
				RoomSeatId:player.GetIdx(),
				RandomSeed:f.randSeed,
			}
			//loop players
			sf := func(k, v interface{}) bool {
				p, ok := v.(iface.IPlayer)
				if ok && p != nil {
					if p.GetId() != player.GetId() {
						msg.Others = append(msg.Others, p.GetId())
						msg.Pros = append(msg.Pros, p.GetProgress())
					}
				}
				return true
			}
			f.players.Range(sf)

			//notify player
			player.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_JoinRoom), msg))
		}
	case pb.ID_MSG_Progress://progress
		{
			if f.state > define.GameReady {
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
			packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Progress),
						&pb.S2C_ProgressMsg{
							Id:player.GetId(),
							Pro:msg.Pro,
						})
			//broadcast
			f.broadcastExclude(packet, player.GetId())
		}

	case pb.ID_MSG_Heartbeat://heart beat
		{
			player.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_Heartbeat), nil))
			player.RefreshHeartbeatTime()
		}

	case pb.ID_MSG_Ready://ready
		{
			if f.state == define.GameReady {
				f.doReady(player)
			}else if f.state == define.Gaming {
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
			f.Lock()
			f.result[player.GetId()] = msg.GetWinnerID()
			f.Unlock()
			log.Printf("[game(%d)] ID_MSG_Result player[%d] winner=[%d]\n",
						f.id, player.GetId(), msg.GetWinnerID())
			player.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_Result), nil))
		}
	default:
		{
			log.Printf("[game(%d)] processMsg unknown message id[%d]\n", f.id, int32(messageId))
		}
	}

	return true
}

//tick main logic
func (f *Game) Tick(now int64) bool {
	//do relate opt by game state
	switch f.state {
	case define.GameReady:
		{
			delta := now - f.startTime
			if delta < define.MaxReadyTime {
				if f.checkReady() {
					//start
					f.doStart()
					f.state = define.Gaming
				}
			}else{
				if f.getOnlinePlayerCount() > 0 {
					//up to ready time, if player online, force start
					f.doStart()
					f.state = define.Gaming
					fmt.Printf("[game(%d)] force start game because ready state is timeout\n", f.id)
				}else{
					//all not join game, force finished
					f.state = define.GameOver
					log.Printf("[game(%d)] game over! nobody ready\n", f.id)
				}
			}
			return true
		}
	case define.Gaming:
		{
			if f.checkOver() {
				f.state = define.GameOver
				log.Printf("[game(%d)] game over successfully!!\n", f.id)
				return true
			}

			if f.isTimeOut() {
				f.state = define.GameOver
				log.Printf("[game(%d)] game timeout\n", f.id)
				return true
			}

			//other logic
			f.logic.Tick()
			f.broadcastFrameData()
			return true
		}
	case define.GameOver:
		{
			f.doGameOver()
			f.state = define.GameStop
			log.Printf("[game(%d)] do game over\n", f.id)
			return true
		}
	case define.GameStop:
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
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Close), nil)
	f.broadcast(packet)
}

//clean up
func (f *Game) CleanUp() {
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			player.CleanUp()
		}
		f.players.Delete(k)
		return true
	}
	f.players.Range(sf)
	f.players = sync.Map{}
}

////////////////
//private func
////////////////

//do ready
func (f *Game) doReady(p iface.IPlayer) {
	//check
	if p.IsReady() {
		return
	}

	//set status
	p.SetReady()

	//init message
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Ready), nil)
	p.SendMessage(packet)
}

//check game is ready
func (f *Game) checkReady() bool {
	isRead := true
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			if player.IsReady() {
				isRead = false
				return false
			}
		}
		return true
	}
	f.players.Range(sf)
	return isRead
}

//game start
func (f *Game) doStart() {
	//init for game start
	f.frameCount = 0
	f.logic.Reset()

	//init players
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			player.SetReady()
			player.SetProgress(100)
		}
		return true
	}
	f.players.Range(sf)

	//init message
	f.startTime = time.Now().Unix()

	msg := &pb.S2C_StartMsg{
		TimeStamp:f.startTime,
	}
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Start), msg)

	//broadcast to all
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
		Id:		p.GetId(),
		Sid:	msg.GetSid(),
		X:		msg.GetX(),
		Y:		msg.GetY(),
		RoomSeatId:p.GetIdx(),
	}
	f.logic.PushCommand(input)
	return true
}

//client reconnect
func (f *Game) doReconnect(p iface.IPlayer) bool {
	//init message
	msg := &pb.S2C_StartMsg{
		TimeStamp:f.startTime,
	}
	packet := protocol.NewPacketWithPara(uint8(pb.ID_MSG_Start), msg)

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
			FrameID:i,
		}
		if frameData != nil {
			fd.Input = frameData.GetData()
		}
		frameMsg.Frames = append(frameMsg.Frames, fd)
		c++

		if c >= define.KMaxFrameDataPerMsg || i == (framesCount - 1) {
			p.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_Frame), frameMsg))
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
	if !f.dirty && (frameCount - f.frameCount < define.BroadcastOffsetFrames) {
		return
	}

	//set key data
	now := time.Now().Unix()
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if !ok || player == nil {
			return true
		}
		//check online
		if !player.IsOnline() {
			return true
		}
		//check status
		if !player.IsReady() {
			return true
		}
		//check heart beat
		diff := now - player.GetLastHeartbeatTime()
		if diff >= define.KBadNetworkThreshold {
			return true
		}
		//check player last frame
		i := player.GetSendFrameCount()
		c := int64(0)
		msg := &pb.S2C_FrameMsg{}

		for ; i < frameCount; i++ {
			frameData := f.logic.GetFrame(i)
			if frameData == nil && i != (frameCount - 1) {
				continue
			}

			//init frame data
			fd := &pb.FrameData{
				FrameID:i,
			}
			if frameData != nil {
				fd.Input = frameData.GetData()
			}
			msg.Frames = append(msg.Frames, fd)
			c++

			//if last frame or up to max frame, send them
			if i == (frameCount - 1) || c >= define.KMaxFrameDataPerMsg {
				player.SendMessage(protocol.NewPacketWithPara(uint8(pb.ID_MSG_Frame), msg))
				c = 0
				msg = &pb.S2C_FrameMsg{}
			}
		}

		//set frame count
		player.SetSendFrameCount(frameCount)
		return true
	}
	f.players.Range(sf)
}

//broad cast
func (f *Game) broadcast(packet iface.IPacket) {
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			player.SendMessage(packet)
		}
		return true
	}
	f.players.Range(sf)
}

//broad cast exclude
func (f *Game) broadcastExclude(packet iface.IPacket, id uint64) {
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			if player.GetId() != id {
				player.SendMessage(packet)
			}
		}
		return true
	}
	f.players.Range(sf)
}

//get one player
func (f *Game) getPlayer(id uint64) iface.IPlayer {
	if id <= 0 {
		return nil
	}
	v, ok := f.players.Load(id)
	if !ok || v == nil {
		return nil
	}
	player, ok := v.(iface.IPlayer)
	if ok && player != nil {
		return player
	}
	return nil
}

//get player count
func (f *Game) getPlayerCount() int {
	return int(f.playerCount)
}

//get online player count
func (f *Game) getOnlinePlayerCount() int {
	num := 0
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			if player.IsOnline() {
				num++
			}
		}
		return true
	}
	f.players.Range(sf)
	return num
}

//check over
func (f *Game) checkOver() bool {
	var (
		checkResult = true
	)
	//check
	if f.playerCount <= 0 {
		return false
	}
	//if someone online and get result, will not over
	sf := func(k, v interface{}) bool {
		player, ok := v.(iface.IPlayer)
		if ok && player != nil {
			if player.IsOnline() {
				f.Lock()
				_, subOk := f.result[player.GetId()]
				f.Unlock()
				if !subOk {
					checkResult = false
					return false
				}
			}
		}
		return true
	}
	f.players.Range(sf)
	return checkResult
}

//is time out
func (f *Game) isTimeOut() bool {
	return f.logic.GetFrameCount() > define.MaxGameFrame
}