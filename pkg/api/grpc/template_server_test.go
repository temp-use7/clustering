package grpcapi

import (
	"clustering/pkg/api"
	"clustering/pkg/store"
	"context"
	"testing"

	templatepb "clustering/api/proto/template"
)

type fakeFSM struct{ st api.ClusterState }

func (f *fakeFSM) GetStateCopy() api.ClusterState { return f.st }

func TestTemplateServerListAndInstantiate(t *testing.T) {
	// Manager with nil raft is allowed (no-op apply)
	m := store.NewManager(nil)
	fsm := &fakeFSM{st: api.ClusterState{Templates: map[string]api.VMTemplate{
		"tpl1": {ID: "tpl1", Name: "small", Resources: api.Resources{CPU: 200, Memory: 256, Disk: 5}},
	}, Nodes: map[string]api.Node{"n1": {ID: "n1", Status: "Alive", Capacity: api.Resources{CPU: 1000, Memory: 1024}}}}}
	ts := NewTemplateServer(m, fsm)
	// list
	resp, err := ts.ListTemplates(context.Background(), &templatepb.Empty{})
	if err != nil || resp == nil {
		t.Fatalf("list err=%v resp=%v", err, resp)
	}
	// instantiate (will no-op apply but should return success)
	if _, err := ts.Instantiate(context.Background(), &templatepb.InstantiateRequest{TemplateId: "tpl1", NewId: "vm-x"}); err != nil {
		t.Fatalf("instantiate: %v", err)
	}
}
