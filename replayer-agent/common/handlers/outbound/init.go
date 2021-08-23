package outbound

import (
	"bytes"
	"log"
	"net"

	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/common/handlers/ignore"
	"github.com/didi/sharingan/replayer-agent/logic/outbound"
	"github.com/didi/sharingan/replayer-agent/logic/outbound/match"
)

var BasePort int

func Init() {
	addrStr := conf.Handler.GetString("outbound.server_addr")
	if addrStr == "" {
		addrStr = "127.0.0.1:3515"
	}
	outboundAddr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		log.Fatal("can not resolve outbound addr: " + err.Error())
	}
	BasePort = outboundAddr.Port

	match.AddMatcher(match.FuncMatcher(func(request []byte) bool {
		if len(request) > ignore.MaxProxyHttpLen+8 {
			request = request[:ignore.MaxProxyHttpLen+8]
		}
		for s := range ignore.ProxyHttp {
			if bytes.Contains(request, []byte(s)) {
				return true
			}
		}
		return false
	}))

	// start outbound servers
	outbound.OutboundServer.Handlers = make(map[string]*outbound.Handler)
	// outbound.OutboundServer.Handlers[BasePort] = &outbound.Handler{}
	go outbound.Start(outboundAddr)
}
