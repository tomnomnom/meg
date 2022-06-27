package main

import (
	"sync"
	"time"

	"github.com/adamdrake/tokenbucket"
)

// a rateLimiter allows you to delay operations
// on a per-key basis. I.e. only one operation for
// a given key can be done within the delay time
type rateLimiter struct {
	sync.Mutex
	delay time.Duration
	ops   map[string]*tokenbucket.TokenBucket
}

// newRateLimiter returns a new *rateLimiter for the
// provided delay
func newRateLimiter(delay time.Duration) *rateLimiter {
	return &rateLimiter{
		delay: delay,
		ops:   make(map[string]*tokenbucket.TokenBucket),
	}
}

// Block blocks until an operation for key is
// allowed to proceed
func (r *rateLimiter) Block(key string) {
	r.Lock()
	defer r.Unlock()

	// if there's nothing in the map we can
	// initialize the new rate limiter and
	// return straight away
	if _, ok := r.ops[key]; !ok {
		r.ops[key] = tokenbucket.New(1, 1, r.delay)
		return
	}

	// if time is up we can return straight away
	if r.ops[key].Take() {
		return
	}

	// otherwise, be sure we wait for a new token to be available,
	// take it, and then return
	time.Sleep(r.delay)
	r.ops[key].Take()
	return
}
