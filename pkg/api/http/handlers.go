package httphandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"clustering/pkg/api"
	"clustering/pkg/store"
)

type fsmReader interface{ GetStateCopy() api.ClusterState }
type applier interface {
	Apply(context.Context, store.Command) error
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

// Config
func ConfigGet(fsm fsmReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, fsm.GetStateCopy().Config) }
}
func ConfigPost(st applier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cfg api.ClusterConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if cfg.DesiredVoters < 1 {
			http.Error(w, "desiredVoters must be >= 1", 400)
			return
		}
		if cfg.DesiredNonVoters < 0 {
			http.Error(w, "desiredNonVoters must be >= 0", 400)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("SetConfig", cfg)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

// Networks
func NetworksGet(fsm fsmReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, fsm.GetStateCopy().Networks) }
}
func NetworksPost(st applier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var nw api.Network
		if err := json.NewDecoder(r.Body).Decode(&nw); err != nil || nw.ID == "" || nw.CIDR == "" {
			http.Error(w, "id and cidr required", 400)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("UpsertNetwork", nw)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

// Storage Pools
func StoragePoolsGet(fsm fsmReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, fsm.GetStateCopy().StoragePools) }
}
func StoragePoolsPost(st applier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sp api.StoragePool
		if err := json.NewDecoder(r.Body).Decode(&sp); err != nil || sp.ID == "" {
			http.Error(w, "id required", 400)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("UpsertStoragePool", sp)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

// Volumes
func VolumesGet(fsm fsmReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, fsm.GetStateCopy().Volumes) }
}
func VolumesPost(st applier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var vol api.Volume
		if err := json.NewDecoder(r.Body).Decode(&vol); err != nil || vol.ID == "" {
			http.Error(w, "id required", 400)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("UpsertVolume", vol)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	}
}

