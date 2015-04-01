package ratelimit

import "time"

type basicLimiter struct {
	t   *time.Ticker
	bc  ByteCount
	cbc chan ByteCount
}

func (bl *basicLimiter) Start() {
	for {
		<-bl.t.C
		bl.cbc <- bl.bc
	}
}

func (bl basicLimiter) GetLimit() <-chan ByteCount {
	return bl.cbc
}

const TIME_UNIT = 50 * time.Millisecond

//BasicLimiter will divvy up the bytes into 100 smaller parts to spread the load
//across time
func BasicLimiter(b ByteCount, t time.Duration) Limiter {
	bl := &basicLimiter{
		t:   time.NewTicker(TIME_UNIT),
		bc:  b / ByteCount(t/TIME_UNIT),
		cbc: make(chan ByteCount),
	}
	go bl.Start()
	return bl
}
