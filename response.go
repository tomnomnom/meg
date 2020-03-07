package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"encoding/json"
	"strings"
)

// a response is a wrapper around an HTTP response;
// it contains the request value for context.
type response struct {
	request request

	status     string
	statusCode int
	headers    []string
	body       []byte
	err        error
}

// String returns a string representation of the request and response
func (r response) jsonString(noHeaders bool) string {
	if noHeaders{
		// Falsch TODO
		b := &bytes.Buffer{}
		b.Write(r.body)
		return b.String()
	}

	b := &bytes.Buffer{}

	b.WriteString(r.request.URL())
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%s %s HTTP/1.1\n", r.request.method, r.request.path))

	// request headers
	for _, h := range r.request.headers {
		b.WriteString(fmt.Sprintf("%s\n", h))
	}
	b.WriteString("\n")

	// status line
	b.WriteString(fmt.Sprintf("HTTP/1.1 %s\n", r.status))

	// response headers
	for _, h := range r.headers {
		b.WriteString(fmt.Sprintf("%s\n", h))
	}
	b.WriteString("\n")

	// body
	/*
	r.headers
	r.status
	r.err
	*/
	type JsonHeader struct{
		JsonHeaderKey string `json:"headerkey"`
		JsonHeaderValue string `json:"headervalue"`
	}
	type JsonRequest struct {
		JsonRequestUrl string `json:"url"`
		JsonRequestHostname string `json:"hostname"`
		JsonRequestMethod string `json:"method"`
		JsonRequestPath    string `json:"path"`
		JsonRequestHost    string `json:"host"`
		JsonRequestHeaders    []JsonHeader `json:"headers"`
		JsonRequestBody    string `json:"body"`
		JsonRequestFollow    bool `json:"follow"`
	}
	r.request.Hostname()
	type JsonResponse struct {
		JsonResponseStatusCode    int `json:"statuscode"`
		JsonResponseStatus    string `json:"status"`
		JsonResponseHeader    []JsonHeader `json:"headers"`
		JsonBody    string `json:"body"`
	}
	type JsonOut struct {
		JsonRequest JsonRequest `json:"request"`
		JsonResponse JsonResponse `json:"response"`
	}
	var jsonRequestHeaders []JsonHeader
	for _, h := range r.request.headers {
		x := strings.Split(h, ":")
		jsonRequestHeaders = append(jsonRequestHeaders, JsonHeader{x[0], strings.TrimSpace(x[1])})
	}
	var jsonResponseHeaders []JsonHeader
	for _, h := range r.headers {
		x := strings.Split(h, ":")
		jsonResponseHeaders = append(jsonResponseHeaders, JsonHeader{x[0], strings.TrimSpace(x[1])})
	}
	//r.h
	//	var jsonRequestHeader []JsonHeader
	//	for _, h := range r.request.headers {
	//		x := strings.Split(h, ":")
	//		jsonRequestHeader = append(jsonRequestHeader, JsonHeader{x[0], strings.TrimSpace(x[1])})
	//	}
	//	//r.headers
	x := &bytes.Buffer{}
	x.Write(r.body)
	x.String()
	resp := JsonOut{
		JsonRequest{
			r.request.URL(),
			r.request.Hostname(),
			r.request.method,
			r.request.path,
			r.request.host,
			jsonRequestHeaders,
			r.request.body,
			r.request.followLocation,
			},
			JsonResponse{
				r.statusCode,
				r.status,
				jsonResponseHeaders,
				x.String(),
			},
	}
	ba, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%s", ba)
	return b.String()
}

// String returns a string representation of the request and response
func (r response) megString(noHeaders bool) string {
	if noHeaders{
		b := &bytes.Buffer{}
		b.Write(r.body)
		return b.String()
	}

	b := &bytes.Buffer{}

	b.WriteString(r.request.URL())
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("> %s %s HTTP/1.1\n", r.request.method, r.request.path))

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


// save write a request and response output to disk
func (r response) save(pathPrefix string, noHeaders bool, json bool) (string, error) {
	var content []byte

	if json{
		content = []byte(r.jsonString(noHeaders))
	}else{
		content = []byte(r.megString(noHeaders))
	}
/*
	if noHeaders {
		content = []byte(r.StringNoHeaders())
	}
*/
	checksum := sha1.Sum(content)
	parts := []string{pathPrefix}

	parts = append(parts, r.request.Hostname())
	parts = append(parts, fmt.Sprintf("%x", checksum))

	p := path.Join(parts...)

	if _, err := os.Stat(path.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(p), 0750)
		if err != nil {
			return p, err
		}
	}

	err := ioutil.WriteFile(p, content, 0640)
	if err != nil {
		return p, err
	}

	return p, nil
}
