package ratelimit

import (
	"bufio"
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
