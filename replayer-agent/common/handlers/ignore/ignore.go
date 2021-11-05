package ignore

import (
	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
)

const (
	NoiseMatch = iota
	NoisePrefix
	NoiseSuffix
	NoiseContains
	NoiseProxyHttp
)

type NoiseMeta struct {
	Typ   int    // type of noise
	Scope string // working scope for this noise
}

// 常规请求级别噪音
var Noise = map[string]bool{}

// 定制请求级别噪音
var SeniorNoise = map[string]NoiseMeta{}

// 接口级别噪音
var OutboundNoise = map[string]NoiseMeta{}

// not matched接口级噪音，目前仅用于go模块
var NotMatchedNoise = map[string]NoiseMeta{}

// 对 agent 收到的 http 请求不进行匹配，直接走代理
var ProxyHttp = map[string]NoiseMeta{}

// 直接走代理的 HTTP 请求中最长的路径长度
var MaxProxyHttpLen int

func Init() {
	noises := conf.HandlerInfo.GetStringSlice("ignore.noise")
	for _, noise := range noises {
		Noise[noise] = true
	}

	seniorNoises := conf.HandlerInfo.GetStringSlice("ignore.seniorNoise")
	for _, seniorNoise := range seniorNoises {
		SeniorNoise[seniorNoise] = NoiseMeta{NoiseSuffix, ""}
	}

	OutboundNoise = map[string]NoiseMeta{
		"\x01\x00\x00\x00\x01": NoiseMeta{NoiseMatch, ""},
	}

	NotMatchedNoise = map[string]NoiseMeta{
		"mysql_native_password": NoiseMeta{NoiseContains, ""},
		"SET NAMES utf8":        NoiseMeta{NoiseContains, ""},
		"SET autocommit=1":      NoiseMeta{NoiseContains, ""},
	}

	notMatches := conf.HandlerInfo.GetStringSlice("ignore.notMatch")
	for _, notMatch := range notMatches {
		NotMatchedNoise[notMatch] = NoiseMeta{NoiseContains, ""}
	}

	proxyHttps := conf.HandlerInfo.GetStringSlice("ignore.proxyHttp")
	for _, proxyHttp := range proxyHttps {
		ProxyHttp[proxyHttp] = NoiseMeta{NoiseProxyHttp, ""}
		if len(proxyHttp) > MaxProxyHttpLen {
			MaxProxyHttpLen = len(proxyHttp)
		}
	}
}
