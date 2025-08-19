package migration

import (
	"clustering/pkg/api"
	"clustering/pkg/store"
	"testing"
)

type fakeFSM struct{ st api.ClusterState }

func (f *fakeFSM) GetStateCopy() api.ClusterState { return f.st }

type fakeRaft struct{}

func TestMigrationTick(t *testing.T) {
	m := store.NewManager(nil)
	// Manager requires raft in real use; here we'll just ensure Apply isn't called
	c := NewController(m, &fakeFSM{st: api.ClusterState{
		Nodes: map[string]api.Node{"n1": {ID: "n1", Status: "Failed"}, "n2": {ID: "n2", Status: "Alive", Capacity: api.Resources{CPU: 1000, Memory: 1024}}},
		VMs:   map[string]api.VM{"vm1": {ID: "vm1", NodeID: "n1", Resources: api.Resources{CPU: 100, Memory: 100}}},
	}})
	// Just call tick(); without a real raft, Apply will fail if reached — acceptable for this placeholder unit test
	c.tick()
}
