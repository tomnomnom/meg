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

	b.WriteString(r.request.url)
	b.WriteString("\n\n")

	// request headers

	// response headers

	// body
	b.Write(r.body)

	return b.String()
}

func (r response) save(pathPrefix string) (string, error) {

	checksum := sha1.Sum([]byte(r.request.url))
	parts := []string{pathPrefix}

	//parts = append(parts, j.req.Hostname)
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
