package limio

import (
	"io"
	"time"
)

//An EqualLimiter is itself a limiter and will evenly distribute the limits
//it is given across all its managed Limiters.
type EqualLimiter struct {
	rmap   map[Limiter](chan uint64)
	remain uint64
	rate   <-chan uint64
}

func (el *EqualLimiter) run() {
	for {
		lim := <-el.rate

		//Split out the write across each channel we're managing
		perChan := uint64(float64(lim) / float64(len(el.rmap)))
		for l := range el.rmap {
			select {
			case el.rmap[l] <- perChan:
			case <-time.After(100 * time.Millisecond):
				//Clean up if channel is blocked
				//This likely would mean channel is done reading
				delete(el.rmap, l)
			}
		}
	}
}

func NewEqualLimiter() *EqualLimiter {
	el := EqualLimiter{
		rmap: make(map[Limiter]chan uint64),
	}
	go el.run()

	return &el
}

//NewReader is a convenience that will automatically wrap an io.Reader with the
//internal Limiter implementation.
func (el *EqualLimiter) NewReader(r io.Reader) *Reader {
	lr := NewReader(r)
	el.ManageLimiter(lr)
	return lr
}

//Limit for the EqualLimiter sets a shared rate limit for all Limiters managed
//by the EqualLimiter
func (el *EqualLimiter) Limit(n uint64, t time.Duration) {
	n, t = Distribute(n, t, window)
	ch := make(chan uint64)

	//TODO make sure there's no memory leak
	go func() {
		for {
			time.Sleep(t)
			ch <- n
		}
	}()
	el.rate = ch
}

func (el *EqualLimiter) LimitChan(rl <-chan uint64) {
	el.rate = rl
}

//ManageLimiter accepts a Limiter to be managed under the new "scope"
//established by this parent Limiter.
func (el *EqualLimiter) ManageLimiter(lr LimitWaiter) {
	ch := make(chan uint64)
	lr.LimitChan(ch)

	el.rmap[lr] = ch

	go func() {
		lr.Wait()
		close(ch)
		delete(el.rmap, lr)
	}()
}
