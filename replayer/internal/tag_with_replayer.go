// +build replayer

package internal

import (
	"net"
	"runtime"
	"syscall"
	"time"
)

// TimeNow origin time.Now
func TimeNow() time.Time {
	return time.Now2()
}

// SyscallConnect origin syscall.Connect
func SyscallConnect(fd int, sa syscall.Sockaddr) (err error) {
	return syscall.Connect2(fd, sa)
}

// ConnRead origin conn.Read
func ConnRead(conn *net.TCPConn, b []byte) (int, error) {
	return conn.Read2(b)
}

// ConnWrite origin conn.Write
func ConnWrite(conn *net.TCPConn, b []byte) (int, error) {
	return conn.Write2(b)
}

// GetConnFD get fd
func GetConnFD(conn *net.TCPConn) int {
	return conn.GetSysFD()
}

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return runtime.GetCurrentGoRoutineId()
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
	runtime.SetDelegatedFromGoRoutineId(gID)
}

// RegisterOnClose RegisterOnClose
func RegisterOnClose(callback func(fd int)) {
	net.OnClose = callback
}
