package recording

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	callOutboundAction  = "callOutbound"
	returnInboundAction = "returnInbound"
)

// Action Action
type Action interface {
	GetActionIndex() int
	GetOccurredAt() int64
	GetActionType() string
}

type action struct {
	ActionIndex int
	OccurredAt  int64
	ActionType  string
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

// CallFromInbound Inbound Request
type CallFromInbound struct {
	action
	Peer     net.TCPAddr
	Request  []byte
	UnixAddr net.UnixAddr
}

// MarshalJSON MarshalJSON
func (callFromInbound *CallFromInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallFromInbound
		Request json.RawMessage
	}{
		CallFromInbound: *callFromInbound,
		Request:         EncodeAnyByteArray(callFromInbound.Request),
	})
}

// ReturnInbound Inbound Response
type ReturnInbound struct {
	action
	Response []byte
}

// MarshalJSON MarshalJSON
func (returnInbound *ReturnInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ReturnInbound
		Response json.RawMessage
	}{
		ReturnInbound: *returnInbound,
		Response:      ParesResponse(returnInbound.Response, returnInboundAction),
	})
}

// CallOutbound Outbound Request and Response
type CallOutbound struct {
	action
	SocketFD     int
	Peer         net.TCPAddr
	Local        *net.TCPAddr `json:"-"`
	Request      []byte
	ResponseTime int64
	Response     []byte
	UnixAddr     net.UnixAddr
	CSpanID      []byte `json:"CSpanId"`
}

// MarshalJSON MarshalJSON
func (callOutbound *CallOutbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallOutbound
		Request  json.RawMessage
		Response json.RawMessage
		CSpanID  json.RawMessage `json:"CSpanId"`
	}{
		CallOutbound: *callOutbound,
		Request:      EncodeAnyByteArray(callOutbound.Request),
		Response:     ParesResponse(callOutbound.Response, callOutboundAction),
		CSpanID:      EncodeAnyByteArray(callOutbound.CSpanID),
	})
}

// AppendFile AppendFile
type AppendFile struct {
	action
	FileName string
	Content  []byte
}

// MarshalJSON MarshalJSON
func (appendFile *AppendFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AppendFile
		Content json.RawMessage
	}{
		AppendFile: *appendFile,
		Content:    EncodeAnyByteArray(appendFile.Content),
	})
}

// SendUDP SendUDP
type SendUDP struct {
	action
	Peer    net.UDPAddr
	Content []byte
}

// MarshalJSON MarshalJSON
func (sendUDP *SendUDP) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SendUDP
		Content json.RawMessage
	}{
		SendUDP: *sendUDP,
		Content: EncodeAnyByteArray(sendUDP.Content),
	})
}

// ReadStorage ReadStorage
type ReadStorage struct {
	action
	Content []byte
}

// MarshalJSON MarshalJSON
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

// EncodeAnyByteArray custom Marshal
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

// ParesResponse ...
func ParesResponse(s []byte, action string) json.RawMessage {
	encoded := []byte{'"'}
	if !bytes.Contains(s, []byte("Content-Encoding: gzip")) {
		return EncodeAnyByteArray(s)
	}

	// handle gzip response
	reader := bufio.NewReader(strings.NewReader(string(s)))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		fmt.Println("反序列化HTTP响应出错：", err)
		return encoded
	}
	defer resp.Body.Close()

	// 解析响应体
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应体出错：", err)
		return encoded
	}

	// 检查Content-Encoding，并解压缩响应体
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
		if err != nil {
			fmt.Println("创建gzip解压缩读取器出错：", err)
			return encoded
		}
		defer reader.Close()

		bodyBytes, err = ioutil.ReadAll(reader)
		if err != nil {
			fmt.Println("读取解压缩后的内容出错：", err)
			return encoded
		}

		switch action {
		case returnInboundAction:
			return bodyBytes
		case callOutboundAction:
			// 将GBK编码转换为UTF-8编码
			utf8Bytes, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(bodyBytes), simplifiedchinese.GBK.NewDecoder()))
			if err != nil {
				fmt.Println("err: ", err)
				return encoded
			}

			return utf8Bytes
		}
	}

	return encoded
}
