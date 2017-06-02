package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

func recordJob(j job, pathPrefix string) (string, error) {

	checksum := sha1.Sum([]byte(j.prefix + j.suffix))
	parts := []string{pathPrefix}

	// we need the host as part of the path. The suffix might
	// fail to parse, but the prefix shouldn't so we should
	// be ok to call url.Parse here
	u, err := url.Parse(j.prefix)
	if err != nil {
		return "", err
	}

	parts = append(parts, u.Host)
	parts = append(parts, fmt.Sprintf("%x", checksum))

	p := path.Join(parts...)

	if _, err := os.Stat(path.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(p), 0750)
		if err != nil {
			return p, err
		}
	}

	err = ioutil.WriteFile(p, []byte(j.String()), 0640)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (j job) String() string {
	buf := &bytes.Buffer{}

	buf.WriteString(j.prefix + j.suffix)
	buf.WriteString("\n\n")
	buf.WriteString(j.resp.status)
	buf.WriteString("\n")

	for name, values := range j.resp.headers {
		buf.WriteString(
			fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")),
		)
	}

	buf.WriteString("\n\n")
	buf.Write(j.resp.body)
	buf.WriteString("\n")

	return buf.String()
}
