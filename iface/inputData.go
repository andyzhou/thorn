package iface

/*
 * interface of frame input data
 */

type IInputData interface {
	GetId() uint64
	GetRaw() interface{}
}
