package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
)

// a request is a wrapper for a URL that we want to request
type request struct {
	method  string
	url     *url.URL
	headers []string
}

// a response is a wrapper around an HTTP response;
// it contains the request value for context.
type response struct {
	request request
	status  string
	headers []string
	body    []byte
	err     error
}

// a requester is a function that makes HTTP requests
type requester func(request) response

type headerArgs []string

func (h *headerArgs) Set(val string) error {
	*h = append(*h, val)
	return nil
}

func (h headerArgs) String() string {
	return "string"
}

func (h headerArgs) StringSlice() []string {
	return []string(h)
}

func main() {

	// headers param
	var headers headerArgs
	flag.Var(&headers, "header", "")
	flag.Var(&headers, "H", "")

	// method param
	method := "GET"
	flag.StringVar(&method, "method", "GET", "")
	flag.StringVar(&method, "X", "GET", "")

	flag.Parse()

	// suffixes might be in a file, or it might be a single value
	suffixArg := flag.Arg(0)
	if suffixArg == "" {
		suffixArg = "suffixes"
	}
	var suffixes []string
	if f, err := os.Stat(suffixArg); err == nil && f.Mode().IsRegular() {
		lines, err := readLines(suffixArg)
		if err != nil {
			log.Fatal(err)
		}
		suffixes = lines
	} else if suffixArg != "suffixes" {
		// don't treat the default suffixes filename as a literal value
		suffixes = []string{suffixArg}
	}

	// prefixes are always in a file
	prefixFile := flag.Arg(1)
	if prefixFile == "" {
		prefixFile = "prefixes"
	}
	prefixes, err := readLines(prefixFile)
	if err != nil {
		log.Fatal(err)
	}

	// default the output directory to ./out
	outputDir := flag.Arg(2)
	if outputDir == "" {
		outputDir = "./out"
	}

	// the request and response channels for
	// the worker pool
	requests := make(chan request)
	responses := make(chan response)

	// spin up some workers to do the requests
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			for req := range requests {
				responses <- doRequest(req)
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
			path, err := res.save(outputDir)
			if err != nil {
				fmt.Printf("failed to save file: %s\n", err)
			}
			fmt.Printf("%s %s (%s)\n", path, res.request.url, res.status)
		}
		owg.Done()
	}()

	// send requests for each suffix for every prefix
	for _, suffix := range suffixes {
		for _, prefix := range prefixes {
			u, err := url.Parse(prefix + suffix)
			if err != nil {
				fmt.Printf("failed to parse url: %s\n", err)
				continue
			}
			requests <- request{method: method, url: u, headers: headers}
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

func init() {
	flag.Usage = func() {
		h := "Request many paths (suffixes) for many hosts (prefixes)\n\n"

		h += "Usage:\n"
		h += "  meg [suffix|suffixFile] [prefixFile] [outputDir]\n\n"

		h += "Options:\n"
		h += "  -H, --header  Send a custom HTTP header\n"
		h += "  -X, --method  HTTP method (default: GET)\n\n"

		h += "Defaults:\n"
		h += "  suffixFile: ./suffixes\n"
		h += "  prefixFile: ./prefixes\n"
		h += "  outputDir:  ./out\n\n"

		h += "Suffix file format:\n"
		h += "  /robots.txt\n"
		h += "  /package.json\n"
		h += "  /security.txt\n\n"

		h += "Prefix file format:\n"
		h += "  http://example.com\n"
		h += "  https://example.edu\n"
		h += "  https://example.net\n\n"

		h += "Examples:\n"
		h += "  meg /robots.txt\n"
		h += "  meg hosts.txt paths.txt output\n"

		fmt.Fprintf(os.Stderr, h)
	}
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
