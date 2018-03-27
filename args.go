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
	concurrency    int
	delay          int
	headers        headerArgs
	followLocation bool
	noPath         bool
	method         string
	saveStatus     int
	timeout        int
	verbose        bool

	paths  string
	hosts  string
	output string

	requester requester
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

	// follow location param
	followLocation := false
	flag.BoolVar(&followLocation, "location", false, "")
	flag.BoolVar(&followLocation, "L", false, "")

	// no path part
	noPath := false
	flag.BoolVar(&noPath, "nopath", false, "")

	// method param
	method := "GET"
	flag.StringVar(&method, "method", "GET", "")
	flag.StringVar(&method, "X", "GET", "")

	// savestatus param
	saveStatus := 0
	flag.IntVar(&saveStatus, "savestatus", 0, "")
	flag.IntVar(&saveStatus, "s", 0, "")

	// timeout param
	timeout := 10000
	flag.IntVar(&timeout, "timeout", 10000, "")
	flag.IntVar(&timeout, "t", 10000, "")

	// rawhttp param
	rawHTTP := false
	flag.BoolVar(&rawHTTP, "rawhttp", false, "")
	flag.BoolVar(&rawHTTP, "r", false, "")

	// verbose param
	verbose := false
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&verbose, "v", false, "")

	flag.Parse()

	// paths might be in a file, or it might be a single value
	paths := flag.Arg(0)
	if paths == "" {
		paths = defaultPathsFile
	}

	// hosts are always in a file
	hosts := flag.Arg(1)
	if hosts == "" {
		hosts = defaultHostsFile
	}

	// default the output directory to ./out
	output := flag.Arg(2)
	if output == "" {
		output = defaultOutputDir
	}

	// NASTY HACK ALERT
	// If you're using the --nopath option then there's no paths file,
	// so you can't specify a hosts file. Shift things along a bit, to fix it
	output = hosts
	hosts = paths

	// set the requester function to use
	requesterFn := goRequest
	if rawHTTP {
		requesterFn = rawRequest
	}

	return config{
		concurrency:    concurrency,
		delay:          delay,
		headers:        headers,
		followLocation: followLocation,
		noPath:         noPath,
		method:         method,
		saveStatus:     saveStatus,
		timeout:        timeout,
		requester:      requesterFn,
		verbose:        verbose,
		paths:          paths,
		hosts:          hosts,
		output:         output,
	}
}

func init() {
	flag.Usage = func() {
		h := "Request many paths for many hosts\n\n"

		h += "Usage:\n"
		h += "  meg [path|pathsFile] [hostsFile] [outputDir]\n\n"

		h += "Options:\n"
		h += "  -c, --concurrency <val>    Set the concurrency level (defaut: 20)\n"
		h += "  -d, --delay <millis>       Milliseconds between requests to the same host (defaut: 5000)\n"
		h += "  -H, --header <header>      Send a custom HTTP header\n"
		h += "  -L, --location             Follow redirects / location header\n"
		h += "		--nopath			   Treat the hosts file as complete URLs\n"
		h += "  -r, --rawhttp              Use the rawhttp library for requests (experimental)\n"
		h += "  -s, --savestatus <status>  Save only responses with specific status code\n"
		h += "  -t, --timeout <millis>     Set the HTTP timeout (default: 10000)\n"
		h += "  -v, --verbose              Verbose mode\n"
		h += "  -X, --method <method>      HTTP method (default: GET)\n\n"

		h += "Defaults:\n"
		h += "  pathsFile: ./paths\n"
		h += "  hostsFile: ./hosts\n"
		h += "  outputDir:  ./out\n\n"

		h += "Paths file format:\n"
		h += "  /robots.txt\n"
		h += "  /package.json\n"
		h += "  /security.txt\n\n"

		h += "Hosts file format:\n"
		h += "  http://example.com\n"
		h += "  https://example.edu\n"
		h += "  https://example.net\n\n"

		h += "Examples:\n"
		h += "  meg /robots.txt\n"
		h += "  meg paths.txt hosts.txt output\n"

		fmt.Fprintf(os.Stderr, h)
	}
}
