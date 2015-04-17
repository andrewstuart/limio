package limio

import (
	"io"
	"sync"
	"time"
)

type Reader struct {
	r   io.Reader
	eof bool

	limitedM *sync.RWMutex
	limited  bool

	rate     chan int
	used     chan int
	newLimit chan *limit
	cls      chan bool
}

type limit struct {
	lim  <-chan int
	rate rate
	done chan<- bool
}

type rate struct {
	n int
	t time.Duration
}

func NewReader(r io.Reader) *Reader {
	lr := Reader{
		r:        r,
		limitedM: &sync.RWMutex{},
		newLimit: make(chan *limit),
		rate:     make(chan int, 10),
		used:     make(chan int),
		cls:      make(chan bool),
	}
	go lr.run()
	return &lr
}

func (r *Reader) Unlimit() {
	r.newLimit <- nil
}

func (r *Reader) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool, 1)
	r.newLimit <- &limit{
		rate: rate{n, t},
		done: done,
	}
	return done
}

func (r *Reader) LimitChan(lch chan int) <-chan bool {
	done := make(chan bool, 1)
	r.newLimit <- &limit{
		lim:  lch,
		done: done,
	}
	return done
}

func (r *Reader) Close() error {
	r.cls <- true
	return nil
}

func (r *Reader) Read(p []byte) (written int, err error) {
	if r.eof {
		err = io.EOF
		return
	}

	var n int
	var lim int
	for written < len(p) && err == nil {

		r.limitedM.RLock()
		if r.limited {
			r.limitedM.RUnlock()
			select {
			case lim = <-r.rate:
			default:
				if written > 0 {
					return
				} else {
					lim = <-r.rate
				}
			}
		} else {
			r.limitedM.RUnlock()
			lim = len(p[written:])
		}

		if lim > len(p[written:]) {
			lim = len(p[written:])
		}

		n, err = r.r.Read(p[written:][:lim])
		written += n

		if err != nil {
			if err == io.EOF {
				r.eof = true
			}
			return
		}
	}
	return
}

const DefaultWindow = 10 * time.Millisecond

func (r *Reader) send(i int) {
	select {
	case r.rate <- i:
	default:
	}
}

func (r *Reader) run() {
	er := rate{}
	cl := &limit{}

	currTicker := &time.Ticker{}

	for {
		select {
		case <-r.cls:
			r.limitedM.Lock()
			r.limited = false
			r.limitedM.Unlock()

			currTicker.Stop()
			go notify(cl.done, true)

			close(r.newLimit)
			close(r.used)
			close(r.rate)

			return
		case l := <-cl.lim:
			r.send(l)
		case <-currTicker.C:
			r.send(cl.rate.n)
		case l := <-r.newLimit:
			go notify(cl.done, false)
			currTicker.Stop()
			cl = &limit{}

			if l != nil {
				cl = l
				r.limitedM.Lock()
				r.limited = true
				r.limitedM.Unlock()

				if cl.rate != er && cl.rate.n != 0 {
					cl.rate.n, cl.rate.t = Distribute(cl.rate.n, cl.rate.t, DefaultWindow)
					currTicker = time.NewTicker(cl.rate.t)
				}
			} else {
				r.limitedM.Lock()
				r.limited = false
				r.limitedM.Unlock()

				r.rate <- 0 //Unlock any readers waiting for a value
			}
		}
	}
}
