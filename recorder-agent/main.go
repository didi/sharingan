package main

import (
	"context"
	"log"
	"net"
	"runtime/debug"

	"github.com/didichuxing/sharingan/recorder-agent/common/conf"
	"github.com/didichuxing/sharingan/recorder-agent/common/zap"
	"github.com/didichuxing/sharingan/recorder-agent/proto"
	"github.com/didichuxing/sharingan/recorder-agent/record"

	"google.golang.org/grpc"
)

var (
	svr      = grpc.NewServer()
	grpcAddr = conf.Handler.GetString("grpc.grpc_addr")
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] %s\n%s", r, debug.Stack())
		}
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("[grpc] Listen failed! error:%s.", err)
	}

	proto.RegisterAgentServer(svr, new(Controller))

	log.Printf("[grpc] Server Running! addr:%s.", grpcAddr)
	if err := svr.Serve(lis); err != nil {
		log.Fatalf("[grpc] Server failed! error:%s.", err)
	}
}

// Controller 业务逻辑
type Controller struct{}

// Record Record
func (s *Controller) Record(ctx context.Context, req *proto.RecordReq) (*proto.RecordRsp, error) {
	// 异常
	isFilter, err := record.Fliter(req.EsData)
	if err != nil {
		return nil, err
	}

	// 正常，不过滤情况下才录制
	res := &proto.RecordRsp{Data: "FLITER"}
	if !isFilter {
		res.Data = "OK"
		zap.Logger.Info(req.EsData) // 日志收集，最终入ES
	}

	return res, nil
}
