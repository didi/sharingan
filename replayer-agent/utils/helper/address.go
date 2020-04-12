package helper

import (
	"net"

	"github.com/pkg/errors"
)

var LocalIp = GetLocalIP()
var PortVal = "8998" //默认值，会被app.toml覆盖

func GetLocalIP() string {
	ip := "127.0.0.1"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ip
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}

	return ip
}

func AssignLocalAddr() (*net.TCPAddr, error) {
	// golang does not provide api to bind before connect
	// this is a hack to assign 127.0.0.1:0 to pre-determine a local port
	listener, err := net.Listen("tcp", "127.0.0.1:0") // ask for new port
	if err != nil {
		return nil, errors.Wrap(err, "listen failed")
	}
	localAddr := listener.Addr().(*net.TCPAddr)
	err = listener.Close()
	if err != nil {
		return nil, errors.Wrap(err, "close failed")
	}
	return localAddr, nil
}
