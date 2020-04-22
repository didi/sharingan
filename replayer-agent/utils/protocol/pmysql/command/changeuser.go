package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ChangeUserBody change user body
type ChangeUserBody struct {
	User      string
	ExtraData string
}

func (q *ChangeUserBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *ChangeUserBody) Map() model.Map {
	r := make(model.Map)
	r["user"] = q.User
	r["extra"] = q.ExtraData
	return r
}

// DecodeChangeUserReq 解码query请求
// doc: https://dev.mysql.com/doc/internals/en/com-change-user.html
func DecodeChangeUserReq(src *parse.Source) (*ChangeUserBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x11 {
		return nil, errors.New("packet isn't a change user request")
	}

	res := new(ChangeUserBody)
	data := src.ReadN(pkLen - 1)
	isNUL, strNUL, posNUL := common.ReadNULPacket(data)
	if isNUL {
		res.User = string(strNUL)
		if posNUL >= len(data)-1 {
			res.ExtraData = ""
		} else {
			res.ExtraData = string(data[posNUL+1:])
		}
	} else {
		return nil, errors.New("error happens at parsing a change user request")
	}
	return res, nil
}
