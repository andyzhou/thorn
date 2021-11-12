package protocol

import "time"

/*
 * kcp config, implement of `IConfig`
 */

//config info
type Config struct {
	packetSendChanLimit    uint32        // the limit of packet send channel
	packetReceiveChanLimit uint32        // the limit of packet receive channel
	connReadTimeout        time.Duration // read timeout
	connWriteTimeout       time.Duration // write timeout
}

//construct
func NewConfig(
			sendChanLimit, receiveChanLimit uint32,
			readTimeOut, writeTimeOut time.Duration,
		) *Config {
	//self init
	this := &Config{
		packetSendChanLimit:sendChanLimit,
		packetReceiveChanLimit:receiveChanLimit,
		connReadTimeout:readTimeOut,
		connWriteTimeout:writeTimeOut,
	}
	return this
}

func (f *Config) GetPacketSendChanLimit() uint32 {
	return f.packetSendChanLimit
}

func (f *Config) GetPacketReceiveChanLimit() uint32 {
	return f.packetReceiveChanLimit
}

func (f *Config) GetConnReadTimeout() time.Duration {
	return f.connReadTimeout
}

func (f *Config) GetConnWriteTimeout() time.Duration {
	return f.connWriteTimeout
}