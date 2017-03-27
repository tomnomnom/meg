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

func httpRequest(method, url string) (response, error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return response{}, err
	}
	req.Close = true

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
