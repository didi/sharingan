package replayer

import (
	"github.com/didi/sharingan/replayer/internal"
)

// GetCurrentGoRoutineID get current goRoutineID incase with delegatedID
func GetCurrentGoRoutineID() int64 {
	return internal.GetCurrentGoRoutineID()
}

// SetDelegatedFromGoRoutineID set goRoutine delegatedID
func SetDelegatedFromGoRoutineID(gID int64) {
	internal.SetDelegatedFromGoRoutineID(gID)
}
