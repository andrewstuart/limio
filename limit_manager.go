package limio

import (
	"io"
	"time"
)

type LimitManager struct {
	rmap map[Limiter]<-chan uint64
}

func (rm *LimitManager) run() {
}

func NewReadManager() *LimitManager {
	rm := LimitManager{
		rmap: make(map[Limiter]<-chan uint64),
	}

	return &rm
}

func (rm *LimitManager) NewReader(r io.Reader) *Reader {
	lr := NewReader(r)

	ch := make(chan uint64)
	lr.LimitChan(ch)

	rm.rmap[lr] = ch

	//When lr closes, close the channel and remove it from the map
	go func() {
		lr.Close()
		close(ch)
		delete(rm.rmap, lr)
	}()

	return nil
}

func (rm *LimitManager) Limit(n uint64, t time.Duration) {
}

func (rm *LimitManager) LimitChan(<-chan uint64) {
}

func (rm *LimitManager) Manage(Limiter) {
}
