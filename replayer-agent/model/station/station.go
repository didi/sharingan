// station is used to decouple replayer and outbound server

package station

import (
	"sync"

	"github.com/didi/sharingan/replayer-agent/logic/replayed"
	"github.com/didi/sharingan/replayer-agent/model/replaying"
)

type meta struct {
	replayingSession *replaying.Session
	replayedSession  *replayed.Session
}

var tmp = map[string]*meta{}
var tmpMutex = &sync.Mutex{}

func Store(traceID string, replaying *replaying.Session, replayed *replayed.Session) {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	tmp[traceID] = &meta{
		replaying,
		replayed,
	}
}

func Load(traceID string) (*replaying.Session, *replayed.Session) {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	m := tmp[traceID]
	// TODO remove
	delete(tmp, traceID)
	return m.replayingSession, m.replayedSession
}

func Remove(traceID string) {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	delete(tmp, traceID)
}
