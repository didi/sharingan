package sut

import (
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/v2pro/plz/countlog"
)

// SocketFD SocketFD
type SocketFD int

// ThreadID ThreadID
type ThreadID int32

//grpc outbound
var globalSocks = map[SocketFD]*socket{}
var globalSocksMutex = &sync.Mutex{}

// grpc inbound
var globalGidSocks = map[SocketFD]*socket{}
var globalGidSocksMutex = &sync.Mutex{}
var globalThreads = map[ThreadID]*Thread{} // real thread id => virtual thread
var globalThreadsMutex = &sync.Mutex{}

func init() {
	if os.Getenv("RECORDER_ENABLED") == "true" {
		go gcStatesInBackground()
	}
}

/* 全局gc */

// gcStatesInBackground 后台gc
func gcStatesInBackground() {
	defer func() {
		if recovered := recover(); recovered != nil {
			countlog.LogPanic(recovered, "sut.gc_states_in_background.panic")
		}
	}()

	for {
		time.Sleep(time.Second * 10)

		expiredSocksCount := gcGlobalSocks()
		expiredRealThreadsCount := gcGlobalRealThreads()

		countlog.Trace("event!sut.gc_global_states",
			"expiredSocksCount", expiredSocksCount,
			"expiredRealThreadsCount", expiredRealThreadsCount)
	}
}

// gcGlobalSocks 回收socket
func gcGlobalSocks() int {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	now := time.Now()
	newMap := map[SocketFD]*socket{}
	expiredSocksCount := 0

	for fd, sock := range globalSocks {
		if now.Sub(sock.lastAccessedAt) < time.Minute*5 {
			newMap[fd] = sock
		} else {
			expiredSocksCount++
		}
	}

	globalSocks = newMap
	return expiredSocksCount
}

// gcGlobalRealThreads 回收thread
func gcGlobalRealThreads() int {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	now := time.Now()
	newMap := map[ThreadID]*Thread{}
	expiredThreadsCount := 0

	for threadID, thread := range globalThreads {
		if now.Sub(thread.lastAccessedAt) < time.Second*5 {
			newMap[threadID] = thread
		} else {
			thread.mutex.Lock()
			thread.OnShutdown()
			thread.mutex.Unlock()
			expiredThreadsCount++
		}
	}

	globalThreads = newMap
	return expiredThreadsCount
}

/* 全局socket管理 */

// RemoveGlobalSock 移除socket， case: Close
func RemoveGlobalSock(socketFD SocketFD) {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	sock := globalSocks[socketFD]
	if sock != nil {
		delete(globalSocks, socketFD)
	}
}

// RemoveGlobalGidSock 移除socket， case: goroutine exit
func RemoveGlobalGidSock(gid SocketFD) {
	globalGidSocksMutex.Lock()
	defer globalGidSocksMutex.Unlock()

	sock := globalGidSocks[gid]
	if sock != nil {
		delete(globalGidSocks, gid)
	}
}

// AddGlobalSock 新增socket， case: Connect
func AddGlobalSock(socketFD SocketFD, remoteAddr net.TCPAddr, isServer bool) {
	sock := &socket{
		socketFD: socketFD,
		isServer: isServer,
		addr:     remoteAddr,
	}
	setGlobalSock(socketFD, sock)
}

// AddGlobalGidSock 新增socket， case: grpc Accept
func AddGlobalGidSock(gid SocketFD, remoteAddr net.TCPAddr, isServer bool) {
	sock := &socket{
		socketFD: gid,
		isServer: isServer,
		addr:     remoteAddr,
	}
	setGlobalGidSock(gid, sock)
}

// setGlobalSock setGlobalSock
func setGlobalSock(socketFD SocketFD, sock *socket) {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	sock.lastAccessedAt = time.Now()
	globalSocks[socketFD] = sock
}

// setGlobalGidSock setGlobalGidSock
func setGlobalGidSock(gid SocketFD, sock *socket) {
	globalGidSocksMutex.Lock()
	defer globalGidSocksMutex.Unlock()

	sock.lastAccessedAt = time.Now()
	globalGidSocks[gid] = sock
}

// GetGlobalSock GetGlobalSock
func GetGlobalSock(socketFD SocketFD) *socket {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	sock := globalSocks[socketFD]
	if sock != nil {
		sock.lastAccessedAt = time.Now()
	}

	return sock
}

// GetGlobalGidSock GetGlobalGidSock
func GetGlobalGidSock(gid SocketFD) *socket {
	globalGidSocksMutex.Lock()
	defer globalGidSocksMutex.Unlock()

	sock := globalGidSocks[gid]
	if sock != nil {
		sock.lastAccessedAt = time.Now()
	}

	return sock
}

// exportSocks 导出全部socket
func exportSocks() map[string]interface{} {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	state := map[string]interface{}{}
	for socketFD, sock := range globalSocks {
		state[strconv.Itoa(int(socketFD))] = sock
	}

	return state
}

/* 全局thread管理 */

// OperateThread create when session notExist
func OperateThread(threadID ThreadID, op func(thread *Thread)) {
	thread := getThread(threadID)

	thread.mutex.Lock()
	defer thread.mutex.Unlock()

	thread.OnAccess()
	thread.lastAccessedAt = time.Now()

	op(thread)
}

// OperateThreadOnRecordingSession only on RecordingSession
func OperateThreadOnRecordingSession(threadID ThreadID, op func(thread *Thread)) {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	thread := globalThreads[threadID]

	// 录制session不存在CallFromInbound提前结束
	if thread == nil ||
		thread.recordingSession == nil ||
		thread.recordingSession.CallFromInbound == nil ||
		len(thread.recordingSession.CallFromInbound.Request) == 0 {
		return
	}

	thread.mutex.Lock()
	defer thread.mutex.Unlock()

	thread.OnAccess()
	thread.lastAccessedAt = time.Now()

	op(thread)
}

// getThread 获取Thread信息，没有会新建一个
func getThread(threadID ThreadID) *Thread {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	thread := globalThreads[threadID]
	if thread == nil {
		thread = newThread(threadID)
		globalThreads[threadID] = thread
	}

	return thread
}

// exportThreads 导出全部Thread
func exportThreads() map[string]interface{} {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	state := map[string]interface{}{}
	for threadID, thread := range globalThreads {
		state[strconv.Itoa(int(threadID))] = thread
	}

	return state
}
