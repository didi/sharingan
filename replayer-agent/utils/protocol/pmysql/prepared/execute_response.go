package prepared

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/command"
	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// StmtExecuteResponse ...
type StmtExecuteResponse struct {
	ColumnNum  int
	ColumnDef  []command.Columndef
	ExtraBytes []byte
}

// Map ...
func (s *StmtExecuteResponse) Map() model.Map {
	r := make(model.Map)
	r["column_num"] = s.ColumnNum
	list := make(model.List, 0, s.ColumnNum)
	for _, def := range s.ColumnDef {
		list = append(list, def.Map())
	}
	r["column_def"] = list
	r["extra_bytes"] = s.ExtraBytes
	return r
}

var (
	errNoStmtExecuteResponsePacket = errors.New("not stmt execute response packet")
)

// DecodeStmtExecuteResonsePacket 解码stmt execute的返回结果
// doc: https://dev.mysql.com/doc/internals/en/com-stmt-execute-response.html
func DecodeStmtExecuteResonsePacket(src *parse.Source) (*StmtExecuteResponse, error) {
	common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp := new(StmtExecuteResponse)
	var err error
	resp.ColumnNum, _, err = common.GetLenencInt(src)
	if nil != err {
		return nil, err
	}
	for i := 0; i < resp.ColumnNum; i++ {
		def, err := command.DecodeColumnDef(src)
		if nil != err {
			return nil, err
		}
		resp.ColumnDef = append(resp.ColumnDef, def)
	}
	if !common.ReadEOFPacket(src) {
		return nil, errNoStmtExecuteResponsePacket
	}
	resp.ExtraBytes = src.ReadAll()
	return resp, nil
}
