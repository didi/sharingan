package fastmock

import (
	"sync"
	"time"

	"github.com/didi/sharingan/replayer/internal"
)

/* global goroutine manager */
var ReplayerGlobalThreads IThreads

func init() {
	ReplayerGlobalThreads = NewThreads()

	// gc
	go func() {
		for {
			time.Sleep(time.Second * 10)
			ReplayerGlobalThreads.Recycle()
		}
	}()
}

// IThreads interface
type IThreads interface {
	Set(threadID int64, traceID string, replayTime int64)
	Get(threadID int64) *Thread
	Access(threadID int64)
	Recycle()
}

// Thread Thread
type Thread struct {
	lastAccessedAt time.Time
	traceID        string
	replayTime     int64
}

// NewThreads new Threads
func NewThreads() IThreads {
	return &Threads{
		mutex: &sync.RWMutex{},
		m:     make(map[int64]*Thread),
	}
}

// Threads Threads
type Threads struct {
	mutex *sync.RWMutex
	m     map[int64]*Thread
}

// Set Set
func (t *Threads) Set(threadID int64, traceID string, replayTime int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.m[threadID] = &Thread{
		lastAccessedAt: internal.TimeNow(),
		traceID:        traceID,
		replayTime:     replayTime,
	}
}

// Get Get
func (t *Threads) Get(threadID int64) *Thread {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if thread := t.m[threadID]; thread != nil {
		return thread
	}

	return nil
}

// Access Access
func (t *Threads) Access(threadID int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if thread := t.m[threadID]; thread != nil {
		thread.lastAccessedAt = internal.TimeNow()
	}
}

// Recycle GC
func (t *Threads) Recycle() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	newGlobalThreads := NewThreads()

	now := internal.TimeNow()
	for threadID, thread := range t.m {
		if now.Sub(thread.lastAccessedAt) < time.Second*5 {
			newGlobalThreads.Set(threadID, thread.traceID, thread.replayTime)
		}
	}

	ReplayerGlobalThreads = newGlobalThreads
}
