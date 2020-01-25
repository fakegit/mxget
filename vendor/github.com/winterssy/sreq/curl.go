package sreq

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

const (
	curlCommand = "curl"
)

type (
	command struct {
		buf strings.Builder
	}
)

func (cmd *command) append(s string) {
	if cmd.buf.Len() > 0 {
		cmd.buf.WriteByte(' ')
	}
	cmd.buf.WriteString(s)
}

func (cmd *command) addFlag(name string, args ...string) {
	if len(args) == 0 {
		cmd.append(name)
	} else {
		for _, arg := range args {
			cmd.append(name)
			cmd.append(bashEscape(arg))
		}
	}
}

func (cmd *command) encode() string {
	return cmd.buf.String()
}

var bashEscaper = strings.NewReplacer(`'`, `'\''`)

func bashEscape(s string) string {
	return `'` + bashEscaper.Replace(s) + `'`
}

func genCURLCommand(req *http.Request) (string, error) {
	var err error
	cmd := command{}
	cmd.append(curlCommand)
	cmd.addFlag("-v")
	cmd.addFlag("-X", req.Method)

	if req.Body != nil {
		var body *bytes.Buffer
		body, err = drainBody(req.Body)
		if err == nil && body.Len() != 0 {
			cmd.addFlag("-d", b2s(body.Bytes()))
			req.Body = ioutil.NopCloser(body)
		}
	}

	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		if !reqWriteExcludeHeaderDump[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	headers := make([]string, 0, len(keys))
	if req.Host != "" && req.Host != req.URL.Host {
		headers = append(headers, fmt.Sprintf("Host: %s", req.Host))
	}
	if len(req.TransferEncoding) > 0 {
		headers = append(headers, fmt.Sprintf("Transfer-Encoding: %s",
			strings.Join(req.TransferEncoding, ",")))
	}
	if req.Close {
		headers = append(headers, "Connection: close")
	}
	for _, k := range keys {
		vs := req.Header[k]
		for _, v := range vs {
			headers = append(headers, fmt.Sprintf("%s: %s", k, v))
		}
	}
	if len(headers) > 0 {
		cmd.addFlag("-H", headers...)
	}

	cmd.append(bashEscape(req.URL.String()))
	return cmd.encode(), err
}
