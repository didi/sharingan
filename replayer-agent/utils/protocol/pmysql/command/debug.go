package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// DebugReq debug req
type DebugReq struct {
}

func (q *DebugReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *DebugReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeDebugReq 解码com_debug请求
// doc: https://dev.mysql.com/doc/internals/en/com-debug.html
func DecodeDebugReq(src *parse.Source) (*DebugReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x0d {
		return nil, errors.New("packet isn't a debug request")
	}
	return &DebugReq{}, nil
}
