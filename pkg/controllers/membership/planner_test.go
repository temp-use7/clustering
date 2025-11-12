package membership

import "testing"

func TestPlanAddAndPromote(t *testing.T) {
	existing := []ExistingServer{{ID: "n1", Suffrage: "voter"}, {ID: "n2", Suffrage: "nonvoter"}}
	alive := map[string]string{"n1": "addr1", "n2": "addr2", "n3": "addr3"}
	add, promote, demote := Plan(existing, alive, 2)
	if len(add) != 1 || add[0] != "n3" {
		t.Fatalf("want add n3 got %v", add)
	}
	if len(promote) != 1 || promote[0] != "n2" {
		t.Fatalf("want promote n2 got %v", promote)
	}
	if len(demote) != 0 {
		t.Fatalf("want no demotions got %v", demote)
	}
}

func TestPlanDemote(t *testing.T) {
	existing := []ExistingServer{{ID: "n1", Suffrage: "voter"}, {ID: "n2", Suffrage: "voter"}, {ID: "n3", Suffrage: "voter"}}
	alive := map[string]string{"n1": "a1", "n2": "a2", "n3": "a3"}
	add, promote, demote := Plan(existing, alive, 2)
	if len(add) != 0 || len(promote) != 0 || len(demote) != 1 || demote[0] != "n3" {
		t.Fatalf("unexpected plan: add=%v promote=%v demote=%v", add, promote, demote)
	}
}

func TestPlanTreatsDesiredVotersMinimumOne(t *testing.T) {
	existing := []ExistingServer{{ID: "n1", Suffrage: "voter"}}
	alive := map[string]string{"n1": "a1"}
	add, promote, demote := Plan(existing, alive, 0)
	if len(add) != 0 || len(promote) != 0 || len(demote) != 0 {
		t.Fatalf("unexpected plan for desired=0: add=%v promote=%v demote=%v", add, promote, demote)
	}
}

func TestPlanAddsNonvoterForNewAlive(t *testing.T) {
	existing := []ExistingServer{{ID: "n1", Suffrage: "voter"}}
	alive := map[string]string{"n1": "a1", "nX": "aX"}
	add, promote, demote := Plan(existing, alive, 1)
	if len(add) != 1 || add[0] != "nX" || len(promote) != 0 || len(demote) != 0 {
		t.Fatalf("unexpected plan: add=%v promote=%v demote=%v", add, promote, demote)
	}
}
