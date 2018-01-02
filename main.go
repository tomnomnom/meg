package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const userAgent = "Mozilla/5.0 (compatible; meg/0.1; +https://github.com/tomnomnom/meg)"

// a requester is a function that makes HTTP requests
type requester func(request) response

func main() {

	// get the config struct
	c := processArgs()

	// if the suffix argument is a file, read it; otherwise
	// treat it as a literal value
	var suffixes []string
	if isFile(c.suffix) {
		lines, err := readLines(c.suffix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open suffix file: %s\n", err)
			os.Exit(1)
		}
		suffixes = lines
	} else if c.suffix != "suffixes" {
		suffixes = []string{c.suffix}
	}

	// read the prefix file
	prefixes, err := readLines(c.prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open prefix file: %s\n", err)
		os.Exit(1)
	}

	// make the output directory
	err = os.MkdirAll(c.output, 0750)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory: %s\n", err)
		os.Exit(1)
	}

	// open the index file
	indexFile := filepath.Join(c.output, "index")
	index, err := os.OpenFile(indexFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open index file for writing: %s\n", err)
		os.Exit(1)
	}

	// set up a rate limiter
	rl := newRateLimiter(time.Duration(c.delay * 1000000))

	// the request and response channels for
	// the worker pool
	requests := make(chan request)
	responses := make(chan response)

	// spin up some workers to do the requests
	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)

		go func() {
			for req := range requests {
				rl.Block(req.Hostname())
				responses <- c.requester(req)
			}
			wg.Done()
		}()
	}

	// start outputting the response lines; we need a second
	// WaitGroup so we know the outputting has finished
	var owg sync.WaitGroup
	owg.Add(1)
	go func() {
		for res := range responses {
			if c.saveStatus != 0 && c.saveStatus != res.statusCode {
				continue
			}

			if res.err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %s\n", res.err)
				continue
			}

			path, err := res.save(c.output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to save file: %s\n", err)
			}

			line := fmt.Sprintf("%s %s (%s)\n", path, res.request.URL(), res.status)
			fmt.Fprintf(index, "%s", line)
			if c.verbose {
				fmt.Printf("%s", line)
			}
		}
		owg.Done()
	}()

	// send requests for each suffix for every prefix
	for _, suffix := range suffixes {
		for _, prefix := range prefixes {

			requests <- request{
				method:  c.method,
				prefix:  prefix,
				suffix:  suffix,
				headers: c.headers,
			}
		}
	}

	// once all of the requests have been sent we can
	// close the requests channel
	close(requests)

	// wait for all the workers to finish before closing
	// the responses channel
	wg.Wait()
	close(responses)

	owg.Wait()

}

// readLines reads all of the lines from a text file in to
// a slice of strings, returning the slice and any error
func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	lines := make([]string, 0)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	return lines, sc.Err()
}

// isFile returns true if its argument is a regular file
func isFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().IsRegular()
}
