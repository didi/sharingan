package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// InitDBBody init db body
type InitDBBody struct {
	Table string
}

func (q *InitDBBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *InitDBBody) Map() model.Map {
	r := make(model.Map)
	r["table"] = q.Table
	return r
}

// DecodeInitDBReq 解码init db请求
// doc: https://dev.mysql.com/doc/internals/en/com-init-db.html
func DecodeInitDBReq(src *parse.Source) (*InitDBBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x02 {
		return nil, errors.New("packet isn't a init db request")
	}
	return &InitDBBody{Table: string(src.ReadN(pkLen - 1))}, nil
}
