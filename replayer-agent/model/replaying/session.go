package replaying

import (
	"github.com/didi/sharingan/replayer-agent/model/recording"
)

type Session struct {
	Context         string
	SessionId       string
	CallFromInbound *recording.CallFromInbound
	ReturnInbound   *recording.ReturnInbound
	CallOutbounds   []*recording.CallOutbound
	RedirectDirs    map[string]string
	MockFiles       map[string][][]byte
	AppendFiles     []*recording.AppendFile
	TracePaths      []string
}

func NewSession() *Session {
	return &Session{}
}
