package ratelimit

import (
	"bufio"
	"fmt"
	"io"
)

type ByteCount int64

//A Limiter should implement some strategy for providing access to a shared io resource.
//The GetLimit() function must return a channel of ByteCount. When it is appropriate for
//the new limited io.Reader to read some amount of data, that amount should be sent
//through the channel, at which point the io.Reader will "burstily" read until it has
//exhausted the number of bytes it was told to read.
//
//Caution is recommended when implementing a Limiter if this bursty behavior is
//undesireable. If undesireable, make sure that any large ByteCounts are broken up
//into smaller values sent at shorter intervals.
type Limiter interface {
	GetLimit() <-chan ByteCount
}

//NewReader takes an io.Reader and a Limiter and returns an io.Reader that will
//be limited via the strategy that Limiter choses to implement.
func NewReader(r io.Reader, l Limiter) io.Reader {
	return &limitedReader{
		br: bufio.NewReader(r),
		c:  l.GetLimit(),
	}
}

type limitedReader struct {
	br        *bufio.Reader
	c         <-chan ByteCount
	eof       bool
	remaining int
}

func (lr *limitedReader) Read(p []byte) (written int, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	for written <= len(p) {
		if lr.remaining == 0 {
			if written > 0 {
				return
			}

			lr.remaining = int(<-lr.c)
		}

		var b byte
		for i := 0; i < int(lr.remaining) && i < len(p); i++ {
			b, err = lr.br.ReadByte()

			if err == io.EOF {
				lr.eof = true
				return
			}

			p[written] = b
			lr.remaining--
		}

	}
	fmt.Println(written)

	return
}
