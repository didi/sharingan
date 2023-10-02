package recording

import (
	"context"

	"github.com/v2pro/plz/countlog"
)

// AsyncRecorder AsyncRecorder
type AsyncRecorder struct {
	Context      context.Context
	realRecorder Recorder
	recordChan   chan *Session
}

// NewAsyncRecorder NewAsyncRecorder
func NewAsyncRecorder(realRecorder Recorder) *AsyncRecorder {
	return &AsyncRecorder{
		recordChan:   make(chan *Session, 100),
		realRecorder: realRecorder,
	}
}

// Start Start
func (recorder *AsyncRecorder) Start() {
	go recorder.backgroundRecord()
}

// backgroundRecord backgroundRecord
func (recorder *AsyncRecorder) backgroundRecord() {
	defer func() {
		if recovered := recover(); recovered != nil {
			countlog.LogPanic(recovered, "recording.panic")
		}
	}()
	for {
		session := <-recorder.recordChan
		countlog.Debug("event!recording.record_session",
			"ctx", recorder.Context,
			"session", session)
		recorder.realRecorder.Record(session)
	}
}

// Record Record
func (recorder *AsyncRecorder) Record(session *Session) {
	select {
	case recorder.recordChan <- session:
	default:
		countlog.Debug("event!recording.record_chan_overflow",
			"ctx", recorder.Context)
	}
}
