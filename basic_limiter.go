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

func BasicLimiter(b ByteCount, t time.Duration) Limiter {
	bl := &basicLimiter{
		t:   time.NewTicker(t),
		bc:  b,
		cbc: make(chan ByteCount),
	}
	go bl.Start()
	return bl
}
