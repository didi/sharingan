package fastcgi

import (
	"bytes"
	"errors"
	"io"
)

var (
	FastCGIRequestHeader  = []byte{0x1, 0x1, 0x0, 0x1, 0x0, 0x8}
	FastCGIResponseHeader = []byte{0x1, 0x6, 0x0, 0x1}
)

func Decode(origin []byte) (string, error) {
	http := NewHttp()

	buf := bytes.NewBuffer(origin)
	for {
		hdr, err := ParseHeader(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		record, err := hdr.ParseRecord(buf)
		if err != nil {
			return "", err
		}
		if record == nil {
			return "", errors.New("unknown type")
		}
		switch record.GetType() {
		case TypeBeginRequest:
			http.IsRequest = true
		case TypeParams:
			http.Header = merge(http.Header, record.(*Params).Maps)
		case TypeStdin:
			http.BodyIn = append(http.BodyIn, record.(*StdinRequest).Data...)
		case TypeStdout:
			http.BodyOut = append(http.BodyOut, record.(*StdoutResponse).Data...)
		case TypeStderr:
			http.BodyErr = append(http.BodyErr, record.(*StderrResponse).Data...)
		case TypeEndRequest:
			http.IsRequest = false
		}
	}

	return http.Generate()
}
