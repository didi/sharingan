package replayed

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/didi/sharingan/replayer-agent/common/handlers/ignore"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/model/nuwaplt"
	"github.com/didi/sharingan/replayer-agent/model/pool"
	"github.com/didi/sharingan/replayer-agent/model/protocol"
	"github.com/didi/sharingan/replayer-agent/model/recording"
	"github.com/didi/sharingan/replayer-agent/model/replaying"
	"github.com/didi/sharingan/replayer-agent/utils/helper"
)

const (
	ReqStatusIgnored    = "ignored..."
	ReqStatusNotMatched = "not matched..."
	ReqStatusMissed     = "missed..."

	LogStatusRepeated = "match repeated"
	LogStatusMissed   = "match missed"
)

type Session struct {
	SessionId         string
	Context           string
	Request           []byte
	OnlineResponse    []byte
	TestResponse      []byte
	Outbounds         []*replaying.CallOutbound
	OnlineOutbounds   []*recording.CallOutbound
	OnlineAppendFiles []*recording.AppendFile
}

type DiffRecord struct {
	Id              int           `json:"id"`
	Project         string        `json:"project"`
	RequestMark     string        `json:"requestMark"`
	Noise           string        `json:"noise"`
	Diff            string        `json:"diff"`
	FormatDiff      []*FormatDiff `json:"formatDiff"`
	OnlineReq       string        `json:"onlineReq"`
	BinaryOnlineReq string        `json:"binaryOnlineReq"`
	TestReq         string        `json:"testReq"`
	BinaryTestReq   string        `json:"binaryTestReq"`
	OnlineRes       string        `json:"onlineRes"`
	BinaryOnlineRes string        `json:"binaryOnlineRes"`
	TestRes         string        `json:"testRes"`
	BinaryTestRes   string        `json:"binaryTestRes"`
	MockedRes       string        `json:"mockedRes"`
	BinaryMockedRes string        `json:"binaryMockedRes"`
	IsDiff          int           `json:"isDiff"`
	Protocol        string        `json:"protocol"`
	ScorePercentage string        `json:"scorePercentage"`
	NoWebDisplay    bool          `json:"noWebDisplay"`
	MatchedIndex    int           `json:"matchedIndex"`
}

/**
 * 回放结果进行diff、binary、missing、extend转换
 *
 * @Return
 */
func DiffReplayed(ctx context.Context, sess *Session, project string) []*DiffRecord {
	if sess == nil {
		return make([]*DiffRecord, 0)
	}
	// 项目名
	//project, _ := nuwaplt.Host2Module[sess.Context]
	// 接口uri
	requestMarkDiff := &Diff{A: helper.BytesToString(sess.Request), B: ""}
	_, requestMark, _, _ := requestMarkDiff.CompareProtocol()
	//远程获取噪音数据
	noiseInfo := nuwaplt.GetNoise(project, requestMark)
	// 已匹配outbound信息
	matchedIndex := make(map[int]bool)

	c := &Composer{
		Project:      project,
		RequestMark:  requestMark,
		NoiseInfo:    noiseInfo,
		MatchedIndex: matchedIndex,
		Sess:         sess,
	}

	// 防止后续Outbounds追加数据引发panic
	outbounds := sess.Outbounds

	ajaxs := make([]*DiffRecord, len(outbounds)+1)
	// 等待所有inbound和outbound比对结束
	c.Add(len(outbounds) + 1)
	c.DiffInbound(ctx, ajaxs)
	if len(outbounds) > 0 {
		// requests send by sut
		for i := range outbounds {
			c.DiffOutbounds(ctx, ajaxs, i)
		}
	}
	c.Wait()

	// public log
	for i := range sess.OnlineAppendFiles {
		c.addPublicLog(i)
	}
	for i := range sess.OnlineAppendFiles {
		ajaxs = c.DiffAppendFile(ctx, ajaxs, i)
	}

	// fill the missing requests
	for i := range sess.OnlineOutbounds {
		ajaxs = c.DiffOther(ctx, ajaxs, i)
	}
	return ajaxs
}

