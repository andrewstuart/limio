package limio

import (
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"

	_ "net/http/pprof"
)

func verifyIsManager(m Manager) {}

func TestManager(t *testing.T) {
	asrt := assert.New(t)
	lmr := NewSimpleManager()

	asrt.Implements((*Manager)(nil), lmr)

	ch := make(chan int, 1)
	lmr.Limit(ch)

	l1 := lmr.NewReader(strings.NewReader(testText))
	l2 := lmr.NewReader(strings.NewReader(testText))

	p := make([]byte, len(testText)+1)

	ch <- 20

	glog.Info("Reading 1")

	n, err := l1.Read(p)

	glog.Infof("Read 1: %d", n)

	asrt.NoError(err)
	asrt.Equal(10, n)

	glog.Info("Reading 2")

	n, err = l2.Read(p)

	glog.Infof("Read 2: %d", n)

	asrt.NoError(err)
	asrt.Equal(10, n)

	l3 := lmr.NewReader(strings.NewReader(testText))
	asrt.NoError(lmr.Manage(l3))

	ch <- 30

	glog.Info("Reading 3")

	n, err = l3.Read(p)
	assert.NoError(t, err)
	asrt.Equal(n, 10)

	glog.Infof("Read 3: %d", n)

	_, err = l1.Read(p)
	asrt.NoError(err)
	_, err = l2.Read(p)
	asrt.NoError(err)
	lmr.Unmanage(l3)

	glog.Info("unmanage done")

	lmr.SimpleLimit(KB, 10*time.Millisecond)

	//Drain channel
	n, err = l1.Read(p)
	asrt.NoError(err)
	m, err := l2.Read(p)

	asrt.NoError(err)
	asrt.Equal(1024, n+m, "Wrong number of total bytes read")
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
