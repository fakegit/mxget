package request

import "github.com/winterssy/sreq"

const (
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
)

var (
	gClient *sreq.Client
)

func init() {
	gClient, _ = sreq.New(nil,
		sreq.EnableSession(),
		sreq.WithTimeout(sreq.DefaultTimeout),
	)
}

func Client() *sreq.Client {
	return gClient
}
