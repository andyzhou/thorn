package room

import (
	"github.com/andyzhou/thorn/pb"
	"sync"
)

/*
 * frame data face, implement of IFrame
 */

//face info
type Frame struct {
	idx  uint32
	data []*pb.InputData
	sync.RWMutex
}

//construct
func NewFrame(idx uint32) *Frame {
	//self init
	this := &Frame{
		idx:idx,
		data:make([]*pb.InputData, 0),
	}
	return this
}

func (f *Frame) GetData() []*pb.InputData {
	f.Lock()
	defer f.Unlock()
	return f.data
}

func (f *Frame) AddData(data *pb.InputData)bool {
	f.Lock()
	defer f.Unlock()
	f.data = append(f.data, data)
	return true
}

func (f *Frame) GetIdx() uint32 {
	return f.idx
}
