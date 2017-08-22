package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// job is a wrapper around an HTTP request,
// the response for that request and any
// associated error.
type job struct {
	req  *http.Request
	resp *http.Response

	err error
}

func worker(jobs <-chan job, results chan<- job) {
	for j := range jobs {
		results <- httpRequest(j)
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

	suffixes, err := readLines(suffixPath)
	if err != nil {
		fmt.Println(err)
		return
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
				req, err := http.NewRequest(method, prefix+suffix, nil)
				if err != nil {
					continue
				}
				req.Close = true

				// Because we sometimes want to send some fairly dodgy paths,
				// like /%%0a0afoo for example, we need to set the path on
				// req.URL's Opaque field where it won't be parsed or encoded
				//req.URL.Opaque = suffix

				// It feels super nasty doing this, but some sites act differently
				// when they don't recognise the user agent. E.g. some will just
				// 302 any request to a 'browser not found' page, which makes the
				// tool kind of useless. It's not about being 'stealthy', it's
				// about making things work as expected.
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

				// Set any user provided headers
				for _, header := range headers {
					p := strings.SplitN(header, ":", 2)
					req.Header.Set(strings.TrimSpace(p[0]), strings.TrimSpace(p[1]))
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
		fmt.Printf("%s %s (%s)\n", filename, r.req.URL.String(), r.resp.Status)
	}

}
