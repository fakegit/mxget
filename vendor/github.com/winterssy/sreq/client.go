package sreq

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
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

	// ErrUnexpectedTransport can be returned if assert a RoundTripper as an *http.Transport failed.
	ErrUnexpectedTransport = errors.New("sreq: current transport is not an *http.Transport instance")
)

type (
	// Client wraps the raw HTTP client.
	Client struct {
		RawClient *http.Client
	}
)

func init() {
	DefaultCookieJar, _ = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	DefaultClient = New()
}

// New returns a new sreq client.
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

func (c *Client) transport() (*http.Transport, error) {
	t, ok := c.RawClient.Transport.(*http.Transport)
	if !ok {
		return nil, ErrUnexpectedTransport
	}
	if t == nil {
		return t, errors.New("sreq: nil transport")
	}

	return t, nil
}

// SetTransport sets transport of the HTTP client.
func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	c.RawClient.Transport = transport
	return c
}

// SetRedirectPolicy sets policy of the HTTP client for handling redirects.
func (c *Client) SetRedirectPolicy(policy func(req *http.Request, via []*http.Request) error) *Client {
	c.RawClient.CheckRedirect = policy
	return c
}

// SetCookieJar sets cookie jar of the HTTP client.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.RawClient.Jar = jar
	return c
}

// SetTimeout sets timeout of the HTTP client.
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.RawClient.Timeout = timeout
	return c
}

// DisableRedirect makes the HTTP client not follow redirects.
func (c *Client) DisableRedirect() *Client {
	policy := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c.SetRedirectPolicy(policy)
}

// SetProxy sets proxy of the HTTP client.
func (c *Client) SetProxy(proxy func(*http.Request) (*stdurl.URL, error)) *Client {
	t, err := c.transport()
	if err != nil {
		log.Print(err)
		return c
	}

	t.Proxy = proxy
	c.RawClient.Transport = t
	return c
}

// ProxyFromURL sets proxy of the HTTP client from a url.
func (c *Client) ProxyFromURL(url string) *Client {
	fixedURL, err := stdurl.Parse(url)
	if err != nil {
		log.Print(err)
		return c
	}
	return c.SetProxy(http.ProxyURL(fixedURL))
}

// DisableProxy makes the HTTP client not use proxy.
func (c *Client) DisableProxy() *Client {
	return c.SetProxy(nil)
}

// SetTLSClientConfig sets TLS configuration of the HTTP client.
func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	t, err := c.transport()
	if err != nil {
		log.Print(err)
		return c
	}

	t.TLSClientConfig = config
	c.RawClient.Transport = t
	return c
}

// AppendClientCertificates appends client certificates to the HTTP client.
func (c *Client) AppendClientCertificates(certs ...tls.Certificate) *Client {
	t, err := c.transport()
	if err != nil {
		log.Print(err)
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
func (c *Client) AppendRootCAs(pemFilePath string) *Client {
	t, err := c.transport()
	if err != nil {
		log.Print(err)
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
		log.Print(err)
		return c
	}

	t.TLSClientConfig.RootCAs.AppendCertsFromPEM(pemCerts)
	c.RawClient.Transport = t
	return c
}

// DisableVerify makes the HTTP client not verify the server's TLS certificate.
func (c *Client) DisableVerify() *Client {
	t, err := c.transport()
	if err != nil {
		log.Print(err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}

	t.TLSClientConfig.InsecureSkipVerify = true
	c.RawClient.Transport = t
	return c
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

// Do sends a raw HTTP request and returns its response.
func Do(req *http.Request) *Response {
	return DefaultClient.Do(req)
}

// Do sends a raw HTTP request and returns its response.
func (c *Client) Do(req *http.Request) *Response {
	rawResponse, err := c.RawClient.Do(req)
	return &Response{
		RawResponse: rawResponse,
		Err:         err,
	}
}
