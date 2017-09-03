package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func recordJob(j job, pathPrefix string) (string, error) {

	// TODO: Improve the entropy of the save path to include method, headers etc
	// ...maybe add the current time too just for the hell of it
	checksum := sha1.Sum([]byte(j.req.URL()))
	parts := []string{pathPrefix}

	parts = append(parts, j.req.Hostname)
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
	buf.WriteString(j.req.URL())
	buf.WriteString("\n\n")

	buf.WriteString(
		fmt.Sprintf("> %s\n", j.req.RequestLine()),
	)

	// Request Headers
	for _, h := range j.req.Headers {
		buf.WriteString(
			fmt.Sprintf("> %s\n", h),
		)
	}
	buf.WriteString("\n")

	if j.resp != nil {

		// Response Status
		buf.WriteString("< ")
		buf.WriteString(j.resp.StatusLine())
		buf.WriteString("\n")

		// Response Headers
		for _, h := range j.resp.Headers() {
			buf.WriteString(
				fmt.Sprintf("< %s\n", h),
			)
		}

		// Response Body
		buf.WriteString("\n\n")
		buf.Write(j.resp.Body())
		buf.WriteString("\n")
	}

	return buf.String()
}
