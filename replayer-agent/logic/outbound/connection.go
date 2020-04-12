package outbound

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didichuxing/sharingan/replayer-agent/model/pool"
	"github.com/didichuxing/sharingan/replayer-agent/model/recording"
	"github.com/didichuxing/sharingan/replayer-agent/model/replaying"
	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"
)

// 回放流量标识，示例：/*{"rid":"7f0000015e7885919ead09b93a768bb0","addr":"127.0.0.1:8888"}*/

var (
	contentLengthRegex = regexp.MustCompile("Content-Length: [0-9]+\r\n")
	traficRegex        = regexp.MustCompile(`/\*{"rid":"(.*?)","addr":"(.*?)"}\*/`)
	mysqlGreetingTrace = "ca4bc2ca79c2f79729b322fbfbd91ef3" // md5("MYSQL_GREETING")
)

// ConnState 连接管理
type ConnState struct {
	LastMatchedIndex int

	// 原始连接信息
	conn    *net.TCPConn
	tcpAddr *net.TCPAddr

	// 代理连接信息
	proxyer   *Proxyer
	proxyAddr string

	// 回放请求traceID
	traceID string

	*Handler
}

// ProcessRequest 处理请求, false：关闭conn，true：保持conn
func (cs *ConnState) ProcessRequest(ctx context.Context, requestID int) bool {
	var request []byte
	var err error

	request, err = cs.readRequest(ctx)

	// EOF、timeout
	if err != nil {
		tlog.Handler.Debugf(ctx, tlog.DebugTag,
			"%s||from=%s||requestID=%v||request==%s||err==%v", helper.CInfo("<<<request of outbound||err<<<"),
			cs.tcpAddr.String(), requestID, request, err)

		return false
	}

	// case: mysqlGreeting, proxy
	if cs.traceID == mysqlGreetingTrace {
		tlog.Handler.Debugf(ctx, tlog.DebugTag,
			"%s||from=%s||requestID=%v", helper.CInfo("<<<request of outbound||mysqlGreeting<<<"),
			cs.tcpAddr.String(), requestID)

		if err := cs.proxyer.Write(ctx, cs.proxyAddr, []byte{}); err != nil {
			return false
		}
		return true
	}

	// case：empty
	if len(request) == 0 {
		tlog.Handler.Debugf(ctx, tlog.DebugTag,
			"%s||from=%s||requestID=%d||request=%s", helper.CInfo("<<<request of empty<<<"),
			cs.tcpAddr.String(), requestID, request)
		return true
	}

	tlog.Handler.Infof(ctx, tlog.DebugTag,
		"%s||from=%s||requestID=%v||request=%s||traceID=%s", helper.CInfo("<<<request of outbound<<<"),
		cs.tcpAddr.String(), requestID, strconv.Quote(helper.BytesToString(request)), cs.traceID)

	// 1、非回放阶段, 代理请求
	// 2、回放阶段，匹配请求
	if cs.traceID == "" {
		err = cs.proxyer.Write(ctx, cs.proxyAddr, request)
	} else {
		err = cs.match(ctx, request)
	}

	if err != nil || terminated(request) {
		return false
	}

	return true
}

// readRequest 获取请求
func (cs *ConnState) readRequest(ctx context.Context) ([]byte, error) {
	buf := pool.GetBuf(1024, false)
	defer pool.PutBuf(buf)

	request := pool.GetBuf(81920, true)

	bytesRead, err := cs.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	request = append(request, buf[:bytesRead]...)
	helper.SetQuickAck(cs.conn)

	// 可能还有数据没读完
	if bytesRead >= len(buf) {
		for {
			cs.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 25))
			bytesRead, err := cs.conn.Read(buf)
			if err != nil {
				break
			}
			helper.SetQuickAck(cs.conn)
			request = append(request, buf[:bytesRead]...)
			if bytesRead < len(buf) {
				break
			}
		}
	}

	request = cs.rmTrafixPrefix(ctx, request)

	return request[:len(request)], nil
}

// rmTrafixPrefix 去除流量标识
func (cs *ConnState) rmTrafixPrefix(ctx context.Context, request []byte) []byte {
	if ss := traficRegex.FindAllSubmatch(request, -1); len(ss) >= 1 {
		// 去掉前缀
		request = bytes.TrimPrefix(request, ss[0][0])
		cs.traceID = string(ss[0][1])
		cs.proxyAddr = string(ss[0][2])

		// 1、分段传输的场景，要把所有的前缀相关内容去掉
		request = bytes.Replace(request, ss[0][0], []byte(""), -1)

		// 2、部分Outbound请求会携带Inbound的所有header信息，需要剔除
		// 示例：\"haringan-Replayer-Traceid\":\"64469c315e78d24e87b25cc10750e302\",
		request = bytes.Replace(request, []byte("%22Sharingan-Replayer-Traceid%22%3A%22"+cs.traceID+"%22%2C"), []byte(""), -1)

		// fmt.Printf("cs.traceID:%s, cs.proxyAddr:%s\n", cs.traceID, cs.proxyAddr)
		// fmt.Printf("buf:%s\n", string(request))
	}

	return request
}

