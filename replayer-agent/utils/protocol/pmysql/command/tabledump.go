package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// TableDumpBody ...
type TableDumpBody struct {
	DatabaseLen int
	Database    string
	TableLen    int
	Table       string
}

func (p *TableDumpBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *TableDumpBody) Map() model.Map {
	r := make(model.Map)
	r["database_len"] = p.DatabaseLen
	r["database"] = p.Database
	r["table_len"] = p.TableLen
	r["table"] = p.Table
	return r
}

// DecodeTableDumpReq
// doc: https://dev.mysql.com/doc/internals/en/com-table-dump.html
func DecodeTableDumpReq(src *parse.Source) (*TableDumpBody, error) {
	common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x13 {
		return nil, errors.New("packet isn't table dump query")
	}
	resp := new(TableDumpBody)
	var err error
	resp.DatabaseLen, err = common.GetIntN(src.ReadN(1), 1)
	if err != nil {
		return nil, err
	}
	resp.Database = common.GetLenencString(src)
	resp.TableLen, err = common.GetIntN(src.ReadN(1), 1)
	if err != nil {
		return nil, err
	}
	resp.Table = common.GetLenencString(src)
	return resp, nil
}
