package sreq

import (
	"errors"
	"fmt"
)

var (
	// ErrUnexpectedTransport can be used if assert a RoundTripper as a non-nil *http.Transport instance failed.
	ErrUnexpectedTransport = errors.New("current transport isn't a non-nil *http.Transport instance")

	// ErrNilCookieJar can be used when the cookie jar is nil.
	ErrNilCookieJar = errors.New("nil cookie jar")

	// ErrNoCookie can be used when a cookie not found in the HTTP response or cookie jar.
	ErrNoCookie = errors.New("sreq: named cookie not present")
)

type (
	// ClientError records a client error, can be used when sreq builds Client failed.
	ClientError struct {
		Cause string
		Err   error
	}

	// RequestError records a request error, can be used when sreq builds Request failed.
	RequestError struct {
		Cause string
		Err   error
	}
)

// Error implements error interface.
func (c *ClientError) Error() string {
	return fmt.Sprintf("sreq [Client] [%s]: %s", c.Cause, c.Err.Error())
}

// Unwrap unpacks and returns the wrapped err of c.
func (c *ClientError) Unwrap() error {
	return c.Err
}

// Error implements error interface.
func (req *RequestError) Error() string {
	return fmt.Sprintf("sreq [Request] [%s]: %s", req.Cause, req.Err.Error())
}

// Unwrap unpacks and returns the wrapped err of req.
func (req *RequestError) Unwrap() error {
	return req.Err
}
