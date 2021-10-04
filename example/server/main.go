package main

import (
	"fmt"
	"github.com/andyzhou/thorn"
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
	//wg := new(sync.WaitGroup)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic happened, err:", err)
		}
	}()

	//init
	server := thorn.NewServer(ServerHost, ServerPort, Password, Salt)

	//set callback
	server.SetCallback(NewRoomCallBack())

	//try create room
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

	//create room
	server.CreateRoom(
			roomId,
			roomPlayers,
			1,
			SecretKey,
		)
	fmt.Printf("create room %d success\n", roomId)
}