func Judge(diffs []*DiffRecord) bool {
	for _, diff := range diffs {
		if diff.IsDiff == 2 {
			return false
		}
	}
	return true
}

type Composer struct {
	// guard MatchedIndex
	sync.RWMutex
	// control workflow
	sync.WaitGroup

	Project      string
	RequestMark  string
	NoiseInfo    map[string]nuwaplt.NoiseInfo
	MatchedIndex map[int]bool
	PublogCnt    map[string]int

	Sess *Session
}

func (c *Composer) GetMatchedIndex(id int) bool {
	c.RLock()
	defer c.RUnlock()

	val, ok := c.MatchedIndex[id]
	return ok && val
}

func (c *Composer) SetMatchedIndex(id int) {
	c.Lock()
	defer c.Unlock()

	c.MatchedIndex[id] = true
}

func (c *Composer) DiffInbound(ctx context.Context, ajaxs []*DiffRecord) {
	defer func() {
		if err := recover(); err != nil {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "panic in %s goroutine||errmsg=%s||stack info=%s", "DiffInbounds", err, strconv.Quote(string(debug.Stack())))
		}
	}()
	defer c.Done()
	ajax := new(DiffRecord)
	// 项目名称:passenger,driver...
	ajax.Id = 0
	ajax.Project = c.Project
	ajax.MatchedIndex = -1

	unzipOnlineResponse := UnzipHttpRepsonse(ctx, c.Sess.OnlineResponse)
	unzipTestResponse := UnzipHttpRepsonse(ctx, c.Sess.TestResponse)

	bytesDiff := &Diff{
		A:                          helper.BytesToString(unzipOnlineResponse),
		B:                          helper.BytesToString(unzipTestResponse),
		Noise:                      c.NoiseInfo,
		CallFromInboundRequestMark: c.RequestMark,
	}

	formatDiff, _, diffErr, protocol := bytesDiff.CompareProtocol()
	if diffErr == HasDiffErr {
		ajax.IsDiff = 2
	} else if diffErr == HasDiffButIgnoreErr {
		ajax.IsDiff = 1
	}
	ajax.RequestMark = c.RequestMark
	ajax.Noise = c.RequestMark
	ajax.FormatDiff = formatDiff

	ajax.OnlineReq = helper.BytesToString(c.Sess.Request)
	binaryOnlineReq := base64.StdEncoding.EncodeToString(c.Sess.Request)
	ajax.BinaryOnlineReq = binaryOnlineReq

	ajax.TestRes = helper.BytesToString(unzipTestResponse)
	binaryTestRes := base64.StdEncoding.EncodeToString(unzipTestResponse)
	ajax.BinaryTestRes = binaryTestRes

	ajax.OnlineRes = helper.BytesToString(unzipOnlineResponse)
	binaryOnlineRes := base64.StdEncoding.EncodeToString(unzipOnlineResponse)
	ajax.BinaryOnlineRes = binaryOnlineRes
	ajax.Protocol = protocol

	ajaxs[0] = ajax
}

// UnzipHttpRepsonse 尝试解压gzip数据，忽略失败
func UnzipHttpRepsonse(ctx context.Context, data []byte) []byte {
	var err error
	var contents [][]byte

	bodySplit := []byte("\r\n\r\n")

	if !bytes.Contains(data, []byte("Content-Encoding: gzip")) {
		return data
	}

	if contents = bytes.Split(data, bodySplit); len(contents) != 2 {
		return data
	}

	if contents[1], err = ParseGzip(ctx, contents[1]); err == nil {
		return bytes.Join(contents, bodySplit)
	}

	return data
}

