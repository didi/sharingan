package protocol

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"
	phelper "github.com/didichuxing/sharingan/replayer-agent/utils/protocol/helper"
	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql"
	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pthrift"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/parse/model"
)

var NonsupportErr = errors.New("don't support the protocol parsing")

const (
	UNKNOWN_PRO    = "UNKNOWN"
	HTTP_PRO       = "HTTP"
	MYSQL_PRO      = "MYSQL"
	REDIS_PRO      = "REDIS"
	BIN_THRIFT_PRO = "BinaryThrift"
	COM_THRIFT_PRO = "CompactThrift"
	PUB_PRO        = "PUBLIC"
)

type Protocol interface {
	Parse(string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string)
}

type nextProtocol struct {
	Next Protocol
}

type HTTP struct {
	nextProtocol
}

func (this *HTTP) Parse(body string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string) {
	protocal = UNKNOWN_PRO
	if !strings.Contains(body, "HTTP/1") {
		if this.Next != nil {
			pairs, requestMark, err, protocal = this.Next.Parse(body)
			return
		}
		err = NonsupportErr
		return
	}

	protocal = HTTP_PRO
	pairs, requestMark, err = ParseHTTP(body)
	return
}

type BinaryThrift struct {
	nextProtocol
}

func (this *BinaryThrift) Parse(body string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string) {
	protocal = UNKNOWN_PRO
	// 忽略4字节packet长度
	err = NonsupportErr
	if len(body) <= 4 {
		return
	}

	thrift, DecErr := pthrift.DecodeBinary([]byte(body)[4:])
	if DecErr != nil && this.Next != nil {
		pairs, requestMark, err, protocal = this.Next.Parse(body)
		return
	}

	protocal = BIN_THRIFT_PRO
	if name, exists := thrift["name"]; exists {
		requestMark, _ = name.(string)
	}
	if param, exists := thrift["param"]; exists {
		var jsonEncode []byte
		if paramMap, ok := param.(model.Map); ok {
			jsonEncode, _ = phelper.MarshalMap(paramMap)
		} else {
			var json = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonEncode, _ = json.Marshal(param)
		}
		pairs, err = helper.Json2SingleLayerMap(jsonEncode)
		return
	}
	return
}

type CompactThrift struct {
	nextProtocol
}

func (this *CompactThrift) Parse(body string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string) {
	protocal = UNKNOWN_PRO
	// 忽略4字节packet长度
	err = NonsupportErr
	if len(body) <= 4 {
		return
	}

	thrift, DecErr := pthrift.DecodeCompact([]byte(body)[4:])
	if DecErr != nil && this.Next != nil {
		pairs, requestMark, err, protocal = this.Next.Parse(body)
		return
	}

	protocal = COM_THRIFT_PRO
	if name, exists := thrift["name"]; exists {
		requestMark, _ = name.(string)
	}
	if param, exists := thrift["param"]; exists {
		var jsonEncode []byte
		if paramMap, ok := param.(model.Map); ok {
			jsonEncode, _ = phelper.MarshalMap(paramMap)
		} else {
			var json = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonEncode, _ = json.Marshal(param)

		}
		pairs, err = helper.Json2SingleLayerMap(jsonEncode)
		return
	}
	return
}

type Mysql struct {
	nextProtocol
}

func (this *Mysql) Parse(body string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string) {
	protocal = UNKNOWN_PRO
	err = NonsupportErr
	mysql := pmysql.DecodePacketWithoutHeader([]byte(body))
	if mysql == nil && this.Next != nil {
		pairs, requestMark, err, protocal = this.Next.Parse(body)
		return
	} else if mysql == nil {
		return
	}

	var data, sql interface{}
	var dataMap model.Map
	var sqlStr string
	var exists, ok bool

	if data, exists = mysql["data"]; !exists {
		return
	}
	if dataMap, ok = data.(model.Map); !ok {
		return
	}
	protocal = MYSQL_PRO
	if sql, exists = dataMap["sql"]; !exists {
		return
	}
	if sqlStr, ok = sql.(string); !ok {
		return
	}

	// mysql注释忽略
	splitSql := strings.Split(sqlStr, "/*")
	if len(splitSql) > 0 {
		if len(splitSql[0]) > 30 {
			requestMark = string([]byte(splitSql[0])[:30])
		} else {
			requestMark = splitSql[0]
		}
		formatSql := strings.Split(strings.Trim(splitSql[0], " "), " ")
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonEncode, _ := json.Marshal(formatSql)
		pairs, err = helper.Json2SingleLayerMap(jsonEncode)
		return
	}

	return
}

type Redis struct {
	nextProtocol
}

func (this *Redis) Parse(body string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string) {
	protocal = REDIS_PRO
	if strings.HasPrefix(body, "*") {
		if len(body) > 30 {
			requestMark = string([]byte(body)[:30])
		} else {
			requestMark = body
		}
		redis := strings.Split(body, "\r\n")
		// 过滤非redis协议
		if len(redis) <= 1 || !strings.Contains(body, "$") {
			goto nextPro
		}
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonEncode, _ := json.Marshal(redis)
		pairs, err = helper.Json2SingleLayerMap(jsonEncode)
		return
	}
nextPro:
	if this.Next != nil {
		pairs, requestMark, err, protocal = this.Next.Parse(body)
		return
	}

	protocal = UNKNOWN_PRO
	err = NonsupportErr
	return
}

type Public struct {
	nextProtocol
}

func (this *Public) Parse(body string) (pairs map[string]json.RawMessage, requestMark string, err error, protocal string) {
	protocal = PUB_PRO
	if strings.Contains(body, "||") {
		pairs, requestMark, err = ParsePublic(body)
		return
	} else if this.Next != nil {
		pairs, requestMark, err, protocal = this.Next.Parse(body)
		return
	}

	protocal = UNKNOWN_PRO
	err = NonsupportErr
	return
}
