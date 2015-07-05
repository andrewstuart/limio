package limio

import "time"

func NewBasicChan(n int, t time.Duration) <-chan int {
	ch := make(chan int)

	n2, t2 := Distribute(n, t, DefaultWindow)

	tkr := time.NewTicker(t2)

	return ch
}
