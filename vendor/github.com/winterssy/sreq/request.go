package sreq

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	stdurl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MethodGet represents the GET method for HTTP.
	MethodGet = "GET"

	// MethodHead represents the HEAD method for HTTP.
	MethodHead = "HEAD"

	// MethodPost represents the POST method for HTTP.
	MethodPost = "POST"

	// MethodPut represents the PUT method for HTTP.
	MethodPut = "PUT"

	// MethodPatch represents the PATCH method for HTTP.
	MethodPatch = "PATCH"

	// MethodDelete represents the DELETE method for HTTP.
	MethodDelete = "DELETE"

	// MethodConnect represents the CONNECT method for HTTP.
	MethodConnect = "CONNECT"

	// MethodOptions represents the OPTIONS method for HTTP.
	MethodOptions = "OPTIONS"

	// MethodTrace represents the TRACE method for HTTP.
	MethodTrace = "TRACE"
)

type (
	// Request wraps the raw HTTP request.
	Request struct {
		RawRequest *http.Request
		Err        error

		retry retry
	}

	// RequestOption specifies the request options, like params, form, etc.
	RequestOption func(*Request) *Request

	retry struct {
		enable     bool
		attempts   int
		delay      time.Duration
		conditions []func(*Response) bool
	}
)

func (req *Request) raiseError(cause string, err error) {
	if req.Err != nil {
		return
	}

	req.Err = fmt.Errorf("sreq [%s]: %s", cause, err.Error())
}

// NewRequest returns a new Request given a method, URL and optional body.
func NewRequest(method string, url string, body io.Reader) *Request {
	req := new(Request)
	rawRequest, err := http.NewRequest(method, url, body)
	if err != nil {
		req.raiseError("NewRequest", err)
		return req
	}

	rawRequest.Header.Set("User-Agent", "sreq "+Version)
	req.RawRequest = rawRequest
	return req
}

// Resolve resolves req and returns its raw HTTP request.
func (req *Request) Resolve() (*http.Request, error) {
	return req.RawRequest, req.Err
}

// SetHost specifies the host on which the URL is sought.
func (req *Request) SetHost(host string) *Request {
	req.RawRequest.Host = host
	return req
}

// SetHeaders sets headers for the HTTP request.
func (req *Request) SetHeaders(headers Headers) *Request {
	for k, v := range headers {
		req.RawRequest.Header.Set(k, v)
	}
	return req
}

// SetUserAgent sets User-Agent header value for the HTTP request.
func (req *Request) SetUserAgent(userAgent string) *Request {
	req.RawRequest.Header.Set("User-Agent", userAgent)
	return req
}

// SetQuery sets query params for the HTTP request.
func (req *Request) SetQuery(params Params) *Request {
	query := req.RawRequest.URL.Query()
	for k, v := range params {
		query.Set(k, v)
	}

	req.RawRequest.URL.RawQuery = query.Encode()
	return req
}

// SetRaw sets raw bytes payload for the HTTP request.
func (req *Request) SetRaw(raw []byte, contentType string) *Request {
	r := bytes.NewBuffer(raw)
	req.RawRequest.Body = ioutil.NopCloser(r)
	req.RawRequest.ContentLength = int64(r.Len())
	buf := r.Bytes()
	req.RawRequest.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(buf)
		return ioutil.NopCloser(r), nil
	}

	req.RawRequest.Header.Set("Content-Type", contentType)
	return req
}

// SetText sets plain text payload for the HTTP request.
func (req *Request) SetText(text string) *Request {
	r := bytes.NewBufferString(text)
	req.RawRequest.Body = ioutil.NopCloser(r)
	req.RawRequest.ContentLength = int64(r.Len())
	buf := r.Bytes()
	req.RawRequest.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(buf)
		return ioutil.NopCloser(r), nil
	}

	req.RawRequest.Header.Set("Content-Type", "text/plain")
	return req
}

// SetForm sets form payload for the HTTP request.
func (req *Request) SetForm(form Form) *Request {
	data := stdurl.Values{}
	for k, v := range form {
		data.Set(k, v)
	}

	r := strings.NewReader(data.Encode())
	req.RawRequest.Body = ioutil.NopCloser(r)
	req.RawRequest.ContentLength = int64(r.Len())
	snapshot := *r
	req.RawRequest.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return ioutil.NopCloser(&r), nil
	}

	req.RawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// SetJSON sets json payload for the HTTP request.
func (req *Request) SetJSON(data JSON, escapeHTML bool) *Request {
	b, err := jsonMarshal(data, "", "", escapeHTML)
	if err != nil {
		req.raiseError("Request.SetJSON", err)
		return req
	}

	r := bytes.NewReader(b)
	req.RawRequest.Body = ioutil.NopCloser(r)
	req.RawRequest.ContentLength = int64(r.Len())
	snapshot := *r
	req.RawRequest.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return ioutil.NopCloser(&r), nil
	}

	req.RawRequest.Header.Set("Content-Type", "application/json")
	return req
}

