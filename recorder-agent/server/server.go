package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// 默认时间
	defaultHTTPQuitTimeout = 5 * time.Second
)

// Server Server
type Server struct {
	httpSvr  *http.Server // http 服务
	httpAddr string       // http 地址

	httpServerMux *http.ServeMux // http ServeMux

	httpQuitTimeout    time.Duration
	httpHandlerTimeout time.Duration
	httpReadTimeout    time.Duration
	httpWriteTimeout   time.Duration
	httpIdleTimeout    time.Duration

	errChan chan error
}

// New Server实例
func New() *Server {
	return &Server{
		httpServerMux: http.DefaultServeMux,

		httpQuitTimeout:    defaultHTTPQuitTimeout,
		httpHandlerTimeout: 0,
		httpReadTimeout:    0,
		httpWriteTimeout:   0,
		httpIdleTimeout:    0,

		errChan: make(chan error),
	}
}

// SetHTTPAddr 设置地址
func (s *Server) SetHTTPAddr(addr string) {
	s.httpAddr = addr
}

// SetHTTPHandlerTimeout 设置程序处理超时时间，不包含读写时间
func (s *Server) SetHTTPHandlerTimeout(t int) {
	s.httpHandlerTimeout = time.Millisecond * time.Duration(t)
}

// SetHTTPReadTimeout 设置读取header+body的整体超时时间
func (s *Server) SetHTTPReadTimeout(t int) {
	s.httpReadTimeout = time.Millisecond * time.Duration(t)
}

// SetHTTPWriteTimeout 设置写操作超时时间
func (s *Server) SetHTTPWriteTimeout(t int) {
	s.httpWriteTimeout = time.Millisecond * time.Duration(t)
}

// SetHTTPIdleTimeout 设置连接空闲超时时间
func (s *Server) SetHTTPIdleTimeout(t int) {
	s.httpIdleTimeout = time.Millisecond * time.Duration(t)
}

// SetHTTPQuitTimeout 设置http退出时间
func (s *Server) SetHTTPQuitTimeout(t int) {
	s.httpQuitTimeout = time.Millisecond * time.Duration(t)
}

// AddHTTPHandle 增加http handle
func (s *Server) AddHTTPHandle(pattern string, handler http.Handler) {
	s.httpServerMux.Handle(pattern, handler)
}

// AddHTTPHandleFunc 增加http handle func
func (s *Server) AddHTTPHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.httpServerMux.HandleFunc(pattern, handler)
}

// Run 启动Server
func (s *Server) Run() error {
	go func() {
		s.errChan <- s.runHTTPServer()
	}()

	return s.wait()
}

// runHTTPServer 启动http服务
func (s *Server) runHTTPServer() error {
	s.httpSvr = &http.Server{
		Addr:    s.httpAddr,
		Handler: s.httpServerMux,
	}

	if s.httpHandlerTimeout != 0 {
		s.httpSvr.Handler = http.TimeoutHandler(s.httpSvr.Handler, s.httpHandlerTimeout, "503 Handler timeout")
	}

	if s.httpReadTimeout != 0 {
		s.httpSvr.ReadTimeout = s.httpReadTimeout
	}

	if s.httpWriteTimeout != 0 {
		s.httpSvr.WriteTimeout = s.httpWriteTimeout
	}

	if s.httpIdleTimeout != 0 {
		s.httpSvr.IdleTimeout = s.httpIdleTimeout
	}

	return s.httpSvr.ListenAndServe()
}

// wait
func (s *Server) wait() error {
	var err error

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-c:
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), s.httpQuitTimeout)
			defer cancel()
			s.httpSvr.Shutdown(ctx)
		case syscall.SIGQUIT:
			s.httpSvr.Close()
		}

		log.Printf("[http] server closed got signal %s, shutdown success\n", sig)
		return fmt.Errorf("server closed got signal %s, shutdown success", sig)
	case err = <-s.errChan:
		return fmt.Errorf("server closed %s", err)
	}
}
