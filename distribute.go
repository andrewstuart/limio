package limio

import (
	"math"
	"time"
)

//Distribute takes a rate (n, t) and window (w), evenly distributes the n/t to
//n'/t' (n'<=n && t'>=w)
func Distribute(n uint64, t, w time.Duration) (uint64, time.Duration) {
	ratio := float64(t) / float64(window)
	nPer := float64(n) / ratio
	n = uint64(nPer)

	if nPer < 1.0 {
		t = time.Duration(math.Pow(nPer, -1))
		n = 1
	} else {
		t = w
	}

	return n, t
}
