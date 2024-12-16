package define

//global
const (
	MaxReadyTime          int64  = 120           //max prepare time, force close if no any login
	MaxGameFrame          uint32 = 30*60*3 + 100 //max frames per game round
	BroadcastOffsetFrames        = 3             //cast per frames
	KMaxFrameDataPerMsg          = 60            //max message packet per frame
	KBadNetworkThreshold         = 2             //max time for no heart beat
	PlayerSendChanSize           = 1024
)

//game state
const (
	GameReady = iota
	Gaming
	GameCountDown
	GameOver
	GameStop
)

