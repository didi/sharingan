package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ProcessReq process req
type ProcessReq struct {
}

func (q *ProcessReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ProcessReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeProcessReq 解码com_process_info请求
// doc: https://dev.mysql.com/doc/internals/en/com-process-info.html
func DecodeProcessReq(src *parse.Source) (*ProcessReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x0a {
		return nil, errors.New("packet isn't a process info request")
	}
	return &ProcessReq{}, nil
}
