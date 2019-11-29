package sreq

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	stdurl "net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	// DefaultTimeout is the timeout that sreq uses for the global client.
	DefaultTimeout = 120 * time.Second
)

var (
	// ErrUnexpectedTransport can be returned if assert a RoundTripper as an *http.Transport failed.
	ErrUnexpectedTransport = errors.New("sreq: current transport is not an *http.Transport instance")

	gClient *Client
	gJar    http.CookieJar
)

type (
	// Client wraps the raw HTTP client.
	Client struct {
		RawClient *http.Client
	}

	// ClientOption specifies the client options.
	ClientOption func(*Client) (*Client, error)
)

func init() {
	gJar, _ = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	gClient, _ = New(nil,
		EnableSession(),
		WithTimeout(DefaultTimeout),
	)
}

// New allows you to customize a sreq client via a RoundTripper.
// If the transport not specified, sreq would use defaults.
func New(transport http.RoundTripper, opts ...ClientOption) (*Client, error) {
	if transport == nil {
		transport = DefaultTransport()
	}

	rawClient := &http.Client{
		Transport: transport,
	}
	client := &Client{
		RawClient: rawClient,
	}

	var err error
	for _, opt := range opts {
		client, err = opt(client)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// FilterCookies returns the cookies to send in a request for the given URL.
func FilterCookies(url string) ([]*http.Cookie, error) {
	return gClient.FilterCookies(url)
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
	return gClient.FilterCookie(url, name)
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

// Do sends a raw HTTP request and returns its response.
func Do(rawRequest *http.Request) *Response {
	return gClient.Do(rawRequest)
}

// Do sends a raw HTTP request and returns its response.
func (c *Client) Do(rawRequest *http.Request) *Response {
	rawResponse, err := c.RawClient.Do(rawRequest)
	return &Response{
		RawResponse: rawResponse,
		Err:         err,
	}
}

// WithRedirectPolicy sets policy of the HTTP client for handling redirects.
func WithRedirectPolicy(policy func(req *http.Request, via []*http.Request) error) ClientOption {
	return func(c *Client) (*Client, error) {
		c.RawClient.CheckRedirect = policy
		return c, nil
	}
}

// WithCookieJar sets cookie jar of the HTTP client.
func WithCookieJar(jar http.CookieJar) ClientOption {
	return func(c *Client) (*Client, error) {
		c.RawClient.Jar = jar
		return c, nil
	}
}

// WithTimeout sets timeout of the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) (*Client, error) {
		c.RawClient.Timeout = timeout
		return c, nil
	}
}

// EnableSession enables session support of the HTTP client by using the global cookie jar.
// It would be useful if you want to keep session across the lifecycle of sreq.
func EnableSession() ClientOption {
	return WithCookieJar(gJar)
}

// DisableRedirect makes the HTTP client not follow redirects.
func DisableRedirect() ClientOption {
	policy := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return WithRedirectPolicy(policy)
}

// WithProxy sets proxy of the HTTP client.
func WithProxy(proxy func(*http.Request) (*stdurl.URL, error)) ClientOption {
	return func(c *Client) (*Client, error) {
		transport, ok := c.RawClient.Transport.(*http.Transport)
		if !ok {
			return nil, ErrUnexpectedTransport
		}

		transport.Proxy = proxy
		c.RawClient.Transport = transport
		return c, nil
	}
}

// ProxyFromEnvironment sets proxy of the HTTP client from the environment variables.
// This is the default behavior of sreq if you do not specify the transport manually.
func ProxyFromEnvironment() ClientOption {
	return WithProxy(http.ProxyFromEnvironment)
}

// ProxyFromURL sets proxy of the HTTP client from a url.
func ProxyFromURL(url string) ClientOption {
	fixedURL, err := stdurl.Parse(url)
	if err != nil {
		return func(c *Client) (*Client, error) {
			return nil, err
		}
	}
	return WithProxy(http.ProxyURL(fixedURL))
}

// DisableProxy makes the HTTP client not use proxy.
func DisableProxy() ClientOption {
	return WithProxy(nil)
}

// WithTLSClientConfig sets TLS configuration of the HTTP client.
func WithTLSClientConfig(config *tls.Config) ClientOption {
	return func(c *Client) (*Client, error) {
		transport, ok := c.RawClient.Transport.(*http.Transport)
		if !ok {
			return nil, ErrUnexpectedTransport
		}

		transport.TLSClientConfig = config
		c.RawClient.Transport = transport
		return c, nil
	}
}

// WithClientCertificates appends client certificates to the HTTP client.
func WithClientCertificates(certs ...tls.Certificate) ClientOption {
	return func(c *Client) (*Client, error) {
		transport, ok := c.RawClient.Transport.(*http.Transport)
		if !ok {
			return nil, ErrUnexpectedTransport
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}

		transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, certs...)
		c.RawClient.Transport = transport
		return c, nil
	}
}

// WithRootCA appends root certificate authorities to the HTTP client.
func WithRootCA(pemFilePath string) ClientOption {
	return func(c *Client) (*Client, error) {
		pemCerts, err := ioutil.ReadFile(pemFilePath)
		if err != nil {
			return nil, err
		}

		transport, ok := c.RawClient.Transport.(*http.Transport)
		if !ok {
			return nil, ErrUnexpectedTransport
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		if transport.TLSClientConfig.RootCAs == nil {
			transport.TLSClientConfig.RootCAs = x509.NewCertPool()
		}

		transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(pemCerts)
		c.RawClient.Transport = transport
		return c, nil
	}
}

// DisableVerify makes the HTTP client not verify the server's TLS certificate.
func DisableVerify() ClientOption {
	return func(c *Client) (*Client, error) {
		transport, ok := c.RawClient.Transport.(*http.Transport)
		if !ok {
			return nil, ErrUnexpectedTransport
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}

		transport.TLSClientConfig.InsecureSkipVerify = true
		c.RawClient.Transport = transport
		return c, nil
	}
}
