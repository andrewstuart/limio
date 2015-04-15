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
	pool := make(chan int)

	go func() {
		for {
			rt, ok := <-pool
			if !ok {
				close(r.rate)
				return
			}

			select {
			case r.rate <- rt:
			default:
			}
		}
	}()

	er := rate{}
	cl := &limit{}

	currTicker := &time.Ticker{}

	for {
		select {
		case <-r.cls:
			r.limitedM.Lock()
			r.limited = false
			r.limitedM.Unlock()

			go notify(cl.notify, true)
			currTicker.Stop()
			close(pool)
			close(r.newLimit)
			close(r.used)

			return
		case l := <-cl.lim:
			pool <- l
		case <-currTicker.C:
			pool <- cl.rate.n
		case l := <-r.newLimit:
			go notify(cl.notify, false)
			currTicker.Stop()

			if l != nil {
				r.limitedM.Lock()
				r.limited = true
				r.limitedM.Unlock()
				cl = l

				if cl.rate != er && cl.rate.n != 0 {
					cl.rate.n, cl.rate.t = Distribute(cl.rate.n, cl.rate.t, DefaultWindow)
					currTicker = time.NewTicker(cl.rate.t)
				}
			} else {
				r.limitedM.Lock()
				r.limited = false
				r.limitedM.Unlock()

				cl = &limit{}
				r.rate <- 0 //Unlock any readers waiting for a value
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
	done := make(chan bool)
	r.newLimit <- &limit{
		lim:    lch,
		notify: done,
	}
	return done
}
