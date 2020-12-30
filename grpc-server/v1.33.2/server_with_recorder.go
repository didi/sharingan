// +build recorder_grpc

package grpc

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/grpc/internal/transport"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/didi/sharingan/recorder/koala_grpc/hook"
)

const (
	h2cAuthority = ":authority"
	h2cMethod    = ":method"
	h2cPath      = ":path"
	h2cScheme    = ":scheme"
)

var http2ClientHeader = []string{
	h2cAuthority,
	h2cMethod,
	h2cPath,
	h2cScheme,
}

func registerOnGrpcAccept(stream *transport.Stream) {
	var ip net.IP
	var port int

	p, _ := peer.FromContext(stream.Context())
	addr := p.Addr.String() // p.Addr.String() (for example, "192.0.2.1:25", "[2001:db8::1]:80")
	tInfo := strings.Split(addr, ":")
	if len(tInfo) < 2 {
		hook.RegisterOnGrpcAccept(ip, port)
		return
	}
	regReplace := regexp.MustCompile(`\[|\]`)
	ip = net.ParseIP(regReplace.ReplaceAllString(addr[:len(addr)-len(tInfo[len(tInfo)-1])-1], ""))
	port, _ = strconv.Atoi(tInfo[len(tInfo)-1])
	hook.RegisterOnGrpcAccept(ip, port)
}

func registerOnGrpcRecv(stream *transport.Stream, span []byte) {
	var header string

	md, _ := metadata.FromIncomingContext(stream.Context())
	for _, h2cHeader := range http2ClientHeader {
		if h2cVavlue, ok := md[h2cHeader]; ok {
			header = header + h2cHeader + ": " + strings.Join(h2cVavlue, "") + "\r\n"
			delete(md, h2cHeader)
		} else if h2cHeader == h2cPath {
			header = header + h2cHeader + ": " + stream.Method() + "\r\n"
		} else if h2cHeader == h2cScheme {
			// from grpc-gateway
			//if scheme, sOk := md["x-forwarded-proto"]; sOk {
			//	header = header + h2cHeader + ": " + strings.Join(scheme, "") + "\r\n"
			//} else {
			//	header = header + h2cHeader + ": " + "http" + "\r\n"
			//}
		} else if h2cHeader == h2cMethod {
			//header = header + h2cHeader + ": " + "POST" + "\r\n"
		}
	}
	for hKey, hVal := range md {
		header = header + hKey + ": " + strings.Join(hVal, "") + "\r\n"
	}
	header = header + "\r\n"
	hook.RegisterOnGrpcRecv(append([]byte(header), span...))
}

func registerOnGrpcSend(span []byte) {
	hook.RegisterOnGrpcSend(span)
}
