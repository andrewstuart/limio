package ratelimit

import (
	"bufio"
	"fmt"
	"io"
)

type ByteCount int64

type Limiter interface {
	GetLimit() <-chan ByteCount
}

type ChunkLimiter interface {
	Limiter
}

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
