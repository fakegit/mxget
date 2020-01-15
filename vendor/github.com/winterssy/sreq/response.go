package sreq

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/text/encoding"
)

type (
	// Response wraps the raw HTTP response.
	Response struct {
		RawResponse *http.Response

		err  error
		body []byte
	}

	// AfterResponseHook specifies an after response hook.
	AfterResponseHook func(resp *Response)
)

// RaiseError is used to make sreq consider resp as an error response.
func (resp *Response) RaiseError(err error) {
	resp.err = err
}

// Error returns resp's potential error.
func (resp *Response) Error() error {
	return resp.err
}

// Raw returns the raw HTTP response.
func (resp *Response) Raw() (*http.Response, error) {
	return resp.RawResponse, resp.err
}

// Content decodes the HTTP response body to bytes.
func (resp *Response) Content() ([]byte, error) {
	if resp.err != nil || resp.body != nil {
		return resp.body, resp.err
	}
	defer resp.RawResponse.Body.Close()

	var err error
	resp.body, err = ioutil.ReadAll(resp.RawResponse.Body)
	return resp.body, err
}

// Text decodes the HTTP response body and returns the text representation of its raw data
// given an optional charset encoding.
func (resp *Response) Text(e ...encoding.Encoding) (string, error) {
	b, err := resp.Content()
	if err != nil || len(e) == 0 {
		return b2s(b), err
	}

	b, err = e[0].NewDecoder().Bytes(b)
	return b2s(b), err
}

// JSON decodes the HTTP response body and unmarshals its JSON-encoded data into v.
// v must be a pointer.
func (resp *Response) JSON(v interface{}) error {
	if resp.err != nil {
		return resp.err
	}

	if resp.body != nil {
		return json.Unmarshal(resp.body, v)
	}

	buf := acquireBuffer()
	tee := io.TeeReader(resp.RawResponse.Body, buf)
	defer func() {
		resp.RawResponse.Body.Close()
		resp.body = buf.Bytes()
		releaseBuffer(buf)
	}()

	return json.NewDecoder(tee).Decode(v)
}

// H decodes the HTTP response body and unmarshals its JSON-encoded data into an H instance.
func (resp *Response) H() (H, error) {
	h := make(H)
	return h, resp.JSON(&h)
}

// XML decodes the HTTP response body and unmarshals its XML-encoded data into v.
func (resp *Response) XML(v interface{}) error {
	if resp.err != nil {
		return resp.err
	}

	if resp.body != nil {
		return xml.Unmarshal(resp.body, v)
	}

	buf := acquireBuffer()
	tee := io.TeeReader(resp.RawResponse.Body, buf)
	defer func() {
		resp.RawResponse.Body.Close()
		resp.body = buf.Bytes()
		releaseBuffer(buf)
	}()

	return xml.NewDecoder(tee).Decode(v)
}

// Cookies returns the HTTP response cookies.
func (resp *Response) Cookies() ([]*http.Cookie, error) {
	if resp.err != nil {
		return nil, resp.err
	}

	return resp.RawResponse.Cookies(), nil
}

// Cookie returns the HTTP response named cookie.
func (resp *Response) Cookie(name string) (*http.Cookie, error) {
	cookies, err := resp.Cookies()
	if err != nil {
		return nil, err
	}

	for _, c := range cookies {
		if c.Name == name {
			return c, nil
		}
	}

	return nil, ErrNoCookie
}

// EnsureStatusOk ensures the HTTP response's status code must be 200.
func (resp *Response) EnsureStatusOk() *Response {
	return resp.EnsureStatus(http.StatusOK)
}

// EnsureStatus2xx ensures the HTTP response's status code must be 2xx.
func (resp *Response) EnsureStatus2xx() *Response {
	if resp.err != nil {
		return resp
	}

	if resp.RawResponse.StatusCode/100 != 2 {
		resp.RaiseError(fmt.Errorf("sreq: response status code expected to be 2xx, but got %d",
			resp.RawResponse.StatusCode))
	}
	return resp
}

// EnsureStatus ensures the HTTP response's status code must be code.
func (resp *Response) EnsureStatus(code int) *Response {
	if resp.err != nil {
		return resp
	}

	if resp.RawResponse.StatusCode != code {
		resp.RaiseError(fmt.Errorf("sreq: response status code expected to be %d, but got %d",
			code, resp.RawResponse.StatusCode))
	}
	return resp
}

// Save saves the HTTP response into a file.
// Note: Save won't cache the HTTP response body for reuse.
func (resp *Response) Save(filename string, perm os.FileMode) error {
	if resp.err != nil {
		return resp.err
	}

	if resp.body != nil {
		return ioutil.WriteFile(filename, resp.body, perm)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	defer resp.RawResponse.Body.Close()

	_, err = io.Copy(file, resp.RawResponse.Body)
	return err
}

// Verbose makes the HTTP request and its response more talkative.
// It's similar to "curl -v", used for debug.
// Note: Verbose won't cache the HTTP response body for reuse.
func (resp *Response) Verbose(w io.Writer, withBody bool) (err error) {
	if resp.err != nil {
		return resp.err
	}

	err = dumpRequest(resp.RawResponse.Request, w, withBody)

	fmt.Fprintf(w, "< %s %s\r\n", resp.RawResponse.Proto, resp.RawResponse.Status)
	for k, vs := range resp.RawResponse.Header {
		for _, v := range vs {
			fmt.Fprintf(w, "< %s: %s\r\n", k, v)
		}
	}
	io.WriteString(w, "<\r\n")

	if !withBody || resp.RawResponse.ContentLength == 0 {
		return
	}

	if resp.body != nil {
		fmt.Fprintf(w, "%s\r\n", b2s(resp.body))
		return
	}

	defer resp.RawResponse.Body.Close()
	_, err = io.Copy(w, resp.RawResponse.Body)

	io.WriteString(w, "\r\n")
	return
}
