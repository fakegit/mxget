# log

A simple logger  to replace `log` standard library of Go.

![Test](https://img.shields.io/github/workflow/status/winterssy/log/Test/master?label=Test&logo=appveyor) ![Go Report Card](https://goreportcard.com/badge/github.com/winterssy/log) [![GoDoc](https://godoc.org/github.com/winterssy/log?status.svg)](https://godoc.org/github.com/winterssy/log) [![License](https://img.shields.io/github/license/winterssy/log.svg)](LICENSE)

## Install

```sh
go get -u github.com/winterssy/log
```

## Usage
```go
import "github.com/winterssy/log"
```

## Quick Start
```go
package main

import "github.com/winterssy/log"

func main() {
	log.SetLevel(log.Ldebug)
	log.Debug("hello world")
	log.Info("hello world")
	log.Warn("hello world")
	log.Error("hello world")
}
```

## License
MIT.
