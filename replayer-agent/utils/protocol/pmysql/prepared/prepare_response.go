package prepared

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/command"
	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

type PrepareResponse struct {
	StatementID  int // 4 bytes
	ColumnNumber int // 2 bytes
	ParamNumber  int // 2 bytes
	// filter 0x00
	Warnings  int // 2 bytes
	ColumnDef []command.Columndef
}

func (s *PrepareResponse) String() string {
	data, err := json.Marshal(s)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (s *PrepareResponse) Map() model.Map {
	r := make(model.Map)
	r["statement_id"] = s.StatementID
	r["column_num"] = s.ColumnNumber
	r["param_num"] = s.ParamNumber
	r["warnings"] = s.Warnings
	colDef := make(model.List, 0, s.ColumnNumber)
	for _, def := range s.ColumnDef {
		colDef = append(colDef, def.Map())
	}
	r["column_def"] = colDef
	return r
}

var (
	errNoPrepareResponsePacket = errors.New("not prepare response")
)

// DecodePrepareResponse 解码prepare语句的response
func DecodePrepareResponse(src *parse.Source) (*PrepareResponse, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	if pkLen == 0 {
		return nil, errNoPrepareResponsePacket
	}
	if !src.Expect1(0) {
		return nil, errNoPrepareResponsePacket
	}
	var err error
	resp := new(PrepareResponse)
	resp.StatementID, err = common.GetIntN(src.ReadN(4), 4)
	if nil != err {
		return nil, err
	}
	resp.ColumnNumber, err = common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return nil, err
	}
	resp.ParamNumber, err = common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return nil, err
	}
	// flter 0x00
	if !src.Expect1(0) {
		return nil, errNoPrepareResponsePacket
	}
	resp.Warnings, err = common.GetIntN(src.ReadN(2), 2)
	if resp.ParamNumber > 0 {
		for i := 0; i < resp.ParamNumber; i++ {
			err = readParamDef(src)
			if nil != err {
				return nil, err
			}
		}
		common.ReadEOFPacket(src)
		if src.Error() != nil {
			return nil, src.Error()
		}
	}

	if resp.ColumnNumber > 0 {
		for i := 0; i < resp.ColumnNumber; i++ {
			colDef, err := command.DecodeColumnDef(src)
			if nil != err {
				return nil, err
			}
			resp.ColumnDef = append(resp.ColumnDef, colDef)
		}
		common.ReadEOFPacket(src)
		if src.Error() != nil {
			return nil, src.Error()
		}
	}
	return resp, src.Error()
}

func readParamDef(src *parse.Source) error {
	pkLen, _ := common.GetPacketHeader(src)
	def := src.ReadN(4)
	if src.Error() != nil {
		return src.Error()
	}
	// 0x03 def
	if def[0] != 0x03 || def[1] != 0x64 || def[2] != 0x65 || def[3] != 0x66 {
		return errNoPrepareResponsePacket
	}
	src.ReadN(pkLen - 4)
	return src.Error()
}
