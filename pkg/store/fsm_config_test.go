package store

import (
	"bytes"
	"io"
	"testing"

	"clustering/pkg/api"
)

func TestFSMSetConfigAndSnapshot(t *testing.T) {
	f := NewFSM()
	// set non-default config
	cfg := api.ClusterConfig{DesiredVoters: 3, DesiredNonVoters: 1}
	if r := f.Apply(mkLog(NewCommand("SetConfig", cfg))); r != nil {
		t.Fatalf("apply config: %v", r)
	}
	st := f.GetStateCopy()
	if st.Config.DesiredVoters != 3 || st.Config.DesiredNonVoters != 1 {
		t.Fatalf("unexpected config in state: %+v", st.Config)
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
	st2 := f2.GetStateCopy()
	if st2.Config.DesiredVoters != 3 || st2.Config.DesiredNonVoters != 1 {
		t.Fatalf("unexpected config after restore: %+v", st2.Config)
	}
}

func TestFSMConfigVersionAndRollback(t *testing.T) {
	f := NewFSM()
	v0 := f.GetStateCopy().ConfigVersion
	if v0 < 1 {
		t.Fatalf("expected initial version >=1 got %d", v0)
	}
	cfg1 := api.ClusterConfig{DesiredVoters: 4, DesiredNonVoters: 2}
	_ = f.Apply(mkLog(NewCommand("SetConfig", cfg1)))
	cfg2 := api.ClusterConfig{DesiredVoters: 3, DesiredNonVoters: 1}
	_ = f.Apply(mkLog(NewCommand("SetConfig", cfg2)))
	st := f.GetStateCopy()
	if st.Config.DesiredVoters != 3 || st.ConfigVersion != v0+2 || len(st.ConfigHistory) != 2 {
		t.Fatalf("unexpected state: %+v", st)
	}
	_ = f.Apply(mkLog(NewCommand("RollbackConfig", nil)))
	st2 := f.GetStateCopy()
	if st2.Config.DesiredVoters != 4 || st2.ConfigVersion != v0+1 || len(st2.ConfigHistory) != 1 {
		t.Fatalf("unexpected after rb: %+v", st2)
	}
}
