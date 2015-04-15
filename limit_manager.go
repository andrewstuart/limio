package limio

import (
	"io"
	"time"
)

type Manager interface {
	Limiter
	Manage(Limiter)
	Unmanage(Limiter)
}

type SimpleManager struct {
	m map[Limiter]chan int

	newLimit chan *limit
	cls      chan struct{}

	newLimiter chan Limiter
	clsLimiter chan Limiter
}

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

func (lm *SimpleManager) run() {
	limited := false
	currLim := &limit{}
	currTicker := &time.Ticker{}

	er := rate{}

	for {
		select {
		case <-currTicker.C:
			lm.distribute(currLim.rate.n)
		case tot := <-currLim.lim:
			lm.distribute(tot)
		case newLim := <-lm.newLimit:
			go notify(currLim.notify, false)

			if newLim != nil {
				limited = true

				for l := range lm.m {
					lm.limit(l)
				}

				if newLim.rate == er {
					currTicker.Stop()
					currLim.lim = newLim.lim
				} else {
					if newLim.rate.n == 0 {
						currTicker.Stop()
					} else {
						currLim.rate.n, currLim.rate.t = Distribute(newLim.rate.n, newLim.rate.t, DefaultWindow)
						currTicker = time.NewTicker(newLim.rate.t)
					}
				}
			} else {
				limited = false
				for l := range lm.m {
					l.Unlimit()
				}
				currTicker.Stop()
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
			notify(currLim.notify, true)
			return
		}
	}
}

//NOTE must ONLY be used inside of run() for concurrency safety
func (lm *SimpleManager) distribute(n int) {
	if len(lm.m) > 0 {
		each := n / len(lm.m)
		for _, ch := range lm.m {
			if ch != nil {
				ch <- each
			}
		}
	}
}

//NOTE must ONLY be used inside of run() for concurrency safety
func (lm *SimpleManager) limit(l Limiter) {
	lm.m[l] = make(chan int)
	done := l.LimitChan(lm.m[l])
	go func() {
		if <-done {
			lm.Unmanage(l)
		}
	}()
}

func (lm *SimpleManager) NewReader(r io.Reader) *Reader {
	lr := NewReader(r)
	lm.Manage(lr)
	return lr
}

func (lm *SimpleManager) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool)
	lm.newLimit <- &limit{
		rate:   rate{n, t},
		notify: done,
	}
	return done
}

func (lm *SimpleManager) LimitChan(l chan int) <-chan bool {
	done := make(chan bool)
	lm.newLimit <- &limit{
		lim:    l,
		notify: done,
	}
	return done
}

func (lm *SimpleManager) Unlimit() {
	lm.newLimit <- nil
}

func (lm *SimpleManager) Close() error {
	lm.cls <- struct{}{}
	return nil
}

func (lm *SimpleManager) Unmanage(l Limiter) {
	lm.clsLimiter <- l
}

func (lm *SimpleManager) Manage(l Limiter) {
	lm.newLimiter <- l
}
