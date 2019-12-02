package sreq

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	stdurl "net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	// DefaultTimeout is the timeout used by DefaultClient.
	DefaultTimeout = 120 * time.Second
)

var (
	// DefaultClient is the default sreq Client.
	DefaultClient *Client

	// DefaultCookieJar is the cookie jar used by DefaultClient.
	DefaultCookieJar http.CookieJar

	errUnexpectedTransport = errors.New("current transport isn't a non-nil *http.Transport instance")
)

type (
	// Client wraps the raw HTTP client.
	Client struct {
		RawClient *http.Client
		Err       error
	}
)

func init() {
	DefaultCookieJar, _ = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	DefaultClient = New()
}

// New returns a new Client.
// It's a clone of DefaultClient indeed.
func New() *Client {
	rawClient := &http.Client{}
	client := &Client{
		RawClient: rawClient,
	}

	client.SetTransport(DefaultTransport()).
		SetCookieJar(DefaultCookieJar).
		SetTimeout(DefaultTimeout)
	return client
}

func (c *Client) httpTransport() (*http.Transport, error) {
	t, ok := c.RawClient.Transport.(*http.Transport)
	if !ok || t == nil {
		return nil, errUnexpectedTransport
	}

	return t, nil
}

func (c *Client) raiseError(cause string, err error) {
	if c.Err != nil {
		return
	}

	c.Err = fmt.Errorf("sreq [%s]: %s", cause, err.Error())
}

// Resolve resolves c and returns its raw HTTP client.
func (c *Client) Resolve() (*http.Client, error) {
	return c.RawClient, c.Err
}

// SetTransport sets transport of the HTTP client.
func SetTransport(transport http.RoundTripper) *Client {
	return DefaultClient.SetTransport(transport)
}

// SetTransport sets transport of the HTTP client.
func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	c.RawClient.Transport = transport
	return c
}

// SetRedirectPolicy sets policy of the HTTP client for handling redirects.
func SetRedirectPolicy(policy func(req *http.Request, via []*http.Request) error) *Client {
	return DefaultClient.SetRedirectPolicy(policy)
}

// SetRedirectPolicy sets policy of the HTTP client for handling redirects.
func (c *Client) SetRedirectPolicy(policy func(req *http.Request, via []*http.Request) error) *Client {
	c.RawClient.CheckRedirect = policy
	return c
}

// DisableRedirect makes the HTTP client not follow redirects.
func DisableRedirect() *Client {
	return DefaultClient.DisableRedirect()
}

// DisableRedirect makes the HTTP client not follow redirects.
func (c *Client) DisableRedirect() *Client {
	policy := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c.SetRedirectPolicy(policy)
}

// SetCookieJar sets cookie jar of the HTTP client.
func SetCookieJar(jar http.CookieJar) *Client {
	return DefaultClient.SetCookieJar(jar)
}

// SetCookieJar sets cookie jar of the HTTP client.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.RawClient.Jar = jar
	return c
}

// DisableSession makes the HTTP client not use cookie jar.
// Only use if you don't want to keep session for the next HTTP request.
func DisableSession() *Client {
	return DefaultClient.DisableSession()
}

// DisableSession makes the HTTP client not use cookie jar.
// Only use if you don't want to keep session for the next HTTP request.
func (c *Client) DisableSession() *Client {
	return c.SetCookieJar(nil)
}

// SetTimeout sets timeout of the HTTP client.
func SetTimeout(timeout time.Duration) *Client {
	return DefaultClient.SetTimeout(timeout)
}

// SetTimeout sets timeout of the HTTP client.
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.RawClient.Timeout = timeout
	return c
}

// SetProxy sets proxy of the HTTP client.
func SetProxy(proxy func(*http.Request) (*stdurl.URL, error)) *Client {
	return DefaultClient.SetProxy(proxy)
}

