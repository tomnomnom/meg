package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"
)

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var checkRedirect = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

var httpClient = &http.Client{
	Transport:     transport,
	CheckRedirect: checkRedirect,
	Timeout:       time.Second * 10,
}

func doRequest(r request) response {

	req, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		return response{request: r, err: err}
	}
	req.Close = true

	// It feels super nasty doing this, but some sites act differently
	// when they don't recognise the user agent. E.g. some will just
	// 302 any request to a 'browser not found' page, which makes the
	// tool kind of useless. It's not about being 'stealthy', it's
	// about making things work as expected.
	req.Header.Set("User-Agent", "User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	resp, err := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return response{request: r, err: err}
	}
	body, _ := ioutil.ReadAll(resp.Body)

	return response{
		request: r,
		status:  resp.Status,
		body:    body,
	}
}
