package worker

import (
	"context"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didichuxing/sharingan/replayer-agent/idl"
	"github.com/didichuxing/sharingan/replayer-agent/logic/search"
	"github.com/didichuxing/sharingan/replayer-agent/model/esmodel"
	"github.com/didichuxing/sharingan/replayer-agent/model/replaying"
)

func FetchSessions(ctx context.Context, sessionId string, project string) []*replaying.Session {
	var esSessions []esmodel.Session

	session := search.GetRawSessions(ctx, &idl.SearchReq{SessionId: sessionId, Size: 1, Project: project})
	if session != nil {
		esSessions = append(esSessions, *session)
	}

	t := &Transformer{}
	sessions, err := t.BuildSessions(esSessions)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, "errmsg=fetch session failed||err=%s", err)
	}

	return sessions
}

type Record struct {
	Session esmodel.Session `json:"data"`
}
