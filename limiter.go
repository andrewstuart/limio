package limio

import "io"

//A Limiter is an interface that meters some underlying discretely quantifiable
//operation with respect to time.
type Limiter interface {
	Limit(chan int) <-chan bool //The channel is useful for knowing that the channel has been unlimited
	Unlimit()
	io.Closer
}
