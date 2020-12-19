package iface

/*
 * interface of lock step
 */

type ILockStep interface {
	Reset()
	GetRangeFrames(from, to uint32) []IFrame
	GetFrame(idx uint32) IFrame
	GetFrameCount() uint32
	PushCommand(data IInputData) bool
	Tick() uint32
}