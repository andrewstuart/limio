package limio

import "time"

type LimitManager struct {
	m map[Limiter]Limiter
}

func (lm *LimitManager) Limit(n int, t time.Duration) <-chan bool {
	done := make(chan bool)
	return done
}

func (lm *LimitManager) LimitChan(chan int) <-chan bool {
	done := make(chan bool)
	return done
}

func (lm *LimitManager) Close() error {
	return nil
}

func (lm *LimitManager) Manager(l Limiter) {
	lim := make(chan int)
	done := l.LimitChan(lim)
	<-done
}
