// +build !recorder
// +build !replayer

package sharingan

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return 0
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
}

func init() {
}
