// +build recorder

package sharingan

import (
	"runtime"

	_ "github.com/didichuxing/sharingan/recorder"
)

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return runtime.GetCurrentGoRoutineId()
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
	runtime.SetDelegatedFromGoRoutineId(gID)
}
