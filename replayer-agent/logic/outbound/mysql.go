package outbound

import (
	"bytes"
	"context"
	"encoding/hex"
	"strconv"

	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
)

//var mysqlGreeting = []byte{53, 0, 0, 0, 10, 53, 46, 48, 46, 53, 49, 98, 0, 1, 0, 0, 0, 47, 85, 62, 116, 80, 114, 109, 75, 0, 12, 162, 33, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 86, 76, 87, 84, 124, 52, 47, 46, 55, 107, 55, 110, 0}

var mysqlGreeting []byte
var maxAllowedPacketResp []byte

func init() {
	mysqlGreeting, _ = hex.DecodeString("4e0000000a352e372e31362d6c6f67009f943f003663266e4a62520d00fff7530200ff8115000000000000000000003e5529546501277e7c430f1a006d7973716c5f6e61746976655f70617373776f726400")
	maxAllowedPacketResp, _ = hex.DecodeString(
		"0100000101" +
			"2a000002036465660000001440406d61785f616c6c6f7765645f7061636b6574000c3f001500000008a000000000" +
			"05000003fe00000200" +
			"09000004083637313038383634" +
			"05000005fe00000200")
}

func simulateMysql(ctx context.Context, request []byte) []byte {
	resp := _simulateMysql(ctx, request)
	if resp != nil {
		resp[3] = request[3] + 1
	}
	return resp
}

func _simulateMysql(ctx context.Context, request []byte) []byte {
	if bytes.Index(request, []byte("mysql_native_password")) != -1 {
		tlog.Handler.Debugf(ctx, tlog.DebugTag, "simulated_mysql||requestKeyword=mysql_native_password||content=%s", strconv.Quote(string(request)))
		return []byte{0x07, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}
	}
	if bytes.Index(request, []byte("@@max_allowed_packet")) != -1 {
		tlog.Handler.Debugf(ctx, tlog.DebugTag, "simulated_mysql||requestKeyword=@@max_allowed_packet||content=%s", strconv.Quote(string(request)))
		return maxAllowedPacketResp
	}

	if bytes.Index(request, []byte("SET NAMES utf8")) != -1 {
		tlog.Handler.Debugf(ctx, tlog.DebugTag, "simulated_mysql||requestKeyword=SET NAMES utf8||content=%s", strconv.Quote(string(request)))
		return []byte{0x07, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}
	}

	//if bytes.Index(request, []byte("START TRANSACTION")) != -1 {
	//    tlog.Handler.Debugf(ctx, tlog.DebugTag, "simulated_mysql||requestKeyword=START TRANSACTION||content=%s", strconv.Quote(string(request)))
	//    return []byte{0x07, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00}
	//}
	return nil
}
