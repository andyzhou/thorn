package room

import (
	"github.com/andyzhou/thorn/iface"
)

/*
 * game face, implement of IGame
 */

//game state
const (
	GameReady = iota
	Gaming
	GameOver
	GameStop
)

//face info
type Game struct {
	id uint64
	startTime int64
	randSeed int32
	state int
	gl iface.IGameListener
	players map[uint64]iface.IPlayer
	frameCount uint32
	result map[uint64]uint64
}

//construct
func NewGame(
			id uint64,
			players []uint64,
			randSeed int32,
			gl iface.IGameListener,
		) *Game {
	//self init
	this := &Game{
		id:id,
		randSeed:randSeed,
		gl:gl,
		players:make(map[uint64]iface.IPlayer),
		result:make(map[uint64]uint64),
	}
	//init players
	for idx, v := range players {
		this.players[v] = NewPlayer(v, uint64(idx + 1))
	}
	return this
}

//close
func (f *Game) Close() {
}

//join game
func (f *Game) JoinGame(id uint64, conn iface.IConn) bool {
	//basic check
	if id <= 0 || conn == nil {
		return false
	}

	//check player
	p, ok := f.players[id]
	if !ok {
		return false
	}

	//check status
	if f.state != GameReady && f.state != Gaming {
		//notify client
		return false
	}

	//check conn
	if p.GetConn() != nil {

	}

	//sync conn
	p.Connect(conn)

	//call cb of game listener
	f.gl.OnJoinGame(f.id, id)

	return true
}

//leave game
func (f *Game) LeaveGame(id uint64) bool {
	//basic check
	if id <= 0 {
		return false
	}

	//check player
	p, ok := f.players[id]
	if !ok {
		return false
	}

	//clean up
	p.CleanUp()

	//call cb of game listener
	f.gl.OnLeaveGame(f.id, id)

	return true
}

//process message
func (f *Game) ProcessMessage(id uint64, packet iface.IPacket) bool {
	//basic check
	if id <= 0 || packet == nil {
		return false
	}

	//check player
	_, ok := f.players[id]
	if !ok {
		return false
	}

	//get message id
	//messageId := int32(packet.GetMessageId())

	//do relate opt by message id

	return true
}

//tick main logic
func (f *Game) Tick(now int64) bool {
	//do relate opt by game state
	switch f.state {
	case GameReady:
	case Gaming:
	case GameOver:
	case GameStop:
	}
	return true
}

//get result
func (f *Game) GetResult() map[uint64]uint64 {
	return f.result
}