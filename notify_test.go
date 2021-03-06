package limio

import "testing"

func TestNotify(t *testing.T) {
	ch := make(chan bool)

	notify(ch, true)

	if s, ok := <-ch; ok {
		t.Errorf("Did not close channel: %t, %t", s, ok)
	}

	ch = make(chan bool, 1)
	notify(ch, true)
	if !<-ch {
		t.Errorf("Did not send `true` down channel")
	}

	if _, ok := <-ch; ok {
		t.Errorf("Did not then close channel dch")
	}
}
