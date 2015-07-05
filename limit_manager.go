package limio

import (
	"errors"
	"io"
	"time"
)

//A Manager enables consumers to treat a group of Limiters as a single Limiter,
//enabling hierarchies of limiters. For example, a network interface could have
//a global limit that is distributed across connections, each of which can
//manage their own distribution of the bandwidth they are allocated.
type Manager interface {
	Limiter
	Manage(Limiter) error
	Unmanage(Limiter)
}

//A SimpleManager is an implementation of the limio.Manager interface. It
//allows simple rate-based and arbitrary channel-based limits to be set.
//
//A SimpleManager is designed so that Limit and Manage may be called
//concurrently.
type SimpleManager struct {
	m map[Limiter]chan int

	newLimit chan *limit
	cls      chan struct{}

	newLimiter chan Limiter
	clsLimiter chan Limiter
}

//NewSimpleManager creates and initializes a SimpleManager.
func NewSimpleManager() *SimpleManager {
	lm := SimpleManager{
		m:          make(map[Limiter]chan int),
		newLimit:   make(chan *limit),
		cls:        make(chan struct{}),
		newLimiter: make(chan Limiter),
		clsLimiter: make(chan Limiter),
	}
	go lm.run()
	return &lm
}

//DefaultWindow is the window used to smooth SimpleLimit rates. That is,
//SimpleLimit distributes the given quantity evenly into buckets of size t.
//This is useful for avoiding tcp silly window syndrome and providing
//predictable resource usage.
const DefaultWindow = 10 * time.Millisecond

func (lm *SimpleManager) run() {
	limited := false
	cl := &limit{}
	ct := &time.Ticker{}

	er := rate{}

	for {
		select {
		case <-ct.C:
			lm.distribute(cl.rate.n)
		case tot := <-cl.lim:
			lm.distribute(tot)
		case newLim := <-lm.newLimit:
			go notify(cl.done, false)
			cl = &limit{}
			ct.Stop()

			if newLim != nil {
				limited = true
				cl = newLim

				for l := range lm.m {
					lm.limit(l)
				}

				if newLim.rate != er && cl.rate.n > 0 {
					cl.rate.n, cl.rate.t = Distribute(cl.rate.n, cl.rate.t, DefaultWindow)
					ct = time.NewTicker(cl.rate.t)
				}
			} else {
				limited = false
				for l := range lm.m {
					l.Unlimit()
				}
			}
		case l := <-lm.newLimiter:
			if limited {
				lm.limit(l)
			} else {
				l.Unlimit()
				lm.m[l] = nil
			}
		case toClose := <-lm.clsLimiter:
			close(lm.m[toClose])
			delete(lm.m, toClose)
		case <-lm.cls:
			for l := range lm.m {
				l.Unlimit()
			}
			go notify(cl.done, true)
			return
		}
	}
}

//NOTE must ONLY be used synchonously with the run() goroutine for concurrency
//safety.
//func distribute(int) takes a number and iterates over each channel in the map of
//managed Limiters, sending an evenly-distriuted limit to each "sublimiter".
//distribute takes a number to distribute and returns the number of bytes
//remaining
func (lm *SimpleManager) distribute(n int) int {
	if len(lm.m) > 0 {
		each := n / len(lm.m)
		for _, ch := range lm.m {
			if ch != nil {
				select {
				case ch <- each:
					n -= each
				default:
					//Skip if not ready
				}
			}
		}
	}
	return n
}

//NOTE must ONLY be used inside of run() for concurrency safety
//limit sets up a new channel for each limiter in the map. It then waits on the
//newly returned bool channel so that limiters can be removed when closed.
func (lm *SimpleManager) limit(l Limiter) {
	lm.m[l] = make(chan int)
	done := l.Limit(lm.m[l])
	go func() {
		//If `true` passed on channel, limiter is closed
		if <-done {
			lm.Unmanage(l)
		}
	}()
}

//NewReader takes an io.Reader and Limits it according to its limit
//policy/strategy
func (lm *SimpleManager) NewReader(r io.Reader) *Reader {
	lr := NewReader(r)
	lm.Manage(lr)
	return lr
}

//SimpleLimit takes an int and time.Duration that will be distributed evenly
//across all managed Limiters.
func (lm *SimpleManager) SimpleLimit(n int, t time.Duration) <-chan bool {
	done := make(chan bool, 1)
	lm.newLimit <- &limit{
		rate: rate{n, t},
		done: done,
	}
	return done
}

//Limit implements the limio.Limiter interface.
func (lm *SimpleManager) Limit(l chan int) <-chan bool {
	done := make(chan bool, 1)
	lm.newLimit <- &limit{
		lim:  l,
		done: done,
	}
	return done
}

//Unlimit implements the limio.Limiter interface.
func (lm *SimpleManager) Unlimit() {
	lm.newLimit <- nil
}

//Close allows the SimpleManager to free any resources it is using if the
//consumer has no further need for the SimpleManager.
func (lm *SimpleManager) Close() error {
	lm.cls <- struct{}{}
	return nil
}

//Unmanage allows consumers to remove a specific Limiter from its management
//strategy
func (lm *SimpleManager) Unmanage(l Limiter) {
	lm.clsLimiter <- l
}

//Manage takes a Limiter that will be adopted under the management policy of
//the SimpleManager.
func (lm *SimpleManager) Manage(l Limiter) error {
	if l == lm {
		return errors.New("A manager cannot manage itself.")
	}

	lm.newLimiter <- l
	return nil
}
