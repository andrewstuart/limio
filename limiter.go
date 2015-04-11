package limio

import "time"

type Limiter interface {
	Limit(n uint64, t time.Duration) <-chan bool //The channel is useful for knowing that the channel has been unlimited
	LimitChan(chan uint64) <-chan bool
	Unlimit()
}
