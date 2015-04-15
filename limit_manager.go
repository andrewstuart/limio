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
			go notify(cl.notify, false)
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
			go notify(cl.notify, true)
			return
		}
	}
}

//NOTE must ONLY be used synchonously with the run() goroutine for concurrency
//safety
func (lm *SimpleManager) distribute(n int) {
	if len(lm.m) > 0 {
		each := n / len(lm.m)
		for _, ch := range lm.m {
			if ch != nil {
				select {
				case ch <- each:
				default:
					//Assume Close() will be called soon
				}
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
