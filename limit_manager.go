package limio

import (
	"io"
	"time"
)

type LimitManager struct {
	m map[Limiter]chan int

	newLimit chan *limit
	unlimit  chan struct{}
	cls      chan struct{}

	newLimiter chan Limiter
	clsLimiter chan Limiter
}

func NewLimitManager() *LimitManager {
	lm := LimitManager{
		m:          make(map[Limiter]chan int),
		newLimit:   make(chan *limit),
		unlimit:    make(chan struct{}),
		cls:        make(chan struct{}),
		newLimiter: make(chan Limiter),
		clsLimiter: make(chan Limiter),
	}
	go lm.run()
	return &lm
}

func (lm *LimitManager) run() {
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
				} else {
					n, t := Distribute(newLim.rate.n, newLim.rate.t, DefaultWindow)
					newLim.rate = rate{n, t}
					currTicker = time.NewTicker(t)
				}

				currLim = newLim
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
		case <-lm.unlimit:
			for l := range lm.m {
				l.Unlimit()
			}
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
func (lm *LimitManager) distribute(n int) {
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
func (lm *LimitManager) limit(l Limiter) {
	lm.m[l] = make(chan int)
	done := l.LimitChan(lm.m[l])

	go func() {
		if <-done {
			lm.clsLimiter <- l
		}
	}()
}

func (lm *LimitManager) NewReader(r io.Reader) *Reader {
	lr := NewReader(r)
	lm.Manage(lr)
	return lr
}

func (lm *LimitManager) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool)
	lm.newLimit <- &limit{
		rate:   rate{n, t},
		notify: done,
	}
	return done
}

func (lm *LimitManager) LimitChan(l chan int) <-chan bool {
	done := make(chan bool)
	lm.newLimit <- &limit{
		lim:    l,
		notify: done,
	}
	return done
}

func (lm *LimitManager) Unlimit() {
	lm.unlimit <- struct{}{}
}

func (lm *LimitManager) Close() error {
	lm.cls <- struct{}{}
	return nil
}

func (lm *LimitManager) Manage(l Limiter) {
	lm.newLimiter <- l
}
