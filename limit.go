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

const DefaultWindow = 10 * time.Millisecond

func (r *Reader) limit() {
	pool := make(chan int, 10)

	go func() {
		for {
			rt, ok := <-pool
			if !ok {
				close(r.rate)
				return
			}

			select {
			case r.rate <- rt:
			case <-time.After(2 * DefaultWindow):
			}
		}
	}()

	er := rate{}
	currLim := &limit{}

	currTicker := &time.Ticker{}

	for {
		select {
		case <-r.cls:
			r.limitedM.Lock()
			r.limited = false
			r.limitedM.Unlock()

			go notify(currLim.notify, true)
			currTicker.Stop()
			close(pool)
			close(r.newLimit)
			close(r.used)

			return
		case l := <-currLim.lim:
			pool <- l
		case <-currTicker.C:
			pool <- currLim.rate.n
		case l := <-r.newLimit:
			go notify(currLim.notify, false)

			if l != nil {
				r.limitedM.Lock()
				r.limited = true
				r.limitedM.Unlock()

				if l.rate != er {
					l.rate.n, l.rate.t = Distribute(l.rate.n, l.rate.t, DefaultWindow)
					currTicker = time.NewTicker(l.rate.t)
				} else {
					currTicker.Stop()
				}

				currLim = l
			} else {
				currTicker.Stop()
				currLim = &limit{}

				r.limitedM.Lock()
				r.limited = false
				r.limitedM.Unlock()

				//Potentially need to unblock reader waiting for a value
				pool <- 1
			}
		}
	}
}

func notify(n chan<- bool, v bool) {
	if n == nil {
		return
	}

	select {
	case n <- v:
	case <-time.After(30 * time.Second):
	}
	close(n)
}

func (r *Reader) Unlimit() {
	r.newLimit <- nil
}

func (r *Reader) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool)
	r.newLimit <- &limit{
		rate:   rate{n, t},
		notify: done,
	}
	return done
}

func (r *Reader) LimitChan(lch chan int) <-chan bool {
	done := make(chan bool, 1)
	r.newLimit <- &limit{
		lim:    lch,
		notify: done,
	}
	return done
}
