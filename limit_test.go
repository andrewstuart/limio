package limio

import (
	"io"
	"strings"
	"testing"
)

func TestUnlimit(t *testing.T) {
	r := NewReader(strings.NewReader(testText))

	ch := make(chan int, 1)
	ch <- 20
	r.LimitChan(ch)

	p := make([]byte, len(testText))

	n, err := r.Read(p)

	if err != nil {
		t.Errorf("Got an error reading: %v", err)
	}

	if n != 20 {
		t.Errorf("Read wrong number of bytes: %d", n)
	}

	r.Unlimit()

	n, err = r.Read(p)

	if err != io.EOF {
		t.Errorf("Error reading: %v", err)
	}

	if n != len(testText)-20 {
		t.Errorf("Wrong number of bytes read: %d, should be %d", n, len(testText)-20)
	}
}

func TestClose(t *testing.T) {
	r := NewReader(strings.NewReader(testText))

	ch := make(chan int, 1)
	r.LimitChan(ch)
	err := r.Close()

	if err != nil {
		t.Fatalf("Close did not work: %v", err)
	}

	p := make([]byte, len(testText))

	n, err := r.Read(p)

	if n != len(p) {
		t.Errorf("Wrong number of bytes reported read: %d", n)
	}

	if err != nil && err != io.EOF {
		t.Errorf("Non-EOF error reported for closed limiter: %v", err)
	}
}
