package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ConnectReq connect req
type ConnectReq struct {
}

func (q *ConnectReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ConnectReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeConnectReq 解码com_connect请求
// doc: https://dev.mysql.com/doc/internals/en/com-connect.html
func DecodeConnectReq(src *parse.Source) (*ConnectReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x0b {
		return nil, errors.New("packet isn't a connect request")
	}
	return &ConnectReq{}, nil
}
