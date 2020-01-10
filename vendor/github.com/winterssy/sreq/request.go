package sreq

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
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
		Body       io.Reader
		Host       string
		Headers    Headers
		Params     Params
		Form       Form
		Cookies    Cookies

		ctx           context.Context
		errBackground chan error
	}

	// RequestOption provides a convenient way to setup Request.
	RequestOption func(req *Request)

	// RequestInterceptor specifies a request interceptor.
	// If the returned error isn't nil, sreq will stop sending the request.
	RequestInterceptor func(req *Request) error
)

func (req *Request) raiseError(cause string, err error) {
	if req.Err != nil {
		return
	}

	req.Err = &RequestError{
		Cause: cause,
		Err:   err,
	}
}

// NewRequest returns a new Request given a method, URL.
func NewRequest(method string, url string) *Request {
	rawRequest, err := http.NewRequest(method, url, nil)
	if err != nil {
		err = &RequestError{
			Cause: "NewRequest",
			Err:   err,
		}
	}
	req := &Request{
		RawRequest: rawRequest,
		Err:        err,
		Headers:    make(Headers),
		Params:     make(Params),
		Form:       make(Form),
		Cookies:    make(Cookies),
	}
	return req
}

// Raw returns the raw HTTP request.
func (req *Request) Raw() (*http.Request, error) {
	return req.RawRequest, req.Err
}

func (req *Request) setup() error {
	if req.Err != nil {
		return req.Err
	}

	req.setHost()
	req.setQuery()
	req.setForm()
	req.setHeaders()
	req.setCookies()

	req.setBody() // must after setForm
	return nil
}

func (req *Request) setHost() {
	req.RawRequest.Host = req.Host
}

func (req *Request) setHeaders() {
	req.Headers.SetDefault("User-Agent", defaultUserAgent)
	for k, vs := range req.Headers.Decode() {
		req.RawRequest.Header.Del(k) // remove existing value
		for _, v := range vs {
			req.RawRequest.Header.Add(k, v)
		}
	}
}

func (req *Request) setQuery() {
	if len(req.Params) == 0 {
		return
	}

	for k, v := range req.RawRequest.URL.Query() {
		req.Params.SetDefault(k, v)
	}
	req.RawRequest.URL.RawQuery = req.Params.Encode(true)
}

func (req *Request) setForm() {
	if len(req.Form) == 0 {
		return
	}

	req.SetContentType("application/x-www-form-urlencoded")
	req.SetBody(strings.NewReader(req.Form.Encode(true)))
}

func (req *Request) setCookies() {
	if len(req.Cookies) == 0 {
		return
	}

	for _, c := range req.Cookies.Decode() {
		req.RawRequest.AddCookie(c)
	}
}

