package room

import "github.com/andyzhou/thorn/iface"

/*
 * frame data face, implement of IFrame
 */

//face info
type Frame struct {
	idx uint32
	data []iface.IInputData
}

//construct
func NewFrame(idx uint32) *Frame {
	//self init
	this := &Frame{
		idx:idx,
		data:make([]iface.IInputData, 0),
	}
	return this
}

func (f *Frame) GetData() []iface.IInputData {
	return f.data
}

func (f *Frame) AddData(data iface.IInputData)bool {
	f.data = append(f.data, data)
	return true
}

func (f *Frame) GetIdx() uint32 {
	return f.idx
}
