package limio

import "time"

type rate struct {
	n int
	t time.Duration
}

type limit struct {
	lim    <-chan int
	rate   rate
	notify chan<- bool
}

func (r *Reader) limit() {
	var remaining int

	var el = limit{}
	var er = rate{}
	var currLim limit
	var currNotify chan<- bool
	currTicker := &time.Ticker{}

	//Keep internal channel for sending limits to Read func
	//Keep internal ticker for n,t limits
	for {
		select {
		case <-r.cls:
			go notify(currNotify)
			currTicker.Stop()
			close(r.newLimit)
			close(r.rate)
			close(r.used)
			return
		case r.rate <- remaining:
			used := <-r.used
			remaining -= used
		case addl := <-currLim.lim:
			remaining += addl
		case <-currTicker.C:
			remaining += currLim.rate.n
		case l := <-r.newLimit:
			go notify(currNotify)

			if l != el {
				if l.rate != er {
					l.rate.n, l.rate.t = Distribute(l.rate.n, l.rate.t, DefaultWindow)
					currTicker = time.NewTicker(l.rate.t)
				} else {
					currTicker.Stop()
				}
				currLim = l
			} else {
				currTicker.Stop()
			}
		}
	}
}

func notify(n chan<- bool) {
	if n == nil {
		return
	}

	select {
	case n <- true:
	case <-time.After(30 * time.Second):
	}
	close(n)
}

func (r *Reader) Unlimit() {
	r.newLimit <- limit{}
}

func (r *Reader) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool)

	r.newLimit <- limit{
		rate:   rate{n, t},
		notify: done,
	}

	return done
}

func (r *Reader) LimitChan(lch chan int) <-chan bool {
	done := make(chan bool)
	r.newLimit <- limit{
		lim:    lch,
		notify: done,
	}
	return done
}
