package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"

	_ "github.com/didi/sharingan"
)

var systemTest *bool
var endRunning chan bool

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
	systemTest = flag.Bool("systemTest", false, "Set to true when running system tests")
}

// Test started when the test binary is started. Only calls main.
func Test_main(t *testing.T) {
	if *systemTest {
		signalHandler()
		endRunning = make(chan bool, 1)
		go main()
		<-endRunning
	}
}
