package main

import (
	"crypto/sha1"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"sync"
	"time"
)

//macro define
const (
	UdpServerAddr = "127.0.0.1:6100"
	Password = "test"
	Salt = "abc"
)

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	//go server()
	time.AfterFunc(time.Second * 2, client)
	//server()
	wg.Wait()
}

//server
func server() {
	key := pbkdf2.Key([]byte(Password), []byte(Salt), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	if listener, err := kcp.ListenWithOptions(UdpServerAddr, block, 10, 3); err == nil {
		log.Println("server created...")
		go serverLoop(listener)

	} else {
		log.Fatal(err)
	}
}

func serverLoop(l *kcp.Listener) {
	for {
		s, err := l.AcceptKCP()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("session:", s.RemoteAddr())
		go handleEcho(s)
	}
}

// handleEcho send back everything it received
func handleEcho(conn *kcp.UDPSession) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("handleEcho, read buf:", string(buf))

		n, err = conn.Write(buf[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func client() {
	key := pbkdf2.Key([]byte(Password), []byte(Salt), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	// wait for server to become ready
	time.Sleep(time.Second)

	// dial to the echo server
	if sess, err := kcp.DialWithOptions(UdpServerAddr, block, 10, 3); err == nil {
		log.Print("client created, session:", sess.RemoteAddr())
		for {
			data := time.Now().String()
			//buf := make([]byte, len(data))
			if _, err := sess.Write([]byte(data)); err == nil {
				log.Println("sent:", data)
				// read back the data
				//if _, err := io.ReadFull(sess, buf); err == nil {
				//	log.Println("recv:", string(buf))
				//} else {
				//	log.Fatal(err)
				//}
			} else {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	} else {
		log.Fatal(err)
	}
}
