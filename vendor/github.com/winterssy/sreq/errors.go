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
	ErrNoCookie = errors.New("named cookie not present")
)

type (
	// Error records an error with more details and makes it more readable.
	Error struct {
		Op  string
		Err error
	}
)

// Error implements error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("sreq [%s]: %s", e.Op, e.Err.Error())
}

// Unwrap unpacks and returns the wrapped err of e.
func (e *Error) Unwrap() error {
	return e.Err
}
