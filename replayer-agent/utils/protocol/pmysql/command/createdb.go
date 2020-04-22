package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// CreateDBBody create db body
type CreateDBBody struct {
	Database string
}

func (q *CreateDBBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *CreateDBBody) Map() model.Map {
	r := make(model.Map)
	r["database"] = q.Database
	return r
}

// DecodeCreateDBReq 解码create db请求
// doc: https://dev.mysql.com/doc/internals/en/com-create-db.html
func DecodeCreateDBReq(src *parse.Source) (*CreateDBBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x05 {
		return nil, errors.New("packet isn't a create db request")
	}
	return &CreateDBBody{Database: string(src.ReadN(pkLen - 1))}, nil
}
