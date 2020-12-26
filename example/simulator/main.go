package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"net"
	"sync"
)

/*
 * client simulator
 */


//inter macro define
const (
	UdpServerAddr = "127.0.0.1:6100"
	Password = "test"
	Salt = "abc"
)

func main()  {
	wg := new(sync.WaitGroup)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic happened, err:", err)
			wg.Done()
		}
	}()

	wg.Add(1)

	key := pbkdf2.Key([]byte(Password), []byte(Salt), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	//dial server
	client, err := kcp.DialWithOptions(UdpServerAddr, block, 10, 3)
	if err != nil {
		log.Println("connect server failed, err:", err)
		return
	}

	go runMainProcess(client)
	wg.Wait()
}


//main process
func runMainProcess(con net.Conn) {
	fmt.Println("connect server success")
}