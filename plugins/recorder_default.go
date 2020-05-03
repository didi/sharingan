package plugins

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/didi/sharingan/plugins/proto"
	"github.com/didi/sharingan/recorder"
	"github.com/didi/sharingan/recorder/recording"
	"github.com/didi/sharingan/recorder/utils"

	"github.com/v2pro/plz/countlog"
	"google.golang.org/grpc"
)

// NewDefaultRecorder New DefaultRecorder
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

// DefaultRecorder DefaultRecorder recorder plugin
type DefaultRecorder struct {
	hostname   string
	localDir   string
	localFile  string
	grpcAddr   string
	grpcClient proto.AgentClient
}

// Record implement Record
func (r *DefaultRecorder) Record(session *recording.Session) {
	defer func() {
		if err := recover(); err != nil {
			countlog.Fatal("event!kafka_recorder.panic", "err", err,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()

	var (
		b   []byte // record session byte
		err error
	)

	// set trace Info
	{
		http := utils.NewHTTP()
		http.ParseRequest(session.CallFromInbound.Request)
		session.TraceID = http.Header["xxx-header-traceid"] // cam custom
		session.SpanID = http.Header["xxx-header-spanid"]   // can custom
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

	// case 1: recorder to local dir [offline use]
	if r.localDir != "" {
		err := ioutil.WriteFile(path.Join(r.localDir, session.SessionID), b, 0666)
		if err != nil {
			countlog.Error("event!recorder.failed to record to localDir", "err", err, "localDir", r.localDir)
			return
		}
	}

	// case 2: recorder to local file [offline use]
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

	// case 3: recorder to remote agent [online use, recommend]
	if r.grpcAddr != "" {
		recorder.SetDelegatedFromGoRoutineID(-1)
		defer recorder.SetDelegatedFromGoRoutineID(0)

		req := &proto.RecordReq{EsData: string(b)}
		_, err := r.grpcClient.Record(context.Background(), req)
		if err != nil {
			countlog.Warn("event!recorder.failed to record to agent", "err", err, "grpcAddr", r.grpcAddr)
			return
		}
	}

	// default console output
	if r.localDir == "" && r.localFile == "" && r.grpcAddr == "" {
		log.Println(string(b))
	}
}
