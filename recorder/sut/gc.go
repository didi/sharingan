package sut

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/v2pro/plz/countlog"
)

// SocketFD SocketFD
type SocketFD int

// ThreadID ThreadID
type ThreadID int32

var globalSocks = map[SocketFD]*socket{}
var globalSocksMutex = &sync.Mutex{}
var globalThreads = map[ThreadID]*Thread{} // real thread id => virtual thread
var globalThreadsMutex = &sync.Mutex{}

// StartGC gc global socket and thread
func StartGC() {
	go gcStatesInBackground()
}

/* global gc */

// gcStatesInBackground background gc
func gcStatesInBackground() {
	defer func() {
		if recovered := recover(); recovered != nil {
			countlog.Fatal("event!sut.gc_states_in_background.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
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

// gcGlobalSocks gc socket
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

// gcGlobalRealThreads gc thread
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

/* global socket manager */

// RemoveGlobalSock rm socket， case: Close
func RemoveGlobalSock(socketFD SocketFD) {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	sock := globalSocks[socketFD]
	if sock != nil {
		delete(globalSocks, socketFD)
	}
}

// AddGlobalSock add socket， case: Accept、Connect
func AddGlobalSock(socketFD SocketFD, remoteAddr net.TCPAddr, isServer bool) {
	sock := &socket{
		socketFD: socketFD,
		isServer: isServer,
		addr:     remoteAddr,
	}
	setGlobalSock(socketFD, sock)
}

// setGlobalSock setGlobalSock
func setGlobalSock(socketFD SocketFD, sock *socket) {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	sock.lastAccessedAt = time.Now()
	globalSocks[socketFD] = sock
}

// getGlobalSock getGlobalSock
func getGlobalSock(socketFD SocketFD) *socket {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	sock := globalSocks[socketFD]
	if sock != nil {
		sock.lastAccessedAt = time.Now()
	}

	return sock
}

// exportSocks export all sockets
func exportSocks() map[string]interface{} {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()

	state := map[string]interface{}{}
	for socketFD, sock := range globalSocks {
		state[strconv.Itoa(int(socketFD))] = sock
	}

	return state
}

/* global thread manager */

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

	// if no CallFromInbound.Request in session，return first
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

// getThread get thread Info，can new incase not found
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

// exportThreads export all threads
func exportThreads() map[string]interface{} {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()

	state := map[string]interface{}{}
	for threadID, thread := range globalThreads {
		state[strconv.Itoa(int(threadID))] = thread
	}

	return state
}
