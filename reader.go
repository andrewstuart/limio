package limio

import (
	"io"
	"math"
	"sync"
	"time"
)

const (
	window  = 100 * time.Microsecond
	Max64   = 1<<64 - 1
	bufsize = KB << 3
)

//Distribute takes a rate (n, t) and window (w), evenly distributes the n/t to
//n'/t' (n'<=n && t'>=w)
func Distribute(n uint64, t, w time.Duration) (uint64, time.Duration) {
	ratio := float64(t) / float64(window)
	nPer := float64(n) / ratio
	n = uint64(nPer)

	if nPer < 1.0 {
		t = time.Duration(math.Pow(nPer, -1))
		n = 1
	} else {
		t = w
	}

	return n, t
}

//A Limiter gets some stream of things. It only lets N things through per T time.
//It can either be given both T and N or block and wait for N on a channel.
//In this case, T is handled externally to the limiter.
//
//Most simply, underlying implementation should always use a channel. It can be
//replaced with a new channel directly or by providing N and T at which point
//an internal channel is set up.
type Limiter interface {
	Limit(n uint64, t time.Duration)
	LimitChan(<-chan uint64)
}

//LimitWaiter is a limiter that also exposes the ability to know when the
//underlying data has been completed. Wait() in this context is identical to
//sync.WaitGroup.Wait() (and is most likely implemented with a sync.WaitGroup)
//in that it blocks until it is freed by the underlying implementation.
type LimitWaiter interface {
	Limiter
	Wait()
}

//Reader is the main implentation of interest for probably most people. It
//takes an io.Reader and returns Reader object which implements, among others,
//the io.Reader and Limiter interfaces
type Reader struct {
	r      io.Reader
	buf    []byte
	eof    bool
	done   *sync.WaitGroup
	remain uint64

	rMut sync.RWMutex
	rate <-chan uint64
}

func (r *Reader) rater() <-chan uint64 {
	r.rMut.RLock()
	defer r.rMut.RUnlock()
	return r.rate
}

func (r *Reader) setRate(ra <-chan uint64) {
	r.rMut.Lock()
	r.rate = ra
	r.rMut.Unlock()
}

func (r *Reader) Read(p []byte) (written int, err error) {
	if r.eof {
		err = io.EOF
		return
	}

	if r.buf == nil {
		r.buf = make([]byte, bufsize)
	}

	for written < len(p) && err == nil {
		var lim uint64
		if r.rater() != nil {
			if r.remain == 0 {
				select {
				case r.remain = <-r.rater():
				default:
					if written > 0 {
						return
					}
					r.remain = <-r.rater()
				}
			}
			lim = r.remain
		} else {
			lim = Max64
		}

		if lim > uint64(len(p[written:])) {
			lim = uint64(len(p[written:]))
		}

		if lim > uint64(len(r.buf)) {
			lim = uint64(len(r.buf))
		}

		var n int
		n, err = r.r.Read(r.buf[:lim])

		copy(p[written:], r.buf[:n])
		written += n

		if r.rater() != nil {
			r.remain -= uint64(n)
		}

		if err != nil {
			if err == io.EOF {
				r.eof = true
				r.done.Done()
			}
			return
		}
	}
	return
}

//Limit provides a basic means for limiting a Reader. Given n bytes per t
//time, it does its best to maintain a constant rate with a high degree of
//accuracy to allow other algorithms (such as TCP window sizing, e.g.) to
//self-adjust.
//
//It is safe to assume that Reader.Limit can be called concurrently with a Read
//operation, though the Read operation will continue to use the prior rate until
//it is requests a rate update.
func (r *Reader) Limit(n uint64, t time.Duration) {
	//Skip calc if unneeded
	if t != window {
		n, t = Distribute(n, t, window)
	}

	//TODO make sure no memory leaks
	ch := make(chan uint64)
	r.setRate(ch)

	go func() {
		tkr := time.NewTicker(t)

		for !r.eof {
			<-tkr.C
			ch <- n
		}

		close(ch)
		tkr.Stop()
	}()
}

//LimitChan exposes the ability to manage the underlying reader with more
//precision and flexibiity. Each uint64 sent on the channel provided is treated
//as a number of bytes to burstily/greedily allow to be Read. It is recommended
//that most implementations take caution to avoid thrashing especially when
//using a TCP connection, as TCP has algorithms to manage its window size.
//http://en.wikipedia.org/wiki/Silly_window_syndrome
func (r *Reader) LimitChan(c <-chan uint64) {
	r.setRate(c)
}

//Wait will block until eof is reached. Once reached, any errors will be
//returned. It is intended to provide synchronization for external channel
//managers
func (r *Reader) Wait() {
	r.done.Wait()
	return
}

//NewReader takes an io.Reader and returns a Limitable Reader.
func NewReader(r io.Reader) *Reader {
	switch r := r.(type) {
	case *Reader:
		return r
	default:
		nr := Reader{
			r:    r,
			buf:  make([]byte, bufsize),
			done: &sync.WaitGroup{},
			rMut: sync.RWMutex{},
		}
		nr.done.Add(1)
		return &nr
	}
}
