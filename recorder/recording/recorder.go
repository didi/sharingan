package recording

// Recorder Recorder
type Recorder interface {
	Record(session *Session)
}

// Recorders Recorders
var Recorders = []Recorder{}

// ShouldRecordAction ShouldRecordAction
var ShouldRecordAction = func(action Action) bool {
	return true
}
