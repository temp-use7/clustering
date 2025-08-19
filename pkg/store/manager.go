package store

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/raft"
)

type Manager struct {
	r *raft.Raft
}

func NewManager(r *raft.Raft) *Manager { return &Manager{r: r} }

func (m *Manager) Apply(ctx context.Context, cmd Command) error {
	// Allow nil raft for tests; no-op apply
	if m.r == nil {
		return nil
	}
	data, _ := json.Marshal(cmd)
	f := m.r.Apply(data, 0)
	return f.Error()
}
