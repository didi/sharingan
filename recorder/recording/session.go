package recording

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/v2pro/plz/countlog"
)

// TraceHeader TraceHeader
type TraceHeader []byte

// Session Session
type Session struct {
	Context             string           `json:"Context"`
	SessionID           string           `json:"SessionId"`
	ThreadID            int32            `json:"ThreadId"`
	TraceHeader         TraceHeader      `json:"TraceHeader"`
	TraceID             []byte           `json:"TraceId"`
	SpanID              []byte           `json:"SpanId"`
	NextSessionID       string           `json:"NextSessionId"`
	CallFromInbound     *CallFromInbound `json:"CallFromInbound"`
	ReturnInbound       *ReturnInbound   `json:"ReturnInbound"`
	Actions             []Action         `json:"Actions"`
	currentAppendFiles  map[string]*AppendFile
	currentCallOutbound *CallOutbound
}

// NewSession NewSession
func NewSession(threadID int32) *Session {
	return &Session{
		ThreadID:  threadID,
		SessionID: fmt.Sprintf("%d-%d", time.Now().UnixNano(), threadID),
	}
}

// MarshalJSON MarshalJSON
func (session *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Session
		TraceID json.RawMessage `json:"TraceId"`
		SpanID  json.RawMessage `json:"SpanId"`
	}{
		Session: *session,
		TraceID: EncodeAnyByteArray(session.TraceID),
		SpanID:  EncodeAnyByteArray(session.SpanID),
	})
}

// AppendFile AppendFile
func (session *Session) AppendFile(ctx context.Context, content []byte, fileName string) {
	if session == nil {
		return
	}
	if session.currentAppendFiles == nil {
		session.currentAppendFiles = map[string]*AppendFile{}
	}
	appendFile := session.currentAppendFiles[fileName]
	if appendFile == nil {
		appendFile = &AppendFile{
			action:   session.newAction("AppendFile"),
			FileName: fileName,
		}
		session.currentAppendFiles[fileName] = appendFile
		session.addAction(appendFile)
	}
	appendFile.Content = append(appendFile.Content, content...)
}

// ReadStorage ReadStorage
func (session *Session) ReadStorage(ctx context.Context, span []byte) {
	if session == nil {
		return
	}
	session.addAction(&ReadStorage{
		action:  session.newAction("ReadStorage"),
		Content: append([]byte(nil), span...),
	})
}

// RecvFromInbound Inbound Request
func (session *Session) RecvFromInbound(ctx context.Context, span []byte, peer net.TCPAddr, unix net.UnixAddr) {
	if session == nil {
		return
	}
	if session.CallFromInbound == nil {
		session.CallFromInbound = &CallFromInbound{
			action:   session.newAction("CallFromInbound"),
			Peer:     peer,
			UnixAddr: unix,
		}
	}

	span = bytes.Replace(span, []byte("Expect: 100-continue\r\n"), []byte(""), -1)
	session.CallFromInbound.Request = append(session.CallFromInbound.Request, span...)
}

// SendToInbound Inbound Response
func (session *Session) SendToInbound(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}

	if bytes.HasPrefix(span, []byte("HTTP/1.1 100 Continue")) {
		return
	}

	if session.ReturnInbound == nil {
		session.ReturnInbound = &ReturnInbound{
			action: session.newAction("ReturnInbound"),
		}
		session.addAction(session.ReturnInbound)
	}
	session.ReturnInbound.Response = append(session.ReturnInbound.Response, span...)
}

// SendToOutbound Outbound TCP Request
func (session *Session) SendToOutbound(ctx context.Context, span []byte, peer net.TCPAddr, local *net.TCPAddr, socketFD int) {
	if session == nil {
		return
	}

	if (session.currentCallOutbound == nil) ||
		(session.currentCallOutbound.Peer.String() != peer.String()) ||
		(session.currentCallOutbound.SocketFD != socketFD) ||
		(len(session.currentCallOutbound.Response) > 0) {
		session.newCallOutbound(peer, local, socketFD)
	} else if session.currentCallOutbound != nil && session.currentCallOutbound.ResponseTime > 0 {
		// last request get a bad response, e.g., timeout
		session.newCallOutbound(peer, local, socketFD)
	}

	session.currentCallOutbound.Request = append(session.currentCallOutbound.Request, span...)
}

