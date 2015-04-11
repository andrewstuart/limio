package limio

import (
	"io"
	"time"
)

type Limiter interface {
	Limit(n uint64, t time.Duration) <-chan bool //The channel is useful for knowing that the channel should be unlimited
	LimitChan(chan uint64) <-chan bool
	io.Closer
}
