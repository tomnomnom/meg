package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"
)

func httpRequest(method, url string) (response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// We don't want to follow redirects
	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client := &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       time.Second * 10,
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return response{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return response{}, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return response{resp.Status, resp.Header, body}, nil
}
