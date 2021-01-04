package sut

import (
	"context"
	"net"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/didi/sharingan/recorder/koala/recording"

	"github.com/v2pro/plz/countlog"
)

// Thread Thread
type Thread struct {
	context.Context
	mutex            *sync.Mutex
	threadID         ThreadID
	socks            map[SocketFD]*socket
	recordingSession *recording.Session
	lastAccessedAt   time.Time
	helperResponse   []byte
	ignoreSocks      map[SocketFD]bool
}

// RecvFlags RecvFlags
type RecvFlags int

// ThreadIDKey ThreadIDKey
type ThreadIDKey string

// newThread new Thread
func newThread(threadID ThreadID) *Thread {
	thread := &Thread{
		Context:          context.WithValue(context.Background(), ThreadIDKey("threadID"), threadID),
		mutex:            &sync.Mutex{},
		threadID:         threadID,
		socks:            map[SocketFD]*socket{},
		lastAccessedAt:   time.Now(),
		ignoreSocks:      map[SocketFD]bool{},
		recordingSession: recording.NewSession(int32(threadID)),
	}

	return thread
}

// OnAccept OnAccept
func (thread *Thread) OnAccept(serverSocketFD SocketFD, clientSocketFD SocketFD, addr net.TCPAddr) {
	// thread.shutdownRecordingSession()
	thread.socks[clientSocketFD] = &socket{
		socketFD: clientSocketFD,
		isServer: true,
		addr:     addr,
	}
}

// OnConnect add global socketFD
func (thread *Thread) OnConnect(socketFD SocketFD, remoteAddr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     remoteAddr,
	}
}

// OnSend OnSend
func (thread *Thread) OnSend(socketFD SocketFD, span []byte, extraHeaderSentSize int, raddr net.Addr) {
	if len(span) == 0 || thread.ignoreSocks[socketFD] {
		return
	}

	sock, err := thread.lookupSocket(socketFD)
	if sock == nil || err != nil {
		// not connect
		if err == syscall.ENOTCONN {
			countlog.Warn("event!sut.unknown-send",
				"threadID", thread.threadID,
				"socketFD", socketFD,
				"content", span,
				"err", err)
			return
		}

		// if net set timeout in connect pool, maybe gc
		sock = &socket{
			socketFD: socketFD,
			isServer: false,
			addr:     *raddr.(*net.TCPAddr),
		}
		thread.socks[socketFD] = sock
		setGlobalSock(socketFD, sock)
		countlog.Warn("event!sut.unknown-send",
			"threadID", thread.threadID,
			"socketFD", socketFD,
			"addr", &sock.addr,
			"content", span,
			"err", err)
	}

	event := "event!sut.inbound_send"
	if sock.isServer {
		thread.recordingSession.SendToInbound(thread, span, sock.addr)
	} else {
		event = "event!sut.outbound_send"
		thread.recordingSession.SendToOutbound(thread, span, sock.addr, sock.localAddr, int(sock.socketFD))
	}
	countlog.Debug(event,
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"recordingSessionPtr", uintptr(unsafe.Pointer(thread.recordingSession)),
		"addr", &sock.addr,
		"content", span)
}

// OnSendTo OnSendTo
func (thread *Thread) OnSendTo(socketFD SocketFD, span []byte, addr net.UDPAddr) {
	countlog.Debug("event!sut.sendto",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &addr,
		"content", span)
	thread.recordingSession.SendUDPToOutbound(thread, span, addr)
}

