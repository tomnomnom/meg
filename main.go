package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
	"strings"
)

const (
	userAgent = "Mozilla/5.0 (compatible; meg/0.2; +https://github.com/tomnomnom/meg)"

	// argument defaults
	defaultPathsFile = "./paths"
	defaultHostsFile = "./hosts"
	defaultOutputDir = "./out"
)

// a requester is a function that makes HTTP requests
type requester func(request) response

func main() {

	// get the config struct
	c := processArgs()

	// read the paths file
	paths, err := readLinesOrLiteral(c.paths, defaultPathsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open paths file: %s\n", err)
		os.Exit(1)
	}

	// read the hosts file
	hosts, err := readLinesOrLiteral(c.hosts, defaultHostsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open hosts file: %s\n", err)
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
			if len(c.saveStatus) > 0 && !c.saveStatus.Includes(res.statusCode) {
				continue
			}

			if c.saveHeader != "" && !res.hasHeader(c.saveHeader) {
				continue
			}

			if res.err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %s\n", res.err)
				continue
			}

			path, err := res.save(c.output, c.noHeaders)
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

	// send requests for each path for every host
	for _, path := range paths {
		for _, host := range hosts {

			// the host portion may contain a path prefix,
			// so we should strip that off and add it to
			// the beginning of the path.
			u, err := url.Parse(host)

			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to parse host: %s\n", err)
				continue
			}

			if len(c.trackerLink) > 0 {
				path = replaceTracker(path, u.Host, c.trackerLink)
			}

			prefixedPath := u.Path + path
			u.Path = ""

			// stripping off a path means we need to
			// rebuild the host portion too
			host = u.String()

			requests <- request{
				method:         c.method,
				host:           host,
				path:           prefixedPath,
				headers:        c.headers,
				followLocation: c.followLocation,
				body:           c.body,
				timeout:        time.Duration(c.timeout * 1000000),
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

func replaceTracker(path string, host string, trackerLink string) string {
	now := time.Now()
	seconds := now.Unix()

	host = strings.Split(host, ":")[0]

	tracker := fmt.Sprintf("%d.%s.%s", seconds, host, trackerLink)

	newPath := strings.Replace(path, "{tracker}", tracker, -1)

	return newPath
}

// readLinesOrLiteral tries to read lines from a file, returning
// the arg in a string slice if the file doesn't exist, unless
// the arg matches its default value
func readLinesOrLiteral(arg, argDefault string) ([]string, error) {
	if isFile(arg) {
		return readLines(arg)
	}

	// if the argument isn't a file, but it is the default, don't
	// treat it as a literal value
	if arg == argDefault {
		return []string{}, fmt.Errorf("file %s not found", arg)
	}

	return []string{arg}, nil
}

// isFile returns true if its argument is a regular file
func isFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().IsRegular()
}
