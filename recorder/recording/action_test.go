package recording

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_marshal_append_file(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&AppendFile{
		FileName: "/abc",
		Content:  []byte("hello"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
}

func Test_marshal_call_outbound(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&CallOutbound{
		Request:  []byte("hello"),
		Response: []byte("world"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
	should.Contains(string(bytes), "world")
}

func Test_marshal_return_inbound(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&ReturnInbound{
		Response: []byte("hello"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
}

func Test_marshal_call_from_inbound(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&CallFromInbound{
		Request: []byte("hello"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
}

func Test_marshal_session(t *testing.T) {
	session := Session{
		CallFromInbound: &CallFromInbound{
			Request: []byte("hello"),
		},
		ReturnInbound: &ReturnInbound{
			Response: []byte("hello"),
		},
		Actions: []Action{
			&CallOutbound{
				Request:  []byte("hello"),
				Response: []byte("world"),
			},
			&AppendFile{
				FileName: "/abc",
				Content:  []byte("hello"),
			},
		},
	}
	bytes, err := json.MarshalIndent(session, "", "  ")
	should := require.New(t)
	should.Nil(err)
	should.NotContains(string(bytes), "=") // no base64
}

func Test_encode_any_byte_array(t *testing.T) {
	should := require.New(t)
	should.Equal(`"hello"`, string(EncodeAnyByteArray([]byte("hello"))))
	should.Equal(`"hel\nlo"`, string(EncodeAnyByteArray([]byte("hel\nlo"))))
	should.Equal(`"hel\rlo"`, string(EncodeAnyByteArray([]byte("hel\rlo"))))
	should.Equal(`"hel\tlo"`, string(EncodeAnyByteArray([]byte("hel\tlo"))))
	should.Equal(`"hel\"lo"`, string(EncodeAnyByteArray([]byte("hel\"lo"))))
	should.Equal(`"hel\\x5cx00lo"`, string(EncodeAnyByteArray([]byte(`hel\x00lo`))))
	should.Equal(`"hel\\x00lo"`, string(EncodeAnyByteArray([]byte("hel\u0000lo"))))
	should.Equal(`"\\x01\\x02\\x03"`, string(EncodeAnyByteArray([]byte{1, 2, 3})))
	should.Equal(`"中文"`, string(EncodeAnyByteArray([]byte("中文"))))
	should.Equal(`"\\xef\\xbf\\xbdBEEF"`,
		string(EncodeAnyByteArray([]byte{239, 191, 189, 66, 69, 69, 70})))
}
