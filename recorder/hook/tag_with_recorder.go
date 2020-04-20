// +build recorder

package hook

import (
	"net"
	"runtime"
	"syscall"
)

// GetCurrentGoRoutineID GetCurrentGoRoutineID
func GetCurrentGoRoutineID() int64 {
	return runtime.GetCurrentGoRoutineId()
}

// RegisterOnConnect RegisterOnConnect
func RegisterOnConnect(callback func(fd int, sa syscall.Sockaddr)) {
	syscall.OnConnect = callback
}

// RegisterOnAccept RegisterOnAccept
func RegisterOnAccept(callback func(serverSocketFD int, clientSocketFD int, sa syscall.Sockaddr)) {
	syscall.OnAccept = callback
}

// RegisterOnRecv RegisterOnRecv
func RegisterOnRecv(callback func(fd int, net string, raddr net.Addr, span []byte)) {
	net.OnRead = callback
}

// RegisterOnSend RegisterOnSend
func RegisterOnSend(callback func(fd int, net string, raddr net.Addr, span []byte)) {
	net.OnWrite = callback
}

// RegisterOnClose RegisterOnClose
func RegisterOnClose(callback func(fd int)) {
	net.OnClose = callback
}

// RegisterOnGoRoutineExit RegisterOnGoRoutineExit
func RegisterOnGoRoutineExit(callback func(goid int64)) {
	runtime.OnGoRoutineExit = callback
}
