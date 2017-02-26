package limio

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/golang/glog"
)

//Reader implements an io-limited reader that conforms to the io.Reader and
//limio.Limiter interface. Reader can have its limits updated concurrently with
//any Read() calls.
type Reader struct {
	r   io.Reader
	eof bool

	limitedM *sync.RWMutex
	limited  bool

	timeoutM *sync.Mutex
	timeout  time.Duration

	rate     chan int
	used     chan int
	newLimit chan *limit
	cls      chan bool
}

type limit struct {
	lim   <-chan int
	rate  rate
	ready chan<- struct{}
	done  chan<- bool
}

type rate struct {
	n int
	t time.Duration
}

//NewReader takes any io.Reader and returns a limio.Reader.
func NewReader(r io.Reader) *Reader {
	lr := Reader{
		r:        r,
		limitedM: &sync.RWMutex{},
		timeoutM: &sync.Mutex{},
		newLimit: make(chan *limit),
		rate:     make(chan int, 10),
		used:     make(chan int),
		cls:      make(chan bool),
	}
	go lr.run()
	return &lr
}

//Unlimit removes any restrictions on the underlying io.Reader.
func (r *Reader) Unlimit() {
	r.newLimit <- nil
}

//SimpleLimit takes an integer and a time.Duration and limits the underlying
//reader non-burstily (given rate is averaged over a small time).
func (r *Reader) SimpleLimit(n int, t time.Duration) <-chan bool {
	done := make(chan bool, 1)
	ready := make(chan struct{})
	r.newLimit <- &limit{
		rate:  rate{n, t},
		done:  done,
		ready: ready,
	}
	<-ready
	return done
}

//Limit can be used to precisely control the limit at which bytes can be Read,
//whether burstily or not.
func (r *Reader) Limit(lch chan int) <-chan bool {
	done := make(chan bool, 1)
	ready := make(chan struct{})
	r.newLimit <- &limit{
		lim:   lch,
		done:  done,
		ready: ready,
	}
	<-ready
	return done
}

//Close allows the goroutines that were managing limits and reads to shut down
//and free up memory. It should be called by any clients of the limio.Reader,
//much as http.Response.Body should be closed to free up system resources.
func (r *Reader) Close() error {
	r.cls <- true
	return nil
}

//ErrTimeoutExceeded will be returned upon a timeout lapsing without a read occuring
var ErrTimeoutExceeded error = errors.New("Timeout Exceeded")

//Read implements io.Reader in a blocking manner according to the limits of the
//limio.Reader.
func (r *Reader) Read(p []byte) (written int, err error) {
	if r.eof {
		err = io.EOF
		return
	}

	var n int
	var lim int
	for written < len(p) && err == nil {

		r.limitedM.RLock()
		isLimited := r.limited
		r.limitedM.RUnlock()

		if isLimited {

			r.timeoutM.Lock()
			timeLimit := r.timeout
			r.timeoutM.Unlock()

			//TODO consolidate two cases if possible. Dynamic select via reflection?
			if timeLimit > 0 {
				select {
				case <-time.After(timeLimit):
					err = ErrTimeoutExceeded
					return
				case lim = <-r.rate:
				default:
					if written > 0 {
						return
					}
					lim = <-r.rate
				}
			} else {
				select {
				case lim = <-r.rate:
				default:
					if written > 0 {
						return
					}
					lim = <-r.rate
				}
			}
		} else {
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

func (r *Reader) sendIfReady(i int) {
	select {
	case r.rate <- i:
	default:
	}
}

//SetTimeout takes some time.Duration t and configures the underlying Reader to
//return a limio.TimedOut error if the timeout is exceeded while waiting for a
//read operation.
func (r *Reader) SetTimeout(t time.Duration) error {
	r.timeoutM.Lock()
	r.timeout = t
	r.timeoutM.Unlock()
	return nil
}

func (r *Reader) run() {
	emptyRate := rate{}
	currLim := &limit{}

	rateTicker := &time.Ticker{}

	//This loop is important for serializing access to the limits and the
	//io.Reader being managed
	for {
		select {
		case <-r.cls:
			r.limitedM.Lock()
			r.limited = false
			r.limitedM.Unlock()

			rateTicker.Stop()
			go notify(currLim.done, true)

			close(r.newLimit)
			close(r.used)
			close(r.rate)

			return
		case l := <-currLim.lim:
			r.sendIfReady(l)
		case <-rateTicker.C:
			r.sendIfReady(currLim.rate.n)
		case l := <-r.newLimit:
			glog.V(9).Infof("Reader got a new limit: %#v", l)
			go notify(currLim.done, false)
			rateTicker.Stop()
			currLim = &limit{}

			if l != nil {
				currLim = l
				r.limitedM.Lock()
				r.limited = true
				r.limitedM.Unlock()

				if currLim.rate != emptyRate && currLim.rate.n != 0 {
					currLim.rate.n, currLim.rate.t = Distribute(currLim.rate.n, currLim.rate.t, DefaultWindow)
					rateTicker = time.NewTicker(currLim.rate.t)
				}

				close(l.ready)
			} else {
				r.limitedM.Lock()
				r.limited = false
				r.limitedM.Unlock()

				r.rate <- 0 //Unlock any readers waiting for a value
			}
		}
	}
}
