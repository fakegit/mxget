package request

import "github.com/winterssy/sreq"

const (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
)

var (
	DefaultClient *sreq.Client
)

func init() {
	DefaultClient = sreq.NewWithSession().OnBeforeRequest(sreq.SetDefaultUserAgent(defaultUserAgent))
}
