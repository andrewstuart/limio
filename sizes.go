package ratelimit

const (
	B ByteCount = 1 << (10 * (iota))
	KB
	MB
	GB
	TB
	PB
)
