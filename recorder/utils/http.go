package utils

import (
	"bytes"
)

// HTTP HTTP
type HTTP struct {
	General []byte
	Header  map[string][]byte
	Body    []byte
}

// NewHTTP New
func NewHTTP() *HTTP {
	return &HTTP{Header: make(map[string][]byte)}
}

// ParseRequest 解析请求
func (http *HTTP) ParseRequest(content []byte) error {
	if len(content) < 4 {
		return nil
	}
	index := bytes.Index(content, []byte("\r\n\r\n"))
	if index != -1 {
		allHeader := content[:index]
		splitHeader := bytes.Split(allHeader, []byte("\r\n"))
		http.General = splitHeader[0]
		for _, v := range splitHeader[1:] {
			kv := bytes.Split(v, []byte(":"))
			key := string(bytes.Trim(kv[0], " "))
			value := bytes.Trim(kv[1], " ")
			http.Header[key] = value
		}
		if bytes.Contains(http.General, []byte("POST")) && len(content[index:]) > 4 {
			http.Body = content[index+4:]
		}
	}
	return nil
}

// ParseResponse 解析返回
func (http *HTTP) ParseResponse(content []byte) error {
	if len(content) < 4 {
		return nil
	}
	index := bytes.Index(content, []byte("\r\n\r\n"))
	if index != -1 {
		allHeader := content[:index]
		splitHeader := bytes.Split(allHeader, []byte("\r\n"))
		http.General = splitHeader[0]
		for _, v := range splitHeader[1:] {
			kv := bytes.Split(v, []byte(":"))
			key := string(bytes.Trim(kv[0], " "))
			value := bytes.Trim(kv[1], " ")
			http.Header[key] = value
		}
		if len(content[index:]) > 4 {
			http.Body = content[index+4:]
		}
	}
	return nil
}
