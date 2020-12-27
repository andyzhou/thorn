package main

import (
	"fmt"
	"github.com/andyzhou/thorn"
	"log"
	"time"
)

//inter macro define
const (
	UdpServerAddr = "127.0.0.1:6100"
	Password = "test"
	Salt = "abc"
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
	server := thorn.NewServer(UdpServerAddr, Password, Salt)

	//set callback
	server.SetCallback(NewRoomCallBack())

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
		)
	fmt.Printf("create room %d success\n", roomId)
}

