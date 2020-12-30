// +build !recorder_grpc

package grpc

import (
	"google.golang.org/grpc/internal/transport"
)

func registerOnGrpcAccept(stream *transport.Stream) {
}

func registerOnGrpcRecv(stream *transport.Stream, span []byte) {
}

func registerOnGrpcSend(span []byte) {
}
