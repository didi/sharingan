package outbound

import (
	"log"
	"net"

	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/logic/outbound"
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

	// start outbound servers
	outbound.OutboundServer.Handlers = make(map[string]*outbound.Handler)
	// outbound.OutboundServer.Handlers[BasePort] = &outbound.Handler{}
	go outbound.Start(outboundAddr)
}