// ParseGzip 解析gzip数据
func ParseGzip(ctx context.Context, data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, data)

	r, err := gzip.NewReader(b)
	if err != nil {
		tlog.Handler.Debugf(ctx, tlog.DebugTag, "[ParseGzip] NewReader error: %v, maybe data is ungzip", err)
		return nil, err
	}
	defer r.Close()

	undatas, err := ioutil.ReadAll(r)
	if err != nil {
		tlog.Handler.Debugf(ctx, tlog.DebugTag, "[ParseGzip] ioutil.ReadAll error: %v", err)
		return nil, err
	}

	return undatas, nil
}

func (c *Composer) DiffOutbounds(ctx context.Context, ajaxs []*DiffRecord, cnt int) {
	defer func() {
		if err := recover(); err != nil {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "panic in %s goroutine||errmsg=%s||stack info=%s", "DiffOutbounds", err, strconv.Quote(string(debug.Stack())))
		}
	}()
	defer c.Done()

	ajax := new(DiffRecord)
	ajax.Id = cnt + 1
	ajax.Project = c.Project

	out := c.Sess.Outbounds[cnt]
	// put back to buffer pool
	defer pool.PutBuf(out.Request)

	if out.MatchedActionIndex < 0 {
		if mockNotMatched(out, "", ajax.Project) {
			setOnlineStatus(ajax, 1, ReqStatusIgnored)
			// 不展示ignore.NotMatchedNoise导致的ignore
			ajax.NoWebDisplay = true
		} else {
			setOnlineStatus(ajax, 2, ReqStatusNotMatched)
			//增加not match请求的协议信息,用于回放结果页的展示
			bytesDiff := &Diff{A: "", B: helper.BytesToString(out.Request), Noise: nil}
			_, _, _, ajax.Protocol = bytesDiff.ParseProtocol(bytesDiff.B)
			ajax.ScorePercentage = "0%"
			ajax.MatchedIndex = -1
		}
	} else {
		if c.GetMatchedIndex(out.MatchedActionIndex) {
			tlog.Handler.Debugf(ctx, tlog.DebugTag, "%s||outboundIndex=%v||actionIndex=%v", helper.CInfo(LogStatusRepeated), cnt, out.MatchedActionIndex)
		}
		c.SetMatchedIndex(out.MatchedActionIndex)
		ajax.MatchedIndex = out.MatchedActionIndex

		ajax.OnlineReq = helper.BytesToString(out.MatchedRequest)
		binaryOnlineReq := base64.StdEncoding.EncodeToString(out.MatchedRequest)
		ajax.BinaryOnlineReq = binaryOnlineReq

		ajax.OnlineRes = helper.BytesToString(out.MatchedResponse)
		binaryOnlineRes := base64.StdEncoding.EncodeToString(out.MatchedResponse)
		ajax.BinaryOnlineRes = binaryOnlineRes

		onlineHttpHeader, testHttpHeader := "", ""
		if strings.HasPrefix(ajax.OnlineRes, "HTTP/1") && !strings.Contains(ajax.OnlineReq, "HTTP/1") {
			onlineHttpHeader, testHttpHeader = c.getLatestHTTPHeader(cnt)
		}
		bytesDiff := &Diff{A: onlineHttpHeader + helper.BytesToString(out.MatchedRequest), B: testHttpHeader + helper.BytesToString(out.Request), Noise: c.NoiseInfo}
		formatDiff, noise, diffErr, protocol := bytesDiff.CompareProtocol()
		if diffErr == HasDiffErr {
			ajax.IsDiff = 2
		} else if diffErr == HasDiffButIgnoreErr {
			ajax.IsDiff = 1
		}
		ajax.RequestMark = c.RequestMark
		ajax.Noise = noise
		ajax.FormatDiff = formatDiff
		ajax.Protocol = protocol
		ajax.ScorePercentage = strconv.Itoa(int(out.MatchedMark*100/2)) + "%"
	}

	ajax.TestReq = helper.BytesToString(out.Request)
	binaryTestReq := base64.StdEncoding.EncodeToString(out.Request)
	ajax.BinaryTestReq = binaryTestReq

	ajax.MockedRes = helper.BytesToString(out.MockedResponse)
	binaryMockedRes := base64.StdEncoding.EncodeToString(out.MockedResponse)
	ajax.BinaryMockedRes = binaryMockedRes

	// 优化相似度打分
	len1 := len(ajax.OnlineReq)
	len2 := len(ajax.TestReq)
	if ajax.OnlineReq == ajax.TestReq {
		ajax.ScorePercentage = "100%"
	} else if strings.Contains(ajax.OnlineReq, ajax.TestReq) || strings.Contains(ajax.TestReq, ajax.OnlineReq) {
		if len1 > len2 {
			ajax.ScorePercentage = strconv.Itoa(len2*100/len1) + "%"
		} else {
			ajax.ScorePercentage = strconv.Itoa(len1*100/len2) + "%"
		}
	} else if ajax.Protocol == protocol.MYSQL_PRO {
		if len1 > 5 && len2 > 5 {
			str1 := ajax.OnlineReq[5:]
			str2 := ajax.TestReq[5:]
			if len1 >= len2 && strings.Contains(str1, str2) {
				ajax.ScorePercentage = strconv.Itoa(len2*100/len1) + "%"
			} else if len2 >= len1 && strings.Contains(str2, str1) {
				ajax.ScorePercentage = strconv.Itoa(len1*100/len2) + "%"
			}
		}
	}

	// 优化Req展示
	ajax.OnlineReq = optimizeDisReq(ajax.Protocol, out.MatchedRequest)
	ajax.TestReq = optimizeDisReq(ajax.Protocol, out.Request)

	ajaxs[cnt+1] = ajax
}