// match 匹配
func (cs *ConnState) match(ctx context.Context, request []byte) error {
	quotedRequest := strconv.Quote(helper.BytesToString(request))

	cs.Handler = loadHandler(ctx, string(cs.traceID))
	if cs.Handler == nil {
		tlog.Handler.Warnf(ctx, tlog.DebugTag, "errmsg=find Handler failed||request=%s||traceID=%s", quotedRequest, string(cs.traceID))
		return nil
	}

	// 去掉COM_STMT_CLOSE
	if request = removeMysqlStmtClose(request); len(request) == 0 {
		return nil
	}

	// new calloutbound
	callOutbound := replaying.NewCallOutbound(*cs.tcpAddr, request)

	// fix http 100 continue
	if err := applySimulation(ctx, simulateHttp, request, cs.conn, callOutbound); err != nil {
		return err
	}
	// some mysql connection setup interaction might not recorded
	if err := applySimulation(ctx, simulateMysql, request, cs.conn, callOutbound); err != nil {
		return err
	}

	var matchedTalk *recording.CallOutbound
	var mark float64
	cs.LastMatchedIndex, mark, matchedTalk = cs.Handler.matcher.MatchOutboundTalk(ctx, cs.Handler.replayingSession, cs.LastMatchedIndex, request)
	if callOutbound.MatchedActionIndex != fakeIndexSimulated {
		if matchedTalk == nil && cs.LastMatchedIndex != 0 {
			cs.LastMatchedIndex, mark, matchedTalk = cs.Handler.matcher.MatchOutboundTalk(ctx, cs.Handler.replayingSession, -1, request)
		}
		if matchedTalk == nil {
			callOutbound.MatchedRequest = nil
			callOutbound.MatchedResponse = nil
			callOutbound.MatchedActionIndex = fakeIndexNotMatched
		} else {
			callOutbound.MatchedRequest = matchedTalk.Request
			callOutbound.MatchedResponse = matchedTalk.Response
			callOutbound.MatchedActionIndex = matchedTalk.ActionIndex
		}
		callOutbound.MatchedMark = mark

		if matchedTalk == nil {
			cs.Handler.replayedSession.Outbounds = append(cs.Handler.replayedSession.Outbounds, callOutbound)
			tlog.Handler.Warnf(ctx, tlog.DebugTag, "errmsg=find matching talk failed||request=%s||traceID=%s", quotedRequest, string(cs.traceID))
			return nil
		}
		response := callOutbound.MatchedResponse
		response = bytes.Replace(response, []byte("Connection: keep-alive\r\n"), []byte("Connection: close\r\n"), -1)
		response = resetContentLength(ctx, response)
		_, err := cs.conn.Write(response)
		if err != nil {
			tlog.Handler.Errorf(ctx, tlog.DebugTag, "errmsg=write back response failed||err=%s", err)
			return err
		}
	}
	//TODO: 为啥不直接放到上面if的elseif里
	// set matched id as ActionIndex for simulateHttp|simulateMysql
	if callOutbound.MatchedActionIndex < 0 && matchedTalk != nil {
		callOutbound.MatchedActionIndex = matchedTalk.ActionIndex
		callOutbound.MatchedRequest = matchedTalk.Request
	}

	// 去掉COM_STMT_CLOSE
	if callOutbound.MatchedRequest != nil {
		callOutbound.MatchedRequest = removeMysqlStmtClose(callOutbound.MatchedRequest)
	}

	cs.Handler.replayedSession.Outbounds = append(cs.Handler.replayedSession.Outbounds, callOutbound)

	tlog.Handler.Infof(ctx, tlog.DebugTag,
		"%s||to=%s||actionId=%s||matchedActionIndex=%v||matchedResponse=%s",
		//"matchedMark", mark,
		helper.CInfo(">>>response of outbound>>>"),
		cs.tcpAddr.String(),
		callOutbound.ActionId,
		callOutbound.MatchedActionIndex,
		strconv.Quote(helper.BytesToString(callOutbound.MatchedResponse)))

	return nil
}

// resetContentLength, 重新计算
func resetContentLength(ctx context.Context, data []byte) []byte {
	var contents [][]byte

	if !bytes.Contains(data, []byte("Content-Encoding: gzip\r\n")) {
		return data
	}

	bodySplit := []byte("\r\n\r\n")
	if contents = bytes.Split(data, bodySplit); len(contents) != 2 {
		return data
	}

	// 因为线上gzip的原因，Content-Length可能会减少
	newLength := fmt.Sprintf("Content-Length: %d\r\n", len(contents[1])) // 计算body长度
	data = contentLengthRegex.ReplaceAll(data, []byte(newLength))
	data = bytes.Replace(data, []byte("Content-Encoding: gzip\r\n"), []byte(""), -1)

	return data
}

// applySimulation
func applySimulation(ctx context.Context, sim func(ctx context.Context, request []byte) []byte,
	request []byte, conn net.Conn, callOutbound *replaying.CallOutbound) error {

	resp := sim(ctx, request) // mysql connection setup might not in the recorded session
	if resp != nil {
		if callOutbound != nil {
			callOutbound.MatchedRequest = request
			callOutbound.MatchedActionIndex = fakeIndexSimulated // to be ignored
			callOutbound.MatchedResponse = resp
		}
		_, err := conn.Write(resp)
		if err != nil {
			tlog.Handler.Errorf(ctx, tlog.DebugTag, "errmsg=write back simulated response failed||err=%s", err)
			return err
		}
		return nil
	}
	return nil
}

// removeMysqlStmtClose COM_STMT_CLOSE经常和其它的包混在一起，统一去掉
func removeMysqlStmtClose(request []byte) []byte {
	if bytes.HasPrefix(request, []byte{0x5, 0x0, 0x0, 0x0, 0x19}) && len(request) >= 9 {
		request = request[9:]
	}
	return request
}

// terminated 终止请求
func terminated(request []byte) bool {
	// mysql close handshake
	if bytes.Equal(request, []byte{0x1, 0x0, 0x0, 0x0, 0x1}) {
		return true
	}
	// mysql close handshake
	if bytes.Equal(request, []byte{0x1, 0x0, 0x0, 0x0, 0x9}) {
		return true
	}
	return false
}
