package limio

import (
	"fmt"
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
		rate:     make(chan int),
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
	done := make(chan bool)
	r.newLimit <- &limit{
		rate: rate{n, t},
		done: done,
	}
	return done
}

func (r *Reader) LimitChan(lch chan int) <-chan bool {
	done := make(chan bool)
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
					fmt.Println(lim)
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

func (r *Reader) run() {
	pool := make(chan int)
	defer close(pool)

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
	cl := &limit{}

	currTicker := &time.Ticker{}

	for {
		select {
		case <-r.cls:
			r.limitedM.Lock()
			r.limited = false
			r.limitedM.Unlock()

			go notify(cl.done, true)
			currTicker.Stop()
			close(r.newLimit)
			close(r.used)

			return
		case l := <-cl.lim:
			pool <- l
		case <-currTicker.C:
			pool <- cl.rate.n
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

func notify(n chan<- bool, v bool) {
	if n == nil {
		return
	}

	select {
	case n <- v:
	default: //Don't wait for somebody to listen on the channel
	}
	close(n)
}
