package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tomnomnom/rawhttp"
)

func rawRequest(r request) response {
	req, err := rawhttp.FromURL(r.method, r.host)
	if err != nil {
		return response{request: r, err: err}
	}

	req.Timeout = r.timeout

	req.Path = r.path

	r.headers = append(r.headers, "Connection: close")

	if !r.HasHeader("Host") {
		// add the host header to the request manually so it shows up in the output
		r.headers = append(r.headers, fmt.Sprintf("Host: %s", r.Hostname()))
	}

	if !r.HasHeader("User-Agent") {
		r.headers = append(r.headers, fmt.Sprintf("User-Agent: %s", userAgent))
	}

	for _, h := range r.headers {
		req.AddHeader(h)
	}

	if r.body != "" {
		req.Body = r.body
	}

	if !r.HasHeader("Content-Length") {
		req.AutoSetContentLength()
	}

	resp, err := rawhttp.Do(req)
	if err != nil {
		return response{request: r, err: err}
	}

	// Silly me. I should have done this in rawhttp
	code, err := strconv.Atoi(resp.StatusCode())
	if err != nil {
		return response{request: r, err: err}
	}

	// This should be done in rawhttp too. Whoops.
	status := resp.StatusLine()
	p := strings.SplitN(resp.StatusLine(), " ", 2)
	if len(p) == 2 {
		status = p[1]
	}

	return response{
		request:    r,
		status:     status,
		statusCode: code,
		headers:    resp.Headers(),
		body:       resp.Body(),
		duration:   resp.Duration,
	}

}
