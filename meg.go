package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type result struct {
	url     string
	code    int
	headers http.Header
}

type results []result

func (r result) printHeaders() {
	if r.headers == nil {
		return
	}
	for name, values := range r.headers {
		fmt.Printf("%s: %s\n", name, strings.Join(values, ", "))
	}
}

func req(url string, rc chan result) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client := &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       time.Second * 10,
	}
	resp, err := client.Get(url)
	if err != nil {
		rc <- result{url, -1, nil}
		return
	}
	rc <- result{url, resp.StatusCode, resp.Header}
}

func reqMulti(urls []string, rc chan result) {
	for _, url := range urls {
		go req(url, rc)
	}
}

func main() {

	flag.Parse()

	prefixes := flag.Arg(0)
	path := flag.Arg(1)

	f, err := os.Open(prefixes)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	urls := make([]string, 0)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		urls = append(urls, sc.Text()+path)
	}
	if err = sc.Err(); err != nil {
		log.Fatal(err)
	}

	rc := make(chan result)
	reqMulti(urls, rc)

	for i := 0; i < len(urls); i++ {
		r := <-rc
		fmt.Printf("%d %s\n", r.code, r.url)
		//r.printHeaders()
		//fmt.Println("")
	}

}
