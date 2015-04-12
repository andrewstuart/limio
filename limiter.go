package limio

import (
	"io"
	"time"
)

type Limiter interface {
	Limit(n int, t time.Duration) <-chan bool //The channel is useful for knowing that the channel has been unlimited
	LimitChan(chan int) <-chan bool
	Unlimit()
	io.Closer
}
