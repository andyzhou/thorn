package define

import "errors"

/*
 * errors declare
 */

var (
	ErrorOfInvalidPara = errors.New("invalid input parameter")

	//for network
	ErrConnClosing   = errors.New("use of closed network connection")
	ErrWriteBlocking = errors.New("write packet was blocking")
	ErrReadBlocking  = errors.New("read packet was blocking")
)