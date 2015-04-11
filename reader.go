package limio

import (
	"bufio"
	"io"
	"time"
)

type rate struct {
	n uint64
	t time.Duration
}

type limit struct {
	lim    <-chan uint64
	rate   rate
	notify chan<- bool
}

type Reader struct {
	br *bufio.Reader

	newLimit chan limit
}

func NewReader(r io.Reader) *Reader {
	lr := Reader{
		br: bufio.NewReader(r),
	}
	go lr.limit()
	return &lr
}

func (r *Reader) limit() {
	if r.newLimit == nil {
		r.newLimit = make(chan limit)
	}

	//Keep internal channel for sending limits to Read func
	//Keep internal ticker for n,t limits
	for {
		select {}
	}
}

func (r *Reader) Unlimit() {
}

func (r *Reader) Limit(n uint64, t time.Duration) <-chan bool {
	done := make(chan bool)

	r.newLimit <- limit{
		rate:   rate{n, t},
		notify: done,
	}

	return done
}

func (r *Reader) LimitChan(lch chan uint64) <-chan bool {
	done := make(chan bool)
	r.newLimit <- limit{
		lim:    lch,
		notify: done,
	}
	return done
}
