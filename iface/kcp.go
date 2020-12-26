package iface

/*
 * interface of kcp server
 */

type IKcpServer interface {
	Quit()
	GetProtocol() IProtocol
	GetConfig() IConfig
	SetConfig(config IConfig) bool
}
