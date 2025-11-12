package nodesync

import (
	"testing"
)

func TestMemberToNodeReadsCapacityTags(t *testing.T) {
	m := MemberInfo{ID: "n1", Addr: "127.0.0.1:9090", Role: "node", Status: "Alive", Tags: map[string]string{"cpu": "16000", "memory": "65536", "disk": "2048"}}
	n := MemberToNode(m)
	if n.Capacity.CPU != 16000 || n.Capacity.Memory != 65536 || n.Capacity.Disk != 2048 {
		t.Fatalf("unexpected capacity: %+v", n.Capacity)
	}
}

func TestMemberToNodeDefaultsWhenNoTags(t *testing.T) {
	m := MemberInfo{ID: "n1", Addr: "127.0.0.1:9090", Role: "node", Status: "Alive"}
	n := MemberToNode(m)
	if n.Capacity.CPU == 0 || n.Capacity.Memory == 0 || n.Capacity.Disk == 0 {
		t.Fatalf("unexpected defaults: %+v", n.Capacity)
	}
}
