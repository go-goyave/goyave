package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLimiter(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	assert.Equal(t, 5, l.config.RequestQuota)
	assert.Equal(t, 0, l.counter)
	assert.NotEqual(t, int64(0), l.resetsAt.UnixNano())
}

func TestLimiterIncrement(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.increment()
	assert.Equal(t, 1, l.counter)
}

func TestLimiterHasExpired(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.resetsAt = time.Now().Add(2 * time.Second)
	assert.False(t, l.hasExpired())

	l.resetsAt = time.Now().Add(-2 * time.Second)
	assert.True(t, l.hasExpired())
}

func TestLimiterHasExceededRequestQuota(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.counter = 4
	assert.False(t, l.hasExceededRequestQuota())

	l.counter = 5
	assert.True(t, l.hasExceededRequestQuota())

	l.counter = 6
	assert.True(t, l.hasExceededRequestQuota())
}

func TestLimiterGetRemainingRequestQuota(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.counter = 3
	assert.Equal(t, 2, l.getRemainingRequestQuota())
}

func TestLimiterReset(t *testing.T) {
	l := newLimiter(Config{
		RequestQuota:  5,
		QuotaDuration: time.Second,
	})
	l.counter = 4
	resetAt := l.resetsAt

	l.reset()
	assert.Equal(t, 0, l.counter)
	assert.NotEqual(t, resetAt, l.counter)
}

func TestLimiterStore(t *testing.T) {
	store := newLimiterStore()
	assert.NotNil(t, store.store)

	l := newLimiter(Config{})
	store.set("key", l)
	limiter, ok := store.store["key"]
	assert.True(t, ok)
	assert.Same(t, l, limiter)
	assert.Same(t, l, store.get("key"))
}
