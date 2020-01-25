# slog

A simple logger for Go, to replace the `log` standard library.

![Test](https://img.shields.io/github/workflow/status/winterssy/slog/Test/master?label=Test&logo=appveyor) [![Go Report Card](https://goreportcard.com/badge/github.com/winterssy/slog)](https://goreportcard.com/report/github.com/winterssy/slog) [![GoDoc](https://godoc.org/github.com/winterssy/slog?status.svg)](https://godoc.org/github.com/winterssy/slog) [![License](https://img.shields.io/github/license/winterssy/slog.svg)](LICENSE)

## Install

```sh
go get -u github.com/winterssy/slog
```

## Usage
```go
import "github.com/winterssy/slog"
```

## Quick Start
```go
package main

import "github.com/winterssy/slog"

func main() {
	slog.SetLevel(slog.Ldebug)
	slog.Debug("hello world")
	slog.Info("hello world")
	slog.Warn("hello world")
	slog.Error("hello world")
}
```

## License

**[MIT](LICENSE)**
