package esmodel

import (
	"net"

	"github.com/didi/sharingan/replayer-agent/utils/helper"
	jsoniter "github.com/json-iterator/go"
)

func RetrieveSessions(data []byte) ([]Session, error) {
	var source DataSource
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(data, &source)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, session := range source.Hits.Hits {
		sessions = append(sessions, session.Sess)
	}

	return sessions, nil
}

// ES存储的完整数据格式
type DataSource struct {
	Hits HitsOutside `json:"hits"`
}

type HitsOutside struct {
	Hits []HitsInside `json:"hits"`
}

type HitsInside struct {
	Sess Session `json:"_source"`
}

type Session struct {
	Context         string
	ThreadId        int32
	SessionId       string
	TraceId         string
	SpanId          string
	NextSessionId   string
	CallFromInbound *CallFromInbound
	ReturnInbound   *ReturnInbound
	Actions         []Action
}

type ActionMeta struct {
	ActionIndex int
	OccurredAt  int64
	ActionType  string
}

type Action struct {
	ActionMeta
	// outbound data
	Content      Raw
	Peer         net.TCPAddr
	Request      Raw
	ResponseTime int64
	Response     Raw
	CSpanId      []byte
	SocketFD     int
}

type CallFromInbound struct {
	ActionMeta
	Peer     net.TCPAddr
	Request  Raw
	UnixAddr net.UnixAddr
}

func (r *Raw) UnmarshalJSON(data []byte) error {
	// step1: unquote string
	// tmp, err := strconv.Unquote(helper.BytesToString(data))
	tmp, err := Unquote(data)
	if err != nil {
		return err
	}
	// step2: stripcslashes
	r.Data = helper.StripcSlashes(helper.StringToBytes(tmp))
	return nil
}

type ReturnInbound struct {
	ActionMeta
	Response Raw
}

type Raw struct {
	Data []byte
}
