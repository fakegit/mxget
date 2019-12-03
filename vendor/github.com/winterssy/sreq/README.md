# sreq

A simple, user-friendly and concurrent safe HTTP request library for Go, 's' means simple. `sreq` provides many convenient APIs to access `net/http` , aims to simplify your work efficiently. 

[![Actions Status](https://github.com/winterssy/sreq/workflows/Test/badge.svg)](https://github.com/winterssy/sreq/actions) [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![codecov](https://codecov.io/gh/winterssy/sreq/branch/master/graph/badge.svg)](https://codecov.io/gh/winterssy/sreq) [![Go Report Card](https://goreportcard.com/badge/github.com/winterssy/sreq)](https://goreportcard.com/report/github.com/winterssy/sreq) [![GoDoc](https://godoc.org/github.com/winterssy/sreq?status.svg)](https://godoc.org/github.com/winterssy/sreq) [![License](https://img.shields.io/github/license/winterssy/sreq.svg)](LICENSE)

## Features

- GET, POST, PUT, PATCH, DELETE, etc.
- Easy set query params, headers and cookies.
- Easy send form, JSON or files payload.
- Easy set basic authentication or bearer token.
- Easy set proxy.
- Easy set context.
- Retry policy support.
- Automatic cookies management.
- Easy decode responses, raw data, text representation and unmarshal the JSON-encoded data.
- Friendly debugging.
- Concurrent safe.

## Install

```sh
go get -u github.com/winterssy/sreq
```

## Usage

```go
import "github.com/winterssy/sreq"
```

## Quick start

The usages of `sreq` are very similar to `net/http` , you can switch from it to `sreq` easily. For example, if your HTTP request code like this:

```go
resp, err := http.Get("http://www.google.com")
```

Use `sreq` you just need to change your code like this:

```go
resp, err := sreq.Get("http://www.google.com").Resolve()
```

You have two ways to access the APIs of `sreq` .

```go
const (
    url       = "http://httpbin.org/get"
    userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
)

params := sreq.Params{
    "k1": "v1",
    "k2": "v2",
}

client := sreq.New()

// 1
req := sreq.
	NewRequest("GET", url, nil).
	SetQuery(params).
	SetUserAgent(userAgent)
err := client.
	Do(req).
	EnsureStatusOk().
	Verbose(ioutil.Discard)
if err != nil {
    panic(err)
}

// 2 (Recommended)
err = client.Get(url,
	sreq.WithQuery(params),
	sreq.WithUserAgent(userAgent),
).
	EnsureStatusOk().
	Verbose(os.Stdout)
if err != nil {
    panic(err)
}

// Output:
// > GET /get?k1=v1&k2=v2 HTTP/1.1
// > Host: httpbin.org
// > User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36
// >
// < HTTP/1.1 200 OK
// < Access-Control-Allow-Origin: *
// < Content-Type: application/json
// < Referrer-Policy: no-referrer-when-downgrade
// < Server: nginx
// < Access-Control-Allow-Credentials: true
// < Date: Mon, 02 Dec 2019 06:24:29 GMT
// < X-Content-Type-Options: nosniff
// < X-Frame-Options: DENY
// < X-Xss-Protection: 1; mode=block
// < Connection: keep-alive
// <
// {
//   "args": {
//     "k1": "v1",
//     "k2": "v2"
//   },
//   "headers": {
//     "Accept-Encoding": "gzip",
//     "Host": "httpbin.org",
//     "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
//   },
//   "origin": "8.8.8.8, 8.8.8.8",
//   "url": "https://httpbin.org/get?k1=v1&k2=v2"
// }
```

[Code examples](examples)

## Projects

`sreq` is used by the following projects.

- [mxget](https://github.com/winterssy/mxget)

## License

[MIT](LICENSE)