func (c *Composer) getLatestHTTPHeader(cnt int) (string, string) {
	for cnt > 0 {
		out := c.Sess.Outbounds[cnt-1]
		if strings.Contains(helper.BytesToString(out.MatchedResponse), "100 Continue") && strings.Contains(helper.BytesToString(out.MatchedRequest), "100-continue") {
			return helper.BytesToString(out.MatchedRequest), helper.BytesToString(out.Request)
		}
		cnt--
	}
	return "", ""
}

func (c *Composer) addPublicLog(cnt int) {
	opKey, err := extractPublicKey(helper.BytesToString(c.Sess.OnlineAppendFiles[cnt].Content), "opera_stat_key=")
	if err != nil {
		return
	}
	if c.PublogCnt == nil {
		c.PublogCnt = make(map[string]int)
	}
	val, ok := c.PublogCnt[opKey]
	if ok {
		c.PublogCnt[opKey] = val + 1
	} else {
		c.PublogCnt[opKey] = 1
	}
}

func (c *Composer) DiffAppendFile(ctx context.Context, ajaxs []*DiffRecord, cnt int) []*DiffRecord {
	onlinePublic := helper.BytesToString(c.Sess.OnlineAppendFiles[cnt].Content)
	testPublic, err := c.getTestPublic(onlinePublic)
	if err != nil && err == PublicLogNotDefinedErr {
		return ajaxs
	}

	ajax := new(DiffRecord)
	ajax.Id = len(ajaxs)
	ajax.Project = c.Project
	ajax.MatchedIndex = -1

	bytesDiff := &Diff{
		A:                          onlinePublic,
		B:                          testPublic,
		Noise:                      c.NoiseInfo,
		CallFromInboundRequestMark: c.RequestMark,
	}
	formatDiff, noise, diffErr, protocol := bytesDiff.CompareProtocol()
	if diffErr == HasDiffButIgnoreErr {
		ajax.IsDiff = 1
	} else if diffErr != nil {
		ajax.IsDiff = 2
	}

	// filter out request mark
	for i := 0; i < len(formatDiff); i++ {
		if formatDiff[i].Key == "out.req" {
			if i < len(formatDiff)-1 {
				copy(formatDiff[i:], formatDiff[i+1:])
			}
			formatDiff = formatDiff[:len(formatDiff)-1]
			break
		}
	}

	ajax.RequestMark = c.RequestMark
	ajax.Noise = noise
	ajax.FormatDiff = formatDiff

	ajax.OnlineReq = onlinePublic
	binaryOnlineReq := base64.StdEncoding.EncodeToString(helper.StringToBytes(onlinePublic))
	ajax.BinaryOnlineReq = binaryOnlineReq

	ajax.TestReq = testPublic
	binaryTestReq := base64.StdEncoding.EncodeToString(helper.StringToBytes(testPublic))
	ajax.BinaryTestReq = binaryTestReq

	ajax.Protocol = protocol
	//ajax.ScorePercentage =

	return append(ajaxs, ajax)
}

