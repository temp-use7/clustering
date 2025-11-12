package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	clusterpb "clustering/api/proto/cluster"
	nodepb "clustering/api/proto/node"
	templatepb "clustering/api/proto/template"
	vmpb "clustering/api/proto/vm"
	"clustering/pkg/api"
	grpcapi "clustering/pkg/api/grpc"
	"clustering/pkg/consensus"
	fsctrl "clustering/pkg/controllers/failover"
	hcctrl "clustering/pkg/controllers/health"
	mc "clustering/pkg/controllers/membership"
	nsync "clustering/pkg/controllers/nodesync"
	"clustering/pkg/membership"
	"clustering/pkg/metrics"
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
			}
		}
	}

	// Store manager
	storeManager := store.NewManager(rft)
	storeManager.SetFSM(fsm)

	// Controllers
	membershipCtrl := mc.NewController(rft, func() []mc.AliveMember {
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

	nodesyncCtrl := nsync.NewController(func() []nsync.MemberInfo {
		var out []nsync.MemberInfo
		for _, m := range s.Members() {
			role := m.Tags["role"]
			status := m.Status.String()
			if role == "node" && joinToken != "" && m.Tags["token"] != joinToken {
				continue
			}
			addr := m.Addr.String()
			if m.Tags["http"] != "" {
				addr = m.Addr.String() + ":" + m.Tags["http"]
			}
			out = append(out, nsync.MemberInfo{ID: m.Name, Addr: addr, Role: role, Status: status, Tags: m.Tags})
		}
		return out
	}, storeManager, func() bool { return rft.State() == raft.Leader })

	healthCtrl := hcctrl.NewController(func() error {
		state := storeManager.GetStateCopy()
		for _, n := range state.Nodes {
			if n.Role != "node" {
				continue
			}
			// Simple health check - in production would make HTTP request
			if n.Status != "Alive" {
				// Update node status
				n.Status = "Failed"
				storeManager.Apply(context.Background(), store.NewCommand("UpsertNode", n))
			}
		}
		return nil
	})

	failoverCtrl := fsctrl.NewController(storeManager, func() bool { return rft.State() == raft.Leader })

	// Start controllers
	stopCh := make(chan struct{})
	go membershipCtrl.Run(stopCh)
	go nodesyncCtrl.Run(stopCh)
	go healthCtrl.Run(stopCh)
	go failoverCtrl.Run(stopCh)

	// HTTP server
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/nodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.Nodes)
		}
	})

	mux.HandleFunc("/api/vms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.VMs)
		} else if r.Method == "POST" {
			var vm api.VM
			if err := json.NewDecoder(r.Body).Decode(&vm); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("UpsertVM", vm)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// Config endpoints
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.Config)
		} else if r.Method == "POST" {
			var cfg api.ClusterConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("SetConfig", cfg)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// Config versioning endpoints
	mux.HandleFunc("/api/config/version", func(w http.ResponseWriter, r *http.Request) {
		state := storeManager.GetStateCopy()
		json.NewEncoder(w).Encode(map[string]int{"version": state.ConfigVersion})
	})

	mux.HandleFunc("/api/config/history", func(w http.ResponseWriter, r *http.Request) {
		state := storeManager.GetStateCopy()
		json.NewEncoder(w).Encode(state.ConfigHistory)
	})

	mux.HandleFunc("/api/config/rollback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				Version int `json:"version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("RollbackConfig", req.Version)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// Networks endpoints
	mux.HandleFunc("/api/networks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.Networks)
		} else if r.Method == "POST" {
			var network api.Network
			if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("UpsertNetwork", network)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// Storage pools endpoints
	mux.HandleFunc("/api/storagepools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.StoragePools)
		} else if r.Method == "POST" {
			var pool api.StoragePool
			if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("UpsertStoragePool", pool)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// Volumes endpoints
	mux.HandleFunc("/api/volumes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.Volumes)
		} else if r.Method == "POST" {
			var volume api.Volume
			if err := json.NewDecoder(r.Body).Decode(&volume); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("UpsertVolume", volume)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// Templates endpoints
	mux.HandleFunc("/api/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			state := storeManager.GetStateCopy()
			json.NewEncoder(w).Encode(state.Templates)
		} else if r.Method == "POST" {
			var template api.VMTemplate
			if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := storeManager.Apply(r.Context(), store.NewCommand("UpsertVMTemplate", template)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(204)
		}
	})

	// VM operations
	mux.HandleFunc("/api/vms/clone", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				SourceID string `json:"sourceId"`
				NewID    string `json:"newId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			// Clone logic would go here
			w.WriteHeader(204)
		}
	})

	mux.HandleFunc("/api/vms/migrate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				VMID       string `json:"vmId"`
				TargetNode string `json:"targetNode"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			// Migration logic would go here
			w.WriteHeader(204)
		}
	})

	mux.HandleFunc("/api/vms/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				VMID       string `json:"vmId"`
				SnapshotID string `json:"snapshotId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			// Snapshot logic would go here
			w.WriteHeader(204)
		}
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(metrics.RenderPrometheus()))
	})

	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	// Audit endpoint
	mux.HandleFunc("/api/audit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(storeManager.Audit())
	})

	// Serve UI
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui/dist"))))

	// Start HTTP server
	httpSrv := &http.Server{Addr: uiAddr, Handler: mux}
	go func() {
		log.Printf("Starting HTTP server on %s", uiAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	clusterpb.RegisterClusterServiceServer(grpcServer, grpcapi.NewClusterServer(rft))
	nodepb.RegisterNodeServiceServer(grpcServer, grpcapi.NewNodeServer(storeManager))
	vmpb.RegisterVMServiceServer(grpcServer, grpcapi.NewVMServer(storeManager, fsm))
	templatepb.RegisterTemplateServiceServer(grpcServer, grpcapi.NewTemplateServer(storeManager, fsm))

	// Health service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Start gRPC server
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	go func() {
		log.Printf("Starting gRPC server on %s", grpcAddr)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")

	// Create snapshot before shutdown
	if err := rft.Snapshot().Error(); err != nil {
		log.Printf("Failed to create snapshot: %v", err)
	}

	// Shutdown components with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop controllers
	close(stopCh)

	// Shutdown HTTP server
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("Failed to shutdown HTTP server: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	// Shutdown Serf
	if err := s.Shutdown(); err != nil {
		log.Printf("Failed to shutdown Serf: %v", err)
	}

	// Shutdown Raft
	if err := rft.Shutdown().Error(); err != nil {
		log.Printf("Failed to shutdown Raft: %v", err)
	}

	log.Println("Shutdown complete")
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
