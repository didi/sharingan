package fastmock

import (
	"testing"
	"time"
)

func TestThreads_Set(t *testing.T) {
	type args struct {
		threadID int64

		traceID    string
		replayTime int64
	}
	tests := []struct {
		name string
		args args
	}{
		{"1", args{threadID: 1, traceID: "1111111", replayTime: time.Now().Unix()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tr := ReplayerGlobalThreads
			tr.Set(tt.args.threadID, tt.args.traceID, tt.args.replayTime)

			thread := tr.Get(tt.args.threadID)
			if thread.traceID != tt.args.traceID {
				t.Errorf("Threads.Get() = %v, want %v", thread.traceID, tt.args.traceID)
			}

			acessTime := thread.lastAccessedAt
			tr.Access(tt.args.threadID)
			if thread := tr.Get(tt.args.threadID); thread.lastAccessedAt == acessTime {
				t.Errorf("Threads.Access() want change acessTime")
			}
		})
	}
}
