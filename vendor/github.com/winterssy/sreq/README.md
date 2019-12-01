# sreq

A simple, user-friendly and concurrent safe HTTP request library for Go, 's' means simple.

[![Actions Status](https://github.com/winterssy/sreq/workflows/Test/badge.svg)](https://github.com/winterssy/sreq/actions) [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![codecov](https://codecov.io/gh/winterssy/sreq/branch/master/graph/badge.svg)](https://codecov.io/gh/winterssy/sreq) [![Go Report Card](https://goreportcard.com/badge/github.com/winterssy/sreq)](https://goreportcard.com/report/github.com/winterssy/sreq) [![GoDoc](https://godoc.org/github.com/winterssy/sreq?status.svg)](https://godoc.org/github.com/winterssy/sreq) [![License](https://img.shields.io/github/license/winterssy/sreq.svg)](LICENSE)

## Notes

`sreq` now is under an alpha test state, its APIs may be changed in future so it's not recommended to use in production. Welcome to give advise to the project.

## Features

- GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS, etc.
- Easy set query params, headers and cookies.
- Easy send form, JSON or files payload.
- Easy set basic authentication or bearer token.
- Easy set proxy.
- Easy set context.
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

The usages of `sreq` are very similar to `net/http` library, you can switch from it to `sreq` easily. For example, if your HTTP request code like this:

```go
resp, err := http.Get("http://www.google.com")
```

Use `sreq` you just need to change your code like this:

```go
resp, err := sreq.Get("http://www.google.com").Resolve()
```

[Code examples](examples)

## License

MIT.
