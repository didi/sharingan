package replaying

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/didichuxing/sharingan/replayer-agent/model/recording"
)

type replayedAction struct {
	ActionId   string
	OccurredAt int64
	ActionType string
}

type ReplayedAction interface {
	GetActionId() string
	GetActionType() string
	GetOccurredAt() int64
}

func (action *replayedAction) GetActionType() string {
	return action.ActionType
}

func (action *replayedAction) GetActionId() string {
	return action.ActionId
}

func (action *replayedAction) GetOccurredAt() int64 {
	return action.OccurredAt
}

func newReplayedAction(actionType string) replayedAction {
	occurredAt := time.Now().UnixNano()
	actionId := strconv.FormatInt(occurredAt, 10)
	return replayedAction{
		ActionId:   actionId,
		OccurredAt: occurredAt,
		ActionType: actionType,
	}
}

type CallFromInbound struct {
	replayedAction
	OriginalRequestTime int64
	OriginalRequest     []byte
}

func (callFromInbound *CallFromInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallFromInbound
		OriginalRequest json.RawMessage
	}{
		CallFromInbound: *callFromInbound,
		OriginalRequest: recording.EncodeAnyByteArray(callFromInbound.OriginalRequest),
	})
}

type ReturnInbound struct {
	replayedAction
	OriginalResponse []byte
	Response         []byte
}

func (returnInbound *ReturnInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ReturnInbound
		OriginalResponse json.RawMessage
		Response         json.RawMessage
	}{
		ReturnInbound:    *returnInbound,
		OriginalResponse: recording.EncodeAnyByteArray(returnInbound.OriginalResponse),
		Response:         recording.EncodeAnyByteArray(returnInbound.Response),
	})
}

type CallOutbound struct {
	replayedAction
	MatchedRequest     []byte
	MatchedResponse    []byte
	MatchedActionIndex int
	MatchedMark        float64
	MockedResponse     []byte
	Request            []byte
	Peer               net.TCPAddr
}

func (callOutbound *CallOutbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallOutbound
		MatchedRequest  json.RawMessage
		MatchedResponse json.RawMessage
		Request         json.RawMessage
	}{
		CallOutbound:    *callOutbound,
		MatchedRequest:  recording.EncodeAnyByteArray(callOutbound.MatchedRequest),
		MatchedResponse: recording.EncodeAnyByteArray(callOutbound.MatchedResponse),
		Request:         recording.EncodeAnyByteArray(callOutbound.Request),
	})
}

func NewCallOutbound(peer net.TCPAddr, request []byte) *CallOutbound {
	return &CallOutbound{
		replayedAction: newReplayedAction("CallOutbound"),
		Peer:           peer,
		Request:        request,
	}
}

type CallFunction struct {
	replayedAction
	CallIntoFile string
	CallIntoLine int
	FuncName     string
	Args         []interface{}
}

type ReturnFunction struct {
	replayedAction
	CallFunctionId string
	ReturnValue    interface{}
}

type AppendFile struct {
	replayedAction
	FileName string
	Content  []byte
}

func (appendFile *AppendFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AppendFile
		Content json.RawMessage
	}{
		AppendFile: *appendFile,
		Content:    recording.EncodeAnyByteArray(appendFile.Content),
	})
}

type SendUDP struct {
	replayedAction
	Peer    net.UDPAddr
	Content []byte
}

func (sendUDP *SendUDP) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SendUDP
		Content json.RawMessage
	}{
		SendUDP: *sendUDP,
		Content: recording.EncodeAnyByteArray(sendUDP.Content),
	})
}
