package room

import (
	"github.com/andyzhou/thorn/conf"
	"github.com/andyzhou/thorn/define"
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

//face info
type Room struct {
	cfg        *conf.RoomConf //room config
	game       iface.IGame    //game instance
	inChan     chan iface.IConn
	outChan    chan iface.IConn
	packetChan chan iface.IPlayerPacket
	closeChan  chan bool
	closeFlag  int32
	wg         sync.WaitGroup
}

//construct
func NewRoom(cfg *conf.RoomConf) *Room {
	//self init
	this := &Room{
		cfg: cfg,
		inChan: make(chan iface.IConn, define.RoomInOutChanSize),
		outChan: make(chan iface.IConn, define.RoomInOutChanSize),
		packetChan: make(chan iface.IPlayerPacket, define.RoomMessageChanSize),
		closeChan: make(chan bool, 1),
	}

	//check default values
	if cfg.Frequency <= 0 {
		cfg.Frequency = define.RoomFrequency
	}

	//if room has time limit, setup timer func
	if cfg.TimeLimit > 0 {
		duration := time.Duration(cfg.TimeLimit) * time.Second
		time.AfterFunc(duration, this.cbForCountDown)
	}

	//init game instance
	this.game = NewGame(
					cfg.RoomId,
					cfg.Players,
					cfg.RandomSeed,
					this,
				)

	//spawn main process
	go this.runMainProcess()
	return this
}

func (f *Room) Stop() {
	var (
		m any = nil
	)
	defer func() {
		if err := recover(); err != m {
			log.Println("Room:Stop panic, err:", err)
		}
	}()
	select {
	case f.closeChan <- true:
	}
}

func (f *Room) GetId() uint64 {
	return f.cfg.RoomId
}

func (f *Room) GetSecretKey() string {
	return f.cfg.SecretKey
}

func (f *Room) IsOver() bool {
	return atomic.LoadInt32(&f.closeFlag) != 0
}

func (f *Room) HasPlayer(playerId uint64) bool {
	if playerId <= 0 || f.cfg.Players == nil {
		return false
	}
	for _, v := range f.cfg.Players {
		if v == playerId {
			return true
		}
	}
	return false
}

func (f *Room) VerifyToken(token string) bool {
	if token != f.cfg.SecretKey {
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
	//async send to chan
	select {
	case f.inChan <- conn:
	}
	return true
}

//cb for OnMessage
func (f *Room) OnMessage(conn iface.IConn, packet iface.IPacket) (bRet bool) {
	var (
		m any = nil
	)
	//try get player id from extra data
	playerId, ok := conn.GetExtraData().(uint64)
	if !ok {
		bRet = false
		return
	}

	//catch panic
	defer func() {
		if err := recover(); err != m {
			bRet = false
			return
		}
	}()

	//init player packet
	playerPacket := protocol.NewPlayerPacket()
	playerPacket.SetId(playerId)
	playerPacket.SetPacket(packet)

	//async send to chan
	select {
	case f.packetChan <- playerPacket:
	}
	bRet = true
	return
}

//cb for OnClose
func (f *Room) OnClose(conn iface.IConn) {
	log.Println("Room.OnClose")
	//async send to chan
	select {
	case f.outChan <- conn:
	}
}

func (f *Room) OnJoinGame(conn iface.IConn, roomId, playerId uint64) {
	log.Printf("room %d OnJoinGame %d\n", roomId, playerId)
}

func (f *Room) OnStartGame(roomId uint64) {
	log.Printf("room %d OnStartGame\n", roomId)
}

func (f *Room) OnLeaveGame(roomId, playerId uint64) {
	log.Printf("room %d OnLeaveGame %d\n", roomId, playerId)
}

func (f *Room) OneGameOver(roomId uint64) {
	log.Printf("room %d OneGameOver\n", roomId)
	atomic.StoreInt32(&f.closeFlag, 1)
}

////////////////
//private func
////////////////

//cb func for timer out
//notify all players and start timer for end
func (f *Room) cbForCountDown() {
}

//main process
func (f *Room) runMainProcess() {
	var (
		//ticker = time.NewTicker(define.RoomTickTimer)
		ticker *time.Ticker
		conn iface.IConn
		message iface.IPlayerPacket
		isOk, bRet bool
	)

	//init key data
	seconds := int64(time.Second) / int64(f.cfg.Frequency)
	duration := time.Duration(seconds)
	ticker = time.NewTicker(duration)

	//defer
	defer func() {
		//clean up
		f.game.Close()
		ticker.Stop()
		close(f.inChan)
		close(f.outChan)
		close(f.packetChan)
		close(f.closeChan)
	}()

	//loop
	for {
		select {
		case <- f.closeChan:
			//closed
			return

		case <- ticker.C:
			{
				//game ticker
				if !f.game.Tick(time.Now().Unix()) {
					break
				}
			}

		case message, isOk = <- f.packetChan:
			if isOk {
				//input message from player
				f.game.ProcessMessage(message.GetId(), message.GetPacket())
			}

		case conn, isOk = <- f.inChan:
			if isOk {
				//join room
				//get player id
				playerId, ok := conn.GetExtraData().(uint64)
				if ok {
					//join game
					bRet = f.game.JoinGame(playerId, conn)
					if !bRet {
						conn.Close()
					}
				}else{
					conn.Close()
				}
			}

		case conn, isOk = <- f.outChan:
			if isOk {
				//leave room
				//get player id
				playerId, ok := conn.GetExtraData().(uint64)
				if ok {
					//leave game
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