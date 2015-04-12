package limio

import "io"

type Reader struct {
	r   io.Reader
	eof bool

	rate     chan int
	used     chan int
	newLimit chan limit
	cls      chan bool
}

func NewReader(r io.Reader) *Reader {
	lr := Reader{
		r:        r,
		newLimit: make(chan limit),
		rate:     make(chan int, 1000),
		used:     make(chan int),
		cls:      make(chan bool),
	}
	go lr.limit()
	return &lr
}

func (r *Reader) Close() error {
	r.Unlimit()
	r.cls <- true
	return nil
}

func (r *Reader) Read(p []byte) (written int, err error) {
	if r.eof {
		err = io.EOF
		return
	}

	for written < len(p) && err == nil {
		lim := <-r.rate

		if lim > len(p[written:]) {
			lim = len(p[written:])
		}

		var n int
		n, err = r.r.Read(p[written:][:lim])
		written += n
		r.used <- n

		if err != nil {
			if err == io.EOF {
				r.eof = true
			}
			return
		}
	}
	return
}
