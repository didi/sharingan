package sut

import (
	"net"
	"time"
)

type socket struct {
	socketFD       SocketFD
	isServer       bool
	addr           net.TCPAddr
	localAddr      *net.TCPAddr
	unixAddr       net.UnixAddr
	lastAccessedAt time.Time
}
