package prepared

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// CloseBody ...
type CloseBody struct {
	StatementID int
}

func (p *CloseBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *CloseBody) Map() model.Map {
	r := make(model.Map)
	r["statement_id"] = p.StatementID
	return r
}

// DecodeStmtClose ...
func DecodeStmtClose(src *parse.Source) (*CloseBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	if pkLen != 0x05 {
		return nil, errors.New("packet isn't close prepared query")
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x19 {
		return nil, errors.New("packet isn't close prepared query")
	}

	resp := new(CloseBody)
	var err error
	resp.StatementID, err = common.GetIntN(src.ReadN(4), 4)
	if nil != err {
		return nil, err
	}

	return resp, nil
}
