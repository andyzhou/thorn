package iface

/*
 * interface of kcp server
 */

type IKcpServer interface {
	Quit()
	GetManager() IManager
	GetRouter() IConnCallBack
	GetProtocol() IProtocol
	GetConfig() IConfig
	SetCallback(cb IRoomCallback) bool
	SetConfig(config IConfig) bool
}
