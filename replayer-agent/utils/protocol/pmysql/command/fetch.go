package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// FetchBody ...
type FetchBody struct {
	StatementID int
	NumRows     int
}

func (p *FetchBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *FetchBody) Map() model.Map {
	r := make(model.Map)
	r["statement_id"] = p.StatementID
	r["num_rows"] = p.NumRows
	return r
}

// DecodeFetchReq
// doc: https://dev.mysql.com/doc/internals/en/com-stmt-fetch.html
func DecodeFetchReq(src *parse.Source) (*FetchBody, error) {
	common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x1c {
		return nil, errors.New("packet isn't fetch query")
	}
	resp := new(FetchBody)
	var err error
	resp.StatementID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.NumRows, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
