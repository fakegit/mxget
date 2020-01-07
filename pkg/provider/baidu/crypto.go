package baidu

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/winterssy/mxget/pkg/cryptography"
	"github.com/winterssy/sreq"
)

const (
	Input = "2012171402992850"
	IV    = "2012061402992850"
)

var (
	key string
)

func init() {
	hash := fmt.Sprintf("%X", md5.Sum([]byte(Input)))
	key = hash[len(hash)/2:]
}

func aesCBCEncrypt(songId string) sreq.Params {
	params := sreq.Params{
		"songid": songId,
		"ts":     time.Now().UnixNano() / 1e6,
	}

	e := base64.StdEncoding.EncodeToString(cryptography.AESCBCEncrypt([]byte(params.Encode(false)), []byte(key), []byte(IV)))
	params.Set("e", e)

	return params
}

func signPayload(params sreq.Params) sreq.Params {
	q := params.Encode(false)
	ts := time.Now().Unix()
	r := fmt.Sprintf("baidu_taihe_music_secret_key%d", ts)
	key := fmt.Sprintf("%x", md5.Sum([]byte(r)))[8:24]
	param := base64.StdEncoding.EncodeToString(cryptography.AESCBCEncrypt([]byte(q), []byte(key), []byte(key)))
	sign := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("baidu_taihe_music%s%d", param, ts))))
	return sreq.Params{
		"timestamp": ts,
		"param":     param,
		"sign":      sign,
	}
}
