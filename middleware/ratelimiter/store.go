package ratelimiter

import (
	"sync"
	"time"
)

type limiter struct {
	requestQuota  int
	counter       int
	quotaDuration time.Duration
	updatedAt     time.Time
}

func newLimiter(requestQuota int, quotaDuration time.Duration) *limiter {
	return &limiter{requestQuota: requestQuota, quotaDuration: quotaDuration, counter: 0, updatedAt: time.Now()}
}

func (l *limiter) reset() {
	l.updatedAt = time.Now()
	l.counter = 0
}

func (l *limiter) increment() {
	l.counter++
}

type limiterStore struct {
	mx    *sync.RWMutex
	store map[string]*limiter
	s     *sync.Map
}

func newLimiterStore() limiterStore {
	return limiterStore{
		mx:    new(sync.RWMutex),
		store: make(map[string]*limiter),
	}
}

func (ls *limiterStore) set(key string, limiter *limiter) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	ls.store[key] = limiter
}

func (ls *limiterStore) get(key string) *limiter {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	return ls.store[key]
}

func (ls *limiterStore) reset(key string) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	if l := ls.store[key]; l != nil {
		l.reset()
	}
}

func (ls *limiterStore) increment(key string) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	ls.store[key].increment()
}
