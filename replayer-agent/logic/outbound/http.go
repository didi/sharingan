package outbound

import (
	"bytes"
	"context"
	"strconv"

	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
)

var http100req = []byte("Expect: 100-continue")
var http100resp = []byte("HTTP/1.1 100 Continue\r\n\r\n")

func simulateHttp(ctx context.Context, request []byte) []byte {
	if bytes.Contains(request, http100req) {
		return simulateHttp100(ctx, request)
	}
	return nil
}

func simulateHttp100(ctx context.Context, request []byte) []byte {
	tlog.Handler.Debugf(ctx, tlog.DebugTag, "simulated_http||requestKeyword=100-continue||content=%s", strconv.Quote(string(request)))
	return http100resp
}
