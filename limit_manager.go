package limio

import "time"

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
			if len(lm.m) > 0 {
				//Distribute
				each := currLim.rate.n / len(lm.m)
				for _, lch := range lm.m {
					lch <- each
				}
			}
		case tot := <-currLim.lim:
			if len(lm.m) > 0 {
				each := tot / len(lm.m)
				for _, lch := range lm.m {
					lch <- each
				}
			}
		case newLim := <-lm.newLimit:
			notify(currLim.notify, false)
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
				//still store in map
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

//NOTE must ONLY be used inside of run()
func (lm *LimitManager) limit(l Limiter) {
	lm.m[l] = make(chan int)
	go func() {
		if <-l.LimitChan(lm.m[l]) {
			lm.clsLimiter <- l
		}
	}()
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
	l.Unlimit()
	lm.newLimiter <- l
}
