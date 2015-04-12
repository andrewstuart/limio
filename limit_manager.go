package limio

import "time"

type managed struct {
	l   Limiter
	lim chan int
}

type LimitManager struct {
	m map[Limiter]*managed

	newLimit chan *limit
	unlimit  chan struct{}

	newLimiter chan *managed
	clsLimiter chan Limiter
}

func NewLimitManager() *LimitManager {
	lm := LimitManager{
		m:          make(map[Limiter]*managed),
		newLimit:   make(chan *limit),
		unlimit:    make(chan struct{}),
		newLimiter: make(chan *managed),
		clsLimiter: make(chan Limiter),
	}
	go lm.run()
	return &lm
}

func (lm *LimitManager) run() {
	currLim := &limit{}
	currTicker := &time.Ticker{}

	er := rate{}

	for {
		select {
		case <-currTicker.C:
			//Distribute
			each := currLim.rate.n / len(lm.m)
			for _, lim := range lm.m {
				lim.lim <- each
			}
		case tot := <-currLim.lim:
			each := tot / len(lm.m)
			for _, lim := range lm.m {
				lim.lim <- each
			}
		case newLim := <-lm.newLimit:
			if newLim != nil {
				if newLim.rate == er {
					currTicker.Stop()
				} else {
					n, t := Distribute(newLim.rate.n, newLim.rate.t, DefaultWindow)
					newLim.rate = rate{n, t}
				}

				currLim = newLim
			} else {
				currTicker.Stop()
			}
		case nl := <-lm.newLimiter:
			lm.m[nl.l] = nl
		case toClose := <-lm.clsLimiter:
			delete(lm.m, toClose)
		case <-lm.unlimit:
			for _, l := range lm.m {
				l.l.Unlimit()
			}
			return
		}
	}
}

func (lm *LimitManager) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool)
	return done
}

func (lm *LimitManager) LimitChan(chan int) <-chan bool {
	done := make(chan bool)
	return done
}

func (lm *LimitManager) Unlimit() {
	lm.unlimit <- struct{}{}
	close(lm.unlimit)
}

func (lm *LimitManager) Manage(l Limiter) {
	lim := make(chan int)
	done := l.LimitChan(lim)

	go func() {
		<-done
		lm.clsLimiter <- l
	}()

	lm.newLimiter <- &managed{l, lim}
}
