package conf

/*
 * conf for room
 */

type RoomConf struct {
	RoomId uint64
	Players []uint64
	RandomSeed int32
	SecretKey string
	MaxPlayers int //0 means no limit
	Frequency int //frame frame, default 30 frames
	TimeLimit int //seconds value, 0 means no limit
	NotifyTime int //seconds value, notify before end
}