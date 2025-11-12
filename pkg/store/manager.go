package store

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/raft"

	"clustering/pkg/api"
	"clustering/pkg/metrics"
)

type Manager struct {
	r *raft.Raft
	// audit ring buffer for state-changing commands
	audit *AuditRing
	fsm   *FSM
}

func NewManager(r *raft.Raft) *Manager { 
	return &Manager{r: r, audit: NewAuditRing(256), fsm: NewFSM()} 
}

func (m *Manager) Apply(ctx context.Context, cmd Command) error {
	// Allow nil raft for tests; no-op apply
	if m.r == nil {
		return nil
	}
	data, _ := json.Marshal(cmd)
	f := m.r.Apply(data, 0)
	err := f.Error()
	if err != nil {
		metrics.IncCounter("raft_apply_errors_total")
		m.audit.Add(AuditEvent{Type: cmd.Type, Info: "error: " + err.Error()})
		return err
	}
	metrics.IncCounter("raft_applies_total")
	m.audit.Add(AuditEvent{Type: cmd.Type, Info: "ok"})
	return nil
}

// GetStateCopy returns a deep copy of the current state for safe reads.
func (m *Manager) GetStateCopy() api.ClusterState {
	if m.fsm == nil {
		return api.ClusterState{}
	}
	return m.fsm.GetStateCopy()
}

// SetFSM sets the FSM reference for state access
func (m *Manager) SetFSM(fsm *FSM) {
	m.fsm = fsm
}

// Audit returns recent audit events from the ring buffer.
func (m *Manager) Audit() []AuditEvent {
	if m.audit == nil {
		return nil
	}
	return m.audit.List()
}
