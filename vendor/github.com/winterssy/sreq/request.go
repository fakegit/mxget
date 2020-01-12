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
		Body       io.Reader
		Host       string
		Headers    Headers
		Params     Params
		Form       Form
		Cookies    Cookies
		Retry      *Retry

		ctx           context.Context
		errBackground chan error
	}

	// RequestOption provides a convenient way to setup Request.
	RequestOption func(req *Request) error

	// BeforeRequestHook is alike to RequestOption, but for specifying a before request hook.
	// Return a non-nil error to prevent requests.
	BeforeRequestHook func(req *Request) error
)

// NewRequest returns a new Request given a method, URL and optional request options.
func NewRequest(method string, url string, opts ...RequestOption) (*Request, error) {
	rawRequest, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req := &Request{
		RawRequest: rawRequest,
		Headers:    make(Headers),
		Params:     make(Params),
		Form:       make(Form),
		Cookies:    make(Cookies),
	}
	for _, opt := range opts {
		if err = opt(req); err != nil {
			break
		}
	}
	return req, err
}

func (req *Request) bindHTTPRequest() {
	req.bindHost()
	req.bindQuery()
	req.bindForm()
	req.bindHeaders()
	req.bindCookies()
	req.bindBody() // must after bindForm
}

func (req *Request) bindHost() {
	req.RawRequest.Host = req.Host
}

func (req *Request) bindHeaders() {
	req.Headers.SetDefault("User-Agent", defaultUserAgent)
	for k, vs := range req.Headers.Decode() {
		req.RawRequest.Header.Del(k) // remove existing value
		for _, v := range vs {
			req.RawRequest.Header.Add(k, v)
		}
	}
}

func (req *Request) bindQuery() {
	if len(req.Params) == 0 {
		return
	}

	for k, v := range req.RawRequest.URL.Query() {
		req.Params.SetDefault(k, v)
	}
	req.RawRequest.URL.RawQuery = req.Params.Encode(true)
}

func (req *Request) bindForm() {
	if len(req.Form) == 0 {
		return
	}

	req.SetContentType("application/x-www-form-urlencoded")
	req.SetBody(strings.NewReader(req.Form.Encode(true)))
}

func (req *Request) bindCookies() {
	if len(req.Cookies) == 0 {
		return
	}

	for _, c := range req.Cookies.Decode() {
		req.RawRequest.AddCookie(c)
	}
}

