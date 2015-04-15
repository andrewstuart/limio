package limio

import (
	"io"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	go http.ListenAndServe(":6060", nil)

	lmr := NewSimpleManager()
	ch := make(chan int, 1)
	lmr.LimitChan(ch)

	l1 := lmr.NewReader(strings.NewReader(testText))
	l2 := lmr.NewReader(strings.NewReader(testText))

	p := make([]byte, len(testText))

	ch <- 20

	n, err := l1.Read(p)

	if n != 10 {
		t.Errorf("Wrong number of bytes read by n1: %d, expected 10", n)
	}

	if err != nil {
		t.Errorf("Error reading l1: %v", err)
	}

	n, err = l2.Read(p)

	if n != 10 {
		t.Errorf("Wrong number of bytes read by n2: %d, expected 10", n)
	}

	if err != nil {
		t.Errorf("Error reading l2: %v", err)
	}

	lmr.Limit(KB, time.Second)

	n, err = l1.Read(p)
	m, err := l2.Read(p)

	if n+m != 10 {
		t.Errorf("Wrong number of bytes read: %d, should be 50", n+m)
	}
	if err != nil {
		t.Errorf("Error reading: %v", err)
	}

	lmr.Unlimit()

	n, err = l1.Read(p)

	if n != len(testText)-10-m {
		t.Errorf("Wrong number read after unlimit: %d", n)
	}

	if err != nil && err != io.EOF {
		t.Errorf("Error reading unlimited: %v", err)
	}

	n, err = l2.Read(p)

	if n != len(testText)-10-m {
		t.Errorf("Wrong number read after unlimit: %d", n)
	}

	if err != nil && err != io.EOF {
		t.Errorf("Error reading unlimited: %v", err)
	}

	_, err = l2.Read(p)

	if err != io.EOF {
		t.Errorf("Should have thrown EOF after reached EOF.")
	}
}