// SetProxy sets proxy of the HTTP client.
func (c *Client) SetProxy(proxy func(*http.Request) (*stdurl.URL, error)) *Client {
	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("Client.SetProxy", err)
		return c
	}

	t.Proxy = proxy
	c.RawClient.Transport = t
	return c
}

// SetProxyFromURL sets proxy of the HTTP client from a url.
func SetProxyFromURL(url string) *Client {
	return DefaultClient.SetProxyFromURL(url)
}

// SetProxyFromURL sets proxy of the HTTP client from a url.
func (c *Client) SetProxyFromURL(url string) *Client {
	fixedURL, err := stdurl.Parse(url)
	if err != nil {
		c.raiseError("Client.SetProxyFromURL", err)
		return c
	}
	return c.SetProxy(http.ProxyURL(fixedURL))
}

// DisableProxy makes the HTTP client not use proxy.
func DisableProxy() *Client {
	return DefaultClient.DisableProxy()
}

// DisableProxy makes the HTTP client not use proxy.
func (c *Client) DisableProxy() *Client {
	return c.SetProxy(nil)
}

// SetTLSClientConfig sets TLS configuration of the HTTP client.
func SetTLSClientConfig(config *tls.Config) *Client {
	return DefaultClient.SetTLSClientConfig(config)
}

// SetTLSClientConfig sets TLS configuration of the HTTP client.
func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("Client.SetTLSClientConfig", err)
		return c
	}

	t.TLSClientConfig = config
	c.RawClient.Transport = t
	return c
}

// AppendClientCertificates appends client certificates to the HTTP client.
func AppendClientCertificates(certs ...tls.Certificate) *Client {
	return DefaultClient.AppendClientCertificates(certs...)
}

// AppendClientCertificates appends client certificates to the HTTP client.
func (c *Client) AppendClientCertificates(certs ...tls.Certificate) *Client {
	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("Client.AppendClientCertificates", err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}

	t.TLSClientConfig.Certificates = append(t.TLSClientConfig.Certificates, certs...)
	c.RawClient.Transport = t
	return c
}

// AppendRootCAs appends root certificate authorities to the HTTP client.
func AppendRootCAs(pemFilePath string) *Client {
	return DefaultClient.AppendRootCAs(pemFilePath)
}

// AppendRootCAs appends root certificate authorities to the HTTP client.
func (c *Client) AppendRootCAs(pemFilePath string) *Client {
	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("Client.AppendRootCAs", err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	if t.TLSClientConfig.RootCAs == nil {
		t.TLSClientConfig.RootCAs = x509.NewCertPool()
	}

	pemCerts, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		c.raiseError("Client.AppendRootCAs", err)
		return c
	}

	t.TLSClientConfig.RootCAs.AppendCertsFromPEM(pemCerts)
	c.RawClient.Transport = t
	return c
}

// DisableVerify makes the HTTP client not verify the server's TLS certificate.
func DisableVerify() *Client {
	return DefaultClient.DisableVerify()
}

// DisableVerify makes the HTTP client not verify the server's TLS certificate.
func (c *Client) DisableVerify() *Client {
	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("Client.DisableVerify", err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}

	t.TLSClientConfig.InsecureSkipVerify = true
	c.RawClient.Transport = t
	return c
}

// Get makes a GET HTTP request.
func Get(url string, opts ...RequestOption) *Response {
	return DefaultClient.Get(url, opts...)
}

// Get makes a GET HTTP request.
func (c *Client) Get(url string, opts ...RequestOption) *Response {
	return c.Send(MethodGet, url, opts...)
}

// Head makes a HEAD HTTP request.
func Head(url string, opts ...RequestOption) *Response {
	return DefaultClient.Head(url, opts...)
}

// Head makes a HEAD HTTP request.
func (c *Client) Head(url string, opts ...RequestOption) *Response {
	return c.Send(MethodHead, url, opts...)
}

// Post makes a POST HTTP request.
func Post(url string, opts ...RequestOption) *Response {
	return DefaultClient.Post(url, opts...)
}

