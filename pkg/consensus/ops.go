package consensus

import (
	"context"
	"time"

	"github.com/hashicorp/raft"
)

// GetConfiguration returns the current Raft configuration.
func GetConfiguration(r *raft.Raft) (raft.Configuration, error) {
	f := r.GetConfiguration()
	if err := f.Error(); err != nil {
		return raft.Configuration{}, err
	}
	return f.Configuration(), nil
}

// AddServer adds a server to the Raft cluster.
func AddServer(r *raft.Raft, id raft.ServerID, address raft.ServerAddress, suffrage raft.ServerSuffrage) error {
	fut := r.AddVoter(id, address, 0, 0)
	if suffrage == raft.Nonvoter {
		fut = r.AddNonvoter(id, address, 0, 0)
	}
	return fut.Error()
}

// RemoveServer removes a server from the Raft cluster.
func RemoveServer(r *raft.Raft, id raft.ServerID) error {
	fut := r.RemoveServer(id, 0, 0)
	return fut.Error()
}

// WaitForLeader blocks until a leader is elected or the timeout elapses.
func WaitForLeader(r *raft.Raft, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if r.Leader() != "" {
				return true
			}
		}
	}
}

