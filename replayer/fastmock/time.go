// +build replayer

package fastmock

import (
	"runtime"
	"time"

	"github.com/didi/sharingan/replayer/monkey"
)

func MockTime() {
	MockTimeNow()
}

func MockTimeNow() {
	monkey.MockGlobalFunc(time.Now, func() time.Time {
		threadID := runtime.GetCurrentGoRoutineId()

		replayTime := int64(0)
		globalThreadsMutex.RLock()
		if thread, ok := globalThreads[threadID]; ok {
			replayTime = thread.replayTime
		}
		globalThreadsMutex.RUnlock()

		if replayTime == 0 {
			return time.Now2()
		}

		sec := replayTime / 1000000000
		nsec := replayTime % 1000000000
		return time.Unix(sec, nsec)
	})
}
