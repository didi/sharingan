// +build recorder_grpc

package sharingan

import (
	"context"
	"log"
	"os"
	"runtime"

	recorderPKG "github.com/didi/sharingan/recorder"
	"github.com/didi/sharingan/recorder/koala_grpc/hook"
	"github.com/didi/sharingan/recorder/koala_grpc/logger"
	"github.com/didi/sharingan/recorder/koala_grpc/recording"
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
	recorder := recording.NewAsyncRecorder(recorderPKG.NewRecorderGrpc())
	recorder.Context = context.Background()
	recorder.Start()
	recording.Recorders = append(recording.Recorders, recorder)

	// set action which should record
	recording.ShouldRecordAction = recorderPKG.ShouldRecordActionGrpc

	// setup logger
	logger.Setup()

	// start hook
	hook.Start()

	// log
	log.Println("mode", "=====grpc recorder=====")
}
