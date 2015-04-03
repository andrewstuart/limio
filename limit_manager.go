package limio

import (
	"io"
	"time"
)

//An EqualLimiter is itself a limiter and will evenly distribute the limits
//it is given across all its managed Limiters.
type EqualLimiter struct {
	rmap   map[Limiter]chan uint64
	rate   <-chan uint64
	remain uint64
}

func (rm *EqualLimiter) run() {
	for {
		lim := <-rm.rate

		perChan := uint64(float64(lim) / float64(len(rm.rmap)))

		for _, c := range rm.rmap {
			go func() {
				c <- perChan
			}()
		}
	}
}

func NewEqualLimiter() *EqualLimiter {
	rm := EqualLimiter{
		rmap: make(map[Limiter]chan uint64),
	}

	return &rm
}

//NewReader is a convenience that will automatically wrap an io.Reader with the
//internal Limiter implementation.
func (rm *EqualLimiter) NewReader(r io.Reader) *Reader {
	lr := NewReader(r)
	rm.ManageLimiter(lr)

	return lr
}

func (rm *EqualLimiter) Limit(n uint64, t time.Duration) {
}

func (rm *EqualLimiter) LimitChan(<-chan uint64) {
}

//ManageLimiter accepts a Limiter to be managed under the new "scope"
//established by this parent Limiter.
func (rm *EqualLimiter) ManageLimiter(lr LimitWaiter) {
	ch := make(chan uint64)
	lr.LimitChan(ch)

	rm.rmap[lr] = ch

	//When lr closes, close the channel and remove it from the map
	//TODO figure out a way to know when to remove other Limiters
	//Limiter.Free()? Limiter.Done()? Is it really the Limiter's responsibility
	//to have a mechanism for this?
	go func() {
		lr.Wait()
		close(ch)
		delete(rm.rmap, lr)
	}()
}
