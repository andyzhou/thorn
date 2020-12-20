package network

import "time"

//config info
type Config struct {
	PacketSendChanLimit    uint32        // the limit of packet send channel
	PacketReceiveChanLimit uint32        // the limit of packet receive channel
	ConnReadTimeout        time.Duration // read timeout
	ConnWriteTimeout       time.Duration // write timeout
}

