package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/json-iterator/go"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// OKResp OK的返回
type OKResp struct {
	Header       byte
	AffectedRows int // lenenc_int
	LastInsertID int // lenenc_int
	StatusFlag   int // 2 bytes
	Warnings     int // 2 bytes
	ExtraBytes   []byte
}

func (ok *OKResp) String() string {
	data, err := json.Marshal(ok)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map converts itself to model.Map
func (ok *OKResp) Map() model.Map {
	r := make(model.Map)
	r["header"] = ok.Header
	r["affected_rows"] = ok.AffectedRows
	r["last_insert_id"] = ok.LastInsertID
	r["status_flag"] = ok.StatusFlag
	r["warnings"] = ok.Warnings
	r["extra_bytes"] = ok.ExtraBytes
	return r
}

var (
	errNoOKPacket = errors.New("not ok packet")
)

// DecodeOKPacket 解码OK的response
// doc: https://dev.mysql.com/doc/internals/en/packet-OK_Packet.html
func DecodeOKPacket(src *parse.Source) (*OKResp, error) {
	pkLen, _ := common.GetPacketHeader(src)
	resp := &OKResp{
		Header: src.Read1(),
	}
	if resp.Header == 0xfe && pkLen < 9 {
		return nil, errors.New("Decode a EOF packet as an OK packet")
	}
	if resp.Header != 0x00 && resp.Header != 0xfe {
		return nil, errNoOKPacket
	}
	payLoad := src.ReadN(pkLen - 1)
	src.Peek1()
	if src.Error() == nil {
		return nil, errNoOKPacket
	}
	affectedRows, bts, err := common.GetIntLenc(payLoad)
	if err != nil {
		src.ReportError(err)
		return nil, err
	}
	resp.AffectedRows = affectedRows
	payLoad = payLoad[bts:]
	lastInsertID, bts, err := common.GetIntLenc(payLoad)
	if err != nil {
		src.ReportError(err)
		return nil, err
	}
	resp.LastInsertID = lastInsertID
	payLoad = payLoad[bts:]
	resp.StatusFlag, _ = common.GetIntN(payLoad, 2)
	payLoad = payLoad[2:]
	resp.Warnings, _ = common.GetIntN(payLoad, 2)
	resp.ExtraBytes = payLoad[2:]
	// 剩下的字段不读了
	return resp, nil
}
