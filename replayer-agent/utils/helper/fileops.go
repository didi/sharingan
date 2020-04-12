package helper

import (
	"bufio"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/didichuxing/sharingan/replayer-agent/common/handlers/tlog"
)

func WriteFile(fpath string, content []byte) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0744)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}

// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

// TempFile creates a new temporary file in the directory dir
// with a name beginning with prefix, opens the file for reading
// and writing, and returns the resulting *os.File.
// If dir is the empty string, TempFile uses the default directory
// for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously
// will not choose the same file.  The caller can use f.Name()
// to find the pathname of the file.  It is the caller's responsibility
// to remove the file when no longer needed.
func TempFileWithSuffix(dir, prefix, suffix string) (f *os.File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+nextSuffix()+suffix)
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				rand = reseed()
			}
			continue
		}
		break
	}
	return
}

//ReadLines 逐行读取并返回文件内容
func ReadLines(confFile string) ([]string, error) {
	file, err := os.Open(confFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var contents []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		contents = append(contents, scanner.Text())
	}
	return contents, scanner.Err()
}

//ReadFile 读取并返回整个文件内容
func ReadFileBytes(confFile string) ([]byte, error) {
	conetent, err := ioutil.ReadFile(confFile)

	return conetent, err
}

//WriteFileString 写入并覆盖配置文件
func WriteFileString(confFile, data string) error {
	fd, err := os.OpenFile(confFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer fd.Close()
	if err != nil {
		tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to open "+confFile+", err="+err.Error())
		return err
	}

	if data != "" && strings.TrimRight(data, "\n") == data {
		data = data + "\n"
	}
	_, err = fd.WriteString(data)
	if err != nil {
		tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to write "+confFile+", err="+err.Error())
		return err
	}

	return nil
}
