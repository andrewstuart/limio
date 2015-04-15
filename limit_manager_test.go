package limio

import (
	"strings"
	"testing"
)

func TestManager(t *testing.T) {
	lmr := NewLimitManager()
	ch := make(chan int, 1)
	lmr.LimitChan(ch)

	l1 := lmr.NewReader(strings.NewReader(testText))
	l2 := lmr.NewReader(strings.NewReader(testText))

	p1 := make([]byte, len(testText))
	p2 := make([]byte, len(testText))

	ch <- 20

	n, err := l1.Read(p1)

	if n != 10 {
		t.Errorf("Wrong number of bytes read by n1: %d, expected 10", n)
	}

	if err != nil {
		t.Errorf("Error reading l1: %v", err)
	}

	n, err = l2.Read(p2)

	if n != 10 {
		t.Errorf("Wrong number of bytes read by n2: %d, expected 10", n)
	}

	if err != nil {
		t.Errorf("Error reading l2: %v", err)
	}

}
