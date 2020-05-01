// +build !recorder
// +build !replayer

package sharingan

// GetCurrentGoRoutineID get current goRoutineID incase with delegatedID
func GetCurrentGoRoutineID() int64 {
	return 0
}

// SetDelegatedFromGoRoutineID set goRoutine delegatedID
func SetDelegatedFromGoRoutineID(gID int64) {
}

func init() {
}
