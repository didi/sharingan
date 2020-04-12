package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// BinlogDumpBody ...
type BinlogDumpBody struct {
	BinlogPos      int
	Flags          int
	ServerID       int
	BinlogFilename string
}

func (p *BinlogDumpBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *BinlogDumpBody) Map() model.Map {
	r := make(model.Map)
	r["binlog_pos"] = p.BinlogPos
	r["flags"] = p.Flags
	r["server_id"] = p.ServerID
	r["binlog_filename"] = p.BinlogFilename
	return r
}

// DecodeBinlogDumpReq
// doc: https://dev.mysql.com/doc/internals/en/com-binlog-dump.html
func DecodeBinlogDumpReq(src *parse.Source) (*BinlogDumpBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x12 {
		return nil, errors.New("packet isn't binlog dump query")
	}
	resp := new(BinlogDumpBody)
	var err error
	resp.BinlogPos, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.Flags, err = common.GetIntN(src.ReadN(2), 2)
	if err != nil {
		return nil, err
	}
	resp.ServerID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.BinlogFilename = string(src.ReadN(pkLen - 11))

	return resp, nil
}
