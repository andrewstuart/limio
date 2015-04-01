package ratelimit

import "time"

type basicLimiter struct {
	t   *time.Ticker
	bc  ByteCount
	cbc []chan ByteCount
}

func (bl *basicLimiter) Start() {
	for {
		<-bl.t.C

		perChan := bl.bc / ByteCount(len(bl.cbc))

		for i := range bl.cbc {
			go func(i int) {
				bl.cbc[i] <- perChan
			}(i)
		}
	}
}

func (bl *basicLimiter) GetLimit() <-chan ByteCount {
	ch := make(chan ByteCount)
	bl.cbc = append(bl.cbc, ch)
	return ch
}

const timeSlice = 20 * time.Millisecond

//NewBasicLimiter will appropriately distribute the rate given across 20ms
//windows. If used to create multiple LimitedReaders (or if GetLimit called
//multiple times), it will divvy up the rate across all the readers, at the same
//rate.
func NewBasicLimiter(b ByteCount, t time.Duration) Limiter {
	bl := &basicLimiter{
		t:   time.NewTicker(timeSlice),
		bc:  b / ByteCount(t/timeSlice),
		cbc: make([]chan ByteCount, 0, 1),
	}
	go bl.Start()
	return bl
}
