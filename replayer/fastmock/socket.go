package fastmock

import (
	"sync"
	"time"

	"github.com/didi/sharingan/replayer/internal"
)

/* global socket manager */
var globalSockets ISockets

func init() {
	globalSockets = NewSockets()
}

// ISockets interface
type ISockets interface {
	Set(fd int, remoteAddr string, accessTime time.Time)
	Get(fd int) *Socket
	Access(fd int)
	Remove(fd int)
}

// Socket Socket
type Socket struct {
	lastAccessedAt time.Time
	remoteAddr     string
}

// NewSockets New
func NewSockets() ISockets {
	return &Sockets{
		mutex: &sync.RWMutex{},
		m:     make(map[int]*Socket),
	}
}

// Sockets Sockets
type Sockets struct {
	mutex *sync.RWMutex
	m     map[int]*Socket
}

// Set Set
func (s *Sockets) Set(fd int, remoteAddr string, accessTime time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.m[fd] = &Socket{
		remoteAddr:     remoteAddr,
		lastAccessedAt: accessTime,
	}
}

// Get Get
func (s *Sockets) Get(fd int) *Socket {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if socket, ok := s.m[fd]; ok {
		return socket
	}
	return nil
}

// Access Access
func (s *Sockets) Access(fd int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if socket, ok := s.m[fd]; ok {
		socket.lastAccessedAt = internal.TimeNow()
	}
}

// Remove Remove
func (s *Sockets) Remove(fd int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.m[fd]; ok {
		delete(s.m, fd)
	}
}
