// +build replayer

package sharingan

import (
	"log"
	"runtime"

	"github.com/didi/sharingan/replayer/fastmock"
)

// GetCurrentGoRoutineID get current goRoutineID incase with delegatedID
func GetCurrentGoRoutineID() int64 {
	return runtime.GetCurrentGoRoutineId()
}

// SetDelegatedFromGoRoutineID set goRoutine delegatedID
func SetDelegatedFromGoRoutineID(gID int64) {
	runtime.SetDelegatedFromGoRoutineId(gID)
}

func init() {
	fastmock.MockSyscall()
	fastmock.MockTime()

	log.Println("mode", "=====replayer=====")
}
