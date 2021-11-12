package iface

import "io"

/*
 * interface of protocol
 */

type IProtocol interface {
	ReadPacket(reader io.Reader) (IPacket, error)
	SetEndian(bool)
}
