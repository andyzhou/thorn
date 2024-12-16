package protocol

/*
 * input data face, implement of IInputData
 */

//data face info
type InputData struct {
	id uint64
	x uint32
	y uint32
	raw interface{}
}

//construct
func NewInputData(
		id uint64,
		x uint32,
		y uint32,
		raw interface{},
	) *InputData {
	//self init
	this := &InputData{
		id:id,
		x:x,
		y:y,
		raw:raw,
	}
	return this
}

func (f *InputData) GetId() uint64 {
	return f.id
}

func (f *InputData) GetRaw() interface{} {
	return f.raw
}