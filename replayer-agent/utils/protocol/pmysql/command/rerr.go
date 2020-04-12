package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ErrResp 错误返回包
type ErrResp struct {
	Header     byte
	ErrCode    int
	ExtraBytes []byte
}

func (e *ErrResp) String() string {
	data, err := json.Marshal(e)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map converts itself to a model.Map
func (e *ErrResp) Map() model.Map {
	r := make(model.Map)
	r["header"] = e.Header
	r["errcode"] = e.ErrCode
	r["extra_bytes"] = e.ExtraBytes
	return r
}

// DecodeErrPacket 解码err response包
// doc: https://dev.mysql.com/doc/internals/en/packet-ERR_Packet.html
func DecodeErrPacket(src *parse.Source) (*ErrResp, error) {
	pkLen, _ := common.GetPacketHeader(src)
	header := src.Read1()
	if header != 0xff {
		return nil, errors.New("not Err_Packet")
	}
	resp := &ErrResp{
		Header: header,
	}
	data := src.ReadN(pkLen)
	errCode, err := common.GetIntN(data, 2)
	if nil != err {
		src.ReportError(err)
		return nil, err
	}
	resp.ErrCode = errCode
	resp.ExtraBytes = data[2:]
	return resp, nil
}
