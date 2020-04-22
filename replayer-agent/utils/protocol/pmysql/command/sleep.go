package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// SleepReq sleep req
type SleepReq struct {
}

func (q *SleepReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *SleepReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeSleepReq 解码com_sleep请求
// doc: https://dev.mysql.com/doc/internals/en/com-sleep.html
func DecodeSleepReq(src *parse.Source) (*SleepReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x00 {
		return nil, errors.New("packet isn't a sleep request")
	}
	return &SleepReq{}, nil
}
