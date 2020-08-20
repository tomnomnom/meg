package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// a request is a wrapper for a URL that we want to request
type request struct {
	method  string
	path    string
	host    string
	headers []string
	body    string

	followLocation bool
	timeout        time.Duration
}

// Hostname returns the hostname part of the request
func (r request) Hostname() string {
	u, err := url.Parse(r.host)

	// the hostname part is used only for the rate
	// limiting and the
	if err != nil {
		return "unknown"
	}
	return u.Hostname()
}

// URL returns the full URL to request
func (r request) URL() string {
	return r.host + r.path
}

// hasHeader returns true if the request
// has the provided header
func (r request) HasHeader(h string) bool {
	norm := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}
	for _, candidate := range r.headers {

		p := strings.SplitN(candidate, ":", 2)
		if norm(p[0]) == norm(h) {
			return true
		}
	}
	return false
}

func (r request) GetHeader(h string) string {
	var header string
	norm := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}
	for _, candidate := range r.headers {

		p := strings.SplitN(candidate, ":", 2)
		if norm(p[0]) == norm(h) {
			header := fmt.Sprintf("%s", norm(p[1]))
			return header
		}
	}
	return header
}
