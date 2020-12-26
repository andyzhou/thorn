package main

import (
	"fmt"
	"github.com/andyzhou/thorn"
	"log"
	"sync"
)

//inter macro define
const (
	UdpAddr = ":6100"
)

func main() {
	wg := new(sync.WaitGroup)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic happened, err:", err)
			wg.Done()
		}
	}()

	//init
	server := thorn.NewServer(UdpAddr)

	//set callback
	server.SetCallback(NewRoomCallBack())

	//start
	server.Start()

	wg.Add(1)

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
	fmt.Println("xxx")

	wg.Wait()
}


