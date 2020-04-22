package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// DaemonReq daemon req
type DaemonReq struct {
}

func (q *DaemonReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *DaemonReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeDaemonReq 解码com daemon请求
// doc: https://dev.mysql.com/doc/internals/en/com-daemon.html
func DecodeDaemonReq(src *parse.Source) (*DaemonReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x1d {
		return nil, errors.New("packet isn't a daemon request")
	}
	return &DaemonReq{}, nil
}
