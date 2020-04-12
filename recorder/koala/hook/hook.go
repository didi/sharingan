package hook

import (
	"net"
	"syscall"

	"github.com/didichuxing/sharingan/recorder/koala/sut"

	"github.com/v2pro/plz/countlog"
)

const ignoreThreadID = -1

// Start start hook
func Start() {
	setupAcceptHook()
	setupConnectHook()
	setupSendHook()
	setupRecvHook()
	setupCloseHook()
	setupGoRoutineExitHook()
}

func setupAcceptHook() {
	RegisterOnAccept(func(serverSocketFD int, clientSocketFD int, sa syscall.Sockaddr) {
		gid := sut.ThreadID(GetCurrentGoRoutineID())

		origAddr := sockaddrToTCP(sa)

		countlog.Debug("event!sut.accept",
			"threadID", gid,
			"serverSocketFD", serverSocketFD,
			"socketFD", clientSocketFD,
			"addr", &origAddr,
		)

		sut.AddGlobalSock(sut.SocketFD(clientSocketFD), origAddr, true)
		sut.OperateThreadOnRecordingSession(gid, func(thread *sut.Thread) {
			thread.OnAccept(sut.SocketFD(serverSocketFD), sut.SocketFD(clientSocketFD), origAddr)
		})
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
				thread.OnSend(sut.SocketFD(fd), span, 0, raddr)
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
