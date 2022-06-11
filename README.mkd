# meg

meg is a tool for fetching lots of URLs but still being 'nice' to servers.

It can be used to fetch many paths for many hosts; fetching one path
for all hosts before moving on to the next path and repeating.

You get lots of results quickly, but non of the individual hosts get
flooded with traffic.

## Install

meg is written in Go and has no run-time dependencies. If you have Go 1.9
or later installed and configured you can install meg with `go install`:

```
▶ go install github.com/tomnomnom/meg@latest
```

Or [download a binary](https://github.com/tomnomnom/meg/releases) and
put it somewhere in your `$PATH` (e.g. in `/usr/bin/`).

### Install Errors

If you see an error like this it means your version of Go is too old:

```
# github.com/tomnomnom/rawhttp
/root/go/src/github.com/tomnomnom/rawhttp/request.go:102: u.Hostname undefined (
type *url.URL has no field or method Hostname)
/root/go/src/github.com/tomnomnom/rawhttp/request.go:103: u.Port undefined (type
 *url.URL has no field or method Port)
 /root/go/src/github.com/tomnomnom/rawhttp/request.go:259: undefined: x509.System
 CertPool
```

You should either update your version of Go, or use a binary release
for your platform.

## Basic Usage

Given a file full of paths:

```
/robots.txt
/.well-known/security.txt
/package.json
```

And a file full of hosts (with a protocol):

```
http://example.com
https://example.com
http://example.net
```

`meg` will request each *path* for every *host*:

```
▶ meg --verbose paths hosts
out/example.com/45ed6f717d44385c5e9c539b0ad8dc71771780e0 http://example.com/robots.txt (404 Not Found)
out/example.com/61ac5fbb9d3dd054006ae82630b045ba730d8618 https://example.com/robots.txt (404 Not Found)
out/example.net/1432c16b671043271eab84111242b1fe2a28eb98 http://example.net/robots.txt (404 Not Found)
out/example.net/61deaa4fa10a6f601adb74519a900f1f0eca38b7 http://example.net/.well-known/security.txt (404 Not Found)
out/example.com/20bc94a296f17ce7a4e2daa2946d0dc12128b3f1 http://example.com/.well-known/security.txt (404 Not Found)
...
```

And save the output in a directory called `./out`:

```
▶ head -n 20 ./out/example.com/45ed6f717d44385c5e9c539b0ad8dc71771780e0
http://example.com/robots.txt

> GET /robots.txt HTTP/1.1
> Host: example.com

< HTTP/1.1 404 Not Found
< Expires: Sat, 06 Jan 2018 01:05:38 GMT
< Server: ECS (lga/13A2)
< Accept-Ranges: bytes
< Cache-Control: max-age=604800
< Content-Type: text/*
< Content-Length: 1270
< Date: Sat, 30 Dec 2017 01:05:38 GMT
< Last-Modified: Sun, 24 Dec 2017 06:53:36 GMT
< X-Cache: 404-HIT

<!doctype html>
<html>
<head>
```

Without any arguments, meg will read paths from a file called `./paths`,
and hosts from a file called `./hosts`. There will also be no output:

```
▶ meg
▶
```

But it will save an *index* file to `./out/index`:

```
▶ head -n 2 ./out/index
out/example.com/538565d7ab544bc3bec5b2f0296783aaec25e756 http://example.com/package.json (404 Not Found)
out/example.com/20bc94a296f17ce7a4e2daa2946d0dc12128b3f1 http://example.com/.well-known/security.txt (404 Not Found)
```

You can use the index file to find where the response is stored, but it's
often easier to find what you're looking for with `grep`:

```
▶ grep -Hnri '< Server:' out/
out/example.com/61ac5fbb9d3dd054006ae82630b045ba730d8618:14:< Server: ECS (lga/13A2)
out/example.com/bd8d9f4c470ffa0e6ec8cfa8ba1c51d62289b6dd:16:< Server: ECS (lga/13A3)
```

If you want to request just one path, you can specify it directly as an argument:

```
▶ meg /admin.php
```

## Detailed Usage

meg's help output tries to actually be helpful:

```
▶ meg --help
Request many paths for many hosts

Usage:
  meg [options] [path|pathsFile] [hostsFile] [outputDir]

Options:
  -c, --concurrency <val>    Set the concurrency level (defaut: 20)
  -d, --delay <val>          Milliseconds between requests to the same host (default: 5000)
  -H, --header <header>      Send a custom HTTP header
  -r, --rawhttp              Use the rawhttp library for requests (experimental)
  -s, --savestatus <status>  Save only responses with specific status code
  -v, --verbose              Verbose mode
  -X, --method <method>      HTTP method (default: GET)

Defaults:
  pathsFile: ./paths
  hostsFile: ./hosts
  outputDir:  ./out

Paths file format:
  /robots.txt
  /package.json
  /security.txt

Hosts file format:
  http://example.com
  https://example.edu
  https://example.net

Examples:
  meg /robots.txt
  meg -s 200 -X HEAD
  meg -c 30 /
  meg hosts.txt paths.txt output
```

### Concurrency

By default meg will attempt to make 20 concurrent requests. You can change that
with the `-c` or `--concurrency` option:

```
▶ meg --concurrency 5
```

It's not very friendly to keep the concurrency level higher than the number of
hosts - you may end up sending lots of requests to one host at once.

### Delay
By default meg will wait 5000 milliseconds between requests to the same host.
You can override that with the `-d` or `--delay` option:

```
▶ meg --delay 10000
```

**Warning:** before reducing the delay, ensure that you have permission to make
large volumes of requests to the hosts you're targeting.

### Adding Headers

You can set additional headers on the requests with the `-H` or `--header`
option:

```
▶ meg --header "Origin: https://evil.com"
▶ grep -h '^>' out/example.com/*
> GET /.well-known/security.txt HTTP/1.1
> Origin: https://evil.com
> Host: example.com
...
```

### Raw HTTP (Experimental)

If you want to send requests that aren't valid - for example with invalid URL encoding -
the Go HTTP client will fail:

```
▶ meg /%%0a0afoo:bar
request failed: parse https://example.org/%%0a0afoo:bar: invalid URL escape "%%0"
```

You can use the `-r` or `--rawhttp` flag to enable use of the [rawhttp](https://github.com/tomnomnom/rawhttp)
library, which does little to no validation on the request:

```
▶ meg --verbose --rawhttp /%%0a0afoo:bar
out/example.com/eac3a4978bfb95992e270c311582e6da4568d83d https://example.com/%%0a0afoo:bar (HTTP/1.1 404 Not Found)
```

The `rawhttp` library and its use is experimental. Amongst other things it doesn't
yet support chunked transfer encoding, so you may notice chunk lengths interspersed
with your output if you use it.

### Saving Only Certain Status Codes

If you only want to save results that returned a certain status code, you can
use the `-s` or `--savestatus` option:

```
▶ meg --savestatus 200 /robots.txt
```

### Specifying The Method

You can specify which HTTP method to use with the `-X` or `--method` option:

```
▶ meg --method TRACE
▶ grep -nri 'TRACE' out/
out/example.com/61ac5fbb9d3dd054006ae82630b045ba730d8618:3:> TRACE /robots.txt HTTP/1.1
out/example.com/bd8d9f4c470ffa0e6ec8cfa8ba1c51d62289b6dd:3:> TRACE /.well-known/security.txt HTTP/1.1
...
```
