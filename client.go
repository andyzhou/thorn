package thorn

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"runtime/debug"
	"sync"
)

/*
 * client api face
 */

//inter macro define
const (
	clientWriteChanSize = 1024 * 5
	clientReadBuffSize = 1024 * 4 //default 4KB
)

//inter type
type (
	clientInfo struct {
		session *kcp.UDPSession
		writeChan chan []byte
		readCloseChan chan bool
		writeCloseChan chan bool
	}
)

//face info
type Client struct {
	address string
	password string
	salt string
	readBuffSize int
	block *kcp.BlockCrypt
	cbForRead func(*kcp.UDPSession, []byte) bool
	clients map[string]*clientInfo //tag -> clientInfo
	sync.RWMutex
}

//construct
func NewClient(
			serverHost string,
			serverPort int,
		) *Client {
	this := &Client{
		address: fmt.Sprintf("%v:%v", serverHost, serverPort),
		readBuffSize: clientReadBuffSize,
		clients: map[string]*clientInfo{},
	}
	return this
}

//quit
func (c *Client) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:Quit panic, err:", err)
		}
	}()
	if c.clients == nil {
		return
	}
	for _, v := range c.clients {
		v.readCloseChan <- true
		v.writeCloseChan <- true
	}
	c.clients = make(map[string]*clientInfo)
}

//close client
func (c *Client) CloseClient(tag string) error {
	//check
	if tag == "" {
		return errors.New("invalid parameter")
	}
	c.Lock()
	defer c.Unlock()
	v, ok := c.clients[tag]
	if !ok || v == nil {
		return errors.New("no such client")
	}
	v.writeCloseChan <- true
	v.readCloseChan <- true
	delete(c.clients, tag)
	return nil
}

//write data
func (c *Client) WriteData(tag string, data []byte) error {
	//check
	if tag == "" || data == nil {
		return errors.New("invalid parameter")
	}

	//get client info by tag
	c.Lock()
	defer c.Unlock()
	clientInfo, ok := c.clients[tag]
	if !ok || clientInfo == nil {
		return errors.New("can't get client info by tag")
	}

	//check chan length
	if len(clientInfo.writeChan) >= clientWriteChanSize {
		return errors.New("write chan is full")
	}

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Printf("client:WriteData panic, err:%v\n", err)
			log.Printf("client:WriteData trace:%v\n", string(debug.Stack()))
		}
	}()

	//async send to chan
	select {
	case clientInfo.writeChan <- data:
	}
	return nil
}

//dial server, step-3
func (c *Client) DialServer(tag string) error {
	//check
	if tag == "" || c.address == "" || c.block == nil {
		return errors.New("invalid parameter")
	}

	//dial server
	session, err := kcp.DialWithOptions(c.address, *c.block, 10, 3)
	if err != nil {
		return err
	}

	//start new client process
	c.createClientProcess(tag, session)
	return nil
}

//set cb for read, step-2
func (c *Client) SetCBForRead(
					cb func(*kcp.UDPSession, []byte) bool,
				) bool {
	//check
	if cb == nil || c.cbForRead != nil {
		return false
	}
	//sync
	c.cbForRead = cb
	return true
}

//set security, step-1
func (c *Client) SetSecurity(password, salt string) error {
	//check
	if password == "" || salt == "" {
		return errors.New("invalid parameter")
	}

	//set key para
	c.password = password
	c.salt = salt

	//init kcp block
	key := pbkdf2.Key(
					[]byte(c.password),
					[]byte(c.salt),
					1024,
					32,
					sha1.New,
				)
	block, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		return err
	}

	//sync block value
	c.block = &block
	return nil
}

//set read buff size, option
func (c *Client) SetReadBuffSize(size int) bool {
	if size <= 0 {
		return false
	}
	c.readBuffSize = size
	return true
}

//////////////
//private func
//////////////

//sub process for one udp session
func (c *Client) createClientProcess(
						tag string,
						session *kcp.UDPSession,
					) bool {
	//check
	if tag == "" || session == nil {
		return false
	}

	//init client info
	clientInfo := &clientInfo{
		session: session,
		writeChan: make(chan []byte, clientWriteChanSize),
		readCloseChan: make(chan bool, 1),
		writeCloseChan: make(chan bool, 1),
	}

	//spawn read and write process
	go c.clientReadProcess(clientInfo)
	go c.clientWriteProcess(clientInfo)

	//sync into map
	c.clients[tag] = clientInfo
	return true
}

//process for client writer
func (c *Client) clientWriteProcess(client *clientInfo) bool {
	var (
		req []byte
		isOk bool
	)

	//check
	if client == nil {
		return false
	}

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:clientReadProcess panic, err:", err)
		}
		close(client.writeChan)
		close(client.writeCloseChan)
	}()

	//loop
	for {
		select {
		case req, isOk = <- client.writeChan:
			if isOk {
				client.session.Write(req)
			}
		case <- client.writeCloseChan:
			return false
		}
	}

	return true
}

//process for client reader
func (c *Client) clientReadProcess(client *clientInfo) bool {
	//check
	if client == nil {
		return false
	}

	//set read buff
	readBuff := make([]byte, c.readBuffSize)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:clientReadProcess panic, err:", err)
		}
		close(client.readCloseChan)
		readBuff = make([]byte, 0)
	}()

	//loop
	for {
		//try get close
		select {
		case <-client.readCloseChan://close chan
			return false
		default:
			{
				//try read origin data
				_, err := client.session.Read(readBuff)
				if err != nil {
					log.Println("Client:clientReadProcess failed, err:", err.Error())
					return false
				}

				//call cb
				if c.cbForRead != nil {
					c.cbForRead(client.session, readBuff)
				}
			}
		}
	}
}

//run main process
func (c *Client) runMainProcess() {

}