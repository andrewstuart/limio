package limio

import (
	"io"
	"time"
)

type ReadManager struct {
	rmap map[Limiter]<-chan uint64
}

func (rm *ReadManager) run() {
}

func NewReadManager() *ReadManager {
	rm := ReadManager{
		rmap: make(map[Limiter]<-chan uint64),
	}

	return &rm
}

func (rm *ReadManager) NewReader(r io.Reader) *Reader {
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

func (rm *ReadManager) Limit(n uint64, t time.Duration) {
}

func (rm *ReadManager) LimitChan(<-chan uint64) {
}

func (rm *ReadManager) AddLimiter(Limiter) {
}
