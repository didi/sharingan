package handshake

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/json-iterator/go"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ServerGreeting 服务器发给client的greeting包
// field顺序就是protocol顺序
type ServerGreeting struct {
	ProtocolVersion byte   // 1
	MySQLVersion    string // string<nul>
	ThreadID        int    // int<4>
	Salt_1          []byte // 8 bytes
	// filter 0x00
	Capabilities               int // int<2>
	ServerLanguage             byte
	StatusFlag                 int // 2 bytes
	ExtendedServerCapabilities int // 2 bytes
	AuthPluginLen              int // 1 byte
	// 10 bytes reserved
	Salt_2     []byte
	AuthPlugin string
}

func (sg *ServerGreeting) String() string {
	data, err := json.Marshal(sg)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (sg *ServerGreeting) Map() model.Map {
	r := make(model.Map)
	r["protocol_version"] = sg.ProtocolVersion
	r["mysql_version"] = sg.MySQLVersion
	r["thread_id"] = sg.ThreadID
	r["salt"] = append(append([]byte(nil), sg.Salt_1...), sg.Salt_2...)
	r["capabilities"] = sg.Capabilities
	r["server_language"] = sg.ServerLanguage
	r["extended_capabilities"] = sg.ExtendedServerCapabilities
	r["status_flag"] = sg.StatusFlag
	r["auth_plugin_length"] = sg.AuthPluginLen
	r["auth_plugin"] = sg.AuthPlugin
	return r
}

// DecodeServerGreeting 解码mysql发给client的greet包
// doc: https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::Handshake
func DecodeServerGreeting(src *parse.Source) (*ServerGreeting, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if err := src.Error(); err != nil {
		return nil, err
	}
	return decodeServerGreeting(src, pkLen)
}

func decodeServerGreeting(src *parse.Source, pkLen int) (*ServerGreeting, error) {
	resp := &ServerGreeting{
		ProtocolVersion: src.Read1(),
	}
	if src.Error() != nil {
		return nil, src.Error()
	}
	if resp.ProtocolVersion != 0x0a {
		return nil, errors.New("unsupported protocol version")
	}
	resp.MySQLVersion = common.GetStringNull(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	var err error
	resp.ThreadID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.Salt_1 = src.ReadN(8)
	if src.Error() != nil {
		return nil, src.Error()
	}
	// filter 0x00
	src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp.Capabilities, err = common.GetIntN(src.ReadN(2), 2)
	if err != nil {
		return nil, err
	}
	resp.ServerLanguage = src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp.StatusFlag, err = common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return nil, err
	}
	resp.ExtendedServerCapabilities, err = common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return nil, err
	}
	resp.AuthPluginLen, err = common.GetIntN(src.ReadN(1), 1)
	if nil != err {
		return nil, err
	}
	// 10 bytes reserved
	src.ReadN(10)
	if src.Error() != nil {
		return nil, src.Error()
	}
	saltLen := max(13, resp.AuthPluginLen-8)
	resp.Salt_2 = src.ReadN(saltLen)
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp.AuthPlugin = common.GetStringNull(src)
	return resp, src.Error()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
