package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// EOFResp EOF的返回
type EOFResp struct {
	Warnings   int // 2 bytes
	StatusFlag int // 2 bytes
}

func (eof *EOFResp) String() string {
	data, err := json.Marshal(eof)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map converts itself to model.Map
func (eof *EOFResp) Map() model.Map {
	r := make(model.Map)
	r["warnings"] = eof.Warnings
	r["status_flag"] = eof.StatusFlag
	return r
}

var (
	errNoEOFPacket = errors.New("not eof packet")
)

// DecodeEOFPacket 解码EOF的response
// doc: https://dev.mysql.com/doc/internals/en/packet-EOF_Packet.html
func DecodeEOFPacket(src *parse.Source) (*EOFResp, error) {
	pkLen, _ := common.GetPacketHeader(src)
	header := src.Read1()
	if header != 0xfe || pkLen >= 9 {
		return nil, errNoEOFPacket
	}

	resp := &EOFResp{}
	payLoad := src.ReadN(pkLen - 1)
	resp.Warnings, _ = common.GetIntN(payLoad, 2)
	payLoad = payLoad[2:]
	resp.StatusFlag, _ = common.GetIntN(payLoad, 2)
	return resp, nil
}
