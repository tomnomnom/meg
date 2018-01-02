package main

import (
	"flag"
	"fmt"
	"os"
)

// headerArgs is the type used to store the header arguments
type headerArgs []string

func (h *headerArgs) Set(val string) error {
	*h = append(*h, val)
	return nil
}

func (h headerArgs) String() string {
	return "string"
}

type config struct {
	concurrency int
	delay       int
	headers     headerArgs
	method      string
	saveStatus  int
	requester   requester
	verbose     bool
	suffix      string
	prefix      string
	output      string
}

func processArgs() config {

	// concurrency param
	concurrency := 20
	flag.IntVar(&concurrency, "concurrency", 20, "")
	flag.IntVar(&concurrency, "c", 20, "")

	// delay param
	delay := 5000
	flag.IntVar(&delay, "delay", 5000, "")
	flag.IntVar(&delay, "d", 5000, "")

	// headers param
	var headers headerArgs
	flag.Var(&headers, "header", "")
	flag.Var(&headers, "H", "")

	// method param
	method := "GET"
	flag.StringVar(&method, "method", "GET", "")
	flag.StringVar(&method, "X", "GET", "")

	// savestatus param
	saveStatus := 0
	flag.IntVar(&saveStatus, "savestatus", 0, "")
	flag.IntVar(&saveStatus, "s", 0, "")

	// rawhttp param
	rawHTTP := false
	flag.BoolVar(&rawHTTP, "rawhttp", false, "")
	flag.BoolVar(&rawHTTP, "r", false, "")

	// verbose param
	verbose := false
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&verbose, "v", false, "")

	if verbose {
		fmt.Println("sdgfsd")
	}

	flag.Parse()

	// suffixes might be in a file, or it might be a single value
	suffix := flag.Arg(0)
	if suffix == "" {
		suffix = "suffixes"
	}

	// prefixes are always in a file
	prefix := flag.Arg(1)
	if prefix == "" {
		prefix = "prefixes"
	}

	// default the output directory to ./out
	output := flag.Arg(2)
	if output == "" {
		output = "./out"
	}

	// set the requester function to use
	requesterFn := goRequest
	if rawHTTP {
		requesterFn = rawRequest
	}

	return config{
		concurrency: concurrency,
		delay:       delay,
		headers:     headers,
		method:      method,
		saveStatus:  saveStatus,
		requester:   requesterFn,
		verbose:     verbose,
		suffix:      suffix,
		prefix:      prefix,
		output:      output,
	}
}

func init() {
	flag.Usage = func() {
		h := "Request many paths (suffixes) for many hosts (prefixes)\n\n"

		h += "Usage:\n"
		h += "  meg [suffix|suffixFile] [prefixFile] [outputDir]\n\n"

		h += "Options:\n"
		h += "  -c, --concurrency <val>    Set the concurrency level (defaut: 20)\n"
		h += "  -d, --delay <val>          Milliseconds between requests to the same host (defaut: 5000)\n"
		h += "  -H, --header <header>      Send a custom HTTP header\n"
		h += "  -r, --rawhttp              Use the rawhttp library for requests (experimental)\n"
		h += "  -s, --savestatus <status>  Save only responses with specific status code\n"
		h += "  -v, --verbose              Verbose mode\n"
		h += "  -X, --method <method>      HTTP method (default: GET)\n\n"

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
