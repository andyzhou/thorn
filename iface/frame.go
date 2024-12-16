package iface

import (
	"github.com/andyzhou/thorn/pb"
)

/*
 * interface of frame
 */

type IFrame interface {
	GetData() []*pb.InputData
	AddData(data *pb.InputData) bool
	GetIdx() uint32
}
