package recording

import (
	"encoding/json"
	"net"
	"unicode/utf8"
)

type action struct {
	ActionIndex int
	OccurredAt  int64
	ActionType  string
}

type Action interface {
	GetActionIndex() int
	GetOccurredAt() int64
	GetActionType() string
}

func (action *action) GetActionType() string {
	return action.ActionType
}

func (action *action) GetActionIndex() int {
	return action.ActionIndex
}

func (action *action) GetOccurredAt() int64 {
	return action.OccurredAt
}

func (action *action) SetActionType(actionType string) {
	action.ActionType = actionType
}

func (action *action) SetActionIndex(actionIndex int) {
	action.ActionIndex = actionIndex
}

func (action *action) SetOccurredAt(occurredAt int64) {
	action.OccurredAt = occurredAt
}

type CallFromInbound struct {
	action
	Peer     net.TCPAddr
	Request  []byte // depends on Raw,  http    | http
	Raw      []byte // decides Request, fastcgi | http
	UnixAddr net.UnixAddr
}

func (callFromInbound *CallFromInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallFromInbound
		Request json.RawMessage
	}{
		CallFromInbound: *callFromInbound,
		Request:         EncodeAnyByteArray(callFromInbound.Request),
	})
}

type ReturnInbound struct {
	action
	Response []byte // http format
	Raw      []byte // fastcgi format
}

func (returnInbound *ReturnInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ReturnInbound
		Response json.RawMessage
	}{
		ReturnInbound: *returnInbound,
		Response:      EncodeAnyByteArray(returnInbound.Response),
	})
}

type CallOutbound struct {
	action
	SocketFD     int
	Peer         net.TCPAddr
	Local        *net.TCPAddr `json:"-"`
	Request      []byte
	ResponseTime int64
	Response     []byte
	UnixAddr     net.UnixAddr
	CSpanId      []byte
	IgnoreFlag   bool
}

func (callOutbound *CallOutbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallOutbound
		Request  json.RawMessage
		Response json.RawMessage
		CSpanId  json.RawMessage
	}{
		CallOutbound: *callOutbound,
		Request:      EncodeAnyByteArray(callOutbound.Request),
		Response:     EncodeAnyByteArray(callOutbound.Response),
		CSpanId:      EncodeAnyByteArray(callOutbound.CSpanId),
	})
}

func (callOutbound *CallOutbound) SetIgnoreFlag(flag bool) {
	callOutbound.IgnoreFlag = true
}

type AppendFile struct {
	action
	FileName string
	Content  []byte
}

func (appendFile *AppendFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AppendFile
		Content json.RawMessage
	}{
		AppendFile: *appendFile,
		Content:    EncodeAnyByteArray(appendFile.Content),
	})
}

type SendUDP struct {
	action
	Peer    net.UDPAddr
	Content []byte
}

func (sendUDP *SendUDP) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SendUDP
		Content json.RawMessage
	}{
		SendUDP: *sendUDP,
		Content: EncodeAnyByteArray(sendUDP.Content),
	})
}

type ReadStorage struct {
	action
	Content []byte
}

func (readStorage *ReadStorage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ReadStorage
		Content json.RawMessage
	}{
		ReadStorage: *readStorage,
		Content:     EncodeAnyByteArray(readStorage.Content),
	})
}

// safeSet holds the value true if the ASCII character with the given array
// position can be represented inside a JSON string without any further
// escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), and the backslash character ("\").
var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
var hex = "0123456789abcdef"

func EncodeAnyByteArray(s []byte) json.RawMessage {
	encoded := []byte{'"'}
	i := 0
	start := i
	for i < len(s) {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				encoded = append(encoded, s[start:i]...)
			}
			switch b {
			case '\\':
				encoded = append(encoded, `\\x5c`...)
			case '"':
				encoded = append(encoded, `\"`...)
			case '\n':
				encoded = append(encoded, `\n`...)
			case '\r':
				encoded = append(encoded, `\r`...)
			case '\t':
				encoded = append(encoded, `\t`...)
			default:
				encoded = append(encoded, `\\x`...)
				encoded = append(encoded, hex[b>>4])
				encoded = append(encoded, hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(s[i:])
		if c == utf8.RuneError {
			if start < i {
				encoded = append(encoded, s[start:i]...)
			}
			for _, b := range s[i : i+size] {
				encoded = append(encoded, `\\x`...)
				encoded = append(encoded, hex[b>>4])
				encoded = append(encoded, hex[b&0xF])
			}
			i += size
			start = i
		} else {
			i += size
		}
	}
	if start < len(s) {
		encoded = append(encoded, s[start:]...)
	}
	encoded = append(encoded, '"')
	return json.RawMessage(encoded)
}
