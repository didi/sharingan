package recorder

import (
	"github.com/didi/sharingan/recorder/internal"
)

// GetCurrentGoRoutineID get current goRoutineID incase with delegatedID
func GetCurrentGoRoutineID() int64 {
	return internal.GetCurrentGoRoutineID()
}

// SetDelegatedFromGoRoutineID set goRoutine delegatedID
func SetDelegatedFromGoRoutineID(gID int64) {
	internal.SetDelegatedFromGoRoutineID(gID)
}
