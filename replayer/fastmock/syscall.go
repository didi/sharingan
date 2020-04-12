// +build replayer

package fastmock

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/didichuxing/sharingan/replayer/monkey"
)

var (
	// inbound 流量标识
	traceRegex = regexp.MustCompile(`Sharingan-Replayer-TraceID : (\w{32})\r\n`)
	timeRegex  = regexp.MustCompile(`Sharingan-Replayer-Time : (\d{19})\r\n`)

	// traffic 前缀
	trafficPrefix = `/*{"rid":"%s","addr":"%s"}*/`

	// mysqlGreetingTrace, md5("MYSQL_GREETING")
	mysqlGreetingTrace = "ca4bc2ca79c2f79729b322fbfbd91ef3"

	// fd
	globalSocketsMutex = &sync.RWMutex{}
	globalSockets      = map[int]*Socket{}

	// goroutine
	globalThreadsMutex = &sync.RWMutex{}
	globalThreads      = map[int64]*Thread{}
)

// Socket Socket
type Socket struct {
	mutex          *sync.Mutex
	lastAccessedAt time.Time
	remoteAddr     string
}

// Thread Thread
type Thread struct {
	mutex          *sync.Mutex
	lastAccessedAt time.Time
	traceID        string
	replayTime     int64
}

// MockSyscall MockSyscall
func MockSyscall() {
	MockTCPConnConnect()
	MockTCPConnRead()
	MockTCPConnWrite()
	MockTCPConnOnClose()
}

// MockTCPConnConnect mock syscall.Connect
func MockTCPConnConnect() {
	monkey.MockGlobalFunc(syscall.Connect, func(fd int, sa syscall.Sockaddr) (err error) {
		sockType, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)
		if err == nil && sockType == syscall.SOCK_DGRAM {
			return syscall.Connect2(fd, sa)
		}

		if mockSaAddr != [4]byte{} && mockSaPort != 0 {
			accessTime := time.Now2()

			// set globalSockets
			rsa := sa.(*syscall.SockaddrInet4)
			addr := net.TCPAddr{IP: rsa.Addr[:], Port: rsa.Port}
			globalSocketsMutex.Lock()
			globalSockets[fd] = &Socket{remoteAddr: addr.String(), lastAccessedAt: accessTime}
			globalSocketsMutex.Unlock()

			// to mock server
			msa := &syscall.SockaddrInet4{Addr: mockSaAddr, Port: mockSaPort}
			err := syscall.Connect2(fd, msa)

			// 100ms内，fd没有数据写入, 尝试发送mysqlGreeting
			go func() {
				time.Sleep(time.Millisecond * 100)

				globalSocketsMutex.Lock()
				if ffd, ok := globalSockets[fd]; ok {
					if accessTime == ffd.lastAccessedAt {
						prefix := fmt.Sprintf(trafficPrefix, mysqlGreetingTrace, addr.String())
						syscall.Write(fd, []byte(prefix))
						// fmt.Println("===syscall.Write", n, err, fd, addr.String())
					}
				}
				globalSocketsMutex.Unlock()
			}()

			return err
		}

		return syscall.Connect2(fd, sa)
	})
}

