package ratelimit

import (
	"bufio"
	"bytes"
	"io"
	"time"
)

type ReadLimiter struct {
	Window  time.Duration `json:"window"xml:"window"`
	br      *bufio.Reader
	tokens  chan struct{}
	newRate chan rate
}

type rate struct {
	limit int64
	time  time.Duration
}

func (rl *ReadLimiter) Read(p []byte) (n int64, err error) {
	i := int64(0)

getTokens:
	for {
		select {
		case <-rl.tokens:
			i++
		default:
			break getTokens
		}
	}

	return io.CopyN(bytes.NewBuffer(p), rl.br, i)
}

func (rl *ReadLimiter) SetLimit(lim int64, win time.Duration) {
	rl.newRate <- rate{lim, win}
	return
}

func NewLimitedReader(r io.Reader, lim int64, window time.Duration) *ReadLimiter {
	rl := ReadLimiter{
		br:      bufio.NewReader(r),
		newRate: make(chan rate, 1),
		tokens:  make(chan struct{}, lim),
	}

	rl.SetLimit(lim, window)

	go func() {
		for {
			select {
			case newLimit := <-rl.newRate:
				window = newLimit.time
				lim = newLimit.limit
			default:
				for i := int64(0); i < lim; i++ {
					rl.tokens <- struct{}{}
				}
				time.Sleep(window)
			}
		}
	}()

	return &rl
}

func NewReader(r io.Reader) *ReadLimiter {
	return &ReadLimiter{
		br: bufio.NewReader(r),
	}
}
