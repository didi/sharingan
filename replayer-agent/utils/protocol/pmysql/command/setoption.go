package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// SetOptionBody set option body
type SetOptionBody struct {
	Operation int
}

func (q *SetOptionBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *SetOptionBody) Map() model.Map {
	r := make(model.Map)
	r["operation"] = q.Operation
	return r
}

// DecodeSetOptionReq 解码refresh请求
// doc: https://dev.mysql.com/doc/internals/en/com-set-option.html
func DecodeSetOptionReq(src *parse.Source) (*SetOptionBody, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x1b {
		return nil, errors.New("packet isn't a set option request")
	}
	res := new(SetOptionBody)
	res.Operation, _ = common.GetIntN(src.ReadN(2), 2)
	return res, nil
}
