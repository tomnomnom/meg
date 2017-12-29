package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func (r response) String() string {
	b := &bytes.Buffer{}

	b.WriteString(r.request.url.String())
	b.WriteString("\n\n")

	qs := ""
	if len(r.request.url.Query()) > 0 {
		qs = "?" + r.request.url.Query().Encode()
	}

	b.WriteString(fmt.Sprintf("> %s %s%s HTTP/1.1\n", r.request.method, r.request.url.EscapedPath(), qs))

	// request headers
	for _, h := range r.request.headers {
		b.WriteString(fmt.Sprintf("> %s\n", h))
	}
	b.WriteString("\n")

	// status line
	b.WriteString(fmt.Sprintf("< HTTP/1.1 %s\n", r.status))

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

	checksum := sha1.Sum([]byte(r.request.url.String()))
	parts := []string{pathPrefix}

	parts = append(parts, r.request.url.Hostname())
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
