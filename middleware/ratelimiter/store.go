package ratelimiter

import (
	"math"
	"sync"
	"time"
)

type limiter struct {
	requestQuota  int
	counter       int
	quotaDuration time.Duration
	resetsAt      time.Time
	mx            sync.RWMutex
}

func newLimiter(requestQuota int, quotaDuration time.Duration) *limiter {
	return &limiter{
		requestQuota:  requestQuota,
		quotaDuration: quotaDuration,
		counter:       0,
		resetsAt:      time.Now().Add(quotaDuration),
	}
}

func (l *limiter) reset() {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.resetsAt = time.Now().Add(l.quotaDuration)
	l.counter = 0
}

func (l *limiter) increment() {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.counter++
}

func (l *limiter) hasExpired() bool {
	l.mx.RLock()
	defer l.mx.RUnlock()
	return time.Now().After(l.resetsAt)
}

func (l *limiter) hasExceededRequestQuota() bool {
	l.mx.RLock()
	defer l.mx.RUnlock()
	return l.counter >= l.requestQuota
}

func (l *limiter) getRemainingRequestQuota() int {
	l.mx.RLock()
	defer l.mx.RUnlock()
	return l.requestQuota - l.counter
}

func (l *limiter) getSecondsToQuotaReset() float64 {
	l.mx.RLock()
	defer l.mx.RUnlock()
	return -1 * math.Round(time.Since(l.resetsAt).Seconds())
}

type limiterStore struct {
	mx    sync.RWMutex
	store map[interface{}]*limiter
}

func newLimiterStore() limiterStore {
	return limiterStore{
		store: make(map[interface{}]*limiter),
	}
}

func (ls *limiterStore) set(key interface{}, limiter *limiter) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	ls.store[key] = limiter
}

func (ls *limiterStore) get(key interface{}) *limiter {
	ls.mx.RLock()
	defer ls.mx.RUnlock()
	return ls.store[key]
}