// OnRecv OnRecv
func (thread *Thread) OnRecv(socketFD SocketFD, span []byte, flags RecvFlags) []byte {
	if thread.ignoreSocks[socketFD] {
		return span
	}

	sock, err := thread.lookupSocket(socketFD)
	if sock == nil || err != nil {
		if err == syscall.ENOTCONN {
			countlog.Warn("event!sut.unknown-recv",
				"threadID", thread.threadID,
				"socketFD", socketFD,
				"content", span,
				"err", err)
		} else {
			countlog.Error("event!sut.unknown-recv",
				"threadID", thread.threadID,
				"socketFD", socketFD,
				"content", span,
				"err", err)
		}

		return span
	}

	if !sock.isServer {
		countlog.Debug("event!sut.outbound_recv",
			"threadID", thread.threadID,
			"socketFD", socketFD,
			"recordingSessionPtr", uintptr(unsafe.Pointer(thread.recordingSession)),
			"addr", &sock.addr,
			"content", span)

		thread.recordingSession.RecvFromOutbound(thread, span, sock.addr, sock.localAddr, int(sock.socketFD))
		return span
	}

	countlog.Debug("event!sut.inbound_recv",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"recordingSessionPtr", uintptr(unsafe.Pointer(thread.recordingSession)),
		"addr", &sock.addr,
		"content", span)

	if span == nil {
		return nil
	}

	if thread.recordingSession.HasResponse() {
		countlog.Trace("event!sut.recv_from_inbound_found_responded",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		thread.shutdownRecordingSession()
	}

	thread.recordingSession.RecvFromInbound(thread, span, sock.addr, sock.unixAddr)
	return span
}

// OnShutdown OnShutdown
func (thread *Thread) OnShutdown() {
	thread.shutdownRecordingSession()
}

// OnAccess check action limit, max = 1000
func (thread *Thread) OnAccess() {
	if thread.recordingSession != nil && len(thread.recordingSession.Actions) > 1000 {
		countlog.Warn("event!sut.recorded_too_many_actions",
			"threadID", thread.threadID,
			"sessionId", thread.recordingSession.SessionID)
		thread.shutdownRecordingSession()
	}
}

// shutdownRecordingSession shutdownRecordingSession
func (thread *Thread) shutdownRecordingSession() {
	newSession := recording.NewSession(int32(thread.threadID))
	thread.recordingSession.Shutdown(thread, newSession)
	thread.socks = map[SocketFD]*socket{} // socks on thread is a temp cache
	thread.recordingSession = newSession
}

// IgnoreSocketFD IgnoreSocketFD
func (thread *Thread) IgnoreSocketFD(socketFD SocketFD, remoteAddr net.TCPAddr) {
	countlog.Trace("event!sut.ignoreSocket",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &remoteAddr)

	thread.ignoreSocks[socketFD] = true
}

// check socketFD
func (thread *Thread) lookupSocket(socketFD SocketFD) (*socket, error) {
	sock := thread.socks[socketFD]
	if sock != nil {
		return sock, nil
	}
	sock = getGlobalSock(socketFD)
	if sock == nil {
		return nil, nil
	}
	remoteAddr, err := syscall.Getpeername(int(socketFD))
	if err != nil {
		return nil, err
	}
	remoteAddr4, _ := remoteAddr.(*syscall.SockaddrInet4)
	// if remote address changed, the fd must be closed and reused
	if remoteAddr4 != nil && (remoteAddr4.Port != sock.addr.Port ||
		remoteAddr4.Addr[0] != sock.addr.IP[0] ||
		remoteAddr4.Addr[1] != sock.addr.IP[1] ||
		remoteAddr4.Addr[2] != sock.addr.IP[2] ||
		remoteAddr4.Addr[3] != sock.addr.IP[3]) {
		sock = &socket{
			socketFD: socketFD,
			isServer: false,
			addr: net.TCPAddr{
				Port: remoteAddr4.Port,
				IP:   net.IP(remoteAddr4.Addr[:]),
			},
			lastAccessedAt: time.Now(),
		}
		setGlobalSock(socketFD, sock)
	}
	remoteAddr6, _ := remoteAddr.(*syscall.SockaddrInet6)
	if remoteAddr6 != nil && (remoteAddr6.Port != sock.addr.Port ||
		remoteAddr6.Addr[0] != sock.addr.IP[0] ||
		remoteAddr6.Addr[1] != sock.addr.IP[1] ||
		remoteAddr6.Addr[2] != sock.addr.IP[2] ||
		remoteAddr6.Addr[3] != sock.addr.IP[3] ||
		remoteAddr6.Addr[4] != sock.addr.IP[4] ||
		remoteAddr6.Addr[5] != sock.addr.IP[5]) {
		sock = &socket{
			socketFD: socketFD,
			isServer: false,
			addr: net.TCPAddr{
				Port: remoteAddr6.Port,
				IP:   net.IP(remoteAddr6.Addr[:]),
			},
			lastAccessedAt: time.Now(),
		}
		setGlobalSock(socketFD, sock)
	}
	thread.socks[socketFD] = sock
	return sock, nil
}
