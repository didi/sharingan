package recorder

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/didi/sharingan/recorder/koala_grpc/recording"
	"github.com/didi/sharingan/recorder/utils"

	"github.com/v2pro/plz/countlog"
)

type recorder_grpc struct {
	hostname    string
	localDir    string
	localFile   string
	esURL       string
	agentURL    string
	agentClient *http.Client
}

// NewRecorderGrpc NewRecorderGrpc
func NewRecorderGrpc() recording.Recorder {
	hostname, _ := os.Hostname()
	recorder := &recorder_grpc{
		hostname:  hostname,
		localDir:  os.Getenv("RECORDER_TO_DIR"),
		localFile: os.Getenv("RECORDER_TO_FILE"),
		esURL:     os.Getenv("RECORDER_TO_ES"),
		agentURL:  os.Getenv("RECORDER_TO_AGENT"),
		agentClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       5 * time.Second, // default: 90s
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: time.Second * 1,
		},
	}

	return recorder
}

// Record 自定义录制方法
func (r *recorder_grpc) Record(session *recording.Session) {
	defer func() {
		if err := recover(); err != nil {
			countlog.Fatal("event!kafka_recorder.panic", "err", err,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()

	var (
		b   []byte // 录制的流量
		err error
	)

	// set trace Info
	{
		http := utils.NewHTTP()
		http.ParseRequest(session.CallFromInbound.Request)
		session.TraceId = http.Header["xxx-header-traceid"] // can custom
		session.SpanId = http.Header["xxx-header-spanid"]   // can custom
	}

	// set hostname
	{
		session.Context += r.hostname + " "
	}

	// Marshal session to b
	if b, err = json.Marshal(session); err != nil {
		countlog.Error("event!recorder.failed to marshal session", "err", err, "session", session)
		return
	}

	// 方式一、本地目录存储录制流量
	if r.localDir != "" {
		err := ioutil.WriteFile(path.Join(r.localDir, session.SessionId), b, 0666)
		if err != nil {
			countlog.Error("event!recorder.failed to record to localDir", "err", err, "localDir", r.localDir)
			return
		}
	}

	// 方式二、本地文件存储录制流量
	if r.localFile != "" {
		f, err := os.OpenFile(r.localFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			countlog.Error("event!recorder.failed to record to localFile", "err", err, "localFile", r.localFile)
			return
		}
		defer f.Close()
		f.Write(b)
		f.WriteString("\n")
	}

	// 方式三、远程agent接受流量【推荐，可以控制录制比例】
	if r.agentURL != "" {
		SetDelegatedFromGoRoutineID(-1)
		defer SetDelegatedFromGoRoutineID(0)

		resp, err := r.agentClient.Post(r.agentURL, "application/json", bytes.NewBuffer(b))
		if err != nil {
			countlog.Warn("event!recorder.failed to record to agent", "err", err, "agentURL", r.agentURL)
			return
		}
		resp.Body.Close()
	}

	// 方式四、直接发送ES存储
	if r.esURL != "" {
		var b []byte

		if b, err = json.Marshal(session); err != nil {
			countlog.Error("event!recorder.failed to marshal session", "err", err, "session", session)
			return
		}

		SetDelegatedFromGoRoutineID(-1)
		defer SetDelegatedFromGoRoutineID(0)

		resp, err := r.agentClient.Post(r.esURL, "application/json", bytes.NewBuffer(b))
		if err != nil {
			countlog.Warn("event!recorder.failed to record to es", "err", err, "esURL", r.esURL)
			return
		}
		resp.Body.Close()
	}

	// 默认console输出
	if r.localDir == "" && r.localFile == "" && r.agentURL == "" && r.esURL == "" {
		log.Println(string(b))
	}
}

// ShouldRecordActionGrpc 自定义过滤方法
func ShouldRecordActionGrpc(action recording.Action) bool {
	if action == nil {
		return false
	}

	switch act := action.(type) {
	case *recording.AppendFile:
		if !strings.Contains(act.FileName, "/public.log") {
			return false
		}
	case *recording.SendUDP:
		if !(act.Peer.IP.String() == "127.0.0.1" && act.Peer.Port == 9891) {
			return false
		}
	}

	return true
}