// MockTCPConnRead mock net.Read
func MockTCPConnRead() {
	var c *net.TCPConn

	monkey.MockMemberFunc(reflect.TypeOf(c), "Read", func(conn *net.TCPConn, b []byte) (int, error) {
		// accsess
		threadAccess()

		n, err := conn.Read2(b)

		// 只处理inbound请求
		if !isInBoundFD(conn.GetSysFD()) || err != nil || n <= 0 {
			return n, err
		}

		// Inbound Hook
		newb, newn := b, n
		traceID, replayTime := "", int64(0)

		// remove traceID header
		if ss := traceRegex.FindAllSubmatch(newb, -1); len(ss) >= 1 {
			traceID = string(ss[0][1])
			newb = bytes.Replace(newb, ss[0][0], []byte(""), -1)
			newn -= len(ss[0][0])
		}

		// remove time header
		if ss := timeRegex.FindAllSubmatch(newb, -1); len(ss) >= 1 {
			replayTime, _ = strconv.ParseInt(string(ss[0][1]), 10, 64)
			newb = bytes.Replace(newb, ss[0][0], []byte(""), -1)
			newn -= len(ss[0][0])
		}

		// set globalThreads
		if traceID != "" || replayTime != 0 {
			threadID := runtime.GetCurrentGoRoutineId()
			globalThreadsMutex.Lock()
			globalThreads[threadID] = &Thread{
				mutex:          &sync.Mutex{},
				lastAccessedAt: time.Now2(),
				traceID:        traceID,
				replayTime:     replayTime,
			}
			globalThreadsMutex.Unlock()
		}

		// fmt.Printf("traceID:%s, replayTime:%d, n:%d\n", traceID, replayTime, n)

		// remove header
		if len(b) > len(newb) && n > newn {
			copy(b, newb)
			n = newn
		}

		return n, err
	})
}

// MockTCPConnWrite mock net.Write
func MockTCPConnWrite() {
	var c *net.TCPConn

	monkey.MockMemberFunc(reflect.TypeOf(c), "Write", func(conn *net.TCPConn, b []byte) (int, error) {
		// accsess
		threadAccess()

		// 只处理outbound请求
		if isInBoundFD(conn.GetSysFD()) {
			return conn.Write2(b)
		}

		// traceID标识
		traceID := ""
		threadID := runtime.GetCurrentGoRoutineId()
		globalThreadsMutex.Lock()
		thread := globalThreads[threadID]
		if thread != nil {
			traceID = thread.traceID
		}
		globalThreadsMutex.Unlock()

		// remoteAddr标识
		remoteAddr := ""
		globalSocketsMutex.Lock()
		if fd, ok := globalSockets[conn.GetSysFD()]; ok {
			fd.lastAccessedAt = time.Now2()
			remoteAddr = fd.remoteAddr
		}
		globalSocketsMutex.Unlock()

		// 加流量标识
		prefix := fmt.Sprintf(trafficPrefix, traceID, remoteAddr)
		newb := append([]byte(prefix), b...)
		var newn int
		var err error
		if thread != nil {
			thread.mutex.Lock()
			newn, err = conn.Write2(newb)
			thread.mutex.Unlock()
		} else {
			newn, err = conn.Write2(newb)
		}

		// newn, err := conn.Write2(newb)
		// str := fmt.Sprintf("%s, goid:%d, len(globalThreads):%d, len(globalSockets):%d", prefix, threadID, len(globalThreads), len(globalSockets))
		// fmt.Println(str, string(newb))

		return newn - len(prefix), err
	})
}

// MockTCPConnOnClose mock net.OnClose
func MockTCPConnOnClose() {
	net.OnClose = func(fd int) {
		globalSocketsMutex.Lock()
		if _, ok := globalSockets[fd]; ok {
			delete(globalSockets, fd)
		}
		globalSocketsMutex.Unlock()
	}
}

// threadAccess goroutine access
func threadAccess() {
	threadID := runtime.GetCurrentGoRoutineId()
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	if thread := globalThreads[threadID]; thread != nil {
		thread.lastAccessedAt = time.Now2()
	}
}

// isInBoundFD 判断是不是outbound的fd
func isInBoundFD(fd int) bool {
	globalSocketsMutex.RLock()
	defer globalSocketsMutex.RUnlock()

	// connect存入map的fd都是outbound
	if _, ok := globalSockets[fd]; ok {
		return false
	}

	return true
}

func init() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			gcGlobalhreads()
		}
	}()
}

// gcGlobalhreads 回收thread
func gcGlobalhreads() int {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	now := time.Now2()
	newMap := map[int64]*Thread{}
	expiredThreadsCount := 0

	for threadID, thread := range globalThreads {
		if now.Sub(thread.lastAccessedAt) < time.Second*5 {
			newMap[threadID] = thread
		}
		// else {
		// 	fmt.Println("gc threadID:", threadID)
		// }
	}

	globalThreads = newMap
	return expiredThreadsCount
}
