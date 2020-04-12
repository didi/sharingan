package replayed

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/ignore"
	"github.com/didichuxing/sharingan/replayer-agent/model/nuwaplt"
	"github.com/didichuxing/sharingan/replayer-agent/model/protocol"
	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"
)

var (
	// standard error type
	HasDiffErr          = errors.New("Has Diff")
	HasDiffButIgnoreErr = errors.New("Has Diff but ignore")

	// user-defined error type
	EmptyJsonErr              = errors.New("empty json obj")
	EmptyBodyErr              = errors.New("empty test request or response")
	PublicLogFormatErr        = errors.New("malformed public log")
	PublicLogNotDefinedErr    = errors.New("log path not defined")
	PublicLogKeyNotDefinedErr = errors.New("key not defined")
)

type FormatDiff struct {
	Key      string                    `json:"label"`
	Children [2]map[string]interface{} `json:"children"`
	Is       int                       `json:"is"`
	NoiseId  int                       `json:"noiseId"`
	NoiseUri string                    `json:"noiseUri"`
	Noise    string                    `json:"noiseData"`
	Project  string                    `json:"noiseProject"`
}

type Diff struct {
	A, B                       string
	Noise                      map[string]nuwaplt.NoiseInfo
	CallFromInboundRequestMark string
}

// 解析协议并对比，通过打平数组diff
func (d *Diff) CompareProtocol() (compared []*FormatDiff, requestMark string, err error, protocal string) {
	var (
		aFormat, bFormat     map[string]json.RawMessage
		protocalA, protocalB string
	)
	compared = make([]*FormatDiff, 0)
	aFormat, requestMark, err, protocalA = d.ParseProtocol(d.A)
	protocal = protocalA
	if err != nil {
		if protocal == protocol.UNKNOWN_PRO {
			protocal = protocol.HTTP_PRO
		}
		return
	}
	if d.CallFromInboundRequestMark != "" {
		requestMark = d.CallFromInboundRequestMark
	}
	aFormat["out.req"] = json.RawMessage([]byte(strconv.Quote(requestMark)))

	// fast return
	if d.B == "" {
		err = EmptyBodyErr
		return
	}

	var requestMark1 string
	bFormat, requestMark1, err, protocalB = d.ParseProtocol(d.B)
	if protocalA != protocalB {
		protocal = protocalA + "-" + protocalB
	}
	if d.CallFromInboundRequestMark != "" {
		requestMark1 = d.CallFromInboundRequestMark
	}
	if err == nil || bFormat != nil {
		bFormat["out.req"] = json.RawMessage([]byte(strconv.Quote(requestMark1)))
	}

	compared, err = d.generateDiffs(compared, requestMark, "", aFormat, bFormat)
	return
}

var parserPool sync.Pool

func (d *Diff) ParseProtocol(body string) (map[string]json.RawMessage, string, error, string) {
	if p := parserPool.Get(); p != nil {
		parser := p.(protocol.Protocol)
		return parser.Parse(body)
	}

	parser := getParser()
	defer parserPool.Put(parser)

	return parser.Parse(body)
}

func getParser() protocol.Protocol {
	public := new(protocol.Public)
	mysql := new(protocol.Mysql)
	binaryThrift := new(protocol.BinaryThrift)
	compactThrift := new(protocol.CompactThrift)
	redis := new(protocol.Redis)
	http := new(protocol.HTTP)

	mysql.Next = public
	compactThrift.Next = mysql
	binaryThrift.Next = compactThrift
	redis.Next = binaryThrift
	http.Next = redis

	return http
}

func (d *Diff) generateDiffs(compared []*FormatDiff, prefix, keyPrefix string, aFormat, bFormat map[string]json.RawMessage) ([]*FormatDiff, error) {
	var err error
	for aKey, aValue := range aFormat {
		bValue, ok := bFormat[aKey]
		if ok {
			delete(bFormat, aKey)
		}
		aRaw, bRaw := getRawLabel(aValue, bValue)

		diff := new(FormatDiff)
		diff.Key = keyPrefix + aKey
		diff.Children[0] = map[string]interface{}{"label": helper.BytesToString(aRaw), "OoT": "O"}
		diff.Children[1] = map[string]interface{}{"label": helper.BytesToString(bRaw), "OoT": "T"}

		is := compare(bValue, aValue)
		if is == 0 {
			compared = append(compared, diff)
			continue
		}

		aKey = keyPrefix + aKey
		err1 := d.markNoiseStatus(diff, prefix, aKey)
		if err1 == HasDiffButIgnoreErr {
			compared = append(compared, diff)
			if err == nil {
				err = err1
			}
			continue
		}

		var err2 error
		compared, err2, ok = d.generateNestedDiffs(compared, prefix, aKey+".", aValue, bValue)
		if !ok {
			compared = append([]*FormatDiff{diff}, compared...)
			err = err1
			continue
		}
		if (err2 == HasDiffButIgnoreErr && err == nil) || err2 == HasDiffErr {
			err = err2
		}
	}

	for bKey, bValue := range bFormat {
		aRaw, bRaw := getRawLabel(nil, bValue)

		diff := new(FormatDiff)
		diff.Key = keyPrefix + bKey
		diff.Children[0] = map[string]interface{}{"label": helper.BytesToString(aRaw), "OoT": "O"}
		diff.Children[1] = map[string]interface{}{"label": helper.BytesToString(bRaw), "OoT": "T"}

		err1 := d.markNoiseStatus(diff, prefix, diff.Key)
		if err1 == HasDiffErr {
			compared = append([]*FormatDiff{diff}, compared...)
		} else {
			compared = append(compared, diff)
		}
		if (err == nil && err1 == HasDiffButIgnoreErr) || err1 == HasDiffErr {
			err = err1
		}
	}

	return compared, err
}

