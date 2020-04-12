package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ResetConnnetionReq reset connection req
type ResetConnnetionReq struct {
}

func (q *ResetConnnetionReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ResetConnnetionReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeResetConnectionReq 解码com reset connection请求
// doc: https://dev.mysql.com/doc/internals/en/com-reset-connection.html
func DecodeResetConnectionReq(src *parse.Source) (*ResetConnnetionReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x1f {
		return nil, errors.New("packet isn't a reset connection request")
	}
	return &ResetConnnetionReq{}, nil
}
