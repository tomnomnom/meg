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

func httpRequest(method, prefix, suffix string) (response, error) {

	req, err := http.NewRequest(method, prefix, nil)
	if err != nil {
		return response{}, err
	}
	req.Close = true

	// Because we sometimes want to send some fairly dodgy paths,
	// like /%%0a0afoo for example, we need to set the path on
	// req.URL's Opaque field where it won't be parsed or encoded
	req.URL.Opaque = suffix

	resp, err := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return response{}, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return response{resp.Status, resp.Header, body}, nil
}