func (req *Request) setBody() {
	if req.Body == nil {
		return
	}

	rc, ok := req.Body.(io.ReadCloser)
	if !ok && req.Body != nil {
		rc = ioutil.NopCloser(req.Body)
	}
	req.RawRequest.Body = rc

	if req.Body != nil {
		switch v := req.Body.(type) {
		case *bytes.Buffer:
			req.RawRequest.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.RawRequest.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.RawRequest.ContentLength = int64(v.Len())
			snapshot := *v
			req.RawRequest.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.RawRequest.ContentLength = int64(v.Len())
			snapshot := *v
			req.RawRequest.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		default:
			// This is where we'd set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
		}
		// For client requests, Request.ContentLength of 0
		// means either actually 0, or unknown. The only way
		// to explicitly say that the ContentLength is zero is
		// to set the Body to nil. But turns out too much code
		// depends on NewRequest returning a non-nil Body,
		// so we use a well-known ReadCloser variable instead
		// and have the http package also treat that sentinel
		// variable to mean explicitly zero.
		if req.RawRequest.GetBody != nil && req.RawRequest.ContentLength == 0 {
			req.RawRequest.Body = http.NoBody
			req.RawRequest.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}
}

// SetBody sets body for the HTTP request.
// Notes: SetBody may not support retry since it's unable to read a stream twice.
func (req *Request) SetBody(body io.Reader) *Request {
	req.Body = body
	return req
}

// SetHost sets host for the HTTP request.
func (req *Request) SetHost(host string) *Request {
	req.Host = host
	return req
}

// SetHeaders sets headers for the HTTP request.
func (req *Request) SetHeaders(headers Headers) *Request {
	req.Headers.Update(headers)
	return req
}

// SetContentType sets Content-Type header value for the HTTP request.
func (req *Request) SetContentType(contentType string) *Request {
	req.Headers.Set("Content-Type", contentType)
	return req
}

// SetUserAgent sets User-Agent header value for the HTTP request.
func (req *Request) SetUserAgent(userAgent string) *Request {
	req.Headers.Set("User-Agent", userAgent)
	return req
}

// SetReferer sets Referer header value for the HTTP request.
func (req *Request) SetReferer(referer string) *Request {
	req.Headers.Set("Referer", referer)
	return req
}

// SetQuery sets query parameters for the HTTP request.
func (req *Request) SetQuery(params Params) *Request {
	req.Params.Update(params)
	return req
}

// SetContent sets bytes payload for the HTTP request.
func (req *Request) SetContent(content []byte) *Request {
	return req.SetBody(bytes.NewBuffer(content))
}

// SetText sets plain text payload for the HTTP request.
func (req *Request) SetText(text string) *Request {
	req.SetContentType("text/plain; charset=utf-8")
	return req.SetBody(bytes.NewBufferString(text))
}

// SetForm sets form payload for the HTTP request.
func (req *Request) SetForm(form Form) *Request {
	req.Form.Update(form)
	return req
}

// SetJSON sets JSON payload for the HTTP request.
func (req *Request) SetJSON(data interface{}, escapeHTML bool) *Request {
	b, err := jsonMarshal(data, "", "", escapeHTML)
	if err != nil {
		req.raiseError("SetJSON", err)
		return req
	}

	req.SetContentType("application/json")
	return req.SetBody(bytes.NewReader(b))
}

// SetXML sets XML payload for the HTTP request.
func (req *Request) SetXML(data interface{}) *Request {
	b, err := xml.Marshal(data)
	if err != nil {
		req.raiseError("SetXML", err)
		return req
	}

	req.SetContentType("application/xml")
	return req.SetBody(bytes.NewReader(b))
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func setMultipartFiles(mw *multipart.Writer, files Files) error {
	const (
		fileFormat = `form-data; name="%s"; filename="%s"`
	)

	var (
		part io.Writer
		err  error
	)
	for k, v := range files {
		filename := v.Filename
		if filename == "" {
			return fmt.Errorf("filename of [%s] not specified", k)
		}

		r := bufio.NewReader(v)
		cType := v.MIME
		if cType == "" {
			data, _ := r.Peek(512)
			cType = http.DetectContentType(data)
		}

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(fileFormat, escapeQuotes(k), escapeQuotes(filename)))
		h.Set("Content-Type", cType)
		part, err = mw.CreatePart(h)
		if err != nil {
			return err
		}

		_, err = io.Copy(part, r)
		if err != nil {
			return err
		}

		v.Close()
	}

	return nil
}

func setMultipartForm(mw *multipart.Writer, form Form) {
	for k, vs := range form.Decode() {
		for _, v := range vs {
			mw.WriteField(k, v)
		}
	}
}

// SetMultipart sets multipart payload for the HTTP request.
// Notes: SetMultipart does not support retry since it's unable to read a stream twice.
func (req *Request) SetMultipart(files Files, form Form) *Request {
	req.errBackground = make(chan error, 1)
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer mw.Close()

		err := setMultipartFiles(mw, files)
		if err != nil {
			req.errBackground <- &RequestError{
				Cause: "SetMultipart",
				Err:   err,
			}
			return
		}

		if form != nil {
			setMultipartForm(mw, form)
		}
	}()

	req.SetContentType(mw.FormDataContentType())
	return req.SetBody(pr)
}

