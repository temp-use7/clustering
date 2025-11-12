package integration

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/raft"

	"clustering/pkg/api"
	"clustering/pkg/store"
)

func applyCommand(t *testing.T, fsm *store.FSM, cmd store.Command) {
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}
	log := &raft.Log{Data: data, Type: raft.LogCommand}
	if result := fsm.Apply(log); result != nil {
		if err, ok := result.(error); ok {
			t.Fatalf("apply command failed: %v", err)
		}
	}
}

func TestFSMNodeAndVMLifecycle(t *testing.T) {
	fsm := store.NewFSM()

	// Add a node
	node := api.Node{ID: "node-1", Address: "10.0.0.1", Role: "node", Status: "Alive", Capacity: api.Resources{CPU: 4000, Memory: 8192, Disk: 200}}
	applyCommand(t, fsm, store.NewCommand("UpsertNode", node))

	// Schedule a VM on the node
	vm := api.VM{ID: "vm-1", Name: "web", NodeID: node.ID, Phase: "Running", Resources: api.Resources{CPU: 1000, Memory: 2048, Disk: 50}}
	applyCommand(t, fsm, store.NewCommand("UpsertVM", vm))

	state := fsm.GetStateCopy()
	if _, ok := state.Nodes[node.ID]; !ok {
		t.Fatalf("expected node %s to exist", node.ID)
	}
	if got := state.Nodes[node.ID].Allocated.CPU; got != vm.Resources.CPU {
		t.Fatalf("expected allocated CPU %d, got %d", vm.Resources.CPU, got)
	}
	if _, ok := state.VMs[vm.ID]; !ok {
		t.Fatalf("expected VM %s to exist", vm.ID)
	}

	// Delete the VM and ensure allocation updates
	applyCommand(t, fsm, store.NewCommand("DeleteVM", vm.ID))
	state = fsm.GetStateCopy()
	if _, ok := state.VMs[vm.ID]; ok {
		t.Fatalf("expected VM %s to be removed", vm.ID)
	}
	if got := state.Nodes[node.ID].Allocated.CPU; got != 0 {
		t.Fatalf("expected allocated CPU to be 0 after delete, got %d", got)
	}
}

func TestFSMConfigVersioning(t *testing.T) {
	fsm := store.NewFSM()

	initial := fsm.GetStateCopy().ConfigVersion

	cfg := api.ClusterConfig{DesiredVoters: 3, DesiredNonVoters: 1}
	applyCommand(t, fsm, store.NewCommand("SetConfig", cfg))

	state := fsm.GetStateCopy()
	if state.ConfigVersion != initial+1 {
		t.Fatalf("expected config version %d, got %d", initial+1, state.ConfigVersion)
	}
	if state.Config.DesiredVoters != 3 {
		t.Fatalf("expected desired voters 3, got %d", state.Config.DesiredVoters)
	}
	if len(state.ConfigHistory) == 0 {
		t.Fatal("expected config history to contain previous version")
	}

	applyCommand(t, fsm, store.NewCommand("RollbackConfig", nil))
	state = fsm.GetStateCopy()
	if state.ConfigVersion != initial {
		t.Fatalf("expected config version to rollback to %d, got %d", initial, state.ConfigVersion)
	}
}

func TestFSMNetworkStorageTemplateLifecycle(t *testing.T) {
	fsm := store.NewFSM()

	nw := api.Network{ID: "net-1", CIDR: "10.10.0.0/24"}
	pool := api.StoragePool{ID: "pool-1", Type: "ssd", Size: 1000}
	vol := api.Volume{ID: "vol-1", Size: 50, Node: "node-1"}
	tpl := api.VMTemplate{ID: "tpl-1", Name: "ubuntu", BaseImage: "ubuntu-22.04", Resources: api.Resources{CPU: 500, Memory: 1024, Disk: 20}}

	applyCommand(t, fsm, store.NewCommand("UpsertNetwork", nw))
	applyCommand(t, fsm, store.NewCommand("UpsertStoragePool", pool))
	applyCommand(t, fsm, store.NewCommand("UpsertVolume", vol))
	applyCommand(t, fsm, store.NewCommand("UpsertTemplate", tpl))

	state := fsm.GetStateCopy()
	if _, ok := state.Networks[nw.ID]; !ok {
		t.Fatalf("expected network %s", nw.ID)
	}
	if _, ok := state.StoragePools[pool.ID]; !ok {
		t.Fatalf("expected storage pool %s", pool.ID)
	}
	if _, ok := state.Volumes[vol.ID]; !ok {
		t.Fatalf("expected volume %s", vol.ID)
	}
	if _, ok := state.Templates[tpl.ID]; !ok {
		t.Fatalf("expected template %s", tpl.ID)
	}

	// Ensure GetStateCopy returns a deep copy
	copyState := fsm.GetStateCopy()
	copyState.Networks[nw.ID] = api.Network{ID: nw.ID, CIDR: "192.168.0.0/24"}

	state = fsm.GetStateCopy()
	if state.Networks[nw.ID].CIDR != nw.CIDR {
		t.Fatalf("expected original network CIDR %s, got %s", nw.CIDR, state.Networks[nw.ID].CIDR)
	}

	// Delete resources
	applyCommand(t, fsm, store.NewCommand("DeleteNetwork", nw.ID))
	applyCommand(t, fsm, store.NewCommand("DeleteStoragePool", pool.ID))
	applyCommand(t, fsm, store.NewCommand("DeleteVolume", vol.ID))
	applyCommand(t, fsm, store.NewCommand("DeleteTemplate", tpl.ID))

	state = fsm.GetStateCopy()
	if len(state.Networks) != 0 || len(state.StoragePools) != 0 || len(state.Volumes) != 0 || len(state.Templates) != 0 {
		t.Fatal("expected all resources to be deleted")
	}
}
