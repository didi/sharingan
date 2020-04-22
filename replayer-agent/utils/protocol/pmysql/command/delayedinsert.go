package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// DelayedInsertReq delayed insert req
type DelayedInsertReq struct {
}

func (q *DelayedInsertReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *DelayedInsertReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeDelayedInsertReq 解码com delayed insert请求
// doc: https://dev.mysql.com/doc/internals/en/com-delayed-insert.html
func DecodeDelayedInsertReq(src *parse.Source) (*DelayedInsertReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x10 {
		return nil, errors.New("packet isn't a delayed insert request")
	}
	return &DelayedInsertReq{}, nil
}
