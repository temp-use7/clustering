package health

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestHealthControllerInterval(t *testing.T) {
	var hits int32
	ping := func() error {
		atomic.AddInt32(&hits, 1)
		return nil
	}
	c := NewController(ping).WithInterval(10 * time.Millisecond)
	stop := make(chan struct{})
	go c.Run(stop)
	defer close(stop)
	time.Sleep(30 * time.Millisecond)
	if atomic.LoadInt32(&hits) < 2 {
		t.Fatalf("expected at least 2 pings, got %d", hits)
	}
}

