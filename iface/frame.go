package iface

/*
 * interface of frame
 */

type IFrame interface {
	GetData() []IInputData
	AddData(data IInputData)bool
	GetIdx()uint32
}
