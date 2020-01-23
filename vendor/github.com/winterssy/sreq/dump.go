package sreq

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	reqWriteExcludeHeaderDump = map[string]bool{
		"Host":              true, // not in Header map anyway
		"Transfer-Encoding": true,
		"Trailer":           true,
	}
)

func drainBody(body io.ReadCloser) (*bytes.Buffer, error) {
	defer body.Close()
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)
	return &buf, err
}

func dumpRequestLine(req *http.Request, w io.Writer) {
	fmt.Fprintf(w, "> %s %s %s\r\n", req.Method, req.URL.RequestURI(), req.Proto)
}

func dumpRequestHeaders(req *http.Request, w io.Writer) {
	host := req.Host
	if req.Host == "" && req.URL != nil {
		host = req.URL.Host
	}
	if host != "" {
		fmt.Fprintf(w, "> Host: %s\r\n", host)
	}

	if len(req.TransferEncoding) > 0 {
		fmt.Fprintf(w, "> Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
	}
	if req.Close {
		io.WriteString(w, "> Connection: close\r\n")
	}

	for k, vs := range req.Header {
		if !reqWriteExcludeHeaderDump[k] {
			for _, v := range vs {
				fmt.Fprintf(w, "> %s: %s\r\n", k, v)
			}
		}
	}

	io.WriteString(w, ">\r\n")
}

func dumpRequestBody(req *http.Request, w io.Writer) error {
	const (
		tip = "if you see this message it means the request body has already been consumed and cannot be read twice"
	)

	var err error
	if req.GetBody == nil {
		fmt.Fprintf(w, "<!-- %s -->\r\n", tip)
	} else {
		var rc io.ReadCloser
		rc, err = req.GetBody()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(w, rc)
		if err == nil {
			io.WriteString(w, "\r\n")
		}
	}

	return err
}

func dumpRequest(req *http.Request, w io.Writer, withBody bool) error {
	dumpRequestLine(req, w)
	dumpRequestHeaders(req, w)

	if !withBody || req.Body == nil || req.Body == http.NoBody {
		return nil
	}

	return dumpRequestBody(req, w)
}
