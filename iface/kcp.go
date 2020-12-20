package iface

/*
 * interface of kcp server
 */

type IKcpServer interface {
	Quit()
	Start(cb IConnCallBack, protocol IProtocol)
}
