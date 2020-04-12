//+build linux

package helper

import (
	"net"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func SetQuickAck(conn *net.TCPConn) error {
	var f *os.File
	var err error
	if f, err = conn.File(); err != nil {
		return err
	}
	defer f.Close()

	fd := int(f.Fd())
	if err := unix.SetNonblock(fd, true); err != nil {
		return err
	}
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_QUICKACK, 1)
}
