package ratelimiter

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/System-Glitch/goyave/v3"
)

type limiter struct {
	requestQuota  int
	counter       int
	quotaDuration time.Duration
	resetsAt      time.Time
	mx            sync.Mutex
}

func newLimiter(requestQuota int, quotaDuration time.Duration) *limiter {
	return &limiter{
		requestQuota:  requestQuota,
		quotaDuration: quotaDuration,
		counter:       0,
		resetsAt:      time.Now().Add(quotaDuration),
	}
}

func (l *limiter) validateAndUpdate(response *goyave.Response) bool {

	l.mx.Lock()
	defer l.mx.Unlock()

	if l.hasExpired() {
		l.reset()
	} else if l.hasExceededRequestQuota() {
		l.updateResponseHeaders(response)
		return false
	}

	l.increment()
	l.updateResponseHeaders(response)
	return true
}

func (l *limiter) updateResponseHeaders(response *goyave.Response) {
	response.Header().Set(
		"RateLimit-Limit",
		fmt.Sprintf("%v, %v;w=%v", l.requestQuota, l.requestQuota, l.quotaDuration.Seconds()),
	)

	response.Header().Set(
		"RateLimit-Remaining",
		fmt.Sprintf("%v", l.getRemainingRequestQuota()),
	)

	response.Header().Set(
		"RateLimit-Reset",
		fmt.Sprintf("%v", l.getSecondsToQuotaReset()),
	)
}

func (l *limiter) reset() {
	l.resetsAt = time.Now().Add(l.quotaDuration)
	l.counter = 0
}

func (l *limiter) increment() {
	l.counter++
}

func (l *limiter) hasExpired() bool {
	return time.Now().After(l.resetsAt)
}

func (l *limiter) hasExceededRequestQuota() bool {
	return l.counter >= l.requestQuota
}

func (l *limiter) getRemainingRequestQuota() int {
	return l.requestQuota - l.counter
}

func (l *limiter) getSecondsToQuotaReset() float64 {
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