// Post makes a POST HTTP request.
func (c *Client) Post(url string, opts ...RequestOption) *Response {
	return c.Send(MethodPost, url, opts...)
}

// Put makes a PUT HTTP request.
func Put(url string, opts ...RequestOption) *Response {
	return DefaultClient.Put(url, opts...)
}

// Put makes a PUT HTTP request.
func (c *Client) Put(url string, opts ...RequestOption) *Response {
	return DefaultClient.Send(MethodPut, url, opts...)
}

// Patch makes a PATCH HTTP request.
func Patch(url string, opts ...RequestOption) *Response {
	return DefaultClient.Patch(url, opts...)
}

// Patch makes a PATCH HTTP request.
func (c *Client) Patch(url string, opts ...RequestOption) *Response {
	return c.Send(MethodPatch, url, opts...)
}

// Delete makes a DELETE HTTP request.
func Delete(url string, opts ...RequestOption) *Response {
	return DefaultClient.Delete(url, opts...)
}

// Delete makes a DELETE HTTP request.
func (c *Client) Delete(url string, opts ...RequestOption) *Response {
	return c.Send(MethodDelete, url, opts...)
}

// Send makes an HTTP request using a specified method.
func Send(method string, url string, opts ...RequestOption) *Response {
	return DefaultClient.Send(method, url, opts...)
}

// Send makes an HTTP request using a specified method.
func (c *Client) Send(method string, url string, opts ...RequestOption) *Response {
	req := NewRequest(method, url, nil)
	for _, opt := range opts {
		req = opt(req)
	}
	return c.Do(req)
}

// FilterCookies returns the cookies to send in a request for the given URL.
func FilterCookies(url string) ([]*http.Cookie, error) {
	return DefaultClient.FilterCookies(url)
}

// FilterCookies returns the cookies to send in a request for the given URL.
func (c *Client) FilterCookies(url string) ([]*http.Cookie, error) {
	if c.RawClient.Jar == nil {
		return nil, errors.New("sreq: nil cookie jar")
	}

	u, err := stdurl.Parse(url)
	if err != nil {
		return nil, err
	}
	cookies := c.RawClient.Jar.Cookies(u)
	if len(cookies) == 0 {
		return nil, errors.New("sreq: cookies for the given URL not present")
	}

	return cookies, nil
}

// FilterCookie returns the named cookie to send in a request for the given URL.
func FilterCookie(url string, name string) (*http.Cookie, error) {
	return DefaultClient.FilterCookie(url, name)
}

// FilterCookie returns the named cookie to send in a request for the given URL.
func (c *Client) FilterCookie(url string, name string) (*http.Cookie, error) {
	cookies, err := c.FilterCookies(url)
	if err != nil {
		return nil, err
	}

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}

	return nil, errors.New("sreq: named cookie for the given URL not present")
}

// Do sends a request and returns its response.
func Do(req *Request) *Response {
	return DefaultClient.Do(req)
}

// Do sends a request and returns its  response.
func (c *Client) Do(req *Request) *Response {
	resp := new(Response)

	if c.Err != nil {
		resp.Err = c.Err
		return resp
	}

	if req.Err != nil {
		resp.Err = req.Err
		return resp
	}

	if !req.retry.enable {
		resp.RawResponse, resp.Err = c.RawClient.Do(req.RawRequest)
		return resp
	}

	ctx := req.RawRequest.Context()
	var err error
	for i := req.retry.attempts; i > 0; i-- {
		resp.RawResponse, resp.Err = c.RawClient.Do(req.RawRequest)
		if err = ctx.Err(); err != nil {
			resp.Err = err
			return resp
		}

		shouldRetry := resp.Err != nil
		for _, condition := range req.retry.conditions {
			shouldRetry = condition(resp)
			if shouldRetry {
				break
			}
		}

		if !shouldRetry {
			return resp
		}

		select {
		case <-time.After(req.retry.delay):
		case <-ctx.Done():
			resp.Err = ctx.Err()
			return resp
		}
	}

	return resp
}
