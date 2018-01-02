package main

import (
	"fmt"
	"strconv"

	"github.com/tomnomnom/rawhttp"
)

func rawRequest(r request) response {
	req, err := rawhttp.FromURL(r.method, r.prefix)
	if err != nil {
		return response{request: r, err: err}
	}

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

	resp, err := rawhttp.Do(req)
	if err != nil {
		return response{request: r, err: err}
	}

	// Silly me. I should have done this in rawhttp
	code, err := strconv.Atoi(resp.StatusCode())
	if err != nil {
		return response{request: r, err: err}
	}

	return response{
		request:    r,
		status:     resp.StatusLine(),
		statusCode: code,
		headers:    resp.Headers(),
		body:       resp.Body(),
	}

}
