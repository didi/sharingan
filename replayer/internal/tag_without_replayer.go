// +build !replayer
// +build !recorder

package internal

import (
	"net"
	"syscall"
	"time"
)

// TimeNow origin time.Now
func TimeNow() time.Time {
	return time.Now()
}

// SyscallConnect origin syscall.Connect
func SyscallConnect(fd int, sa syscall.Sockaddr) (err error) {
	return syscall.Connect(fd, sa)
}

// ConnRead origin conn.Read
func ConnRead(conn *net.TCPConn, b []byte) (int, error) {
	return conn.Read(b)
}

// ConnWrite origin conn.Write
func ConnWrite(conn *net.TCPConn, b []byte) (int, error) {
	return conn.Write(b)
}

// GetConnFD get conn fd
func GetConnFD(conn *net.TCPConn) int {
	return 0
}

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return 0
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
}

// RegisterOnClose RegisterOnClose
func RegisterOnClose(callback func(fd int)) {
}
