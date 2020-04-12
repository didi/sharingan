package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// BinlogDumpGtidBody ...
type BinlogDumpGtidBody struct {
	Flags             int
	ServerID          int
	BinlogFilenameLen int
	BinlogFilename    string
	BinlogPos         int
	ExtraData         string
}

func (p *BinlogDumpGtidBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *BinlogDumpGtidBody) Map() model.Map {
	r := make(model.Map)
	r["flags"] = p.Flags
	r["server_id"] = p.ServerID
	r["binlog_filename_len"] = p.BinlogFilenameLen
	r["binlog_filename"] = p.BinlogFilename
	r["binlog_pos"] = p.BinlogPos
	r["extra_data"] = p.ExtraData
	return r
}

// DecodeBinlogDumpGtidReq
// doc: https://dev.mysql.com/doc/internals/en/com-binlog-dump-gtid.html
func DecodeBinlogDumpGtidReq(src *parse.Source) (*BinlogDumpGtidBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x1e {
		return nil, errors.New("packet isn't binlog dump gtid query")
	}
	resp := new(BinlogDumpGtidBody)
	var err error
	resp.Flags, err = common.GetIntN(src.ReadN(2), 2)
	if err != nil {
		return nil, err
	}
	resp.ServerID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.BinlogFilenameLen, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.BinlogFilename = common.GetLenencString(src)
	resp.BinlogPos, err = common.GetIntN(src.ReadN(8), 8)
	if err != nil {
		return nil, err
	}
	resp.ExtraData = string(src.ReadN(pkLen - 19 - len([]byte(resp.BinlogFilename))))
	return resp, nil
}
