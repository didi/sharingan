// +build !recorder

package internal

import (
	"net"
	"syscall"
)

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return 0
}

// SetDelegatedFromGoRoutineID SetDelegatedFromGoRoutineID
func SetDelegatedFromGoRoutineID(gID int64) {
}

// RegisterOnConnect RegisterOnConnect
func RegisterOnConnect(callback func(fd int, sa syscall.Sockaddr)) {
}

// RegisterOnAccept RegisterOnAccept
func RegisterOnAccept(callback func(serverSocketFD int, clientSocketFD int, sa syscall.Sockaddr)) {
}

// RegisterOnRecv RegisterOnRecv
func RegisterOnRecv(callback func(fd int, net string, raddr net.Addr, span []byte)) {
}

// RegisterOnSend RegisterOnSend
func RegisterOnSend(callback func(fd int, net string, raddr net.Addr, span []byte)) {
}

// RegisterOnClose RegisterOnClose
func RegisterOnClose(callback func(fd int)) {
}

// RegisterOnGoRoutineExit RegisterOnGoRoutineExit
func RegisterOnGoRoutineExit(callback func(goid int64)) {
}