func (req *Request) bindBody() {
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
// Note: SetBody may not support retry since it's unable to read a stream twice.
func (req *Request) SetBody(body io.Reader) {
	req.Body = body
}

// SetHost sets host for the HTTP request.
func (req *Request) SetHost(host string) {
	req.Host = host
}

// SetHeaders sets headers for the HTTP request.
func (req *Request) SetHeaders(headers Headers) {
	req.Headers.Update(headers)
}

// SetContentType sets Content-Type header value for the HTTP request.
func (req *Request) SetContentType(contentType string) {
	req.Headers.Set("Content-Type", contentType)
}

// SetUserAgent sets User-Agent header value for the HTTP request.
func (req *Request) SetUserAgent(userAgent string) {
	req.Headers.Set("User-Agent", userAgent)
}

// SetReferer sets Referer header value for the HTTP request.
func (req *Request) SetReferer(referer string) {
	req.Headers.Set("Referer", referer)
}

// SetQuery sets query parameters for the HTTP request.
func (req *Request) SetQuery(params Params) {
	req.Params.Update(params)
}

// SetContent sets bytes payload for the HTTP request.
func (req *Request) SetContent(content []byte) {
	req.SetBody(bytes.NewBuffer(content))
}

// SetText sets plain text payload for the HTTP request.
func (req *Request) SetText(text string) {
	req.SetContentType("text/plain; charset=utf-8")
	req.SetBody(bytes.NewBufferString(text))
}

// SetForm sets form payload for the HTTP request.
func (req *Request) SetForm(form Form) {
	req.Form.Update(form)
}

// SetJSON sets JSON payload for the HTTP request.
func (req *Request) SetJSON(data interface{}, escapeHTML bool) error {
	b, err := jsonMarshal(data, "", "", escapeHTML)
	if err != nil {
		return err
	}

	req.SetContentType("application/json")
	req.SetBody(bytes.NewReader(b))
	return nil
}

// SetXML sets XML payload for the HTTP request.
func (req *Request) SetXML(data interface{}) error {
	b, err := xml.Marshal(data)
	if err != nil {
		return err
	}

	req.SetContentType("application/xml")
	req.SetBody(bytes.NewReader(b))
	return nil
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
// Note: SetMultipart does not support retry since it's unable to read a stream twice.
func (req *Request) SetMultipart(files Files, form Form) {
	req.errBackground = make(chan error, 1)
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer mw.Close()

		err := setMultipartFiles(mw, files)
		if err != nil {
			req.errBackground <- err
			return
		}

		if form != nil {
			setMultipartForm(mw, form)
		}
	}()

	req.SetContentType(mw.FormDataContentType())
	req.SetBody(pr)
}

// SetCookies sets cookies for the HTTP request.
func (req *Request) SetCookies(cookies Cookies) {
	req.Cookies.Update(cookies)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// SetBasicAuth sets basic authentication for the HTTP request.
func (req *Request) SetBasicAuth(username string, password string) {
	req.Headers.Set("Authorization", "Basic "+basicAuth(username, password))
}

// SetBearerToken sets bearer token for the HTTP request.
func (req *Request) SetBearerToken(token string) {
	req.Headers.Set("Authorization", "Bearer "+token)
}

// SetContext sets context for the HTTP request.
func (req *Request) SetContext(ctx context.Context) {
	req.ctx = ctx
}

// SetRetry specifies the retry policy for handling retries.
func (req *Request) SetRetry(retry *Retry) {
	req.Retry = retry
}

// WithBody is a request option to set body for the HTTP request.
// Note: WithBody may not support retry since it's unable to read a stream twice.
func WithBody(body io.Reader) RequestOption {
	return func(req *Request) error {
		req.SetBody(body)
		return nil
	}
}

// WithHost is a request option to set host for the HTTP request.
func WithHost(host string) RequestOption {
	return func(req *Request) error {
		req.SetHost(host)
		return nil
	}
}

// WithHeaders is a request option to set headers for the HTTP request.
func WithHeaders(headers Headers) RequestOption {
	return func(req *Request) error {
		req.SetHeaders(headers)
		return nil
	}
}

// WithContentType is a request option to set Content-Type header value for the HTTP request.
func WithContentType(contentType string) RequestOption {
	return func(req *Request) error {
		req.SetContentType(contentType)
		return nil
	}
}

// WithUserAgent is a request option to set User-Agent header value for the HTTP request.
func WithUserAgent(userAgent string) RequestOption {
	return func(req *Request) error {
		req.SetUserAgent(userAgent)
		return nil
	}
}

// WithReferer is a request option to set Referer header value for the HTTP request.
func WithReferer(referer string) RequestOption {
	return func(req *Request) error {
		req.SetReferer(referer)
		return nil
	}
}

// WithQuery is a request option to set query parameters for the HTTP request.
func WithQuery(params Params) RequestOption {
	return func(req *Request) error {
		req.SetQuery(params)
		return nil
	}
}

// WithContent is a request option to set bytes payload for the HTTP request.
func WithContent(content []byte) RequestOption {
	return func(req *Request) error {
		req.SetContent(content)
		return nil
	}
}

// WithText is a request option to set plain text payload for the HTTP request.
func WithText(text string) RequestOption {
	return func(req *Request) error {
		req.SetText(text)
		return nil
	}
}

// WithForm is a request option to set form payload for the HTTP request.
func WithForm(form Form) RequestOption {
	return func(req *Request) error {
		req.SetForm(form)
		return nil
	}
}

// WithJSON is a request option to set JSON payload for the HTTP request.
func WithJSON(data interface{}, escapeHTML bool) RequestOption {
	return func(req *Request) error {
		return req.SetJSON(data, escapeHTML)
	}
}

// WithXML is a request option to set XML payload for the HTTP request.
func WithXML(data interface{}) RequestOption {
	return func(req *Request) error {
		return req.SetXML(data)
	}
}

// WithMultipart is a request option sets multipart payload for the HTTP request.
// Note: WithMultipart does not support retry since it's unable to read a stream twice.
func WithMultipart(files Files, form Form) RequestOption {
	return func(req *Request) error {
		req.SetMultipart(files, form)
		return nil
	}
}

// WithCookies is a request option to set cookies for the HTTP request.
func WithCookies(cookies Cookies) RequestOption {
	return func(req *Request) error {
		req.SetCookies(cookies)
		return nil
	}
}

// WithBasicAuth is a request option to set basic authentication for the HTTP request.
func WithBasicAuth(username string, password string) RequestOption {
	return func(req *Request) error {
		req.SetBasicAuth(username, password)
		return nil
	}
}

// WithBearerToken is a request option to set bearer token for the HTTP request.
func WithBearerToken(token string) RequestOption {
	return func(req *Request) error {
		req.SetBearerToken(token)
		return nil
	}
}

// WithContext is a request option to set context for the HTTP request.
func WithContext(ctx context.Context) RequestOption {
	return func(req *Request) error {
		req.SetContext(ctx)
		return nil
	}
}

// WithRetry is a request option to specify the retry policy for handling retries.
func WithRetry(retry *Retry) RequestOption {
	return func(req *Request) error {
		req.SetRetry(retry)
		return nil
	}
}
