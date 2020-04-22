package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// FieldListBody field list body
type FieldListBody struct {
	Table string
	Field string
}

func (q *FieldListBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *FieldListBody) Map() model.Map {
	r := make(model.Map)
	r["table"] = q.Table
	r["field"] = q.Field
	return r
}

// DecodeFieldListReq 解码com field list请求
// doc: https://dev.mysql.com/doc/internals/en/com-field-list.html
func DecodeFieldListReq(src *parse.Source) (*FieldListBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x04 {
		return nil, errors.New("packet isn't a field list request")
	}

	res := new(FieldListBody)
	data := src.ReadN(pkLen - 1)
	isNUL, strNUL, posNUL := common.ReadNULPacket(data)
	if isNUL {
		res.Table = string(strNUL)
		if posNUL >= len(data)-1 {
			res.Field = ""
		} else {
			res.Field = string(data[posNUL+1:])
		}

	} else {
		return nil, errors.New("error happens at parsing a field list request")
	}

	return res, nil
}
