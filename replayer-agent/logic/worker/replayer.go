package worker

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/global"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/logic/outbound"
	"github.com/didi/sharingan/replayer-agent/logic/replayed"
	"github.com/didi/sharingan/replayer-agent/model/nuwaplt"
	"github.com/didi/sharingan/replayer-agent/model/replaying"
	"github.com/didi/sharingan/replayer-agent/model/station"
)

var hooks []func()

func ExitHook() {
	for _, hook := range hooks {
		hook()
	}
}

type Replayer struct {
	BasePort int // the base port number of the outbound servers
	OBSIdx   int // stands for the index of the outbound servers

	Language string // language of module
	Protocol string // protocol of module

	ReplayAddr string

	ReplayedSession *replayed.Session // replayed session
}

func (r *Replayer) ReplaySession(ctx context.Context, session *replaying.Session, project string) error {
	stat, err := r.ReplaySessionPreHandle(ctx, session)
	if stat != 0 {
		return err
	}

	traceID := GenTraceID()
	// store session
	station.Store(traceID, session, r.ReplayedSession)
	defer station.Remove(traceID)
	// pass session to outbound-matcher
	outbound.StoreHandler(ctx, traceID)
	defer outbound.RemoveHandler(ctx, traceID)

	stat, err = r.ReplaySessionDoreplay(ctx, session, traceID, project)
	return err
}

func (r *Replayer) ReplaySessionPreHandle(ctx context.Context, session *replaying.Session) (int, error) {
	if session == nil || session.CallFromInbound == nil || session.ReturnInbound == nil {
		err := errors.New("CallFromInbound is nill or ReturnInbound is nil")
		tlog.Handler.Errorf(ctx, tlog.DebugTag, "errmsg=", err)
		return 1, err
	}

	// replayed record
	r.ReplayedSession = new(replayed.Session)
	r.ReplayedSession.SessionId = session.SessionId
	r.ReplayedSession.Context = session.Context
	r.ReplayedSession.OnlineOutbounds = session.CallOutbounds
	r.ReplayedSession.OnlineAppendFiles = session.AppendFiles

	return 0, nil
}

func (r *Replayer) ReplaySessionDoreplay(ctx context.Context, session *replaying.Session, traceID string, project string) (int, error) {
	request := session.CallFromInbound.Request
	conn, err := newConn(ctx, r.ReplayAddr)
	if err != nil {
		return 1, err
	}

	{
		// add header
		traceHeader := fmt.Sprintf("\r\nSharingan-Replayer-Traceid: %s", traceID)
		timeHeader := fmt.Sprintf("\r\nSharingan-Replayer-Time: %d", session.CallFromInbound.OccurredAt)
		s := bytes.Split(request, []byte("\r\n"))
		if strings.Contains(string(s[0]), " HTTP/") {
			// for supporting grpc server
			serverType := nuwaplt.GetValueByKey(project, nuwaplt.KServerType, global.HTTP_SERVER)
			serverHeader := fmt.Sprintf("\r\nSharingan-Replayer-Server: %s", serverType)
			// http: add header at the second line of http inbound. the first line cannot be changed for 400 Bad Request
			s[0] = append(s[0], []byte(traceHeader+timeHeader+serverHeader)...)
			request = bytes.Join(s, []byte("\r\n"))
		} else {
			// thrift: add header at the first line of thrift inbound, for thrift request may be too large
			traceHeader = fmt.Sprintf("Sharingan-Replayer-Traceid: %s\r\n", traceID)
			timeHeader = fmt.Sprintf("Sharingan-Replayer-Time: %d\r\n", session.CallFromInbound.OccurredAt)
			request = append([]byte(traceHeader+timeHeader), request...)
		}
	}

	testResponse, err := handleConn(ctx, conn, request, true, project)

	tlog.Handler.Infof(ctx, tlog.DebugTag, "responseOfTest=%v", strconv.Quote(string(testResponse)))
	tlog.Handler.Infof(ctx, tlog.DebugTag, "responseOfOnline=%v", strconv.Quote(string(session.ReturnInbound.Response)))

	// fill with replayed response
	r.ReplayedSession.Request = []byte(session.CallFromInbound.Request)
	r.ReplayedSession.OnlineResponse = session.ReturnInbound.Response
	r.ReplayedSession.TestResponse = testResponse
	return 0, nil
}

//GenTraceID 生成traceID
func GenTraceID() string {
	ip := "127.0.0.1"

	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()
	b := bytes.Buffer{}

	b.WriteString(hex.EncodeToString(net.ParseIP(ip).To4()))
	b.WriteString(fmt.Sprintf("%x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", rand.Int31n(1<<24)))
	b.WriteString("b0")

	return b.String()
}
