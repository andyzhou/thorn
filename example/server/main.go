package main

import (
	"fmt"
	"github.com/andyzhou/thorn"
	"github.com/andyzhou/thorn/conf"
	"log"
	"time"
)

//inter macro define
const (
	ServerHost = "127.0.0.1"
	ServerPort = 6100
	Password = "test"
	Salt = "abc"
	SecretKey = "testRoom"
)

func main() {
	var (
		m any = nil
	)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Println("panic happened, err:", err)
		}
	}()

	//setup server conf
	serverConf := &thorn.ServerConf{
		Host: ServerHost,
		Port: ServerPort,
		Password: Password,
		Salt: Salt,
	}

	//init server
	server := thorn.NewServer(serverConf)

	//set callback
	server.SetCallback(NewRoomCallBack())

	//create room
	go createRoom(server)

	//start
	server.Start()
}

func createRoom(server *thorn.Server) {
	time.Sleep(time.Second * 2)

	//init room
	roomId := uint64(1)
	roomPlayers := []uint64{
		1,
		2,
	}

	//setup room conf
	roomCfg := &conf.RoomConf{
		RoomId: roomId,
		Players: roomPlayers,
		RandomSeed: int32(time.Now().Unix()),
		SecretKey: SecretKey,
	}

	//create room
	server.CreateRoom(roomCfg)
	fmt.Printf("create room %d success\n", roomId)
}

