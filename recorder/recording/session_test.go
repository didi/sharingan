package recording

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSession_RecvFromInbound(t *testing.T) {
	should := require.New(t)
	ctx := context.Background()

	threadID := int32(1)
	inboundAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:9999")
	inboundRequest1 := []byte("aaa")
	inboundRequest2 := []byte("bbb")
	inboundRequest3 := []byte("Expect: 100-continue\r\nccc")
	expectRequest := []byte("aaabbbccc") // merge request, ignore 100-continue

	session := NewSession(threadID)
	session.RecvFromInbound(ctx, inboundRequest1, *inboundAddr, net.UnixAddr{})
	session.RecvFromInbound(ctx, inboundRequest2, *inboundAddr, net.UnixAddr{})
	session.RecvFromInbound(ctx, inboundRequest3, *inboundAddr, net.UnixAddr{})

	should.Contains(session.SessionID, fmt.Sprintf("-%d", threadID))
	should.Equal(session.CallFromInbound.ActionType, "CallFromInbound")
	should.Equal(session.CallFromInbound.Peer.String(), inboundAddr.String())
	should.Equal(session.CallFromInbound.Request, expectRequest)
}

func TestSession_SendToInbound(t *testing.T) {
	should := require.New(t)
	ctx := context.Background()

	threadID := int32(1)
	inboundAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:9999")
	inboundResponse1 := []byte("HTTP/1.1 100 Continue")
	inboundResponse2 := []byte("aaa")
	expectResponse := []byte("aaa") // merge request, ignore 100-continue

	session := NewSession(threadID)
	session.SendToInbound(ctx, inboundResponse1, *inboundAddr)
	session.SendToInbound(ctx, inboundResponse2, *inboundAddr)

	should.Contains(session.SessionID, fmt.Sprintf("-%d", threadID))
	should.Equal(session.ReturnInbound.ActionType, "ReturnInbound")
	should.Equal(session.ReturnInbound.Response, expectResponse)
}

func TestSession_Outbound(t *testing.T) {
	should := require.New(t)
	ctx := context.Background()

	// 127.0.0.1:8888 -> 127.0.0.1:9999
	threadID := int32(1)
	outboundLocalAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8888")
	outboundRemoteAddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:9999")

	outboundRequest1 := []byte("aaa")
	outboundRequest2 := []byte("bbb")
	outboundResponse1 := []byte("AAA")
	outboundResponse2 := []byte("BBB")
	expectRequest := []byte("aaabbb")
	expectResponse := []byte("AAABBB")

	socketFD := 1

	session := NewSession(threadID)
	session.SendToOutbound(ctx, outboundRequest1, *outboundRemoteAddr, outboundLocalAddr, socketFD)
	session.SendToOutbound(ctx, outboundRequest2, *outboundRemoteAddr, outboundLocalAddr, socketFD)
	session.RecvFromOutbound(ctx, outboundResponse1, *outboundRemoteAddr, outboundLocalAddr, socketFD)
	session.RecvFromOutbound(ctx, outboundResponse2, *outboundRemoteAddr, outboundLocalAddr, socketFD)

	action := session.currentCallOutbound

	should.Equal("CallOutbound", action.ActionType)
	should.Equal(socketFD, action.SocketFD)
	should.Equal(outboundRemoteAddr.String(), action.Peer.String())
	should.Equal(expectRequest, action.Request)
	should.Equal(expectResponse, action.Response)
}