// SetFiles sets files payload for the HTTP request.
func (req *Request) SetFiles(files Files) *Request {
	for fieldName, filePath := range files {
		if _, err := existsFile(filePath); err != nil {
			req.raiseError("Request.SetFiles",
				fmt.Errorf("file for %q not ready: %s", fieldName, err.Error()))
			return req
		}
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer mw.Close()

		for fieldName, filePath := range files {
			fileName := filepath.Base(filePath)
			part, err := mw.CreateFormFile(fieldName, fileName)
			if err != nil {
				return
			}

			file, err := os.Open(filePath)
			if err != nil {
				return
			}

			_, err = io.Copy(part, file)
			if err != nil || file.Close() != nil {
				return
			}
		}
	}()

	req.RawRequest.Body = ioutil.NopCloser(pr)
	req.RawRequest.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func existsFile(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		if fi.Mode().IsDir() {
			return false, fmt.Errorf("%q is a directory", filename)
		}
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, err
	}

	return true, err
}

// SetCookies sets cookies for the HTTP request.
func (req *Request) SetCookies(cookies ...*http.Cookie) *Request {
	for _, c := range cookies {
		req.RawRequest.AddCookie(c)
	}
	return req
}

// SetBasicAuth sets basic authentication for the HTTP request.
func (req *Request) SetBasicAuth(username string, password string) *Request {
	req.RawRequest.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	return req
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// SetBearerToken sets bearer token for the HTTP request.
func (req *Request) SetBearerToken(token string) *Request {
	req.RawRequest.Header.Set("Authorization", "Bearer "+token)
	return req
}

// SetContext sets context for the HTTP request.
func (req *Request) SetContext(ctx context.Context) *Request {
	if ctx == nil {
		req.raiseError("Request.SetContext", errors.New("nil Context"))
		return req
	}

	req.RawRequest = req.RawRequest.WithContext(ctx)
	return req
}

// SetRetry sets retry policy for the HTTP request.
func (req *Request) SetRetry(attempts int, delay time.Duration, conditions ...func(*Response) bool) *Request {
	if attempts > 1 {
		req.retry.enable = true
		req.retry.attempts = attempts
		req.retry.delay = delay
		req.retry.conditions = conditions
	}
	return req
}

// WithHost specifies the host on which the URL is sought.
func WithHost(host string) RequestOption {
	return func(req *Request) *Request {
		return req.SetHost(host)
	}
}

// WithHeaders sets headers for the HTTP request.
func WithHeaders(headers Headers) RequestOption {
	return func(req *Request) *Request {
		return req.SetHeaders(headers)
	}
}

// WithUserAgent sets User-Agent header value for the HTTP request.
func WithUserAgent(userAgent string) RequestOption {
	return func(req *Request) *Request {
		return req.SetUserAgent(userAgent)
	}
}

// WithQuery sets query params for the HTTP request.
func WithQuery(params Params) RequestOption {
	return func(req *Request) *Request {
		return req.SetQuery(params)
	}
}

// WithRaw sets raw bytes payload for the HTTP request.
func WithRaw(raw []byte, contentType string) RequestOption {
	return func(req *Request) *Request {
		return req.SetRaw(raw, contentType)
	}
}

// WithText sets plain text payload for the HTTP request.
func WithText(text string) RequestOption {
	return func(req *Request) *Request {
		return req.SetText(text)
	}
}

// WithForm sets form payload for the HTTP request.
func WithForm(form Form) RequestOption {
	return func(req *Request) *Request {
		return req.SetForm(form)
	}
}

// WithJSON sets json payload for the HTTP request.
func WithJSON(data JSON, escapeHTML bool) RequestOption {
	return func(req *Request) *Request {
		return req.SetJSON(data, escapeHTML)
	}
}

// WithFiles sets files payload for the HTTP request.
func WithFiles(files Files) RequestOption {
	return func(req *Request) *Request {
		return req.SetFiles(files)
	}
}

// WithCookies sets cookies for the HTTP request.
func WithCookies(cookies ...*http.Cookie) RequestOption {
	return func(req *Request) *Request {
		return req.SetCookies(cookies...)
	}
}

// WithBasicAuth sets basic authentication for the HTTP request.
func WithBasicAuth(username string, password string) RequestOption {
	return func(req *Request) *Request {
		return req.SetBasicAuth(username, password)
	}
}

// WithBearerToken sets bearer token for the HTTP request.
func WithBearerToken(token string) RequestOption {
	return func(req *Request) *Request {
		return req.SetBearerToken(token)
	}
}

// WithContext sets context for the HTTP request.
func WithContext(ctx context.Context) RequestOption {
	return func(req *Request) *Request {
		return req.SetContext(ctx)
	}
}

// WithRetry sets retry policy for the HTTP request.
func WithRetry(attempts int, delay time.Duration,
	conditions ...func(*Response) bool) RequestOption {
	return func(req *Request) *Request {
		return req.SetRetry(attempts, delay, conditions...)
	}
}
