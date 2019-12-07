package request

import "github.com/winterssy/sreq"

const (
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
)

func SetUserAgent(req *sreq.Request) error {
	if req.Err != nil {
		return req.Err
	}

	req.RawRequest.Header.Set("User-Agent", UserAgent)
	return nil
}

var (
	DefaultClient = sreq.New().UseRequestInterceptors(SetUserAgent)
)
