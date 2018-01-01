package main

import "net/url"

// a request is a wrapper for a URL that we want to request
type request struct {
	method  string
	prefix  string
	suffix  string
	headers []string
}

// Hostname returns the hostname part of the request
func (r request) Hostname() string {
	u, err := url.Parse(r.prefix)

	// the hostname part is used only for the rate
	// limiting and the
	if err != nil {
		return "unknown"
	}
	return u.Hostname()
}

// URL returns the full URL to request
func (r request) URL() string {
	return r.prefix + r.suffix
}
