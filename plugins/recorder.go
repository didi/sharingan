package plugins

import (
	"context"

	"github.com/didi/sharingan/recorder/koala/recording"
)

var recorders []recording.Recorder

// InitRecorderPlugin init recorder plugin
func InitRecorderPlugin() {
	registerRecorderPlugin(NewDefaultRecorder())
}

// StartRecorder StartRecorder
func StartRecorder() {
	for _, recorder := range recorders {
		// set async recorder
		recorder := recording.NewAsyncRecorder(recorder)
		recorder.Context = context.Background()
		recording.Recorders = append(recording.Recorders, recorder)

		// start
		recorder.Start()
	}
}

// add recorder plugin
func registerRecorderPlugin(recorder recording.Recorder) {
	recorders = append(recorders, recorder)
}
