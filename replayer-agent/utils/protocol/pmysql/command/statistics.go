package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// StatisticsReq statistics req
type StatisticsReq struct {
}

func (q *StatisticsReq) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *StatisticsReq) Map() model.Map {
	r := make(model.Map)
	return r
}

// DecodeStatisticsReq 解码com_statistics请求
// doc: https://dev.mysql.com/doc/internals/en/com-statistics.html
func DecodeStatisticsReq(src *parse.Source) (*StatisticsReq, error) {
	common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x09 {
		return nil, errors.New("packet isn't a statistic request")
	}
	return &StatisticsReq{}, nil
}
