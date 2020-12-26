package iface

import "time"

/*
 * interface of config
 */

type IConfig interface {
	GetPacketSendChanLimit() uint32
	GetPacketReceiveChanLimit() uint32
	GetConnReadTimeout() time.Duration
	GetConnWriteTimeout() time.Duration
}
