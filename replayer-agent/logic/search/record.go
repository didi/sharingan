package search

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didichuxing/sharingan/replayer-agent/model/esmodel"
	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"
)

type SearchRecord struct {
	Project   string              `json:"project"`
	SessionId string              `json:"sessionId"`
	Timestamp string              `json:"timestamp"`
	Req       string              `json:"req"`
	RawReq    string              `json:"-"`
	Res       string              `json:"res"`
	RawRes    string              `json:"-"`
	Actions   []map[string]string `json:"actions"`
}

/**
 * 是否只是获取SessionId
 */
func queryForID(fields []string) bool {
	if len(fields) == 1 && fields[0] == "SessionId" {
		return true
	}
	return false
}

/**
 * 从返回值中恢复SessionId
 */
func retrieveSessionIds(ctx context.Context, body []byte, flowList []*SearchRecord) []*SearchRecord {
	ids, err := esmodel.RetrieveSessionIds(body)
	if err != nil {
		tlog.Handler.Warnf(ctx, tlog.DLTagUndefined, "errmsg=RetrieveSessions error||err=%s", err)
		return flowList
	}
	for _, s := range ids {
		flowList = append(flowList, &SearchRecord{SessionId: s.Id})
	}
	return flowList
}

/**
 * 从返回值中恢复Session
 */
func retrieveSessions(ctx context.Context, project string, body []byte, flowList []*SearchRecord) []*SearchRecord {
	sessions, err := esmodel.RetrieveSessions(body)
	if err != nil {
		tlog.Handler.Warnf(ctx, tlog.DLTagUndefined, "errmsg=RetrieveSessions error||err=%s", err)
		return flowList
	}

	for _, sess := range sessions {
		rec := handleOneSessionRaw(ctx, sess, project)
		if rec != nil {
			flowList = append(flowList, rec)
		}
	}
	return flowList
}

//handleOneSessionRaw 处理原始流量，只处理一个session
func handleOneSessionRaw(ctx context.Context, sess esmodel.Session, project string) *SearchRecord {
	req, res, valid := uniform(ctx, sess)
	if !valid {
		return nil
	}

	rec := &SearchRecord{
		Project:   project,
		SessionId: sess.SessionId,
		Timestamp: getTimestamp(sess.SessionId),
		Req:       req,
		RawReq:    helper.BytesToString(sess.CallFromInbound.Request.Data),
		Res:       res,
		RawRes:    helper.BytesToString(sess.ReturnInbound.Response.Data),
		Actions:   []map[string]string{},
	}

	rec.Actions = append(rec.Actions, map[string]string{
		"req": req,
		"res": res,
	})
	for _, action := range sess.Actions {
		if action.ActionMeta.ActionType == "CallOutbound" {
			ac := map[string]string{
				"req": string(action.Request.Data),
				"res": string(action.Response.Data),
			}
			rec.Actions = append(rec.Actions, ac)
		} else if action.ActionMeta.ActionType == "SendUDP" {
			ac := map[string]string{
				"apollo": string(action.Content.Data),
			}
			rec.Actions = append(rec.Actions, ac)
		}

	}

	return rec
}

func uniform(ctx context.Context, sess esmodel.Session) (string, string, bool) {
	req := helper.BytesToString(sess.CallFromInbound.Request.Data)

	if sess.ReturnInbound == nil {
		tlog.Handler.Warnf(ctx, tlog.DLTagUndefined, "errmsg=returnInboud is nil||sessionid=%s", sess.SessionId)
		return "", "", false
	}
	if sess.ReturnInbound.Response.Data == nil {
		raw := getInboundResponse(sess.Actions)
		if raw == nil {
			tlog.Handler.Warnf(ctx, tlog.DLTagUndefined, "errmsg=returnInboud is nil||sessionid=%s", sess.SessionId)
			return "", "", false
		}
		sess.ReturnInbound.Response.Data = raw
	}

	res := helper.BytesToString(sess.ReturnInbound.Response.Data)
	return req, res, true
}

func getTimestamp(sid string) string {
	splits := strings.Split(sid, "-")
	if len(splits) != 2 {
		return ""
	}
	nanosec, err := strconv.Atoi(splits[0])
	if err != nil {
		return ""
	}
	then := time.Unix(int64(nanosec)/1e9, int64(nanosec)%1e9)
	return then.Format("2006-01-02T15:04:05")
}

func getInboundResponse(actions []esmodel.Action) []byte {
	for _, action := range actions {
		if action.ActionMeta.ActionType == "ReturnInbound" {
			return action.Response.Data
		}
	}
	return nil
}
