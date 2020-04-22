package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ShutdownBody shutdown body
type ShutdownBody struct {
	ShutDown int
}

func (q *ShutdownBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ShutdownBody) Map() model.Map {
	r := make(model.Map)
	r["shutdown_type"] = q.ShutDown
	return r
}

// DecodeShutdownReq 解码shutdown请求
// doc: https://dev.mysql.com/doc/internals/en/com-shutdown.html
func DecodeShutdownReq(src *parse.Source) (*ShutdownBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x08 {
		return nil, errors.New("packet isn't a shutdown request")
	}
	res := new(ShutdownBody)
	if pkLen > 1 {
		res.ShutDown, _ = common.GetIntN(src.ReadN(1), 1)
	}
	return res, nil
}
