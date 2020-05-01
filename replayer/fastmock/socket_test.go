package fastmock

import (
	"testing"
	"time"
)

func TestSockets_Set(t *testing.T) {
	type args struct {
		fd         int
		remoteAddr string
	}
	tests := []struct {
		name string
		args args
	}{
		{"1", args{fd: 1, remoteAddr: "127.0.0.1:8888"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := globalSockets
			s.Set(tt.args.fd, tt.args.remoteAddr, time.Now())

			socket := s.Get(tt.args.fd)
			if socket.remoteAddr != tt.args.remoteAddr {
				t.Errorf("Sockets.Get() = %v, want %v", socket.remoteAddr, tt.args.remoteAddr)
			}

			acessTime := socket.lastAccessedAt
			s.Access(tt.args.fd)
			if socket := s.Get(tt.args.fd); socket.lastAccessedAt == acessTime {
				t.Errorf("Sockets.Access() want change acessTime")
			}

			s.Remove(tt.args.fd)
			if socket := s.Get(tt.args.fd); socket != nil {
				t.Errorf("Sockets.Remove() false, get socket")
			}
		})
	}
}
