package prepared

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// StmtExecuteBody client发起执行stmt的packet
// 由于解析过于复杂，这期先不做
type StmtExecuteBody struct {
	StatementID int
	Flag        byte
	ExtraBytes  []byte
}

func (s *StmtExecuteBody) String() string {
	data, err := json.Marshal(s)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (s *StmtExecuteBody) Map() model.Map {
	r := make(model.Map)
	r["statement_id"] = s.StatementID
	r["flag"] = s.Flag
	r["extra_bytes"] = s.ExtraBytes
	return r
}

var (
	errNoStmtExecutePacket = errors.New("not stmt execute packet")
)

// DecodeStmtExecutePacket 解码stmt execute包
// 解析过于复杂，这期先不做
// doc: https://dev.mysql.com/doc/internals/en/com-stmt-execute.html
func DecodeStmtExecutePacket(src *parse.Source) (*StmtExecuteBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	if !src.Expect1(0x17) {
		return nil, errNoStmtExecutePacket
	}
	resp := new(StmtExecuteBody)
	var err error
	resp.StatementID, err = common.GetIntN(src.ReadN(4), 4)
	if nil != err {
		return nil, err
	}
	resp.Flag = src.Read1()
	if resp.Flag > 0x04 {
		return nil, errNoPrepareResponsePacket
	}
	// The iteration-count(4 bytes) is always 1
	if !src.Expect1(1) {
		return nil, errNoPrepareResponsePacket
	}
	for i := 0; i < 3; i++ {
		if !src.Expect1(0) {
			return nil, errNoStmtExecutePacket
		}
	}
	// ---
	resp.ExtraBytes = src.ReadN(pkLen - 10)
	return resp, src.Error()
}
