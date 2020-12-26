package room

import (
	"github.com/andyzhou/thorn/iface"
	"github.com/andyzhou/thorn/protocol"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * room face, implement of IRoom
 */

//inter macro define
const (
	Frequency = 30 //frame frequency
	TickTimer = time.Second / Frequency
	TimeOut = time.Minute * 5
	InOutChanSize = 64
	MessageChanSize = 2048
)

//face info
type Room struct {
	roomId uint64 //room id
	secretKey string
	closeFlag int32
	timeStamp int64
	players []uint64
	cb iface.IRoomCallback //cb for api client
	inChan chan iface.IConn
	outChan chan iface.IConn
	messageChan chan iface.IMessage
	closeChan chan bool
	game iface.IGame `game instance`
	wg sync.WaitGroup
}

//construct
func NewRoom(
			roomId uint64,
			players []uint64,
			randomSeed int32,
			cb iface.IRoomCallback,
		) *Room {
	//self init
	this := &Room{
		roomId:roomId,
		cb:cb,
		players:make([]uint64, 0),
		inChan:make(chan iface.IConn, InOutChanSize),
		outChan:make(chan iface.IConn, InOutChanSize),
		messageChan:make(chan iface.IMessage, MessageChanSize),
		closeChan:make(chan bool, 1),
	}

	//init game instance
	this.game = NewGame(roomId, players, randomSeed, this)

	//spawn main process
	go this.runMainProcess()

	return this
}

func (f *Room) Stop() {
	f.closeChan <- true
}

func (f *Room) GetId() uint64 {
	return f.roomId
}

func (f *Room) GetSecretKey() string {
	return f.secretKey
}

func (f *Room) GetTimeStamp() int64 {
	return f.timeStamp
}

func (f *Room) GetCB() iface.IRoomCallback {
	return f.cb
}

func (f *Room) IsOver() bool {
	return atomic.LoadInt32(&f.closeFlag) != 0
}

func (f *Room) HasPlayer(playerId uint64) bool {
	if playerId <= 0 || f.players == nil {
		return false
	}
	for _, v := range f.players {
		if v == playerId {
			return true
		}
	}
	return false
}

func (f *Room) VerifyToken(token string) bool {
	if token != f.secretKey {
		return false
	}
	return true
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
func (f *Room) OnMessage(conn iface.IConn, packet iface.IPacket) (bRet bool) {
	//try get data
	playerId, ok := conn.GetExtraData().(uint64)
	if !ok {
		bRet = false
		return
	}

	//catch panic
	defer func() {
		if err := recover(); err != nil {
			bRet = false
			return
		}
	}()

	//init message
	message := protocol.NewMessage(playerId, packet)

	//send to chan
	f.messageChan <- message
	bRet = true
	return
}

//cb for OnClose
func (f *Room) OnClose(conn iface.IConn) {
	f.outChan <- conn
}


//////////////////////
//cb for IGameListener
//////////////////////

func (f *Room) OnJoinGame(roomId, playerId uint64) {
	log.Printf("room %d OnJoinGame %d\n", roomId, playerId)
}

func (f *Room) OnStartGame(roomId uint64) {
	log.Printf("room %d OnStartGame\n", roomId)
}

func (f *Room) OnLeaveGame(roomId, playerId uint64) {
	log.Printf("room %d OnLeaveGame %d\n", roomId, playerId)
}

func (f *Room) OneGameOver(uint64) {
	atomic.StoreInt32(&f.closeFlag, 1)
}

//////////////
//private func
//////////////

//main process
func (f *Room) runMainProcess() {
	var (
		ticker = time.NewTicker(TickTimer)
		timer = time.NewTimer(TimeOut)
		conn iface.IConn
		message iface.IMessage
		isOk, bRet bool
	)

	defer func() {
		//clean up
		f.game.Close()
		ticker.Stop()
		close(f.inChan)
		close(f.outChan)
		close(f.messageChan)
		close(f.closeChan)
	}()

	log.Printf("Room %d is running\n", f.roomId)

	//loop
	for {
		select {
		case <- f.closeChan:
			//closed
			return

		case <- ticker.C:
			{
				if !f.game.Tick(time.Now().Unix()) {
					break
				}
			}

		case <- timer.C:
			//timer out
			return

		case message, isOk = <- f.messageChan://message
			if isOk {
				f.game.ProcessMessage(message.GetId(), message.GetPacket())
			}

		case conn, isOk = <- f.inChan://join
			if isOk {
				playerId, ok := conn.GetExtraData().(uint64)
				if ok {
					bRet = f.game.JoinGame(playerId, conn)
					if !bRet {
						conn.Close()
					}
				}else{
					conn.Close()
				}
			}

		case conn, isOk = <- f.outChan://leave
			if isOk {
				playerId, ok := conn.GetExtraData().(uint64)
				if ok {
					bRet = f.game.LeaveGame(playerId)
					if !bRet {
						conn.Close()
					}
				}else{
					conn.Close()
				}
			}
		}
	}
}