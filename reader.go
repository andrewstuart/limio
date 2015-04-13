package limio

import (
	"io"
	"sync"
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

func NewReader(r io.Reader) *Reader {
	lr := Reader{
		r:        r,
		limitedM: &sync.RWMutex{},
		newLimit: make(chan *limit),
		rate:     make(chan int, 10),
		used:     make(chan int),
		cls:      make(chan bool),
	}
	go lr.limit()
	return &lr
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
			lim = len(p[written:])
		}
		r.limitedM.RUnlock()

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
