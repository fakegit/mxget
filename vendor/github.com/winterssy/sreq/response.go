package sreq

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type (
	// Response wraps the raw HTTP response.
	Response struct {
		RawResponse *http.Response
		Err         error

		body []byte
	}
)

// Resolve resolves r and returns its raw HTTP response.
func (r *Response) Resolve() (*http.Response, error) {
	return r.RawResponse, r.Err
}

// Raw decodes the HTTP response body of r and returns its raw data.
func (r *Response) Raw() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}

	if r.Err != nil {
		return nil, r.Err
	}
	defer r.RawResponse.Body.Close()

	var err error
	r.body, err = ioutil.ReadAll(r.RawResponse.Body)
	return r.body, err
}

// Text decodes the HTTP response body of r and returns the text representation of its raw data.
func (r *Response) Text() (string, error) {
	b, err := r.Raw()
	return string(b), err
}

// JSON decodes the HTTP response body of r and unmarshals its JSON-encoded data into v.
func (r *Response) JSON(v interface{}) error {
	if r.body != nil {
		return json.Unmarshal(r.body, v)
	}

	if r.Err != nil {
		return r.Err
	}
	defer r.RawResponse.Body.Close()

	return json.NewDecoder(r.RawResponse.Body).Decode(v)
}

// Cookies returns the HTTP response cookies.
func (r *Response) Cookies() ([]*http.Cookie, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	cookies := r.RawResponse.Cookies()
	if len(cookies) == 0 {
		return nil, errors.New("sreq: cookies not present")
	}

	return cookies, nil
}

// Cookie returns the HTTP response cookie by name.
func (r *Response) Cookie(name string) (*http.Cookie, error) {
	cookies, err := r.Cookies()
	if err != nil {
		return nil, err
	}

	for _, c := range cookies {
		if c.Name == name {
			return c, nil
		}
	}

	return nil, errors.New("sreq: named cookie not present")
}

// EnsureStatusOk ensures the HTTP response's status code of r must be 200.
func (r *Response) EnsureStatusOk() *Response {
	return r.EnsureStatus(http.StatusOK)
}

// EnsureStatus2xx ensures the HTTP response's status code of r must be 2xx.
func (r *Response) EnsureStatus2xx() *Response {
	if r.Err != nil {
		return r
	}
	if r.RawResponse.StatusCode/100 != 2 {
		r.Err = fmt.Errorf("sreq: bad status: %d", r.RawResponse.StatusCode)
	}
	return r
}

// EnsureStatus ensures the HTTP response's status code of r must be the code parameter.
func (r *Response) EnsureStatus(code int) *Response {
	if r.Err != nil {
		return r
	}
	if r.RawResponse.StatusCode != code {
		r.Err = fmt.Errorf("sreq: bad status: %d", r.RawResponse.StatusCode)
	}
	return r
}

// Save saves the HTTP response into a file.
func (r *Response) Save(filename string, perm os.FileMode) error {
	if r.body != nil {
		return ioutil.WriteFile(filename, r.body, perm)
	}

	if r.Err != nil {
		return r.Err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	defer r.RawResponse.Body.Close()

	_, err = io.Copy(file, r.RawResponse.Body)
	return err
}

// Verbose makes the HTTP request and its response more talkative.
// It's similar to "curl -v", used for debug.
func (r *Response) Verbose(w io.Writer) error {
	if r.Err != nil {
		return r.Err
	}

	req := r.RawResponse.Request
	fmt.Fprintf(w, "> %s %s %s\r\n", req.Method, req.URL.RequestURI(), req.Proto)
	fmt.Fprintf(w, "> Host: %s\r\n", req.Host)
	for k := range req.Header {
		fmt.Fprintf(w, "> %s: %s\r\n", k, req.Header.Get(k))
	}
	fmt.Fprint(w, ">\r\n")

	if req.GetBody != nil && req.ContentLength != 0 {
		rc, err := req.GetBody()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}

		fmt.Fprint(w, "\r\n")
	}

	resp := r.RawResponse
	fmt.Fprintf(w, "< %s %s\r\n", resp.Proto, resp.Status)
	for k := range resp.Header {
		fmt.Fprintf(w, "< %s: %s\r\n", k, resp.Header.Get(k))
	}
	fmt.Fprint(w, "<\r\n")

	if r.body != nil {
		fmt.Fprint(w, string(r.body))
		return nil
	}

	defer resp.Body.Close()
	_, err := io.Copy(w, resp.Body)
	return err
}
