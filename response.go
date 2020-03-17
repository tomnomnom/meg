package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

// jsonString returns a JSON string representation of the request and response
func (r response) jsonString(noHeaders bool) string {
	// Building the structs for the various JSON elements (root element, headers, body, ...)

	// Struct for Headers (Key, Value)
	type JsonHeader struct {
		JsonHeaderKey   string `json:"headerkey"`
		JsonHeaderValue string `json:"headervalue"`
	}
	// Struct representing the request (URL, Hostname, HTTP Method, Path, Host, Headers (struct), Body, Follow (-L))
	type JsonRequest struct {
		JsonRequestUrl      string       `json:"url"`
		JsonRequestHostname string       `json:"hostname"`
		JsonRequestMethod   string       `json:"method"`
		JsonRequestPath     string       `json:"path"`
		JsonRequestHost     string       `json:"host"`
		JsonRequestHeaders  []JsonHeader `json:"headers"`
		JsonRequestBody     string       `json:"body"`
		JsonRequestFollow   bool         `json:"follow"`
	}
	// Struct representing the response (HTTP status code, status text, headers(struct), body)
	type JsonResponse struct {
		JsonResponseStatusCode int          `json:"statuscode"`
		JsonResponseStatus     string       `json:"status"`
		JsonResponseHeader     []JsonHeader `json:"headers"`
		JsonBody               string       `json:"body"`
	}
	// Struct representing the root element (Request, Response)
	type JsonOut struct {
		JsonRequest  JsonRequest  `json:"request"`
		JsonResponse JsonResponse `json:"response"`
	}

	// Prepare the Header structs which will be inserted into the JSON element
	var jsonRequestHeaders []JsonHeader
	var jsonResponseHeaders []JsonHeader
	// If we specify --no-headers we will simply receive "headers":null in both the request and response
	// Otherwise we will extract the headers from the request/response and fill the header elements
	if !noHeaders {
		// Fill the request header struct
		for _, h := range r.request.headers {
			x := strings.Split(h, ":")
			jsonRequestHeaders = append(jsonRequestHeaders, JsonHeader{x[0], strings.TrimSpace(x[1])})
		}
		// Fill the response header struct
		for _, h := range r.headers {
			x := strings.Split(h, ":")
			jsonResponseHeaders = append(jsonResponseHeaders, JsonHeader{x[0], strings.TrimSpace(x[1])})
		}
	}
	// Create the JSON Body element
	jsonBody := &bytes.Buffer{}
	jsonBody.Write(r.body)

	// Create the JSON element
	// Root Element
	resp := JsonOut{
		// Request Element
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
		// Response Element
		JsonResponse{
			r.statusCode,
			r.status,
			jsonResponseHeaders,
			jsonBody.String(),
		},
	}
	// Marshal the JSON struct
	ba, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("error:", err)
	}

	return string(ba)
}

// megString returns a string representation of the request and response in a "traditional" meg format
func (r response) megString(noHeaders bool) string {
	if noHeaders {
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

	// Depending if we want to save a json or meg string we call the corresponding String method
	if json {
		content = []byte(r.jsonString(noHeaders))
	} else {
		content = []byte(r.megString(noHeaders))
	}

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
