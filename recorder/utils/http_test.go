package utils

import (
	"testing"
)

var getReq []byte = []byte("GET /api/v1/get?name=test HTTP/1.1\r\nHost: 100.69.238.11:8000\r\nAccept: */*\r\nxxx-header-rid: b0fb8436baeb2caf7e398d5d361c178c\r\nxxx-header-spanid: 3e645d3d130dbfeb\r\n\r\n")
var postReq []byte = []byte("POST /api/v1/post HTTP/1.1\r\nHost: 100.69.238.59:8000\r\nAccept: */*\r\nxxx-header-rid: 64469d275a178d9d46cd082b0db01b02\r\nxxx-header-spanid: 3e5b1e5144a258c7\r\nContent-Length: 143\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nname=xxx&version=1.0.0")
var response []byte = []byte("HTTP/1.1 200 OK\r\nServer: DFE\r\nDate: Thu, 23 Nov 2017 07:34:54 GMT\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 111\r\nConnection: keep-alive\r\n\r\n{\"errno\":40003,\"errmsg\":\"起点城市未开通\",\"log_id\":\"\"l}")

func TestParseRequest(t *testing.T) {
	httpGet := NewHTTP()
	httpGet.ParseRequest(getReq)
	if host, ok := httpGet.Header["Host"]; ok {
		t.Log(string(host), string(httpGet.General), httpGet)
	} else {
		t.Error("Parse http get failed!!!", httpGet)
	}
	httpPost := NewHTTP()
	httpPost.ParseRequest(postReq)
	if len(httpPost.Body) > 0 {
		t.Log(string(httpPost.Body), httpPost)
	} else {
		t.Error("Parse http get failed!!!")
	}
}

func TestParseResponse(t *testing.T) {
	http := NewHTTP()
	http.ParseResponse(response)
	if serv, ok := http.Header["Server"]; ok {
		t.Log(string(serv), string(http.Body), http)
	} else {
		t.Error("Parse http response failed!!!")
	}
}
