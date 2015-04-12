package limio

import (
	"math"
	"time"
)

const DefaultWindow = 100 * time.Microsecond

//Distribute takes a rate (n, t) and window (w), evenly distributes the n/t to
//n'/t' (n'<=n && t'>=w)
func Distribute(n int, t, w time.Duration) (int, time.Duration) {
	ratio := float64(t) / float64(w)
	nPer := float64(n) / ratio
	n = int(nPer)

	if nPer < 1.0 {
		t = time.Duration(math.Pow(nPer, -1))
		n = 1
	} else {
		t = w
	}

	return n, t
}
