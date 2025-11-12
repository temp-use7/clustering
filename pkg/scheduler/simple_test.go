package scheduler

import (
	"clustering/pkg/api"
	"testing"
)

func TestChooseNode(t *testing.T) {
	st := api.ClusterState{Nodes: map[string]api.Node{
		"n1": {ID: "n1", Status: "Alive", Capacity: api.Resources{CPU: 2000, Memory: 2048}},
		"n2": {ID: "n2", Status: "Alive", Capacity: api.Resources{CPU: 1000, Memory: 1024}, Allocated: api.Resources{CPU: 500}},
	}}
	vm := api.VM{ID: "vm1", Resources: api.Resources{CPU: 300, Memory: 200}}
	id, ok := ChooseNode(st, vm)
	if !ok {
		t.Fatal("expected a placement")
	}
	if id != "n1" {
		t.Fatalf("expected n1 got %s", id)
	}
}

func TestChooseNodeMostFree(t *testing.T) {
	st := api.ClusterState{Nodes: map[string]api.Node{
		"n1": {ID: "n1", Status: "Alive", Capacity: api.Resources{CPU: 4000, Memory: 4096}, Allocated: api.Resources{CPU: 3000}},
		"n2": {ID: "n2", Status: "Alive", Capacity: api.Resources{CPU: 4000, Memory: 4096}, Allocated: api.Resources{CPU: 1000}},
	}}
	vm := api.VM{ID: "vm1", Resources: api.Resources{CPU: 500, Memory: 200}, Policy: api.VMSchedulingPolicy{Spread: false}}
	id, ok := ChooseNode(st, vm)
	if !ok || id != "n2" {
		t.Fatalf("expected n2 got %s ok=%v", id, ok)
	}
}

func TestChooseNodeWithAffinity(t *testing.T) {
	st := api.ClusterState{Nodes: map[string]api.Node{
		"n1": {ID: "n1", Status: "Alive", Capacity: api.Resources{CPU: 2000, Memory: 2048}, Labels: map[string]string{"zone": "a"}},
		"n2": {ID: "n2", Status: "Alive", Capacity: api.Resources{CPU: 2000, Memory: 2048}, Labels: map[string]string{"zone": "b"}},
	}}
	vm := api.VM{ID: "vm1", Resources: api.Resources{CPU: 100, Memory: 100}, Policy: api.VMSchedulingPolicy{Affinity: map[string]string{"zone": "b"}}}
	id, ok := ChooseNode(st, vm)
	if !ok || id != "n2" {
		t.Fatalf("expected n2 got %s ok=%v", id, ok)
	}
}
