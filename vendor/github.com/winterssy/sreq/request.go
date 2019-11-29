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

		retryPolicy retryOption
		ctx         context.Context
	}

	// RequestOption specifies the request options, like params, form, etc.
	RequestOption func(*Request) (*Request, error)

	retryOption struct {
		enable     bool
		attempts   int
		delay      time.Duration
		conditions []func(*Response) bool
	}
)

func (c *Client) newRequest(method string, url string, opts ...RequestOption) (*Request, error) {
	rawRequest, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	rawRequest.Header.Set("User-Agent", "sreq "+Version)
	req := &Request{
		RawRequest: rawRequest,
	}

	for _, opt := range opts {
		req, err = opt(req)
		if err != nil {
			return nil, err
		}
	}

	if req.ctx != nil {
		req.RawRequest = req.RawRequest.WithContext(req.ctx)
	}

	return req, nil
}

// Get makes a GET HTTP request.
func Get(url string, opts ...RequestOption) *Response {
	return gClient.Get(url, opts...)
}

// Get makes a GET HTTP request.
func (c *Client) Get(url string, opts ...RequestOption) *Response {
	return c.Send(MethodGet, url, opts...)
}

// Head makes a HEAD HTTP request.
func Head(url string, opts ...RequestOption) *Response {
	return gClient.Head(url, opts...)
}

// Head makes a HEAD HTTP request.
func (c *Client) Head(url string, opts ...RequestOption) *Response {
	return c.Send(MethodHead, url, opts...)
}

// Post makes a POST HTTP request.
func Post(url string, opts ...RequestOption) *Response {
	return gClient.Post(url, opts...)
}

// Post makes a POST HTTP request.
func (c *Client) Post(url string, opts ...RequestOption) *Response {
	return c.Send(MethodPost, url, opts...)
}

// Put makes a PUT HTTP request.
func Put(url string, opts ...RequestOption) *Response {
	return gClient.Put(url, opts...)
}

// Put makes a PUT HTTP request.
func (c *Client) Put(url string, opts ...RequestOption) *Response {
	return gClient.Send(MethodPut, url, opts...)
}

// Patch makes a PATCH HTTP request.
func Patch(url string, opts ...RequestOption) *Response {
	return gClient.Patch(url, opts...)
}

// Patch makes a PATCH HTTP request.
func (c *Client) Patch(url string, opts ...RequestOption) *Response {
	return c.Send(MethodPatch, url, opts...)
}

// Delete makes a DELETE HTTP request.
func Delete(url string, opts ...RequestOption) *Response {
	return gClient.Delete(url, opts...)
}

// Delete makes a DELETE HTTP request.
func (c *Client) Delete(url string, opts ...RequestOption) *Response {
	return c.Send(MethodDelete, url, opts...)
}

// Connect makes a CONNECT HTTP request.
func Connect(url string, opts ...RequestOption) *Response {
	return gClient.Connect(url, opts...)
}

// Connect makes a CONNECT HTTP request.
func (c *Client) Connect(url string, opts ...RequestOption) *Response {
	return c.Send(MethodConnect, url, opts...)
}

// Options makes an OPTIONS request.
func Options(url string, opts ...RequestOption) *Response {
	return gClient.Options(url, opts...)
}

// Options makes an OPTIONS request.
func (c *Client) Options(url string, opts ...RequestOption) *Response {
	return c.Send(MethodOptions, url, opts...)
}

// Trace makes a TRACE HTTP request.
func Trace(url string, opts ...RequestOption) *Response {
	return gClient.Trace(url, opts...)
}

// Trace makes a TRACE HTTP request.
func (c *Client) Trace(url string, opts ...RequestOption) *Response {
	return c.Send(MethodTrace, url, opts...)
}

// Send makes an HTTP request using a specified method.
func Send(method string, url string, opts ...RequestOption) *Response {
	return gClient.Send(method, url, opts...)
}

// Send makes an HTTP request using a specified method.
func (c *Client) Send(method string, url string, opts ...RequestOption) *Response {
	resp := new(Response)
	req, err := c.newRequest(method, url, opts...)
	if err != nil {
		resp.Err = err
		return resp
	}

	if !req.retryPolicy.enable {
		return c.Do(req.RawRequest)
	}

	ctx := req.RawRequest.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	for i := req.retryPolicy.attempts; i > 0; i-- {
		resp = c.Do(req.RawRequest)
		if err = ctx.Err(); err != nil {
			resp.Err = err
			return resp
		}

		shouldRetry := resp.Err != nil
		for _, condition := range req.retryPolicy.conditions {
			shouldRetry = condition(resp)
			if shouldRetry {
				break
			}
		}

		if !shouldRetry {
			return resp
		}

		select {
		case <-time.After(req.retryPolicy.delay):
		case <-ctx.Done():
			resp.Err = ctx.Err()
			return resp
		}
	}

	return resp
}

