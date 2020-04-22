package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ConnectOutReq connect req
type ConnectOutReq struct {
}

func (q *ConnectOutReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ConnectOutReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeConnectOutReq 解码com_connect_out请求
// doc: https://dev.mysql.com/doc/internals/en/com-connect-out.html
func DecodeConnectOutReq(src *parse.Source) (*ConnectOutReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x14 {
		return nil, errors.New("packet isn't a connect out request")
	}
	return &ConnectOutReq{}, nil
}
