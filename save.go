package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func recordJob(j job, pathPrefix string) (string, error) {

	// TODO: Improve the entropy of the save path to include method, headers etc
	// ...maybe add the current time too just for the hell of it
	checksum := sha1.Sum([]byte(j.req.URL.String()))
	parts := []string{pathPrefix}

	parts = append(parts, j.req.URL.Host)
	parts = append(parts, fmt.Sprintf("%x", checksum))

	p := path.Join(parts...)

	if _, err := os.Stat(path.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(p), 0750)
		if err != nil {
			return p, err
		}
	}

	err := ioutil.WriteFile(p, []byte(j.String()), 0640)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (j job) String() string {
	buf := &bytes.Buffer{}

	// Request URL
	buf.WriteString(j.req.URL.String())
	buf.WriteString("\n\n")

	// Request Headers
	for name, values := range j.req.Header {
		buf.WriteString(
			fmt.Sprintf("> %s: %s\n", name, strings.Join(values, ", ")),
		)
	}
	buf.WriteString("\n")

	if j.resp != nil {
		defer j.resp.Body.Close()

		// Response Status
		buf.WriteString("< ")
		buf.WriteString(j.resp.Status)
		buf.WriteString("\n")

		// Response Headers
		for name, values := range j.resp.Header {
			buf.WriteString(
				fmt.Sprintf("< %s: %s\n", name, strings.Join(values, ", ")),
			)
		}

		// Response Body
		body, _ := ioutil.ReadAll(j.resp.Body)
		buf.WriteString("\n\n")
		buf.Write(body)
		buf.WriteString("\n")
	}

	return buf.String()
}
