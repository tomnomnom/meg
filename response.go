package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

func (r response) String() string {
	b := &bytes.Buffer{}

	b.WriteString(fmt.Sprintf("> %s %s\n", r.request.method, r.request.url))

	// request headers
	for _, h := range r.request.headers {
		b.WriteString(fmt.Sprintf("> %s\n", h))
	}
	b.WriteString("\n")

	// response headers
	for _, h := range r.headers {
		b.WriteString(fmt.Sprintf("< %s\n", h))
	}
	b.WriteString("\n")

	// body
	b.Write(r.body)

	return b.String()
}

func (r response) save(pathPrefix string) (string, error) {

	checksum := sha1.Sum([]byte(r.request.url))
	parts := []string{pathPrefix}

	host := "unknownhost"
	if u, err := url.Parse(r.request.url); err == nil {
		host = u.Hostname()
	}

	parts = append(parts, host)
	parts = append(parts, fmt.Sprintf("%x", checksum))

	p := path.Join(parts...)

	if _, err := os.Stat(path.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(p), 0750)
		if err != nil {
			return p, err
		}
	}

	err := ioutil.WriteFile(p, []byte(r.String()), 0640)
	if err != nil {
		return p, err
	}

	return p, nil
}
