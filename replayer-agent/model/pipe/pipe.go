package pipe

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didichuxing/sharingan/replayer-agent/utils/helper"

	"github.com/pkg/errors"
)

const (
	ApolloSuffix = ".old"
)

func Write(ctx context.Context, fpath string, content [][]byte) error {
	defer func() {
		if err := recover(); err != nil {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "panic in %s goroutine||errmsg=%s||stack info=%s", "PipeWrite", err, strconv.Quote(string(debug.Stack())))
		}
	}()
	defer os.Remove(fpath)

	err := os.MkdirAll(filepath.Dir(fpath), 0744)
	if err != nil {
		return errors.Wrap(err, "os MkdirAll")
	}
	err = syscall.Mkfifo(fpath, 0666)
	if err != nil {
		return errors.Wrap(err, "syscall Mkfifo")
	}

	// another goroutine: force WritePipe to exit
	go exit(ctx, fpath)

	var f *os.File
	for {
		select {
		case <-ctx.Done():
			return err
		default:
			f, err = os.OpenFile(fpath, os.O_WRONLY|os.O_SYNC, os.ModeNamedPipe)
			if err != nil {
				tlog.Handler.Errorf(ctx, tlog.DebugTag, "filename=%s||errmsg=open named pipe failed||err=%s", fpath, err)
				continue
			}
			tlog.Handler.Debugf(ctx, tlog.DebugTag, "%s||toggleName=%s||content=%s", helper.CInfo("[[[replay apollo toggle]]]"), fpath, string(content[0]))
			_, err = f.Write(content[0])
			if err != nil {
				tlog.Handler.Errorf(ctx, tlog.DebugTag, "filename=%s||content=%s||errmsg=write into pipe failed||err=%s", fpath, string(content[0]), err)
			}
			f.Close()
			if len(content) > 1 {
				content = content[1:]
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func exit(ctx context.Context, fpath string) {
	defer func() {
		if err := recover(); err != nil {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "panic in %s goroutine||errmsg=%s||stack info=%s", "PipeExit", err, strconv.Quote(string(debug.Stack())))
		}
	}()
	select {
	case <-ctx.Done():
		if _, err := os.Stat(fpath); err == nil {
			os.OpenFile(fpath, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
		}
	}
}
