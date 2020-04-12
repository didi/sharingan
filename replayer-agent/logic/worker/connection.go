package worker

import (
	"context"
	"net"
	"time"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
)

func newConn(ctx context.Context, remoteAddr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, "errmsg=connect to sut failed||err=%s", err)
	}

	return conn, err
}

func handleConn(ctx context.Context, conn net.Conn, req []byte, closeConn bool, project string) ([]byte, error) {
	defer func() {
		if closeConn {
			conn.Close()
		}
	}()
	_, err := conn.Write(req)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, "errmsg=write to sut failed||err=%s", err)
		return nil, err
	}
	response, err := readResponse(conn, project)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, "errmsg=read from sut failed||err=%s", err)
	}
	return response, err
}

func readResponse(conn net.Conn, project string) ([]byte, error) {
	buf := make([]byte, 1024)
	t1 := time.Second * 60
	t2 := time.Second * 80
	conn.SetReadDeadline(time.Now().Add(t1))
	bytesRead, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	response := []byte{}
	response = append(response, buf[:bytesRead]...)
	if bytesRead < len(buf) {
		return response, nil
	}

	for {
		conn.SetReadDeadline(time.Now().Add(t2))
		bytesRead, err = conn.Read(buf)
		if err != nil {
			break
		}
		response = append(response, buf[:bytesRead]...)
		if bytesRead < len(buf) {
			return response, nil
		}
	}
	return response, nil
}
