package limio

import (
	"bufio"
	"io"
)

//A ByteCount is simply an abstraction over some integer type to provide more
//flexibility should the type need to be changed. Since zettabyte overflows
//uint64, I suppose it may someday need to be changed.
type ByteCount uint64

//A Limiter should implement some strategy for providing access to a shared io
//resource.  The GetLimit() function must return a channel of ByteCount. When
//it is appropriate for the new limited io.Reader to read some amount of data,
//that amount should be sent through the channel, at which point the io.Reader
//will "burstily" read until it has exhausted the number of bytes it was told
//to read.
//
//Caution is recommended when implementing a Limiter if this bursty behavior is
//undesireable. If undesireable, make sure that any large ByteCounts are broken
//up into smaller values sent at shorter intervals. See BasicReader for a good
//example of how this can be achieved.
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
	remaining ByteCount
}

func (lr *limitedReader) Read(p []byte) (written int, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	var b byte
	for written < len(p) {
		if lr.remaining == 0 {
			select {
			case lr.remaining = <-lr.c:
				break
			default:
				if written > 0 {
					return
				} else {
					lr.remaining = <-lr.c
				}
			}
		}

		b, err = lr.br.ReadByte()

		if err == io.EOF {
			lr.eof = true
			return
		}

		p[written] = b
		written++
		lr.remaining--
	}
	return
}
