package recorder

import (
	"github.com/didi/sharingan/recorder/koala/hook"
)

// GetCurrentGoRoutineID get current goRoutineID incase with delegatedID
func GetCurrentGoRoutineID() int64 {
	return hook.GetCurrentGoRoutineID()
}

// SetDelegatedFromGoRoutineID set goRoutine delegatedID
func SetDelegatedFromGoRoutineID(gID int64) {
	hook.SetDelegatedFromGoRoutineID(gID)
}
