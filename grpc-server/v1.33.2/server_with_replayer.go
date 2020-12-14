// +build replayer

package grpc

import (
	"runtime"
	"strconv"
	"strings"

	"google.golang.org/grpc/internal/transport"
	"google.golang.org/grpc/metadata"

    "github.com/didi/sharingan/replayer/fastmock"
)

func handleReplayerHeader(stream *transport.Stream) {
	// get ctx by stream
	ctx := NewContextWithServerTransportStream(stream.Context(), stream)
	// get metadata by ctx
	md, _ := metadata.FromIncomingContext(ctx)
	// get traceid & replaytime from header by metadata
	traceID := md["sharingan-replayer-traceid"]
	replayTime := md["sharingan-replayer-time"]
	// format
	rTime, _ := strconv.ParseInt(strings.Join(replayTime, ""), 10, 64)
	goid := runtime.GetCurrentGoRoutineId()
	// set ReplayerGloabalThreads
	fastmock.ReplayerGlobalThreads.Access(goid)
	fastmock.ReplayerGlobalThreads.Set(goid, strings.Join(traceID, ""), rTime)
}
