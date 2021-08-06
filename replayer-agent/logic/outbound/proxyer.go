package outbound

import (
	"context"
	"net"
	"strconv"

	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/utils/helper"
	"go.uber.org/zap/zapcore"
)

// Proxyer Proxyer
type Proxyer struct {
	srcConn *net.TCPConn
	dstConn *net.TCPConn
}

// NewProxyer 新建代理
func NewProxyer(srcConn *net.TCPConn) *Proxyer {
	return &Proxyer{srcConn: srcConn}
}

// Write 往代理连接写数据
func (p *Proxyer) Write(ctx context.Context, proxyAddr string, request []byte) error {
	// 1、获取或者新建dstConn
	if p.dstConn == nil {
		tcpAddr, err := net.ResolveTCPAddr("tcp", proxyAddr)
		if err != nil {
			tlog.Handler.Errorf(ctx, tlog.DebugTag, "errmsg=DialTCP error||err=%s", err)
			return err
		}
		p.dstConn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			tlog.Handler.Errorf(ctx, tlog.DebugTag, "errmsg=DialTCP error||err=%s", err)
			return err
		}
		tlog.Handler.Debugf(ctx, tlog.DebugTag, "dial tcp %s success", proxyAddr)

		// 1.2、代理
		errch := make(chan error, 1)
		go p.proxy(ctx, errch)
	} else {
		// 校验dstConn信息
		if p.dstConn.RemoteAddr().String() != proxyAddr {
			tlog.Handler.Errorf(ctx, tlog.DebugTag,
				"%s||proxyAddr=%s||dstConn.RemoteAddr()=%s||request=%s",
				helper.CInfo("<<<outbound proxy addr missmatch<<<"),
				proxyAddr,
				p.dstConn.RemoteAddr().String(),
				request)
		}
	}

	// 2、write dstConn
	if len(request) > 0 {
		p.dstConn.Write(request)
	}

	tlog.Handler.Infof(ctx, tlog.DebugTag,
		"%s||proxyAddr=%s||request=%s", helper.CInfo("<<<response of outbound||proxy<<<"), proxyAddr, request)

	return nil
}

func (p *Proxyer) proxy(ctx context.Context, errch chan error) {
	for {
		buffer := make([]byte, 2048)
		rnum, err := p.dstConn.Read(buffer)
		if err != nil || rnum <= 0 {
			errch <- err
			return
		}

		if tlog.Handler.Enable(zapcore.DebugLevel) {
			tlog.Handler.Debugf(ctx, tlog.DebugTag, "proxyRead=%s", strconv.QuoteToASCII(string(buffer[:rnum])))
		}

		_, err = p.srcConn.Write(buffer[:rnum])
		if err != nil {
			errch <- err
			return
		}

		if tlog.Handler.Enable(zapcore.DebugLevel) {
			tlog.Handler.Debugf(ctx, tlog.DebugTag, "proxyWrite=%s", strconv.QuoteToASCII(string(buffer[:rnum])))
		}
	}
}

// Close 关闭代理
func (p *Proxyer) Close() {
	if p.dstConn != nil {
		p.dstConn.Close()
	}
}
