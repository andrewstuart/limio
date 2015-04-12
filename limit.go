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
	pool := make(chan int, 1000)

	go func() {
		for {
			rt, ok := <-pool
			if !ok {
				close(r.rate)
				return
			}
			r.rate <- rt
		}
	}()

	er := rate{}
	currLim := &limit{
		rate: rate{n: 1<<63 - 1},
	}

	var currNotify chan<- bool
	currTicker := time.NewTicker(DefaultWindow)

	for {
		select {
		case <-r.cls:
			go notify(currNotify)
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
			go notify(currNotify)

			if l != nil {
				if l.rate != er {
					l.rate.n, l.rate.t = Distribute(l.rate.n, l.rate.t, DefaultWindow)
					currTicker = time.NewTicker(l.rate.t)
				} else {
					currTicker.Stop()
				}
				currLim = l
			} else {
				currTicker = time.NewTicker(DefaultWindow)
				currLim.rate.n = 1<<63 - 1
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
	done := make(chan bool)
	r.newLimit <- &limit{
		lim:    lch,
		notify: done,
	}
	return done
}
