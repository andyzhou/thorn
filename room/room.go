package room

import (
	"github.com/andyzhou/thorn/game"
	"github.com/andyzhou/thorn/iface"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * room face, implement of IRoom
 */

//inter macro define
const (
	Frequency = 30
	TickTimer = time.Second / Frequency
	TimeOut = time.Minute * 5
	InOutChanSize = 64
	MessageChanSize = 2048
)

//face info
type Room struct {
	id uint64
	secretKey string
	closeFlag int32
	timeStamp int64
	players []uint64
	inChan chan iface.IConn
	outChan chan iface.IConn
	messageChan chan iface.IPacket
	closeChan chan bool
	game iface.IGame `game instance`
	wg sync.WaitGroup
}

//construct
func NewRoom(
			id uint64,
			players []uint64,
			randomSeed int32,
		) *Room {
	//self init
	this := &Room{
		id:id,
		players:make([]uint64, 0),
		inChan:make(chan iface.IConn, InOutChanSize),
		outChan:make(chan iface.IConn, InOutChanSize),
		messageChan:make(chan iface.IPacket, MessageChanSize),
		closeChan:make(chan bool, 1),
	}

	//init game instance
	this.game = game.NewGame(id, players, randomSeed, this)

	return this
}

func (f *Room) Stop() {
	close(f.closeChan)
	f.wg.Done()
}

func (f *Room) Start() {
	go f.runMainProcess()
}

func (f *Room) GetId() uint64 {
	return f.id
}

func (f *Room) GetSecretKey() string {
	return f.secretKey
}

func (f *Room) GetTimeStamp() int64 {
	return f.timeStamp
}

func (f *Room) IsOver() bool {
	return atomic.LoadInt32(&f.closeFlag) != 0
}

func (f *Room) HasPlayer(id uint64) bool {
	if id <= 0 || f.players == nil {
		return false
	}
	for _, v := range f.players {
		if v == id {
			return true
		}
	}
	return false
}

//////////////////
//cb for iConnect
//////////////////

//cb for OnConnect
func (f *Room) OnConnect(conn iface.IConn) bool {
	conn.SetCallBack(f)
	f.inChan <- conn
	return true
}

//cb for OnMessage
func (f *Room) OnMessage(conn iface.IConn, packet iface.IPacket) bool {
	//try get data
	_, ok := conn.GetExtraData().(uint64)
	if !ok {
		return false
	}
	f.messageChan <- packet
	return true
}

//cb for OnClose
func (f *Room) OnClose(conn iface.IConn) {
	f.outChan <- conn
}


//////////////////////
//cb for IGameListener
//////////////////////

func (f *Room) OnJoinGame(uint64, uint64) {

}
func (f *Room) OnStartGame(uint64) {

}
func (f *Room) OnLeaveGame(uint64, uint64) {

}
func (f *Room) OneGameOver(uint64) {

}

//////////////
//private func
//////////////

//main process
func (f *Room) runMainProcess() {
	var (
		ticker = time.NewTicker(TickTimer)
		timer = time.NewTimer(TimeOut)
	)

	f.wg.Add(1)
	defer func() {
		//clean up
		ticker.Stop()
		close(f.closeChan)
		f.wg.Done()
	}()

	//loop
	for {
		select {
		case <- f.closeChan:
			//closed
			return
		case <- timer.C:
			//timer out
			return

		}
	}
}