func (c *Composer) getTestPublic(online string) (string, error) {
	// key of public log
	traceId, err := extractPublicKey(online, "trace_id=")
	if err != nil {
		return "trace_id is not exist", err
	}
	opKey, err := extractPublicKey(online, "opera_stat_key=")
	if err != nil {
		return "opera_stat_key is not exist", err
	}

	// path of public log
	logPath := nuwaplt.GetValueWithProject(c.Project, nuwaplt.KLogPath, "")
	if logPath == "" {
		return "logpath is not set", PublicLogPathErr
	}

	cnt, ok := c.PublogCnt[opKey]
	if !ok {
		return "", PublicLogKeyNotDefinedErr
	}
	c.PublogCnt[opKey] = cnt - 1

	// offline public log
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("grep -rw %s %s/public.log | grep %s | tail -n %d | head -n 1", traceId, logPath, opKey, cnt))
	out, err := cmd.Output()
	return strings.TrimSpace(helper.BytesToString(out)), err
}

func (c *Composer) DiffOther(ctx context.Context, ajaxs []*DiffRecord, cnt int) []*DiffRecord {
	out := c.Sess.OnlineOutbounds[cnt]
	// put back to buffer pool
	defer pool.PutBuf(out.Request)

	// ignore Online Request
	if c.ignoreOnlineRequest(out) {
		return ajaxs
	}

	if c.GetMatchedIndex(out.ActionIndex) {
		return ajaxs
	}

	if match(ignore.NotMatchedNoise, "", helper.BytesToString(out.Request)) {
		out.IgnoreFlag = true
	}

	tlog.Handler.Debugf(ctx, tlog.DebugTag, "%s||actionIndex=%v", helper.CInfo(LogStatusMissed), out.ActionIndex)

	ajax := new(DiffRecord)
	ajax.Id = len(ajaxs)
	ajax.Project = c.Project
	ajax.MatchedIndex = -1

	ajax.OnlineReq = helper.BytesToString(out.Request)
	binaryOnlineReq := base64.StdEncoding.EncodeToString(out.Request)
	ajax.BinaryOnlineReq = binaryOnlineReq
	// 优化Req展示
	ajax.OnlineReq = optimizeDisReq(ajax.Protocol, out.Request)

	ajax.OnlineRes = helper.BytesToString(out.Response)
	binaryOnlineRes := base64.StdEncoding.EncodeToString(out.Response)
	ajax.BinaryOnlineRes = binaryOnlineRes

	if out.IgnoreFlag || mockOutbound(out, c.RequestMark) {
		setTestStatus(ajax, 1, ReqStatusIgnored)
		return append(ajaxs, ajax)
	}

	if len(out.Request) != 0 {
		setTestStatus(ajax, 2, ReqStatusMissed)
		//增加missed...请求的协议信息,用于回放结果页的展示
		bytesDiff := &Diff{A: helper.BytesToString(out.Request), B: "", Noise: nil}
		_, _, _, ajax.Protocol = bytesDiff.ParseProtocol(bytesDiff.A)
		ajax.ScorePercentage = "0%"
		return append(ajaxs, ajax)
	}
	return ajaxs
}

var limit = make(chan int, 1)

func init() {
	limit <- 1
}

