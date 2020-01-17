package sreq

import (
	"fmt"
	"io"
	"net/http"
)

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Transfer-Encoding": true,
	"Trailer":           true,
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

	for k, vs := range req.Header {
		if !reqWriteExcludeHeaderDump[k] {
			for _, v := range vs {
				fmt.Fprintf(w, "> %s: %s\r\n", k, v)
			}
		}
	}

	io.WriteString(w, ">\r\n")
}

func dumpRequestBody(req *http.Request, w io.Writer) (err error) {
	const (
		tip = "if you see this message it means the HTTP request body cannot be read twice, it may be a stream"
	)

	if req.GetBody == nil {
		fmt.Fprintf(w, "<!-- %s -->\r\n", tip)
	} else if req.ContentLength != 0 {
		var rc io.ReadCloser
		rc, err = req.GetBody()
		if err != nil {
			return
		}
		defer rc.Close()

		_, err = io.Copy(w, rc)
		io.WriteString(w, "\r\n")
	}

	return
}

func dumpRequest(req *http.Request, w io.Writer, withBody bool) error {
	dumpRequestLine(req, w)
	dumpRequestHeaders(req, w)

	if !withBody || req.Body == nil {
		return nil
	}

	return dumpRequestBody(req, w)
}
