package hook

import (
	"net"
	"syscall"

	"github.com/didi/sharingan/recorder/koala_grpc/sut"

	"github.com/v2pro/plz/countlog"
)

const ignoreThreadID = -1
const agentAddr = "127.0.0.1:9003"

// Start start hook
func Start() {
	//setupAcceptHook()
	setupConnectHook()
	setupSendHook()
	setupRecvHook()
	setupCloseHook()
	setupGoRoutineExitHook()
}

// RegisterOnGrpcAccept RegisterOnGrpcAccept
// called by grpc/server.go
func RegisterOnGrpcAccept(ip []byte, port int) {
	gid := sut.ThreadID(GetCurrentGoRoutineID())

	sut.AddGlobalGidSock(sut.SocketFD(gid), net.TCPAddr{ip, port, ""}, true)
}

// RegisterOnGrpcRecv RegisterOnGrpcRecv
// called by grpc/server.go
func RegisterOnGrpcRecv(span []byte) {
	gid := sut.ThreadID(GetCurrentGoRoutineID())
	if gid == ignoreThreadID {
		return
	}

	sut.OperateThread(gid, func(thread *sut.Thread) {
		thread.OnRecv(sut.SocketFD(gid), span, 0)
	})
}

// RegisterOnGrpcSend RegisterOnGrpcSend
// called by grpc/server.go
func RegisterOnGrpcSend(span []byte) {
	gid := sut.ThreadID(GetCurrentGoRoutineID())
	if gid == ignoreThreadID {
		return
	}
	raddr := net.TCPAddr{}
	sut.OperateThread(gid, func(thread *sut.Thread) {
		thread.OnSend(sut.SocketFD(gid), span, 0, &raddr, gid)
	})
}

func setupConnectHook() {
	RegisterOnConnect(func(fd int, sa syscall.Sockaddr) {
		gid := sut.ThreadID(GetCurrentGoRoutineID())

		ipv4Addr, _ := sa.(*syscall.SockaddrInet4)
		if ipv4Addr == nil {
			countlog.Info("event!discard non-ipv4 addr on connect", "addr", sa)
			return
		}

		origIP := make([]byte, 4)
		copy(origIP, ipv4Addr.Addr[:]) // ipv4Addr.Addr will be reused
		origAddr := net.TCPAddr{
			IP:   origIP,
			Port: ipv4Addr.Port,
		}

		if origAddr.String() == agentAddr {
			return
		}

		countlog.Debug("event!sut.connect",
			"threadID", gid,
			"socketFD", fd,
			"addr", &origAddr,
		)

		sut.AddGlobalSock(sut.SocketFD(fd), origAddr, false)
		sut.OperateThreadOnRecordingSession(gid, func(thread *sut.Thread) {
			thread.OnConnect(sut.SocketFD(fd), origAddr)
		})
	})
}

func setupSendHook() {
	RegisterOnSend(func(fd int, network string, raddr net.Addr, span []byte) {
		gid := sut.ThreadID(GetCurrentGoRoutineID())
		if gid == ignoreThreadID {
			return
		}

		switch network {
		case "udp", "udp4", "udp6":
			udpAddr := raddr.(*net.UDPAddr)
			sut.OperateThread(gid, func(thread *sut.Thread) {
				thread.OnSendTo(sut.SocketFD(fd), span, *udpAddr)
			})
		default:
			sut.OperateThread(gid, func(thread *sut.Thread) {
				thread.OnSend(sut.SocketFD(fd), span, 0, raddr, gid)
			})
		}
	})
}

func setupRecvHook() {
	RegisterOnRecv(func(fd int, network string, raddr net.Addr, span []byte) {
		gid := sut.ThreadID(GetCurrentGoRoutineID())
		if gid == ignoreThreadID {
			return
		}

		switch network {
		case "udp", "udp4", "udp6":
		default:
			sut.OperateThread(gid, func(thread *sut.Thread) {
				thread.OnRecv(sut.SocketFD(fd), span, 0)
			})
		}
	})
}

func setupCloseHook() {
	RegisterOnClose(func(fd int) {
		countlog.Debug("event!sut.close", "socketFD", fd)
		sut.RemoveGlobalSock(sut.SocketFD(fd))
	})
}

func setupGoRoutineExitHook() {
	// true goroutineID
	RegisterOnGoRoutineExit(func(gid int64) {
		countlog.Debug("event!sut.goroutine_exit", "threadID", gid)
		sut.OperateThreadOnRecordingSession(sut.ThreadID(gid), func(thread *sut.Thread) {
			thread.OnShutdown()
		})
		sut.RemoveGlobalGidSock(sut.SocketFD(gid))
	})
}

func sockaddrToTCP(sa syscall.Sockaddr) net.TCPAddr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return net.TCPAddr{IP: sa.Addr[0:], Port: sa.Port}
	case *syscall.SockaddrInet6:
		return net.TCPAddr{IP: sa.Addr[0:], Port: sa.Port, Zone: ip6ZoneToString(int(sa.ZoneId))}
	}
	return net.TCPAddr{}
}

func ip6ZoneToString(zone int) string {
	if zone == 0 {
		return ""
	}
	if ifi, err := net.InterfaceByIndex(zone); err == nil {
		return ifi.Name
	}
	return itod(uint(zone))
}

// Convert i to decimal string.
func itod(i uint) string {
	if i == 0 {
		return "0"
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}

	return string(b[bp:])
}
