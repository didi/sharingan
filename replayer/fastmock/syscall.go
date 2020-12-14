package fastmock

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/didi/sharingan/replayer/internal"
	"github.com/didi/sharingan/replayer/monkey"
)

const (
	HTTP_SERVER = "HTTP_SERVER"
	GRPC_SERVER = "GRPC_SERVER"
)

var (
	// inbound header feature, include traceID、replay Time
	traceRegex  = regexp.MustCompile(`Sharingan-Replayer-Trace(?:id|ID)\s?: (\w{32})\r\n`)
	timeRegex   = regexp.MustCompile(`Sharingan-Replayer-Time\s?: (\d{19})\r\n`)
	serverRegex = regexp.MustCompile(`Sharingan-Replayer-Server\s?: (\w{11})\r\n`)

	// outbound traffic prefix, include traceID、origin connect addr
	trafficPrefix = `/*{"rid":"%s","addr":"%s"}*/`

	// mysqlGreetingTrace, val === md5("MYSQL_GREETING")
	mysqlGreetingTrace = "ca4bc2ca79c2f79729b322fbfbd91ef3"
)

// MockSyscall mock conn
func MockSyscall() {
	mockTCPConnConnect()
	mockTCPConnRead()
	mockTCPConnWrite()
	mockTCPConnOnClose()
}

// mock syscall.Connect
func mockTCPConnConnect() {
	monkey.MockGlobalFunc(syscall.Connect, func(fd int, sa syscall.Sockaddr) (err error) {
		sockType, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)

		// ignore UDP
		if err == nil && sockType == syscall.SOCK_DGRAM {
			return internal.SyscallConnect(fd, sa)
		}

		// ignore mock info nil
		if mockSaAddr == [4]byte{} || mockSaPort == 0 {
			return internal.SyscallConnect(fd, sa)
		}

		// origin addr info
		rsa := sa.(*syscall.SockaddrInet4)
		addr := net.TCPAddr{IP: rsa.Addr[:], Port: rsa.Port}

		// set globalSockets
		accessTime := internal.TimeNow()
		globalSockets.Set(fd, addr.String(), accessTime)

		// to mock server
		msa := &syscall.SockaddrInet4{Addr: mockSaAddr, Port: mockSaPort}
		err = internal.SyscallConnect(fd, msa)

		// wait for 100ms， if no write(check by accessTime), send mysqlGreeting
		go func() {
			time.Sleep(time.Millisecond * 100)
			if socket := globalSockets.Get(fd); socket != nil {
				if accessTime == socket.lastAccessedAt {
					prefix := fmt.Sprintf(trafficPrefix, mysqlGreetingTrace, addr.String())
					syscall.Write(fd, []byte(prefix))
				}
			}
		}()

		return err
	})
}

// mock conn.Read
func mockTCPConnRead() {
	var c *net.TCPConn

	monkey.MockMemberFunc(reflect.TypeOf(c), "Read", func(conn *net.TCPConn, b []byte) (int, error) {
		threadID := internal.GetCurrentGoRoutineID()
		fd := internal.GetConnFD(conn)

		// accsess
		ReplayerGlobalThreads.Access(threadID)
		// globalSockets.Access(fd)  // ignore Read Access

		// origin Read
		n, err := internal.ConnRead(conn, b)

		// only process inbound Request
		if err != nil || n <= 0 || !isInboundFD(fd) {
			return n, err
		}

		// Inbound Hook
		newb, newn := b, n
		traceID, replayTime, serverType := "", int64(0), ""

		// get and remove server-type header
		if ss := serverRegex.FindAllSubmatch(newb, -1); len(ss) >= 1 {
			serverType = string(ss[0][1])
			newb = bytes.Replace(newb, ss[0][0], []byte(""), -1)
			newn -= len(ss[0][0])
		}

		// get and remove traceID header
		if ss := traceRegex.FindAllSubmatch(newb, -1); GRPC_SERVER != serverType && len(ss) >= 1 {
			traceID = string(ss[0][1])
			newb = bytes.Replace(newb, ss[0][0], []byte(""), -1)
			newn -= len(ss[0][0])
		}

		// get and remove time header
		if ss := timeRegex.FindAllSubmatch(newb, -1); GRPC_SERVER != serverType && len(ss) >= 1 {
			replayTime, _ = strconv.ParseInt(string(ss[0][1]), 10, 64)
			newb = bytes.Replace(newb, ss[0][0], []byte(""), -1)
			newn -= len(ss[0][0])
		}

		if traceID != "" || replayTime != 0 {
			ReplayerGlobalThreads.Set(threadID, traceID, replayTime)
		}

		// remove header
		if len(b) > len(newb) && n > newn {
			copy(b, newb)
			n = newn
		}

		return n, err
	})
}

// mock conn.Write
func mockTCPConnWrite() {
	var c *net.TCPConn

	monkey.MockMemberFunc(reflect.TypeOf(c), "Write", func(conn *net.TCPConn, b []byte) (int, error) {
		threadID := internal.GetCurrentGoRoutineID()
		fd := internal.GetConnFD(conn)

		// accsess
		ReplayerGlobalThreads.Access(threadID)
		globalSockets.Access(fd)

		// ingore inbound
		if isInboundFD(fd) {
			return internal.ConnWrite(conn, b)
		}

		// get traceID
		traceID := ""
		thread := ReplayerGlobalThreads.Get(threadID)
		if thread != nil {
			traceID = thread.traceID
		}

		// get remoteAddr
		remoteAddr := ""
		if socket := globalSockets.Get(fd); socket != nil {
			remoteAddr = socket.remoteAddr
		}

		// add traffic prefix
		prefix := fmt.Sprintf(trafficPrefix, traceID, remoteAddr)
		newb := append([]byte(prefix), b...)

		newn, err := internal.ConnWrite(conn, newb)
		return newn - len(prefix), err
	})
}

// mock conn.OnClose
func mockTCPConnOnClose() {
	internal.RegisterOnClose(func(fd int) {
		globalSockets.Remove(fd)
	})
}

// isInboundFD isInboundFD
func isInboundFD(fd int) bool {
	// if fd exist in globalSockets, judge this outbound（Connect fd）
	if socket := globalSockets.Get(fd); socket != nil {
		return false
	}

	return true
}
