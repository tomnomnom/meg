package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tomnomnom/rawhttp"
)

// job is a wrapper around an HTTP request,
// the response for that request and any
// associated error.
type job struct {
	req  *rawhttp.Request
	resp *rawhttp.Response

	err error
}

func worker(jobs <-chan job, results chan<- job) {
	for j := range jobs {
		resp, err := rawhttp.Do(j.req)

		j.resp = resp
		j.err = err

		results <- j
	}
}

type reqHeaders []string

func (h *reqHeaders) Set(val string) error {
	*h = append(*h, val)
	return nil
}

func (h reqHeaders) String() string {
	return "string"
}

func main() {

	concurrency := 20
	method := "GET"
	sleep := 0
	savePath := "./out"
	prefixPath := "prefixes"
	suffixPath := "suffixes"

	var headers reqHeaders

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: meg [flags]\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&method, "method", "GET", "HTTP method to use")
	flag.StringVar(&savePath, "savepath", "./out", "where to save the output")
	flag.StringVar(&prefixPath, "prefixes", "prefixes", "file containing prefixes")
	flag.StringVar(&suffixPath, "suffixes", "suffixes", "file containing suffixes")
	flag.IntVar(&sleep, "sleep", 0, "sleep duration between each suffix")
	flag.IntVar(&concurrency, "concurrency", 20, "concurrency")
	flag.Var(&headers, "header", "header to add to the request")

	flag.Parse()

	prefixes, err := readLines(prefixPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	path := flag.Arg(0)
	suffixes := []string{path}

	if path == "" {
		suffixes, err = readLines(suffixPath)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	jobs := make(chan job)
	results := make(chan job)

	// spin up the workers
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {

		wg.Add(1)
		go func() {
			worker(jobs, results)
			wg.Done()
		}()
	}

	// close the results channel when all of the
	// workers have finished
	go func() {
		wg.Wait()
		close(results)
	}()

	// feed in the jobs
	go func() {
		for _, suffix := range suffixes {
			for _, prefix := range prefixes {

				// Make a new request
				req, err := rawhttp.FromURL(method, prefix)
				if err != nil {
					continue
				}
				req.Path = suffix

				req.AddHeader("Connection: close")

				// Set any user provided headers
				for _, header := range headers {
					req.AddHeader(header)
				}
				if req.Header("Host") == "" {
					req.AutoSetHost()
				}

				// Feed the job into the requests channel so it can be picked up
				// by a worker
				jobs <- job{req: req}
			}

			// Sleep for a bit before moving onto the next prefix...
			// This responsibility needs to be moved into some kind
			// of controller for the worker pool... Or something.
			time.Sleep(time.Second * time.Duration(sleep))
		}
		close(jobs)
	}()

	// wait for results
	for r := range results {
		filename, err := recordJob(r, savePath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		status := "[error]"
		if r.resp != nil {
			status = r.resp.StatusLine()
		}
		fmt.Printf("%s %s (%s)\n", filename, r.req.URL(), status)
	}

}
