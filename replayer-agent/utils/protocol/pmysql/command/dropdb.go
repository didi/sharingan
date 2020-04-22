package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// DropDBBody drop db body
type DropDBBody struct {
	Database string
}

func (q *DropDBBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *DropDBBody) Map() model.Map {
	r := make(model.Map)
	r["database"] = q.Database
	return r
}

// DecodeDropDBReq 解码drop db请求
// doc: https://dev.mysql.com/doc/internals/en/com-drop-db.html
func DecodeDropDBReq(src *parse.Source) (*DropDBBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x06 {
		return nil, errors.New("packet isn't a drop db request")
	}
	return &DropDBBody{Database: string(src.ReadN(pkLen - 1))}, nil
}
