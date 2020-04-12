package helper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"

	"github.com/pkg/errors"
)

func StartDetached(cmd string, args []string) error {
	var attr = os.ProcAttr{
		Dir: ".",
		Env: os.Environ(),
		Files: []*os.File{
			nil,
			nil,
			nil,
		},
		Sys: &syscall.SysProcAttr{Foreground: false},
	}
	process, err := os.StartProcess(cmd, args, &attr)
	if err != nil {
		return errors.Wrap(err, "start process failed")
	}
	// It is not clear from docs, but Realease actually detaches the process
	err = process.Release()
	if err != nil {
		return errors.Wrap(err, "release process failed")
	}
	return nil
}

// GetPidByName 根据进程名字，读取进程id
func GetPidByName(binName string) (int, string) {
	//读取 pid
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("ps -ef | grep '%s' | grep -v grep | grep -v '/bin/bash -c' | awk '{print $2}'", binName))
	out, err := cmd.Output()
	if err != nil {
		msg := "Errors happened when getting pid of " + binName + "," + err.Error()
		tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, msg)
		return 1, msg //遇到错误
	} else if len(out) == 0 {
		return 2, "Process " + binName + " is not running! " //进程不存在
	}
	pid := BytesToString(out)

	return 0, pid
}

// MkdirForce 强制创建目录
func MkdirForce(dir string) int {
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("mkdir -p %s", dir))
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) != 0 {
		tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to mkdir "+dir+"!")
		return 1
	}

	return 0
}

// MVFile 重命名文件
func MVFile(ori, des string) int {
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("mv -f %s %s", ori, des))
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) != 0 {
		tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to mv -f "+ori+" "+des+"!")
		return 1
	}

	return 0
}

//PathExists 判断路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
