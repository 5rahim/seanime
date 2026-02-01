package limiter

import (
	"sync"
	"time"
)

// https://stackoverflow.com/a/72452542

func NewAnilistLimiter() *Limiter {
	//return NewLimiter(15*time.Second, 18)
	//return NewLimiter(6*time.Second, 8)
	return NewLimiter(10*time.Second, 5)
}

//----------------------------------------------------------------------------------------------------------------------

type Limiter struct {
	tick    time.Duration
	count   uint
	entries []time.Time
	index   uint
	mu      sync.Mutex
}

func NewLimiter(tick time.Duration, count uint) *Limiter {
	l := Limiter{
		tick:  tick,
		count: count,
		index: 0,
	}
	l.entries = make([]time.Time, count)
	before := time.Now().Add(-2 * tick)
	for i := range l.entries {
		l.entries[i] = before
	}
	return &l
}

func (l *Limiter) Wait() {
	l.mu.Lock()
	defer l.mu.Unlock()
	last := &l.entries[l.index]
	next := last.Add(l.tick)
	now := time.Now()
	if now.Before(next) {
		time.Sleep(next.Sub(now))
	}
	*last = time.Now()
	l.index = l.index + 1
	if l.index == l.count {
		l.index = 0
	}
}
