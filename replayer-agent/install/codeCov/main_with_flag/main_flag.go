// Only for the flag requirements in business server!
package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/didichuxing/sharingan/replayer"

	// TODO：最后import其他业务包！
)

var endRunning chan bool

// TODO: you can change the flags below to your own flags!
var flag1 *bool
var flag2 *bool

func stop() {
	endRunning <- true
}

func signalHandler() {
	var callback sync.Once
	// 定义并监听 kill信号, On ^C or SIGTERM
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigc
		callback.Do(stop)
	}()
}

func init() {
	// TODO: you can change the flags below to your own flags!
	flag1 = flag.Bool("flag1", false, "flag demo1")
	flag2 = flag.Bool("flag2", false, "flag demo2")
}

// main
func main() {
	signalHandler()
	endRunning = make(chan bool, 1)
	go func() {
		// TODO: codes here for server init & run!
	}()
	<-endRunning
}
