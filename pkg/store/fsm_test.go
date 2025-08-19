package store

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"clustering/pkg/api"

	"github.com/hashicorp/raft"
)

func TestFSMApplyAndSnapshotRestore(t *testing.T) {
	f := NewFSM()
	na := api.Node{ID: "n1", Capacity: api.Resources{CPU: 1000, Memory: 1024, Disk: 10}, Status: "Alive"}
	if r := f.Apply(mkLog(NewCommand("UpsertNode", na))); r != nil {
		t.Fatalf("apply node: %v", r)
	}

	vm := api.VM{ID: "vm1", Resources: api.Resources{CPU: 200, Memory: 256, Disk: 1}, NodeID: "n1", Phase: "Running"}
	if r := f.Apply(mkLog(NewCommand("UpsertVM", vm))); r != nil {
		t.Fatalf("apply vm: %v", r)
	}

	snap, _ := f.Snapshot()
	var buf bytes.Buffer
	if err := snap.Persist(&sink{&buf}); err != nil {
		t.Fatalf("persist: %v", err)
	}

	f2 := NewFSM()
	if err := f2.Restore(io.NopCloser(bytes.NewReader(buf.Bytes()))); err != nil {
		t.Fatalf("restore: %v", err)
	}
	st := f2.GetStateCopy()
	if len(st.Nodes) != 1 || len(st.VMs) != 1 {
		t.Fatalf("unexpected state: %+v", st)
	}
}

func TestFSMNetworkAndStorageCRUD(t *testing.T) {
	f := NewFSM()
	nw := api.Network{ID: "net1", CIDR: "10.0.0.0/24"}
	if r := f.Apply(mkLog(NewCommand("UpsertNetwork", nw))); r != nil {
		t.Fatalf("upsert network: %v", r)
	}
	sp := api.StoragePool{ID: "pool1", Type: "local", Size: 100}
	if r := f.Apply(mkLog(NewCommand("UpsertStoragePool", sp))); r != nil {
		t.Fatalf("upsert pool: %v", r)
	}
	st := f.GetStateCopy()
	if len(st.Networks) != 1 || len(st.StoragePools) != 1 {
		t.Fatalf("unexpected counts: %+v", st)
	}
	if r := f.Apply(mkLog(NewCommand("DeleteNetwork", "net1"))); r != nil {
		t.Fatalf("delete network: %v", r)
	}
	if r := f.Apply(mkLog(NewCommand("DeleteStoragePool", "pool1"))); r != nil {
		t.Fatalf("delete pool: %v", r)
	}
	st2 := f.GetStateCopy()
	if len(st2.Networks) != 0 || len(st2.StoragePools) != 0 {
		t.Fatalf("unexpected after delete: %+v", st2)
	}
}

// helpers to emulate raft.Log and SnapshotSink
func mkLog(c Command) *raft.Log { b, _ := json.Marshal(c); return &raft.Log{Data: b} }

type sink struct{ b *bytes.Buffer }

func (s *sink) ID() string                  { return "test" }
func (s *sink) Cancel() error               { return nil }
func (s *sink) Close() error                { return nil }
func (s *sink) Write(p []byte) (int, error) { return s.b.Write(p) }

// (no fake raft.Log type needed)
