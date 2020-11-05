package fastcgi

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Http struct {
	IsRequest bool              // true: request, false: response
	Header    map[string]string // http header
	BodyIn    []byte            // http body (in)
	BodyOut   []byte            // http body (out)
	BodyErr   []byte            // http body (err)
}

func NewHttp() *Http {
	var http Http
	http.IsRequest = true
	http.Header = make(map[string]string)
	http.BodyIn = make([]byte, 0)
	http.BodyOut = make([]byte, 0)
	http.BodyErr = make([]byte, 0)
	return &http
}

func (http *Http) Generate() (string, error) {
	data := bytes.NewBuffer(nil)
	if http.IsRequest {
		if uri, ok := http.Header["REQUEST_URI"]; !ok || uri == "" {
			return "", errors.New("empty request uri")
		}
		data.WriteString(fmt.Sprintf("%s %s %s\r\n", http.Header["REQUEST_METHOD"], http.Header["REQUEST_URI"], http.Header["SERVER_PROTOCOL"]))
		for k, v := range http.Header {
			if !strings.HasPrefix(k, "HTTP_") {
				continue
			}
			k = strings.Replace(strings.ToLower(strings.TrimPrefix(k, "HTTP_")), "_", "-", -1)
			data.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		data.WriteString("\r\n")
		data.Write(http.BodyIn)
	} else {
		data.WriteString("HTTP/1.1 200 OK\r\n")
		data.Write(http.BodyOut)
		data.Write(http.BodyErr)
	}
	return data.String(), nil
}
