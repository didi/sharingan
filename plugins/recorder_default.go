package plugins

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/didi/sharingan/recorder"
	"github.com/didi/sharingan/recorder/koala/recording"
	"github.com/didi/sharingan/recorder/utils"

	"github.com/v2pro/plz/countlog"
)

// NewDefaultRecorder New DefaultRecorder
func NewDefaultRecorder() recording.Recorder {
	hostname, _ := os.Hostname()

	agentClient := &http.Client{
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
	}

	recorder := &DefaultRecorder{
		hostname:    hostname,
		localDir:    os.Getenv("RECORDER_TO_DIR"),
		localFile:   os.Getenv("RECORDER_TO_FILE"),
		agentAddr:   os.Getenv("RECORDER_TO_AGENT"),
		agentClient: agentClient,
	}

	return recorder
}

// DefaultRecorder DefaultRecorder recorder plugin
type DefaultRecorder struct {
	hostname    string
	localDir    string
	localFile   string
	agentAddr   string
	agentClient *http.Client
}

// Record implement Record
func (r *DefaultRecorder) Record(session *recording.Session) {
	defer func() {
		if err := recover(); err != nil {
			countlog.LogPanic(err, "kafka_recorder.panic")
		}
	}()

	// set trace Info
	{
		http := utils.NewHTTP()
		http.ParseRequest(session.CallFromInbound.Request)
		session.TraceID = http.Header["xxx-header-traceid"] // can custom
		session.SpanID = http.Header["xxx-header-spanid"]   // can custom
	}

	// set hostname
	{
		session.Context += r.hostname + " "
	}

	var (
		b   []byte // session bytes which tobe recorder
		err error
	)

	// Marshal session to b
	if b, err = json.Marshal(session); err != nil {
		countlog.Error("event!recorder.failed to marshal session", "err", err, "session", session)
		return
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

	// case 3: recorder to remote agent [to recorder-agent or ES]
	if r.agentAddr != "" {
		recorder.SetDelegatedFromGoRoutineID(-1)
		defer recorder.SetDelegatedFromGoRoutineID(0)

		resp, err := r.agentClient.Post(r.agentAddr, "application/json", bytes.NewBuffer(b))
		if err != nil {
			countlog.Warn("event!recorder.failed to record to agent", "err", err, "agentAddr", r.agentAddr)
			return
		}
		resp.Body.Close()
	}

	// default console output
	if r.localDir == "" && r.localFile == "" && r.agentAddr == "" {
		log.Println(string(b))
	}
}
