package consensus

import (
    "log"
    "net"
    "os"
    "path/filepath"
    "time"

    "clustering/pkg/store"

    "github.com/hashicorp/raft"
    raftboltdb "github.com/hashicorp/raft-boltdb"
)

// MustStartRaft initializes and starts a HashiCorp Raft node with a noop FSM.
func MustStartRaft(nodeID, bindAddr, dataDir string) (*raft.Raft, *raft.NetworkTransport, *store.FSM) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("mkdir data dir: %v", err)
	}

	logPath := filepath.Join(dataDir, "raft-log.bolt")
	stablePath := filepath.Join(dataDir, "raft-stable.bolt")
	snapshotsDir := filepath.Join(dataDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0o755); err != nil {
		log.Fatalf("mkdir snapshots: %v", err)
	}

	lstore, err := raftboltdb.NewBoltStore(logPath)
	if err != nil {
		log.Fatalf("open log store: %v", err)
	}
	sstore, err := raftboltdb.NewBoltStore(stablePath)
	if err != nil {
		log.Fatalf("open stable store: %v", err)
	}
	snapStore, err := raft.NewFileSnapshotStore(snapshotsDir, 2, os.Stderr)
	if err != nil {
		log.Fatalf("snapshot store: %v", err)
	}

    bind := ensureHost(bindAddr)
    addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		log.Fatalf("resolve raft addr: %v", err)
	}
    // Fix: Use raft.NewTCPTransport instead of raft.NewNetworkTransport, and ensure advertisable address
    transport, err := raft.NewTCPTransport(bind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatalf("failed to create raft transport: %v", err)
	}

	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(nodeID)

    fsm := store.NewFSM()
    r, err := raft.NewRaft(cfg, fsm, lstore, sstore, snapStore, transport)
	if err != nil {
		log.Fatalf("new raft: %v", err)
	}
    return r, transport, fsm
}

func ensureHost(addr string) string {
    host, port, err := net.SplitHostPort(addr)
    if err != nil {
        return addr
    }
    if host == "" || host == "0.0.0.0" || host == "::" {
        host = "127.0.0.1"
    }
    return net.JoinHostPort(host, port)
}