// SendUDPToOutbound Outbound UDP Request
func (session *Session) SendUDPToOutbound(ctx context.Context, span []byte, peer net.UDPAddr) {
	session.addAction(&SendUDP{
		action:  session.newAction("SendUDP"),
		Peer:    peer,
		Content: append([]byte(nil), span...),
	})
}

// RecvFromOutbound Outbound Response
func (session *Session) RecvFromOutbound(ctx context.Context, span []byte, peer net.TCPAddr, local *net.TCPAddr, socketFD int) {
	if session == nil {
		return
	}

	// match response, find best match action int reverse order, limit 10 times
	searchCnt, searchLimit := 0, 10
	for i := len(session.Actions) - 1; i >= 0; i-- {
		if outbound, ok := session.Actions[i].(*CallOutbound); ok {
			if outbound.Peer.String() == peer.String() && outbound.SocketFD == socketFD {
				outbound.ResponseTime = time.Now().UnixNano()
				outbound.Response = append(outbound.Response, span...)
				return
			}
		}
		searchCnt++
		if searchCnt > searchLimit {
			break
		}
	}

	if session.currentCallOutbound == nil {
		session.newCallOutbound(peer, local, socketFD)
	}
	if (session.currentCallOutbound.Peer.String() != peer.String()) ||
		(session.currentCallOutbound.SocketFD != socketFD) {
		session.newCallOutbound(peer, local, socketFD)
	}
	if session.currentCallOutbound.ResponseTime == 0 {
		session.currentCallOutbound.ResponseTime = time.Now().UnixNano()
	}
	session.currentCallOutbound.Response = append(session.currentCallOutbound.Response, span...)
}

// HasRequest has Request
func (session *Session) HasRequest() bool {
	if session == nil {
		return false
	}

	if session.CallFromInbound == nil {
		return false
	}

	return true
}

// HasResponse has Response
func (session *Session) HasResponse() bool {
	if session == nil {
		return false
	}
	if session.ReturnInbound == nil {
		return false
	}
	return true
}

// newAction new action
func (session *Session) newAction(actionType string) action {
	occurredAt := time.Now().UnixNano()

	return action{
		ActionIndex: len(session.Actions),
		OccurredAt:  occurredAt,
		ActionType:  actionType,
	}
}

// newCallOutbound new call outbound
func (session *Session) newCallOutbound(peer net.TCPAddr, local *net.TCPAddr, socketFD int) {
	session.currentCallOutbound = &CallOutbound{
		action:   session.newAction("CallOutbound"),
		Peer:     peer,
		Local:    local,
		SocketFD: socketFD,
		CSpanID:  []byte(nil),
	}

	session.addAction(session.currentCallOutbound)
}

// addAction addAction
func (session *Session) addAction(action Action) {
	if !session.HasRequest() {
		return
	}

	if !ShouldRecordAction(action) {
		return
	}

	session.Actions = append(session.Actions, action)
}

// Shutdown shotdown and recorder
func (session *Session) Shutdown(ctx context.Context, newSession *Session) {
	if session == nil {
		return
	}

	session.Summary(newSession)

	if session.CallFromInbound == nil {
		return
	}

	if len(session.CallFromInbound.Request) == 0 {
		return
	}

	session.NextSessionID = newSession.SessionID
	for _, recorder := range Recorders {
		recorder.Record(session)
	}

	countlog.Debug("event!recording.session_recorded",
		"ctx", ctx,
		"threadID", session.ThreadID,
		"session", session,
	)
}

// Summary sunmmary
func (session *Session) Summary(newSession *Session) {
	reqLen := 0
	respLen := 0

	if session.CallFromInbound != nil {
		reqLen = len(session.CallFromInbound.Request)
	}

	if session.ReturnInbound != nil {
		respLen = len(session.ReturnInbound.Response)
	}

	countlog.Trace("event!recording.shutdown_recording_session",
		"threadID", session.ThreadID,
		"sessionId", session.SessionID,
		"nextSessionId", newSession.SessionID,
		"callFromInboundBytes",
		reqLen,
		"returnInboundBytes",
		respLen,
		"actionsCount",
		len(session.Actions))
}
