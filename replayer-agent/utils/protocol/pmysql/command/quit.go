package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// QuitReq quit req
type QuitReq struct {
}

func (q *QuitReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *QuitReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeQuitReq 解码com_quit请求
// doc: https://dev.mysql.com/doc/internals/en/com-quit.html
func DecodeQuitReq(src *parse.Source) (*QuitReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x01 {
		return nil, errors.New("packet isn't a quit request")
	}
	return &QuitReq{}, nil
}
