package store

import (
	"encoding/json"
	"io"
	"sync"

	"clustering/pkg/api"

	"github.com/hashicorp/raft"
)

type FSM struct {
	mu    sync.RWMutex
	state api.ClusterState
}

func NewFSM() *FSM {
	return &FSM{state: api.ClusterState{Nodes: map[string]api.Node{}, VMs: map[string]api.VM{}, Networks: map[string]api.Network{}, StoragePools: map[string]api.StoragePool{}, Config: api.ClusterConfig{DesiredVoters: 5, DesiredNonVoters: 2}, ConfigVersion: 1, ConfigHistory: []api.ClusterConfig{}}}
}

func (f *FSM) Apply(l *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		return err
	}
	return f.applyCommand(cmd)
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	// Deep copy for snapshot
	buf, _ := json.Marshal(f.state)
	var copy api.ClusterState
	_ = json.Unmarshal(buf, &copy)
	return &snapshot{state: copy}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	dec := json.NewDecoder(rc)
	var s api.ClusterState
	if err := dec.Decode(&s); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = s
	return nil
}

func (f *FSM) applyCommand(c Command) interface{} {
	switch c.Type {
	case "UpsertNode":
		var n api.Node
		_ = json.Unmarshal(c.Payload, &n)
		f.state.Nodes[n.ID] = n
		// recompute allocations: naive aggregate VMs on node
		n.Allocated = api.Resources{}
		for _, v := range f.state.VMs {
			if v.NodeID == n.ID {
				n.Allocated.CPU += v.Resources.CPU
				n.Allocated.Memory += v.Resources.Memory
				n.Allocated.Disk += v.Resources.Disk
			}
		}
		f.state.Nodes[n.ID] = n
	case "DeleteNode":
		var id string
		_ = json.Unmarshal(c.Payload, &id)
		delete(f.state.Nodes, id)
	case "UpsertVM":
		var v api.VM
		_ = json.Unmarshal(c.Payload, &v)
		f.state.VMs[v.ID] = v
		// adjust allocation for node
		if n, ok := f.state.Nodes[v.NodeID]; ok {
			n.Allocated.CPU += v.Resources.CPU
			n.Allocated.Memory += v.Resources.Memory
			n.Allocated.Disk += v.Resources.Disk
			f.state.Nodes[v.NodeID] = n
		}
	case "DeleteVM":
		var id string
		_ = json.Unmarshal(c.Payload, &id)
		v, ok := f.state.VMs[id]
		delete(f.state.VMs, id)
		if ok {
			if n, ok2 := f.state.Nodes[v.NodeID]; ok2 {
				n.Allocated.CPU -= v.Resources.CPU
				n.Allocated.Memory -= v.Resources.Memory
				n.Allocated.Disk -= v.Resources.Disk
				if n.Allocated.CPU < 0 {
					n.Allocated.CPU = 0
				}
				if n.Allocated.Memory < 0 {
					n.Allocated.Memory = 0
				}
				if n.Allocated.Disk < 0 {
					n.Allocated.Disk = 0
				}
				f.state.Nodes[v.NodeID] = n
			}
		}
	case "SetConfig":
		var cfg api.ClusterConfig
		_ = json.Unmarshal(c.Payload, &cfg)
		if cfg.DesiredVoters <= 0 {
			cfg.DesiredVoters = 1
		}
		if cfg.DesiredNonVoters < 0 {
			cfg.DesiredNonVoters = 0
		}
		// push current to history and set new version
		f.state.ConfigHistory = append(f.state.ConfigHistory, f.state.Config)
		f.state.Config = cfg
		f.state.ConfigVersion++
	case "RollbackConfig":
		if len(f.state.ConfigHistory) > 0 {
			last := f.state.ConfigHistory[len(f.state.ConfigHistory)-1]
			f.state.ConfigHistory = f.state.ConfigHistory[:len(f.state.ConfigHistory)-1]
			f.state.Config = last
			if f.state.ConfigVersion > 1 {
				f.state.ConfigVersion--
			}
		}
	case "UpsertNetwork":
		var nw api.Network
		_ = json.Unmarshal(c.Payload, &nw)
		f.state.Networks[nw.ID] = nw
	case "DeleteNetwork":
		var id string
		_ = json.Unmarshal(c.Payload, &id)
		delete(f.state.Networks, id)
	case "UpsertStoragePool":
		var sp api.StoragePool
		_ = json.Unmarshal(c.Payload, &sp)
		f.state.StoragePools[sp.ID] = sp
	case "DeleteStoragePool":
		var id string
		_ = json.Unmarshal(c.Payload, &id)
		delete(f.state.StoragePools, id)
	}
	return nil
}

type snapshot struct {
	state api.ClusterState
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	enc := json.NewEncoder(sink)
	if err := enc.Encode(s.state); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}
func (s *snapshot) Release() {}

// GetStateCopy returns a deep copy of the current state for safe reads.
func (f *FSM) GetStateCopy() api.ClusterState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	buf, _ := json.Marshal(f.state)
	var copy api.ClusterState
	_ = json.Unmarshal(buf, &copy)
	return copy
}
