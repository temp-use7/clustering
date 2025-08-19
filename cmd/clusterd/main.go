package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	clusterpb "clustering/api/proto/cluster"
	nodepb "clustering/api/proto/node"
	vmpb "clustering/api/proto/vm"
	"clustering/pkg/api"
	grpcapi "clustering/pkg/api/grpc"
	"clustering/pkg/consensus"
	fsctrl "clustering/pkg/controllers/failover"
	hcctrl "clustering/pkg/controllers/health"
	mc "clustering/pkg/controllers/membership"
	nsync "clustering/pkg/controllers/nodesync"
	"clustering/pkg/membership"
	"clustering/pkg/scheduler"
	"clustering/pkg/store"
)

func main() {
	var (
		nodeID    string
		dataDir   string
		raftBind  string
		grpcAddr  string
		uiAddr    string
		serfBind  string
		serfJoin  string
		wipeData  bool
		joinToken string
	)

	flag.StringVar(&nodeID, "node-id", "node-1", "unique node ID")
	flag.StringVar(&dataDir, "data-dir", "./data", "data directory for raft state")
	flag.StringVar(&raftBind, "raft-bind", ":7000", "raft bind address host:port")
	var bootstrap bool
	flag.BoolVar(&bootstrap, "bootstrap", false, "bootstrap single-node raft configuration if empty")
	flag.BoolVar(&wipeData, "wipe-data", false, "DANGEROUS: delete data dir on start (dev reset)")
	flag.StringVar(&grpcAddr, "grpc", ":8081", "gRPC listen address")
	flag.StringVar(&uiAddr, "ui", ":8080", "UI/HTTP listen address")
	flag.StringVar(&serfBind, "serf-bind", ":7946", "serf bind address host:port")
	flag.StringVar(&serfJoin, "serf-join", "", "comma-separated serf peers to join")
	flag.StringVar(&joinToken, "join-token", "", "shared join token that node agents must present")
	flag.Parse()

	if wipeData {
		_ = os.RemoveAll(dataDir)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("mkdir data dir: %v", err)
	}

	// Consensus (HashiCorp Raft)
	rft, transport, fsm := consensus.MustStartRaft(nodeID, raftBind, dataDir)
	defer transport.Close()

	// Membership (Serf)
	sconf := membership.Config{NodeID: nodeID, BindAddr: serfBind}
	s, events := membership.MustStartSerf(sconf)
	// Tag this process as control-plane
	if err := s.SetTags(map[string]string{"role": "control-plane", "raft": string(transport.LocalAddr())}); err != nil {
		log.Printf("serf set tags: %v", err)
	}
	if serfJoin != "" {
		peers := splitCSV(serfJoin)
		if _, err := s.Join(peers, true); err != nil {
			log.Printf("serf join error: %v", err)
		}
	}
	go func() {
		for e := range events {
			log.Printf("serf: %s %s", e.Type, e.Node)
		}
	}()

	// Optional single-node bootstrap if requested and no servers configured
	if bootstrap {
		cfgF := rft.GetConfiguration()
		if cfgF.Error() == nil && len(cfgF.Configuration().Servers) == 0 {
			conf := raft.Configuration{Servers: []raft.Server{{ID: raft.ServerID(nodeID), Address: transport.LocalAddr(), Suffrage: raft.Voter}}}
			if err := rft.BootstrapCluster(conf).Error(); err != nil {
				log.Printf("bootstrap cluster: %v", err)
			} else {
				log.Printf("bootstrapped single-node cluster: %s", nodeID)
			}
		}
	}

	// Start membership controller (logs plan only for now)
	stopCh := make(chan struct{})
	controller := mc.NewController(rft, func() []mc.AliveMember {
		var out []mc.AliveMember
		for _, m := range s.Members() {
			if m.Status == serf.StatusAlive {
				if joinToken != "" && m.Tags["token"] != joinToken {
					continue
				}
				out = append(out, mc.AliveMember{ID: m.Name, RaftAddr: m.Tags["raft"]})
			}
		}
		return out
	}).WithDesiredVotersFunc(func() int { return fsm.GetStateCopy().Config.DesiredVoters })
	go controller.Run(stopCh)

	// gRPC server + health
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen %s: %v", grpcAddr, err)
	}
	grpcServer := grpc.NewServer()
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthSrv)
	// State manager for Raft proposals (needed by gRPC VM server)
	st := store.NewManager(rft)
	// Register gRPC services (stubs until codegen is wired in)
	clusterpb.RegisterClusterServiceServer(grpcServer, grpcapi.NewClusterServer(rft))
	nodepb.RegisterNodeServiceServer(grpcServer, grpcapi.NewNodeServer(fsm))
	vmpb.RegisterVMServiceServer(grpcServer, grpcapi.NewVMServer(st, fsm))
	go func() {
		log.Printf("gRPC listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// State manager already initialized above for gRPC services

	// Node sync controller (reflect Serf into Raft state)
	nc := nsync.NewController(func() []nsync.MemberInfo {
		var out []nsync.MemberInfo
		for _, m := range s.Members() {
			role := m.Tags["role"]
			status := m.Status.String()
			if role == "node" && joinToken != "" && m.Tags["token"] != joinToken {
				continue
			}
			addr := fmt.Sprintf("%s:%s", m.Addr, m.Tags["http"]) // prefer agent's http port if tagged
			if m.Tags["http"] == "" {
				addr = fmt.Sprintf("%s:%d", m.Addr, m.Port)
			}
			out = append(out, nsync.MemberInfo{ID: m.Name, Addr: addr, Role: role, Status: status})
		}
		return out
	}, st, func() bool { return rft.State() == raft.Leader })
	go nc.Run(stopCh)

	// Failover/migration controller (stub)
	fc := fsctrl.NewController(st, func() bool { return rft.State() == raft.Leader })
	go fc.Run(stopCh)

	// Health controller: probe known nodes by reading FSM and hitting /healthz
	healthCtl := hcctrl.NewController(func() error {
		stCopy := fsm.GetStateCopy()
		for _, n := range stCopy.Nodes {
			if n.Role != "node" {
				continue
			}
			url := fmt.Sprintf("http://%s/healthz", n.Address)
			resp, err := http.Get(url)
			if err != nil || resp.StatusCode != 200 {
				// mark node as Failed
				_ = st.Apply(context.Background(), store.NewCommand("UpsertNode", map[string]any{
					"id":       n.ID,
					"address":  n.Address,
					"role":     n.Role,
					"status":   "Failed",
					"capacity": n.Capacity,
				}))
				continue
			}
			_ = resp.Body.Close()
			// mark Alive if previously not Alive
			if n.Status != "Alive" {
				_ = st.Apply(context.Background(), store.NewCommand("UpsertNode", map[string]any{
					"id":       n.ID,
					"address":  n.Address,
					"role":     n.Role,
					"status":   "Alive",
					"capacity": n.Capacity,
				}))
			}
		}
		return nil
	})
	go healthCtl.Run(stopCh)

	// Minimal UI HTTP server to visualize raft+serf state and expose state APIs
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		leader := string(rft.Leader())
		state := rft.State().String()
		members := s.Members()
		fmt.Fprintf(w, "<html><head><title>Cluster</title><style>body{font-family:sans-serif;padding:20px} table{border-collapse:collapse} td,th{border:1px solid #ddd;padding:8px}</style></head><body>")
		fmt.Fprintf(w, "<h1>Cluster Dashboard</h1>")
		fmt.Fprintf(w, "<p><b>Node:</b> %s</p>", nodeID)
		fmt.Fprintf(w, "<p><b>Raft State:</b> %s, <b>Leader:</b> %s</p>", state, leader)
		fmt.Fprintf(w, "<h2>Members (%d)</h2>", len(members))
		fmt.Fprintf(w, "<table><tr><th>Name</th><th>Addr</th><th>Status</th><th>Tags</th></tr>")
		for _, m := range members {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s:%d</td><td>%s</td><td>%v</td></tr>", m.Name, m.Addr, m.Port, m.Status.String(), m.Tags)
		}
		fmt.Fprintf(w, "</table>")
		// Nodes section
		fmt.Fprintf(w, "<h2>Nodes</h2>")
		fmt.Fprintf(w, "<table><tr><th>Name</th><th>Addr</th><th>Status</th></tr>")
		for _, m := range members {
			if m.Tags["role"] == "node" {
				fmt.Fprintf(w, "<tr><td>%s</td><td>%s:%d</td><td>%s</td></tr>", m.Name, m.Addr, m.Port, m.Status.String())
			}
		}
		fmt.Fprintf(w, "</table>")
		fmt.Fprintf(w, "<p style=\"margin-top:16px\"><a href=\"/ui/\">Open React UI</a></p>")
		fmt.Fprintf(w, "</body></html>")
	})

	// Serve built React UI (ui/dist) under /ui/ with SPA fallback
	uiDir := "./ui/dist"
	mux.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusPermanentRedirect)
	})
	mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/ui/")
		if p == "" || p == "/" {
			p = "index.html"
		}
		fp := filepath.Join(uiDir, p)
		if info, err := os.Stat(fp); err == nil && !info.IsDir() {
			http.ServeFile(w, r, fp)
			return
		}
		http.ServeFile(w, r, filepath.Join(uiDir, "index.html"))
	})
	mux.HandleFunc("/api/raft/config", func(w http.ResponseWriter, r *http.Request) {
		cfg, err := consensus.GetConfiguration(rft)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, cfg)
	})
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		stCopy := fsm.GetStateCopy()
		writeJSON(w, stCopy)
	})
	mux.HandleFunc("/api/networks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, fsm.GetStateCopy().Networks)
		case http.MethodPost:
			var nw api.Network
			if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if nw.ID == "" || nw.CIDR == "" {
				http.Error(w, "id and cidr required", 400)
				return
			}
			if err := st.Apply(r.Context(), store.NewCommand("UpsertNetwork", nw)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/networks/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
			http.Error(w, "id required", 400)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("DeleteNetwork", body.ID)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/api/storagepools", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, fsm.GetStateCopy().StoragePools)
		case http.MethodPost:
			var sp api.StoragePool
			if err := json.NewDecoder(r.Body).Decode(&sp); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if sp.ID == "" {
				http.Error(w, "id required", 400)
				return
			}
			if err := st.Apply(r.Context(), store.NewCommand("UpsertStoragePool", sp)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/storagepools/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
			http.Error(w, "id required", 400)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("DeleteStoragePool", body.ID)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, fsm.GetStateCopy().Config)
		case http.MethodPost:
			var cfg api.ClusterConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			// validate
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
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/config/version", func(w http.ResponseWriter, r *http.Request) {
		stCopy := fsm.GetStateCopy()
		writeJSON(w, map[string]any{"version": stCopy.ConfigVersion})
	})
	mux.HandleFunc("/api/config/history", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, fsm.GetStateCopy().ConfigHistory)
	})
	mux.HandleFunc("/api/config/rollback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := st.Apply(r.Context(), store.NewCommand("RollbackConfig", nil)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "clusterd_up 1\n")
		stCopy := fsm.GetStateCopy()
		fmt.Fprintf(w, "nodes_total %d\n", len(stCopy.Nodes))
		fmt.Fprintf(w, "vms_total %d\n", len(stCopy.VMs))
	})
	mux.HandleFunc("/api/vms", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, fsm.GetStateCopy().VMs)
	})
	mux.HandleFunc("/api/nodes/list", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, fsm.GetStateCopy().Nodes)
	})

	// Simple health endpoint for UI checks
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	// CRUD: Nodes
	mux.HandleFunc("/api/nodes/upsert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// pass-through; in real impl, bind to api.Node
		if err := st.Apply(context.Background(), store.NewCommand("UpsertNode", body)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/api/nodes/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := st.Apply(context.Background(), store.NewCommand("DeleteNode", body.ID)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})

	// CRUD: VMs
	mux.HandleFunc("/api/vms/upsert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// naive scheduling if nodeId is empty
		if _, ok := body["nodeId"]; !ok {
			stCopy := fsm.GetStateCopy()
			vm := mapToVM(body)
			if nid, ok := scheduler.ChooseNode(stCopy, vm); ok {
				body["nodeId"] = nid
			}
		}
		if err := st.Apply(context.Background(), store.NewCommand("UpsertVM", body)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/api/vms/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := st.Apply(context.Background(), store.NewCommand("DeleteVM", body.ID)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/api/serf/members", func(w http.ResponseWriter, r *http.Request) {
		type mem struct {
			Name   string
			Addr   string
			Status string
		}
		var list []mem
		for _, m := range s.Members() {
			list = append(list, mem{Name: m.Name, Addr: fmt.Sprintf("%s:%d", m.Addr, m.Port), Status: m.Status.String()})
		}
		writeJSON(w, list)
	})
	mux.HandleFunc("/api/nodes", func(w http.ResponseWriter, r *http.Request) {
		type node struct {
			Name   string            `json:"name"`
			Addr   string            `json:"addr"`
			Status string            `json:"status"`
			Tags   map[string]string `json:"tags"`
		}
		nodes := []node{}
		for _, m := range s.Members() {
			if m.Tags["role"] == "node" {
				nodes = append(nodes, node{Name: m.Name, Addr: fmt.Sprintf("%s:%d", m.Addr, m.Port), Status: m.Status.String(), Tags: m.Tags})
			}
		}
		writeJSON(w, nodes)
	})
	httpSrv := &http.Server{Addr: uiAddr, Handler: mux}
	go func() {
		log.Printf("UI listening on http://%s", uiAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	// Keep process alive; shutdown on SIGTERM would be added later
	<-context.Background().Done()

	// Graceful stops (unreachable in this minimal skeleton)
	_ = httpSrv.Shutdown(context.Background())
	grpcServer.GracefulStop()
}

func splitCSV(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

// helper: convert generic map to VM struct for scheduler
func mapToVM(m map[string]any) api.VM {
	vm := api.VM{ID: str(m["id"]), Name: str(m["name"])}
	vm.NodeID = str(m["nodeId"])
	vm.Resources = api.Resources{CPU: toInt(m["cpu"]), Memory: toInt(m["memory"]), Disk: toInt(m["disk"])}
	return vm
}

func toInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	default:
		return 0
	}
}
func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