// SetCookies sets cookies for the HTTP request.
func (req *Request) SetCookies(cookies Cookies) *Request {
	req.Cookies.Update(cookies)
	return req
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// SetBasicAuth sets basic authentication for the HTTP request.
func (req *Request) SetBasicAuth(username string, password string) *Request {
	req.Headers.Set("Authorization", "Basic "+basicAuth(username, password))
	return req
}

// SetBearerToken sets bearer token for the HTTP request.
func (req *Request) SetBearerToken(token string) *Request {
	req.Headers.Set("Authorization", "Bearer "+token)
	return req
}

// SetContext sets context for the HTTP request.
func (req *Request) SetContext(ctx context.Context) *Request {
	req.ctx = ctx
	return req
}

// WithBody sets body for the HTTP request.
// Notes: WithBody may not support retry since it's unable to read a stream twice.
func WithBody(body io.Reader) RequestOption {
	return func(req *Request) {
		req.SetBody(body)
	}
}

// WithHost sets host for the HTTP request.
func WithHost(host string) RequestOption {
	return func(req *Request) {
		req.SetHost(host)
	}
}

// WithHeaders sets headers for the HTTP request.
func WithHeaders(headers Headers) RequestOption {
	return func(req *Request) {
		req.SetHeaders(headers)
	}
}

// WithContentType sets Content-Type header value for the HTTP request.
func WithContentType(contentType string) RequestOption {
	return func(req *Request) {
		req.SetContentType(contentType)
	}
}

// WithUserAgent sets User-Agent header value for the HTTP request.
func WithUserAgent(userAgent string) RequestOption {
	return func(req *Request) {
		req.SetUserAgent(userAgent)
	}
}

// WithReferer sets Referer header value for the HTTP request.
func WithReferer(referer string) RequestOption {
	return func(req *Request) {
		req.SetReferer(referer)
	}
}

// WithQuery sets query parameters for the HTTP request.
func WithQuery(params Params) RequestOption {
	return func(req *Request) {
		req.SetQuery(params)
	}
}

// WithContent sets bytes payload for the HTTP request.
func WithContent(content []byte) RequestOption {
	return func(req *Request) {
		req.SetContent(content)
	}
}

// WithText sets plain text payload for the HTTP request.
func WithText(text string) RequestOption {
	return func(req *Request) {
		req.SetText(text)
	}
}

// WithForm sets form payload for the HTTP request.
func WithForm(form Form) RequestOption {
	return func(req *Request) {
		req.SetForm(form)
	}
}

// WithJSON sets JSON payload for the HTTP request.
func WithJSON(data interface{}, escapeHTML bool) RequestOption {
	return func(req *Request) {
		req.SetJSON(data, escapeHTML)
	}
}

// WithXML sets XML payload for the HTTP request.
func WithXML(data interface{}) RequestOption {
	return func(req *Request) {
		req.SetXML(data)
	}
}

// WithMultipart sets multipart payload for the HTTP request.
// Notes: WithMultipart does not support retry since it's unable to read a stream twice.
func WithMultipart(files Files, form Form) RequestOption {
	return func(req *Request) {
		req.SetMultipart(files, form)
	}
}

// WithCookies appends cookies for the HTTP request.
func WithCookies(cookies Cookies) RequestOption {
	return func(req *Request) {
		req.SetCookies(cookies)
	}
}

// WithBasicAuth sets basic authentication for the HTTP request.
func WithBasicAuth(username string, password string) RequestOption {
	return func(req *Request) {
		req.SetBasicAuth(username, password)
	}
}

// WithBearerToken sets bearer token for the HTTP request.
func WithBearerToken(token string) RequestOption {
	return func(req *Request) {
		req.SetBearerToken(token)
	}
}

// WithContext sets context for the HTTP request.
func WithContext(ctx context.Context) RequestOption {
	return func(req *Request) {
		req.SetContext(ctx)
	}
}
