// +build recorder

package sharingan

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/didi/sharingan/proto"
	"github.com/didi/sharingan/recorder/hook"
	"github.com/didi/sharingan/recorder/logger"
	"github.com/didi/sharingan/recorder/recording"
	"github.com/didi/sharingan/recorder/utils"

	"github.com/v2pro/plz/countlog"
	"google.golang.org/grpc"
)

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return runtime.GetCurrentGoRoutineId()
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
	runtime.SetDelegatedFromGoRoutineId(gID)
}

func init() {
	if os.Getenv("RECORDER_ENABLED") != "true" {
		return
	}

	// set recorder
	recorder := recording.NewAsyncRecorder(NewDefaultRecorder())
	recorder.Context = context.Background()
	recorder.Start()
	recording.Recorders = append(recording.Recorders, recorder)

	// setup logger
	logger.Setup()

	// start hook
	hook.Start()

	// log
	log.Println("mode", "=====recorder=====")
}

/* =================DefaultRecorder================= */

// DefaultRecorder DefaultRecorder
type DefaultRecorder struct {
	hostname    string
	localDir    string
	localFile   string
	agentURL    string
	grpcAddr    string
	agentClient *http.Client
	grpcClient  proto.AgentClient
}

// NewDefaultRecorder NewDefaultRecorder
func NewDefaultRecorder() recording.Recorder {
	hostname, _ := os.Hostname()

	recorder := &DefaultRecorder{
		hostname:  hostname,
		localDir:  os.Getenv("RECORDER_TO_DIR"),
		localFile: os.Getenv("RECORDER_TO_FILE"),
		grpcAddr:  os.Getenv("RECORDER_TO_AGENT"),
	}

	// set grpcClient
	if recorder.grpcAddr != "" {
		grpcConn, err := grpc.Dial(recorder.grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
		if err != nil {
			log.Fatal(err)
		}
		recorder.grpcClient = proto.NewAgentClient(grpcConn)
	}

	return recorder
}

// Record 自定义录制方法
func (r *DefaultRecorder) Record(session *recording.Session) {
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
		session.TraceId = http.Header["xxx-header-traceid"] // 需要自定义
		session.SpanId = http.Header["xxx-header-spanid"]   // 需要自定义
	}

	// set hostname
	{
		session.Context += r.hostname + " "
	}

	// Marshal session
	{
		if b, err = json.Marshal(session); err != nil {
			countlog.Error("event!recorder.failed to marshal session", "err", err, "session", session)
			return
		}
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
	if r.grpcAddr != "" {
		SetDelegatedFromGoRoutineID(-1)
		defer SetDelegatedFromGoRoutineID(0)

		req := &proto.RecordReq{EsData: string(b)}
		_, err := r.grpcClient.Record(context.Background(), req)
		if err != nil {
			countlog.Warn("event!recorder.failed to record to agent", "err", err, "agentURL", r.agentURL)
			return
		}
	}

	// 默认console输出
	if r.localDir == "" && r.localFile == "" && r.agentURL == "" && r.grpcAddr == "" {
		log.Println(string(b))
	}
}
