package record

import (
	"testing"
	"time"

	"github.com/juju/ratelimit"
)

func TestFliter(t *testing.T) {
	var str string

	str = `{"operate":100001,"data":{"Context":"hostname ","ThreadId":2612231,"SessionId":"1548745599676096093-2612231","TraceHeader":null,"NextSessionId":"1548745599676604413-2612231","CallFromInbound":{"ActionIndex":0,"OccurredAt":1548745599676102064,"ActionType":"CallFromInbound","Peer":{"IP":"127.0.0.1","Port":53442,"Zone":""},"UnixAddr":{"Name":"","Net":""},"Request":"POST /report/set HTTP/1.0\r\nHost: 127.0.0.1:9999\r\nConnection: close\r\nContent-Length: 618\r\nxxx-header-traceid: 645a693a5c4ffb7fae8a4c1f14637602\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nappid=2\u0026action=1"},"ReturnInbound":{"ActionIndex":0,"OccurredAt":1548745599676544387,"ActionType":"ReturnInbound","Response":"HTTP/1.0 200 OK\r\nContent-Type: application/json\r\nDate: Tue, 29 Jan 2019 07:06:39 GMT\r\nContent-Length: 25\r\n\r\n{\"errno\":0,\"errmsg\":\"ok\"}"},"Actions":[{"ActionIndex":0,"OccurredAt":1548745599676544387,"ActionType":"ReturnInbound","Response":"HTTP/1.0 200 OK\r\nContent-Type: application/json\r\nDate: Tue, 29 Jan 2019 07:06:39 GMT\r\nContent-Length: 25\r\n\r\n{\"errno\":0,\"errmsg\":\"ok\"}"}],"TraceId":"645a693a5c4ffb7fae8a4c1f14637602","SpanId":""}}`

	// 0.5r/s，2秒一个请求
	bucketMap.m["/report/set"] = ratelimit.NewBucketWithRate(0.5, defaultBucketCapacity)

	var res bool

	// 初始，不限制
	res, _ = Fliter(str)
	if res == true {
		t.Errorf("Fliter() = %v, want %v", true, false)
	}

	// 1s之后，应该限制
	time.Sleep(1 * time.Second)
	res, _ = Fliter(str)
	if res == false {
		t.Errorf("Fliter() = %v, want %v", false, true)
	}

	// 2s之后，应该解除限制
	time.Sleep(1 * time.Second)
	res, _ = Fliter(str)
	if res == true {
		t.Errorf("Fliter() = %v, want %v", true, false)
	}
}

func Test_matchKeyWords(t *testing.T) {
	type args struct {
		str      string
		keywords []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"1", args{"xxxaxxxbxxx", []string{"a", "b"}}, true},
		{"1", args{"xxxaxxxbxxx", []string{"c"}}, false},
		{"1", args{"xxxaxxxbxxx", []string{"a", "b", "c"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchKeyWords(tt.args.str, tt.args.keywords); got != tt.want {
				t.Errorf("matchKeyWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getURI(t *testing.T) {
	type args struct {
		str string
	}

	session1 := `{"operate":100001,"data":{"Context":"hostname ","ThreadId":2612231,"SessionId":"1548745599676096093-2612231","TraceHeader":null,"NextSessionId":"1548745599676604413-2612231","CallFromInbound":{"ActionIndex":0,"OccurredAt":1548745599676102064,"ActionType":"CallFromInbound","Peer":{"IP":"127.0.0.1","Port":53442,"Zone":""},"UnixAddr":{"Name":"","Net":""},"Request":"POST /report/set HTTP/1.0\r\nHost: 127.0.0.1:9999\r\nConnection: close\r\nContent-Length: 618\r\nxxx-header-traceid: 645a693a5c4ffb7fae8a4c1f14637602\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nappid=2\u0026action=1"},"ReturnInbound":{"ActionIndex":0,"OccurredAt":1548745599676544387,"ActionType":"ReturnInbound","Response":"HTTP/1.0 200 OK\r\nContent-Type: application/json\r\nDate: Tue, 29 Jan 2019 07:06:39 GMT\r\nContent-Length: 25\r\n\r\n{\"errno\":0,\"errmsg\":\"ok\"}"},"Actions":[{"ActionIndex":0,"OccurredAt":1548745599676544387,"ActionType":"ReturnInbound","Response":"HTTP/1.0 200 OK\r\nContent-Type: application/json\r\nDate: Tue, 29 Jan 2019 07:06:39 GMT\r\nContent-Length: 25\r\n\r\n{\"errno\":0,\"errmsg\":\"ok\"}"}],"TraceId":"645a693a5c4ffb7fae8a4c1f14637602","SpanId":""}}`

	tests := []struct {
		name string
		args args
		want string
	}{
		{"GET", args{session1}, "/report/set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getURI(tt.args.str); got != tt.want {
				t.Errorf("getURI() = %v, want %v", got, tt.want)
			}
		})
	}
}
