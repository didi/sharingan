package prepared

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/json-iterator/go"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// QueryBody ...
type QueryBody struct {
	RawQuery string
}

func (p *QueryBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *QueryBody) Map() model.Map {
	r := make(model.Map)
	r["sql"] = p.RawQuery
	return r
}

// DecodePreparedQuery ...
func DecodePreparedQuery(src *parse.Source) (*QueryBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x16 {
		return nil, errors.New("packet isn't prepared query")
	}
	query := src.ReadN(pkLen - 1)
	if src.Error() != nil {
		return nil, src.Error()
	}
	return &QueryBody{RawQuery: string(query)}, nil
}
