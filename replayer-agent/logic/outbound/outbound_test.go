package outbound_test

import (
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/logic/outbound"
	"go.uber.org/zap"
)

func init() {
	tlog.Handler = tlog.NewTLog(zap.NewExample())
}

func TestOutBound(t *testing.T) {

	const address = "127.0.0.1:3515"

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	go outbound.Start(addr)
	time.Sleep(300 * time.Millisecond)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	msgs := []string{
		"2f2a7b22726964223a223766303030303031363066666334386561336666323934626466663761626230222c2261646472223a2231302e3139302e392e3137393a39313031227d2a2f000000e8",
		"2f2a7b22726964223a223766303030303031363066666334386561336666323934626466663761626230222c2261646472223a2231302e3139302e392e3137393a39313031227d2a2f80010001",
	}

	for _, msg := range msgs {
		time.Sleep(5 * time.Millisecond)

		b, err := hex.DecodeString(msg)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
		_, err = conn.Write(b)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
	}

	time.Sleep(300 * time.Millisecond)
	conn.Close()
	time.Sleep(300 * time.Millisecond)
}
