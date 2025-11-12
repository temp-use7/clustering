package httphandlers

import (
	"bytes"
	"clustering/pkg/api"
	"clustering/pkg/store"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeFSM struct{ st api.ClusterState }

func (f *fakeFSM) GetStateCopy() api.ClusterState { return f.st }

type fakeApplier struct{ cmds []store.Command }

func (f *fakeApplier) Apply(_ context.Context, c store.Command) error {
	f.cmds = append(f.cmds, c)
	return nil
}

func TestConfigHandlers(t *testing.T) {
	fsm := &fakeFSM{st: api.ClusterState{Config: api.ClusterConfig{DesiredVoters: 5, DesiredNonVoters: 2}}}
	// GET
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	ConfigGet(fsm)(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status: %d", rr.Code)
	}
	// POST
	ap := &fakeApplier{}
	rr2 := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"desiredVoters":3,"desiredNonVoters":1}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/config", body)
	ConfigPost(ap)(rr2, req2)
	if rr2.Code != 204 {
		t.Fatalf("post status: %d", rr2.Code)
	}
	if len(ap.cmds) != 1 || ap.cmds[0].Type != "SetConfig" {
		t.Fatalf("unexpected cmds: %+v", ap.cmds)
	}
}

func TestNetworksHandlers(t *testing.T) {
	fsm := &fakeFSM{st: api.ClusterState{Networks: map[string]api.Network{"n1": {ID: "n1", CIDR: "10.0.0.0/24"}}}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	NetworksGet(fsm)(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status: %d", rr.Code)
	}
	ap := &fakeApplier{}
	rr2 := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"id":"n2","cidr":"10.0.1.0/24"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/networks", body)
	NetworksPost(ap)(rr2, req2)
	if rr2.Code != 204 {
		t.Fatalf("post status: %d", rr2.Code)
	}
}

func TestStorageHandlers(t *testing.T) {
	fsm := &fakeFSM{st: api.ClusterState{StoragePools: map[string]api.StoragePool{"p1": {ID: "p1", Type: "local", Size: 10}}, Volumes: map[string]api.Volume{"v1": {ID: "v1", Size: 1}}}}
	// pools
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/storagepools", nil)
	StoragePoolsGet(fsm)(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status: %d", rr.Code)
	}
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/storagepools", bytes.NewBufferString(`{"id":"p2","type":"nfs","size":100}`))
	StoragePoolsPost(&fakeApplier{})(rr2, req2)
	if rr2.Code != 204 {
		t.Fatalf("post status: %d", rr2.Code)
	}
	// vols
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/volumes", nil)
	VolumesGet(fsm)(rr3, req3)
	if rr3.Code != 200 {
		t.Fatalf("status: %d", rr3.Code)
	}
	rr4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/api/volumes", bytes.NewBufferString(`{"id":"v2","size":10}`))
	VolumesPost(&fakeApplier{})(rr4, req4)
	if rr4.Code != 204 {
		t.Fatalf("post status: %d", rr4.Code)
	}
}
