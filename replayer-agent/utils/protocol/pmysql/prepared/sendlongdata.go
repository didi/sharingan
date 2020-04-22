package prepared

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// QueryBody ...
type SendLongDataBody struct {
	StatementID int
	ParamID     int
	Data        string
}

func (p *SendLongDataBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *SendLongDataBody) Map() model.Map {
	r := make(model.Map)
	r["statement_id"] = p.StatementID
	r["param_id"] = p.ParamID
	r["data"] = p.Data
	return r
}

// DecodeSendLongDataQuery ... 发送BLOB类型的数据
// doc: https://dev.mysql.com/doc/internals/en/com-stmt-send-long-data.html
func DecodeSendLongDataQuery(src *parse.Source) (*SendLongDataBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x18 {
		return nil, errors.New("packet isn't send long data query")
	}
	resp := new(SendLongDataBody)
	var err error
	resp.StatementID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.ParamID, err = common.GetIntN(src.ReadN(2), 2)
	if err != nil {
		return nil, err
	}
	resp.Data = string(src.ReadN(pkLen - 7))

	return resp, nil
}
