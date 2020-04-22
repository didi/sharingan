package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// TimeReq time req
type TimeReq struct {
}

func (q *TimeReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *TimeReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeTimeReq 解码com_time请求
// doc: https://dev.mysql.com/doc/internals/en/com-time.html
func DecodeTimeReq(src *parse.Source) (*TimeReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x0f {
		return nil, errors.New("packet isn't a time request")
	}
	return &TimeReq{}, nil
}
