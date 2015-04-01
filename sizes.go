package ratelimit

//Some useful constants with the proper typing
const (
	B ByteCount = 1 << (10 * (iota))
	KB
	MB
	GB
	TB
	PB
	EB
)
