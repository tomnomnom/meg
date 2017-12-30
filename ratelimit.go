package main

import (
	"net/url"
	"sync"
	"time"
)

type rateLimiter struct {
	sync.Mutex
	delay time.Duration
	reqs  map[string]time.Time
}

func (r *rateLimiter) Block(u *url.URL) {
	now := time.Now()
	key := u.Hostname()

	r.Lock()

	// if there's nothing in the map we can
	// return straight away
	if _, ok := r.reqs[u.Hostname()]; !ok {
		r.reqs[key] = now
		r.Unlock()
		return
	}

	// if time is up we can return straight away
	t := r.reqs[key]
	deadline := t.Add(r.delay)
	if now.After(deadline) {
		r.reqs[key] = now
		r.Unlock()
		return
	}

	remaining := deadline.Sub(now)

	// Set the time of the request
	r.reqs[key] = now.Add(remaining)
	r.Unlock()

	// Block for the remaining time
	<-time.After(remaining)
}