// WithHost specifies the host on which the URL is sought.
func WithHost(host string) RequestOption {
	return func(req *Request) (*Request, error) {
		req.RawRequest.Host = host
		return req, nil
	}
}

// WithHeaders sets headers for the HTTP request.
func WithHeaders(headers Headers) RequestOption {
	return func(req *Request) (*Request, error) {
		for k, v := range headers {
			req.RawRequest.Header.Set(k, v)
		}
		return req, nil
	}
}

// WithUserAgent sets User-Agent header value for the HTTP request.
func WithUserAgent(userAgent string) RequestOption {
	return func(req *Request) (*Request, error) {
		req.RawRequest.Header.Set("User-Agent", userAgent)
		return req, nil
	}
}

// WithQuery sets query params for the HTTP request.
func WithQuery(params Params) RequestOption {
	return func(req *Request) (*Request, error) {
		query := req.RawRequest.URL.Query()
		for k, v := range params {
			query.Set(k, v)
		}
		req.RawRequest.URL.RawQuery = query.Encode()
		return req, nil
	}
}

// WithRaw sets raw bytes payload for the HTTP request.
func WithRaw(raw []byte, contentType string) RequestOption {
	return func(req *Request) (*Request, error) {
		r := bytes.NewBuffer(raw)
		req.RawRequest.Body = ioutil.NopCloser(r)
		req.RawRequest.ContentLength = int64(r.Len())
		buf := r.Bytes()
		req.RawRequest.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(buf)
			return ioutil.NopCloser(r), nil
		}

		req.RawRequest.Header.Set("Content-Type", contentType)
		return req, nil
	}
}

// WithText sets plain text payload for the HTTP request.
func WithText(text string) RequestOption {
	return func(req *Request) (*Request, error) {
		r := bytes.NewBufferString(text)
		req.RawRequest.Body = ioutil.NopCloser(r)
		req.RawRequest.ContentLength = int64(r.Len())
		buf := r.Bytes()
		req.RawRequest.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(buf)
			return ioutil.NopCloser(r), nil
		}

		req.RawRequest.Header.Set("Content-Type", "text/plain")
		return req, nil
	}
}

// WithForm sets form payload for the HTTP request.
func WithForm(form Form) RequestOption {
	return func(req *Request) (*Request, error) {
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
		return req, nil
	}
}

// WithJSON sets json payload for the HTTP request.
func WithJSON(data JSON, escapeHTML bool) RequestOption {
	return func(req *Request) (*Request, error) {
		b, err := Marshal(data, "", "", escapeHTML)
		if err != nil {
			return nil, err
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
		return req, nil
	}
}

// WithFiles sets files payload for the HTTP request.
func WithFiles(files Files) RequestOption {
	return func(req *Request) (*Request, error) {
		for fieldName, filePath := range files {
			if _, err := ExistsFile(filePath); err != nil {
				return nil, fmt.Errorf("sreq: file for %q not ready: %v", fieldName, err)
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
					continue
				}

				file, err := os.Open(filePath)
				if err != nil {
					continue
				}

				_, err = io.Copy(part, file)
				if err != nil || file.Close() != nil {
					continue
				}
			}
		}()

		req.RawRequest.Body = ioutil.NopCloser(pr)
		req.RawRequest.Header.Set("Content-Type", mw.FormDataContentType())
		return req, nil
	}
}

// WithCookies sets cookies for the HTTP request.
func WithCookies(cookies ...*http.Cookie) RequestOption {
	return func(req *Request) (*Request, error) {
		for _, c := range cookies {
			req.RawRequest.AddCookie(c)
		}
		return req, nil
	}
}

// WithBasicAuth sets basic authentication for the HTTP request.
func WithBasicAuth(username string, password string) RequestOption {
	return func(req *Request) (*Request, error) {
		req.RawRequest.Header.Set("Authorization", "Basic "+basicAuth(username, password))
		return req, nil
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// WithBearerToken sets bearer token for the HTTP request.
func WithBearerToken(token string) RequestOption {
	return func(req *Request) (*Request, error) {
		req.RawRequest.Header.Set("Authorization", "Bearer "+token)
		return req, nil
	}
}

// WithContext sets context for the HTTP request.
func WithContext(ctx context.Context) RequestOption {
	return func(req *Request) (*Request, error) {
		if ctx == nil {
			return nil, errors.New("sreq: nil Context")
		}

		req.ctx = ctx
		return req, nil
	}
}

// WithRetry sets retry policy for the HTTP request.
func WithRetry(attempts int, delay time.Duration, conditions ...func(*Response) bool) RequestOption {
	return func(req *Request) (*Request, error) {
		if attempts > 1 {
			req.retryPolicy.enable = true
			req.retryPolicy.attempts = attempts
			req.retryPolicy.delay = delay
			req.retryPolicy.conditions = conditions
		}
		return req, nil
	}
}