func extractPublicKey(public string, keyfmt string) (string, error) {
	idx := strings.Index(public, keyfmt)
	if idx == -1 {
		return "", PublicLogFormatErr
	}
	key := public[idx:]
	idx = strings.Index(key, "||")
	if idx == -1 {
		idx = len(key)
	}
	key = key[:idx]
	return key, nil
}

// 十六进制后diff整合
func BinaryDiff(a, b []byte) ([]byte, error) {
	<-limit
	aPath := "/tmp/fd_a.bin"
	bPath := "/tmp/fd_b.bin"
	ioutil.WriteFile(aPath, a, 0777)
	ioutil.WriteFile(bPath, b, 0777)
	cmd := exec.Command("/bin/bash", "-c", "diff -U 100 <(xxd -c24 "+
		aPath+" | cut -c 10-) <(xxd -c24 "+bPath+" | cut -c 10-)")
	res, err := cmd.CombinedOutput()
	limit <- 1
	return res, err
}

// 十六进制转换
func ToBinary(text []byte) ([]byte, error) {
	binary := "/tmp/fd_binary.bin"
	err := ioutil.WriteFile(binary, text, 0777)
	if err != nil {
		return []byte{}, err
	}
	cmd := exec.Command("/bin/bash", "-c", "xxd -c24 "+binary+" | cut -c 10-")
	return cmd.Output()
}

func mockNotMatched(out *replaying.CallOutbound, scope, project string) bool {
	return match(ignore.NotMatchedNoise, scope, helper.BytesToString(out.Request))
}

func mockOutbound(out *recording.CallOutbound, scope string) bool {
	return match(ignore.OutboundNoise, scope, helper.BytesToString(out.Request))
}

func match(noises map[string]ignore.NoiseMeta, scope, request string) bool {
	for noise, meta := range noises {
		if meta.Scope != "" && meta.Scope != scope {
			continue
		}
		switch meta.Typ {
		case ignore.NoiseMatch:
			if request == noise {
				return true
			}
		case ignore.NoisePrefix:
			if strings.HasPrefix(request, noise) {
				return true
			}
		case ignore.NoiseSuffix:
			if strings.HasSuffix(request, noise) {
				return true
			}
		case ignore.NoiseContains:
			if strings.Contains(request, noise) {
				return true
			}
		}
	}
	return false
}

func setOnlineStatus(diff *DiffRecord, isDiff int, content string) {
	diff.IsDiff = isDiff
	diff.OnlineReq = content
	diff.BinaryOnlineReq = content
	diff.OnlineRes = content
	diff.BinaryOnlineRes = content
}

func setTestStatus(diff *DiffRecord, isDiff int, content string) {
	diff.IsDiff = isDiff
	diff.TestReq = content
	diff.BinaryTestReq = content
}

func (c *Composer) ignoreOnlineRequest(out *recording.CallOutbound) bool {
	// ignore dns outbounds
	if out.Peer.IP.String() == "127.0.0.1" && out.Peer.Port == 53 {
		return true
	}

	// ignore redis ping
	if bytes.Equal(out.Request, []byte("*1\r\n$4\r\nPING\r\n")) {
		return true
	}

	// ignore mysql COM_STMT_CLOSE
	if bytes.HasPrefix(out.Request, []byte{0x5, 0x0, 0x0, 0x0, 0x19}) && len(out.Request) == 9 {
		return true
	}

	// ignore mysql mysql_native_password
	if bytes.Contains(out.Request, []byte("mysql_native_password")) {
		return true
	}

	// ignore mysql SET autocommit=1
	if bytes.Contains(out.Request, []byte("SET autocommit=1")) {
		return true
	}

	// ignore mysql SET NAMES utf8
	if bytes.Contains(out.Request, []byte("SET NAMES utf8")) {
		return true
	}

	// ignore data what is mysql response like 'set names utf8'
	if bytes.Equal(out.Response, []byte{0x07, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}) {
		return true
	}

	// ignore mysql request quit
	if bytes.Equal(out.Request, []byte{0x01, 0x00, 0x00, 0x00, 0x01}) {
		return true
	}

	return false
}
