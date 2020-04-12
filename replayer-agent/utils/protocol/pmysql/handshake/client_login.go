package handshake

import (
	"bytes"
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// ClientLogin client发送给mysql server的auth请求体
type ClientLogin struct {
	Capabilities         int // int<2>
	ExtendedCapabilities int // int<2>
	MaxPacketSize        int // int<4>
	Charset              byte
	// 23 bytes zero
	UserName   string // string<nul>
	PasswdLen  byte
	Passwd     []byte // string<len> len=$PasswdLen
	Database   string
	AuthPlugin string
	Attr       map[string]string
}

func (c *ClientLogin) String() string {
	data, err := json.Marshal(c)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (c *ClientLogin) Map() model.Map {
	r := make(model.Map)
	r["capabilities"] = c.Capabilities
	r["extended_capabilities"] = c.ExtendedCapabilities
	r["max_packet_size"] = c.MaxPacketSize
	r["charset"] = c.Charset
	r["user_name"] = c.UserName
	r["passwd_len"] = c.PasswdLen
	r["passwd"] = c.Passwd
	r["database"] = c.Database
	r["auth_plugin"] = c.AuthPlugin
	attr := make(model.Map)
	for k, v := range c.Attr {
		attr[k] = v
	}
	r["attr"] = attr
	return r
}

// DecodeClientLoginPacket 解码client发送给mysql server的auth验证请求
// doc: https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::HandshakeResponse41
func DecodeClientLoginPacket(src *parse.Source) (*ClientLogin, error) {
	common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp := &ClientLogin{}
	var err error
	resp.Capabilities, err = common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return nil, err
	}
	resp.ExtendedCapabilities, err = common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return nil, err
	}
	resp.MaxPacketSize, err = common.GetIntN(src.ReadN(4), 4)
	if nil != err {
		return nil, err
	}
	resp.Charset = src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	for i := 0; i < 23; i++ {
		if !src.Expect1(byte(0)) {
			return nil, errNotHandshakeResponsePacket
		}
	}
	resp.UserName = common.GetStringNull(src)
	pl := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp.PasswdLen = pl
	resp.Passwd = src.ReadN(int(pl))
	if src.Error() != nil {
		return nil, src.Error()
	}
	resp.AuthPlugin = common.GetStringNull(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	if idx := bytes.IndexByte(src.PeekAll(), byte(0)); idx != -1 {
		resp.Database = resp.AuthPlugin
		resp.AuthPlugin = common.GetStringNull(src)
		if src.Error() != nil {
			return nil, src.Error()
		}
	}
	kvLen, _, err := common.GetLenencInt(src)
	if nil != err {
		return nil, err
	}
	if kvLen == 0 {
		return resp, nil
	}
	resp.Attr = make(map[string]string)
	consumed := 0
	for consumed < kvLen {
		k := common.GetLenencString(src)
		if src.Error() != nil {
			return nil, src.Error()
		}
		v := common.GetLenencString(src)
		if src.Error() != nil {
			return nil, src.Error()
		}
		resp.Attr[k] = v
		consumed += common.GetLenencStringLength(k) + common.GetLenencStringLength(v)
	}
	return resp, nil
}

var (
	errNotHandshakeResponsePacket = errors.New("not handshake response packet")
)
