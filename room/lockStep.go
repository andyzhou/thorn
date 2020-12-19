package room

import (
	"github.com/andyzhou/thorn/iface"
	"sync/atomic"
)

/*
 * lock step data face, implement of ILockStep
 */

//face info
type LockStep struct {
	frames map[uint32]iface.IFrame
	frameCount uint32
}

//construct
func NewLockStep() *LockStep {
	//self init
	this := &LockStep{
		frames:make(map[uint32]iface.IFrame),
		frameCount:0,
	}
	return this
}


//reset
func (f *LockStep) Reset() {
	f.frames = make(map[uint32]iface.IFrame)
	f.frameCount = 0
}

//push frame data
func (f *LockStep) PushCommand(data iface.IInputData) bool {
	//basic check
	if data == nil {
		return false
	}

	//check frame
	frame, ok := f.frames[f.frameCount]
	if !ok {
		//init new
		frame = NewFrame(f.frameCount)
	}

	//check is same frame id
	for _, v := range frame.GetData() {
		if v.GetId() == data.GetId() {
			//same data id, skipped
			return false
		}
	}

	//add data into frame
	frame.AddData(data)

	return true
}

//get frame count
func (f *LockStep) GetFrameCount() uint32 {
	return f.frameCount
}

//gen tick
func (f *LockStep) Tick() uint32 {
	atomic.AddUint32(&f.frameCount, 1)
	return f.frameCount
}

//get batch frame
func (f *LockStep) GetRangeFrames(from, to uint32) []iface.IFrame {
	//basic check
	if from < 0 || to < 0 {
		return nil
	}
	//init result
	result := make([]iface.IFrame, 0)
	for ; from <= to && from <= f.frameCount; from++ {
		f, ok := f.frames[from]
		if !ok {
			continue
		}
		result = append(result, f)
	}
	return result
}


//get one frame
func (f *LockStep) GetFrame(idx uint32) iface.IFrame {
	//basic check
	if idx < 0 {
		return nil
	}
	v, ok := f.frames[idx]
	if !ok {
		return nil
	}
	return v
}