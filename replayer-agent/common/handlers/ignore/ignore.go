package ignore

const (
	NoiseMatch = iota
	NoisePrefix
	NoiseSuffix
	NoiseContains
)

type NoiseMeta struct {
	Typ   int    // type of noise
	Scope string // working scope for this noise
}

// 常规请求级别噪音
var Noise map[string]bool

// 定制请求级别噪音
var SeniorNoise map[string]NoiseMeta

// 接口级别噪音
var OutboundNoise map[string]NoiseMeta

// not matched接口级噪音，目前仅用于go模块
var NotMatchedNoise map[string]NoiseMeta

func Init() {
	Noise = map[string]bool{
		// global field
		"createTime": true,
		"req_flag":   true,
		"request_id": true,
		"sign":       true,
		"time_stamp": true,
		"timeOffset": true,
		"timeStamp":  true,
		"timestamp":  true,
		"token":      true,

		// http method field

		// thrift method field
		"mget.field_2.field_3": true,
		"mset.field_2.field_3": true,

		// controller field
	}

	SeniorNoise = map[string]NoiseMeta{
		".cyborg_sub_id": NoiseMeta{NoiseSuffix, ""},
		".sample.code":   NoiseMeta{NoiseSuffix, ""},
	}

	OutboundNoise = map[string]NoiseMeta{
		"\x01\x00\x00\x00\x01": NoiseMeta{NoiseMatch, ""},
	}

	NotMatchedNoise = map[string]NoiseMeta{
		"mysql_native_password": NoiseMeta{NoiseContains, ""},
		"SET NAMES utf8":        NoiseMeta{NoiseContains, ""},
	}
}
