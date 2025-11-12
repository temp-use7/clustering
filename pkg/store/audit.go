package store

import (
	"sync"
)

type AuditEvent struct {
	Type string
	Info string
}

type AuditRing struct {
	mu   sync.Mutex
	buf  []AuditEvent
	size int
	head int
	full bool
}

func NewAuditRing(size int) *AuditRing {
	if size <= 0 {
		size = 128
	}
	return &AuditRing{buf: make([]AuditEvent, size), size: size}
}

func (a *AuditRing) Add(ev AuditEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buf[a.head] = ev
	a.head = (a.head + 1) % a.size
	if a.head == 0 {
		a.full = true
	}
}

func (a *AuditRing) List() []AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.full {
		out := make([]AuditEvent, a.head)
		copy(out, a.buf[:a.head])
		return out
	}
	out := make([]AuditEvent, a.size)
	copy(out, a.buf[a.head:])
	copy(out[a.size-a.head:], a.buf[:a.head])
	return out
}

