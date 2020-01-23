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
		*http.Response
		content []byte
		err     error
	}

	// AfterResponseHook specifies an after response hook.
	// If the returned error isn't nil, sreq will consider resp as a bad response.
	AfterResponseHook func(resp *Response) error
)

// Error reports resp's potential error.
func (resp *Response) Error() error {
	return resp.err
}

// Raw returns the raw HTTP response.
func (resp *Response) Raw() (*http.Response, error) {
	return resp.Response, resp.err
}

// Prefetch reads from the HTTP response body until an error or EOF and keeps the data in memory for reuse.
func (resp *Response) Prefetch() *Response {
	if resp.err != nil || resp.content != nil {
		return resp
	}
	defer resp.Body.Close()

	resp.content, resp.err = ioutil.ReadAll(resp.Body)
	return resp
}

// Content decodes the HTTP response body to bytes.
func (resp *Response) Content() ([]byte, error) {
	if resp.err != nil || resp.content != nil {
		return resp.content, resp.err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
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

	if resp.content != nil {
		return json.Unmarshal(resp.content, v)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
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

	if resp.content != nil {
		return xml.Unmarshal(resp.content, v)
	}
	defer resp.Body.Close()

	return xml.NewDecoder(resp.Body).Decode(v)
}

// Cookies returns the HTTP response cookies.
func (resp *Response) Cookies() ([]*http.Cookie, error) {
	if resp.err != nil {
		return nil, resp.err
	}

	return resp.Response.Cookies(), nil
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

	if resp.StatusCode/100 != 2 {
		resp.err = fmt.Errorf("sreq: bad status (%s)", resp.Status)
	}
	return resp
}

// EnsureStatus ensures the HTTP response's status code must be code.
func (resp *Response) EnsureStatus(code int) *Response {
	if resp.err != nil {
		return resp
	}

	if resp.StatusCode != code {
		resp.err = fmt.Errorf("sreq: bad status (%s)", resp.Status)
	}
	return resp
}

// Save saves the HTTP response into a file.
func (resp *Response) Save(filename string, perm os.FileMode) error {
	if resp.err != nil {
		return resp.err
	}

	if resp.content != nil {
		return ioutil.WriteFile(filename, resp.content, perm)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// Verbose makes the HTTP request and its response more talkative.
// It's similar to "curl -v", used for debug.
func (resp *Response) Verbose(w io.Writer, withBody bool) (err error) {
	if resp.err != nil {
		return resp.err
	}

	err = dumpRequest(resp.Request, w, withBody)

	fmt.Fprintf(w, "< %s %s\r\n", resp.Proto, resp.Status)
	for k, vs := range resp.Header {
		for _, v := range vs {
			fmt.Fprintf(w, "< %s: %s\r\n", k, v)
		}
	}
	io.WriteString(w, "<\r\n")

	if !withBody || resp.ContentLength == 0 {
		return
	}

	if resp.content != nil {
		fmt.Fprintf(w, "%s\r\n", b2s(resp.content))
		return
	}

	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)

	io.WriteString(w, "\r\n")
	return
}
