// +build replayer

package sharingan

import (
	"log"

	"github.com/didi/sharingan/replayer"
	"github.com/didi/sharingan/replayer/fastmock"
)

// GetCurrentGoRoutineID get current goroutineID incase SetDelegatedFromGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return replayer.GetCurrentGoRoutineID()
}

// SetDelegatedFromGoRoutineID should be used when this goroutine is doing work for another goroutine
func SetDelegatedFromGoRoutineID(gID int64) {
	replayer.SetDelegatedFromGoRoutineID(gID)
}

func init() {
	fastmock.MockSyscall()
	fastmock.MockTime()

	log.Println("mode", "=====replayer=====")
}
