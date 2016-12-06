package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type result struct {
	url     string
	status  string
	headers http.Header
	body    []byte
}

func (r result) String() string {
	buf := &bytes.Buffer{}

	buf.WriteString(r.url)
	buf.WriteString("\n\n")
	buf.WriteString(r.status)
	buf.WriteString("\n")

	for name, values := range r.headers {
		buf.WriteString(
			fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")),
		)
	}

	buf.WriteString("\n\n")
	buf.Write(r.body)
	buf.WriteString("\n")

	return buf.String()
}

func (r result) printHeaders() {
	if r.headers == nil {
		return
	}
	for name, values := range r.headers {
		fmt.Printf("%s: %s\n", name, strings.Join(values, ", "))
	}
}

type results []result

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
		rc <- result{url, "", nil, nil}
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	rc <- result{url, resp.Status, resp.Header, body}
}

func reqMulti(urls []string, rc chan result) {
	for _, url := range urls {
		go req(url, rc)
	}
}

func writeFile(r result) {

	u, err := url.Parse(r.url)
	if err != nil {
		log.Printf("failed to parse url [%s]", r.url)
		return
	}

	parts := []string{"./out"}
	parts = append(parts, u.Scheme)
	parts = append(parts, u.Host)
	parts = append(parts, u.Path)

	p := path.Join(parts...)

	if _, err := os.Stat(path.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(p), 0750)
		if err != nil {
			log.Printf("failed to create dir for [%s]", p)
			return
		}
	}

	err = ioutil.WriteFile(p, []byte(r.String()), 0640)
	if err != nil {
		log.Printf("failed to write [%s]", p)
		return
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
		urls = append(urls, strings.ToLower(sc.Text())+path)
	}
	if err = sc.Err(); err != nil {
		log.Fatal(err)
	}

	rc := make(chan result)
	reqMulti(urls, rc)

	for i := 0; i < len(urls); i++ {
		r := <-rc
		fmt.Printf("%s %s\n", r.status, r.url)

		writeFile(r)
	}

}
