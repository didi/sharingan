// +build replayer

package sharingan

import (
	"runtime"

	_ "github.com/didichuxing/sharingan/replayer"
)

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return runtime.GetCurrentGoRoutineId()
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
	runtime.SetDelegatedFromGoRoutineId(gID)
}
