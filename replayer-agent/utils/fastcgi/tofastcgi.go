package fastcgi

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
)

func Encode(reqId uint16, http *Http) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	if err := writeBeginRequest(buf, reqId); err != nil {
		return nil, err
	}

	if http.IsRequest {
		if err := writeRequest(buf, reqId, http); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func writeRequest(buf *bytes.Buffer, reqId uint16, http *Http) error {
	if err := writeParams(buf, reqId, http.Header); err != nil {
		return err
	}

	var body io.Reader
	if http.BodyIn != nil {
		body = bytes.NewBuffer(http.BodyIn)
	}
	if err := writeStdin(buf, reqId, body); err != nil {
		return err
	}

	return nil
}

func writeStdin(buf *bytes.Buffer, reqId uint16, body io.Reader) error {
	if body == nil {
		return writeHeader(buf, byte(TypeStdin), reqId, []byte{0, 0, 0, 0})
	}

	data, err := ioutil.ReadAll(body)
	if err != nil || len(data) == 0 {
		return writeHeader(buf, byte(TypeStdin), reqId, []byte{0, 0, 0, 0})
	}

	if err := writeWithPadding(buf, data, byte(TypeStdin), reqId); err != nil {
		return err
	}

	return writeHeader(buf, byte(TypeStdin), reqId, []byte{0, 0, 0, 0})
}

func writeParams(buf *bytes.Buffer, reqId uint16, m map[string]string) error {
	if len(m) == 0 {
		return writeHeader(buf, byte(TypeParams), reqId, []byte{0, 0, 0, 0})
	}

	tmp := bytes.NewBuffer(nil)
	for k, v := range m {
		lk, lv := len(k), len(v)
		if err := writeLen(tmp, lk); err != nil {
			return err
		}
		if err := writeLen(tmp, lv); err != nil {
			return err
		}
		if _, err := tmp.WriteString(k); err != nil {
			return err
		}
		if _, err := tmp.WriteString(v); err != nil {
			return err
		}
	}

	if err := writeWithPadding(buf, tmp.Bytes(), byte(TypeParams), reqId); err != nil {
		return err
	}

	return writeHeader(buf, byte(TypeParams), reqId, []byte{0, 0, 0, 0})
}

func writeBeginRequest(buf *bytes.Buffer, reqId uint16) error {
	if err := writeHeader(buf, byte(TypeBeginRequest), reqId, []byte{0, 8, 0, 0}); err != nil {
		return err
	}
	if _, err := buf.Write([]byte{0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	return nil
}

func writeHeader(buf *bytes.Buffer, typ byte, reqId uint16, len []byte) error {
	if err := buf.WriteByte(Version); err != nil {
		return err
	}
	if err := buf.WriteByte(typ); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, reqId); err != nil {
		return err
	}
	if _, err := buf.Write(len); err != nil {
		return err
	}
	return nil
}

func writeWithPadding(buf *bytes.Buffer, data []byte, typ byte, reqId uint16) error {
	lenArr, paddingLen := genLenBytes(len(data))
	if err := writeHeader(buf, typ, reqId, lenArr); err != nil {
		return err
	}

	if _, err := buf.Write(data); err != nil {
		return err
	}
	for i := 0; i < paddingLen; i++ {
		if err := buf.WriteByte(0); err != nil {
			return err
		}
	}
	return nil
}

func writeLen(buf *bytes.Buffer, l int) error {
	if l >= 256 {
		return binary.Write(buf, binary.BigEndian, int(l))
	}
	return binary.Write(buf, binary.BigEndian, byte(l))
}

func genLenBytes(valueLen int) ([]byte, int) {
	paddingLen := 8 - valueLen%8
	if paddingLen == 8 {
		paddingLen = 0
	}
	return []byte{byte(valueLen >> 8), byte(valueLen & 0xFF), byte(paddingLen), 0}, paddingLen
}
