package store

import "testing"

func TestAuditRing(t *testing.T) {
	a := NewAuditRing(2)
	a.Add(AuditEvent{Type: "A", Info: "ok"})
	a.Add(AuditEvent{Type: "B", Info: "ok"})
	if len(a.List()) != 2 {
		t.Fatalf("want 2 events")
	}
	a.Add(AuditEvent{Type: "C", Info: "ok"})
	evs := a.List()
	if len(evs) != 2 {
		t.Fatalf("want 2 events after wrap")
	}
	if evs[0].Type == "A" {
		t.Fatalf("oldest should be dropped")
	}
}

