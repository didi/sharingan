//+build !linux

package helper

import "net"

func SetQuickAck(conn *net.TCPConn) error {
	return nil
}
