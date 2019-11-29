# sreq

A simple, user-friendly and concurrent safe HTTP request library for Go, 's' means simple.

- [简体中文](README_CN.md)

[![Actions Status](https://github.com/winterssy/sreq/workflows/CI/badge.svg)](https://github.com/winterssy/sreq/actions) [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![codecov](https://codecov.io/gh/winterssy/sreq/branch/master/graph/badge.svg)](https://codecov.io/gh/winterssy/sreq) [![Go Report Card](https://goreportcard.com/badge/github.com/winterssy/sreq)](https://goreportcard.com/report/github.com/winterssy/sreq) [![GoDoc](https://godoc.org/github.com/winterssy/sreq?status.svg)](https://godoc.org/github.com/winterssy/sreq) [![License](https://img.shields.io/github/license/winterssy/sreq.svg)](LICENSE)

## Notes

`sreq` now is under an alpha test state, its APIs may be changed in future so it's not recommended to use in production. Welcome to give advise to the project.

## Features

- GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS, etc.
- Easy set query params, headers and cookies.
- Easy send form, JSON or upload files.
- Easy set basic authentication or bearer token.
- Easy set proxy.
- Easy set context.
- Session support.
- Customize HTTP client.
- Easy decode responses, raw data, text representation and unmarshal the JSON-encoded data.
- Concurrent safe.

## Install

```sh
go get -u github.com/winterssy/sreq
```

## Usage

```go
import "github.com/winterssy/sreq"
```

## Examples

The usages of `sreq` are very similar to `net/http` library, you can switch from it to `sreq` easily. For example, if your HTTP request code like this:

```go
resp, err := http.Get("http://www.google.com")
```

Use `sreq` you just need to change your code like this:

```go
resp, err := sreq.Get("http://www.google.com").Resolve()
```

See more examples as follow.

- [Set Query Params](#Set-Query-Params)
- [Set Headers](#Set-Headers)
- [Set Cookies](#Set-Cookies)
- [Send Form](#Send-Form)
- [Send JSON](#Send-JSON)
- [Upload Files](#Upload-Files)
- [Set Basic Authentication](#Set-Basic-Authentication)
- [Set Bearer Token](#Set-Bearer-Token)
- [Set Proxy](#Set-Proxy)
- [Session Support](#Session-Support)
- [Customize HTTP Client](#Customize-HTTP-Client)
- [Concurrent Safe](#Concurrent-Safe)

### Set Query Params

```go
data, err := sreq.
    Get("http://httpbin.org/get",
        sreq.WithQuery(sreq.Params{
            "k1": "v1",
            "k2": "v2",
        }),
       ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Set Headers

```go
data, err := sreq.
    Get("http://httpbin.org/get",
        sreq.WithHeaders(sreq.Headers{
            "Origin":  "http://httpbin.org",
            "Referer": "http://httpbin.org",
        }),
       ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Set Cookies

```go
data, err := sreq.
    Get("http://httpbin.org/cookies",
        sreq.WithCookies(
            &http.Cookie{
                Name:  "n1",
                Value: "v1",
            },
            &http.Cookie{
                Name:  "n2",
                Value: "v2",
            },
        ),
       ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Send Form

```go
data, err := sreq.
    Post("http://httpbin.org/post",
         sreq.WithForm(sreq.Form{
             "k1": "v1",
             "k2": "v2",
         }),
        ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Send JSON

```go
data, err := sreq.
    Post("http://httpbin.org/post",
         sreq.WithJSON(sreq.JSON{
             "msg": "hello world",
             "num": 2019,
         }, true),
        ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Upload Files

```go
data, err := sreq.
    Post("http://httpbin.org/post",
         sreq.WithFiles(sreq.Files{
             "image1": "./testdata/testimage1.jpg",
             "image2": "./testdata/testimage2.jpg",
         }),
        ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Set Basic Authentication

```go
data, err := sreq.
    Get("http://httpbin.org/basic-auth/admin/pass",
        sreq.WithBasicAuth("admin", "pass"),
       ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Set Bearer Token

```go
data, err := sreq.
    Get("http://httpbin.org/bearer",
        sreq.WithBearerToken("sreq"),
       ).
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Set Proxy

```go
client, _ := sreq.New(nil,
	sreq.ProxyFromURL("socks5://127.0.0.1:1080"),
)
data, err := client.
    Get("https://api.ipify.org?format=json").
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Session Support

if you want to keep session across the lifecycle of `sreq` , you just need to call `EnableSession` method when you construct a `*sreq.Client`.

```go
client, _ := sreq.New(nil,
	sreq.EnableSession(),
)
data, err := client.
    Get("http://httpbin.org/get").
    Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Customize HTTP Client

For advanced use, you can construct a `*sreq.Client` instance via a customized `*http.Client` , `sreq` also provides some useful API to help you do this, please read the API document for more detail.

```go
transport := &http.Transport{
    DialContext: (&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }).DialContext,
    MaxIdleConns:          100,
    IdleConnTimeout:       90 * time.Second,
    TLSHandshakeTimeout:   10 * time.Second,
    ExpectContinueTimeout: 1 * time.Second,
}
policy := func(req *http.Request, via []*http.Request) error {
    return http.ErrUseLastResponse
}
jar, _ := cookiejar.New(&cookiejar.Options{
    PublicSuffixList: publicsuffix.List,
})

client, err := sreq.New(transport,
	sreq.WithRedirectPolicy(policy),
	sreq.WithCookieJar(jar),
	sreq.WithTimeout(120*time.Second),
	sreq.ProxyFromURL("socks5://127.0.0.1:1080"),
)
if err != nil {
    panic(err)
}

data, err := client.
	Get("https://www.google.com").
	Text()
if err != nil {
    panic(err)
}
fmt.Println(data)
```

### Concurrent Safe

`sreq` is concurrent safe, you can easily use it across goroutines.

```go
const MaxWorker = 1000
wg := new(sync.WaitGroup)

for i := 0; i < MaxWorker; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()

        params := sreq.Params{}
        params.Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))

        data, err := sreq.
            Get("http://httpbin.org/get",
                sreq.WithQuery(params),
               ).
            Text()
        if err != nil {
            return
        }

        fmt.Println(data)
    }(i)
}

wg.Wait()
```

## License

MIT.
