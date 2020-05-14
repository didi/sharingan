package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	_ "github.com/didi/sharingan"
)

const waitFlagParseTime = 10

var endRunning chan struct{}
var callback sync.Once

func stop() {
	endRunning <- struct{}{}
}

func signalHandler() {
	// 定义并监听 kill信号, On ^C or SIGTERM
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigc
		callback.Do(stop)
	}()
}

// TestMain Test started when the test binary is started. Only calls main.
func TestMain(m *testing.M) {
	if os.Getenv("BAN_SYSTEM_TEST") == "1" {
		// Original test flow
		os.Exit(m.Run())
		return
	}
	endRunning = make(chan struct{}, 1)
	signalHandler()
	go func() {
		main()
		callback.Do(stop)
	}()
	// Maximum waiting time(10s) for flag.Parse.
	// If the flag still missed to execute after 10 seconds, check your logic with main function.
	checkTime := time.After(waitFlagParseTime * time.Second)
	for {
		if flag.Parsed() {
			break
		}
		select {
		case <-checkTime:
			if !flag.Parsed() {
				flag.Parse()
			}
			break
		case <-endRunning:
			os.Exit(m.Run())
			return
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	<-endRunning
	os.Exit(m.Run())
}
