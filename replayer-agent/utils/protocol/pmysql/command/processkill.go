package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ProcessKillBody process kill body
type ProcessKillBody struct {
	ConnectionID int
}

func (q *ProcessKillBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ProcessKillBody) Map() model.Map {
	r := make(model.Map)
	r["connection_id"] = q.ConnectionID
	return r
}

// DecodeProcessKillReq 解码com process kill请求
// doc: https://dev.mysql.com/doc/internals/en/com-process-kill.html
func DecodeProcessKillReq(src *parse.Source) (*ProcessKillBody, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x0c {
		return nil, errors.New("packet isn't a process kill request")
	}
	res := new(ProcessKillBody)
	res.ConnectionID, _ = common.GetIntN(src.ReadN(4), 4)
	return res, nil
}
