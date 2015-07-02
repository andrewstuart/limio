package limio

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"net/http"
	_ "net/http/pprof"
)

func verifyIsManager(m Manager) {}

func TestManager(t *testing.T) {
	go http.ListenAndServe(":6060", nil)

	lmr := NewSimpleManager()

	verifyIsManager(lmr)

	ch := make(chan int, 1)
	lmr.Limit(ch)

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

	l3 := lmr.NewReader(strings.NewReader(testText))
	lmr.Manage(l3)

	ch <- 30

	n, err = l3.Read(p)

	if n != 10 {
		t.Errorf("Wrong bytes read")
	}

	l1.Read(p)
	l2.Read(p)
	lmr.Unmanage(l3)

	lmr.SimpleLimit(KB, 10*time.Millisecond)

	//Drain channel
	n, err = l1.Read(p)
	m, err := l2.Read(p)

	if n+m != 1024 {
		t.Errorf("Wrong number of bytes read: %d, should be 1024", n+m)
	}
	if err != nil {
		t.Errorf("Error reading: %v", err)
	}

	lmr.Unlimit()

	n, err = l1.Read(p)

	if n != len(testText)-20-m {
		t.Errorf("Wrong number read after unlimit: %d", n)
	}

	if err != nil && err != io.EOF {
		t.Errorf("Error reading unlimited: %v", err)
	}

	n, err = l2.Read(p)

	if n != len(testText)-20-m {
		t.Errorf("Wrong number read after unlimit: %d", n)
	}

	if err != nil && err != io.EOF {
		t.Errorf("Error reading unlimited: %v", err)
	}

	_, err = l2.Read(p)

	if err != io.EOF {
		t.Errorf("Should have thrown EOF after reached EOF.")
	}

	done := lmr.SimpleLimit(KB, time.Second)
	lmr.Manage(l3)

	w := &sync.WaitGroup{}
	w.Add(1)

	go func() {
		if cls := <-done; !cls {
			t.Errorf("Did not close \"done\" and pass true. Done: %t", cls)
		}
		w.Done()
	}()

	err = lmr.Close()

	if err != nil {
		t.Errorf("Error closing lmr: %v", err)
	}

	n, err = l3.Read(p)

	if n != len(testText)-10 {
		t.Errorf("Wrong number of bytes read: %d, should be %d", n, len(testText)-10)
	}
	w.Wait()
}
