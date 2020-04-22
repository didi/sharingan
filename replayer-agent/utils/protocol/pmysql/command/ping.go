package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// PingReq ping req
type PingReq struct {
}

func (q *PingReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *PingReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodePingReq 解码com_ping请求
// doc: https://dev.mysql.com/doc/internals/en/com-ping.html
func DecodePingReq(src *parse.Source) (*PingReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x0e {
		return nil, errors.New("packet isn't a ping request")
	}
	return &PingReq{}, nil
}