func (d *Diff) generateNestedDiffs(compared []*FormatDiff, prefix, keyPrefix string, aValue, bValue json.RawMessage) ([]*FormatDiff, error, bool) {
	saVal, err := strconv.Unquote(format(helper.BytesToString(aValue)))
	if err != nil {
		return compared, err, false
	}
	saFormat, err := helper.Json2SingleLayerMap(helper.StringToBytes(saVal))
	if err != nil {
		return compared, err, false
	}
	if len(saFormat) == 0 {
		return compared, EmptyJsonErr, false
	}
	sbVal, err := strconv.Unquote(format(helper.BytesToString(bValue)))
	if err != nil {
		return compared, err, false
	}
	sbFormat, err := helper.Json2SingleLayerMap(helper.StringToBytes(sbVal))
	if err != nil {
		return compared, err, false
	}
	if len(sbFormat) == 0 {
		return compared, EmptyJsonErr, false
	}
	compared, err = d.generateDiffs(compared, prefix, keyPrefix, saFormat, sbFormat)
	return compared, err, true
}

func (d *Diff) markNoiseStatus(diff *FormatDiff, prefix, key string) error {
	//忽略常见噪音: 直接匹配key
	_, exists := ignore.Noise[strings.Trim(key, ".0")]
	if !exists {
		_, exists = ignore.Noise[prefix+"."+strings.Trim(key, ".0")]
	}

	if !exists {
		for noise, meta := range ignore.SeniorNoise {
			switch meta.Typ {
			case ignore.NoiseMatch:
				if key == noise {
					exists = true
				}
			case ignore.NoisePrefix:
				if strings.HasPrefix(key, noise) {
					exists = true
				}
			case ignore.NoiseSuffix:
				if strings.HasSuffix(key, noise) {
					exists = true
				}
			}
			if exists {
				break
			}
		}
	}

	var noiseInfo nuwaplt.NoiseInfo
	if !exists && len(d.Noise) > 0 {
		noiseInfo, exists = d.Noise[prefix+"."+key]
	}

	if exists {
		diff.Is = 1
		if noiseInfo.Id != 0 || conf.Handler.GetString("http_api.noise_del") == "" {
			diff.NoiseId = noiseInfo.Id
			diff.NoiseUri = noiseInfo.Uri
			diff.Noise = noiseInfo.Noise
			diff.Project = noiseInfo.Project
		}
		return HasDiffButIgnoreErr
	}
	diff.Is = 2
	return HasDiffErr
}

func format(a string) string {
	return strings.Replace(strings.TrimSpace(a), "\\\\\\/\\\\u", "\\\\u", -1)
}

func getRawLabel(aValue, bValue json.RawMessage) (json.RawMessage, json.RawMessage) {
	la := len(aValue)
	lb := len(bValue)
	if la == 0 {
		if bytes.Equal(bValue, []byte{34, 34}) {
			// strconv.Quote("\"\"")
			return aValue, json.RawMessage{34, 92, 34, 92, 34, 34}
		}
	} else if lb == 0 {
		if bytes.Equal(aValue, []byte{34, 34}) {
			// strconv.Quote("\"\"")
			return json.RawMessage{34, 92, 34, 92, 34, 34}, bValue
		}
	} else if la >= 2 && aValue[0] == 34 && bValue[0] != 34 {
		if aValue[la-1] == 34 && bytes.Equal(aValue[1:la-1], bValue) {
			return json.RawMessage(strconv.Quote(string(aValue))), bValue
		}
	} else if aValue[0] != 34 && lb >= 2 && bValue[0] == 34 {
		if bValue[lb-1] == 34 && bytes.Equal(aValue, bValue[1:lb-1]) {
			return aValue, json.RawMessage(strconv.Quote(string(bValue)))
		}
	}
	return aValue, bValue
}

// 忽略因为时间戳导致只有一个字符有区别的噪音，faketime会有1s的误差
// a -> test; b -> online
// @Return 0: a==b  <>0: a!=b
func compare(a, b []byte) (n int) {
	if len(a) != len(b) {
		return 1
	}
	if len(a) < 4 {
		return bytes.Compare(a, b)
	}

	num := 0
	for i := 0; i < len(a); i++ {
		if a[i] == b[i] {
			if unicode.IsDigit(rune(a[i])) {
				num++
			} else {
				num = 0
			}
			continue
		}
		if !unicode.IsDigit(rune(a[i])) || b[i]-a[i] != 1 {
			return 1
		}
		num++
		for i+1 < len(a) && num < 10 && a[i+1] == '9' && b[i+1] == '0' {
			num++
			i++
		}
		if num != 10 && num != 2 && num != 5 {
			return 1
		}
		// 154226xxxx
		if num == 10 && a[i-9] == '1' && a[i-8] >= '5' && a[i-8] <= '6' {
			continue
		}
		// xx:xx:xx
		if num == 2 && i > 7 && a[i-2] == ':' && a[i-5] == ':' {
			continue
		}
		// 86xxx
		if num == 5 && a[i-4] == '8' && a[i-3] == '6' && a[i-2] >= '3' && a[i-2] <= '4' {
			continue
		}
		return 1
	}
	return 0
}
