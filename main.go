package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type job struct {
	// request data
	prefix  string
	suffix  string
	method  string
	headers reqHeaders

	resp response
	err  error
}

type response struct {
	status  string
	headers http.Header
	body    []byte
}

func worker(jobs <-chan job, results chan<- job) {
	for j := range jobs {
		r, err := httpRequest(j)
		j.resp = r
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
	sleep := 30
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
	flag.IntVar(&sleep, "sleep", 30, "sleep duration between each suffix")
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
				jobs <- job{prefix: prefix, suffix: suffix, method: method, headers: headers}
			}
			time.Sleep(time.Second * time.Duration(sleep))
		}
		close(jobs)
	}()

	// wait for results
	for r := range results {
		fn, err := recordJob(r, savePath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s %s%s (%s)\n", fn, r.prefix, r.suffix, r.resp.status)
	}

}
