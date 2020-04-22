package httpserv

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/logic/worker"
	"github.com/didi/sharingan/replayer-agent/router"
	"github.com/didi/sharingan/replayer-agent/utils/helper"
)

var srv *http.Server

func Init() {
	srv = &http.Server{}
	srv.Handler = router.Handler
	srv.Addr = conf.Handler.GetString("http.addr")
	if strings.Trim(srv.Addr, ":") != "" {
		helper.PortVal = strings.Trim(srv.Addr, ":")
	}
	srv.ReadHeaderTimeout = time.Millisecond * time.Duration(conf.Handler.GetInt("http.readHeaderTimeout"))
	srv.ReadTimeout = time.Millisecond * time.Duration(conf.Handler.GetInt("http.readTimeout"))
	srv.WriteTimeout = time.Millisecond * time.Duration(conf.Handler.GetInt("http.writeTimeout"))
	srv.IdleTimeout = time.Millisecond * time.Duration(conf.Handler.GetInt("http.idleTimeout"))
	handlerTimeout := time.Millisecond * time.Duration(conf.Handler.GetInt("http.handlerTimeout"))
	srv.Handler = http.TimeoutHandler(srv.Handler, handlerTimeout, "503 Handler timeout")
}

func Run() {
	errCh := make(chan error)
	go func() {
		log.Printf("[http] Server Running! addr:%s.\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil {
			log.Printf("[http] Server Stoped! errmsg:%s.\n", err.Error())
			tlog.Handler.Fatalf(context.TODO(), tlog.DLTagUndefined, "errmsg="+err.Error())
		} else {
			log.Printf("[http] Server Stoped! \n")
			tlog.Handler.Warnf(context.TODO(), tlog.DLTagUndefined, "errmsg=The server stoped !")
		}

		//tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "errmsg="+err.Error())
		errCh <- err
	}()
	wait(errCh)
}

func wait(errCh chan error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case s := <-c:
		tlog.Handler.Warnf(context.TODO(), tlog.DLTagUndefined, "errmsg=Got signal"+s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// Process On shutdown
			func() {
				worker.ExitHook()
			}()
			srv.Shutdown(context.TODO())
		}
	case <-errCh:
	}
}
