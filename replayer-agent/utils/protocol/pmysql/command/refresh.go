package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// RefreshBody refresh body
type RefreshBody struct {
	SubCMD int
}

func (q *RefreshBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *RefreshBody) Map() model.Map {
	r := make(model.Map)
	r["subCMD"] = q.SubCMD
	return r
}

// DecodeRefreshReq 解码refresh请求
// doc: https://dev.mysql.com/doc/internals/en/com-refresh.html
func DecodeRefreshReq(src *parse.Source) (*RefreshBody, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x07 {
		return nil, errors.New("packet isn't a refresh request")
	}
	res := new(RefreshBody)
	res.SubCMD, _ = common.GetIntN(src.ReadN(1), 1)
	return res, nil
}
