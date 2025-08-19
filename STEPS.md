Local run and test guide (7-node cluster: 5 voters + 2 non-voters)

Prerequisites
- Go 1.22+ installed
- Windows PowerShell (commands below use Windows path separators)
- Free local ports (we’ll use distinct ports per node)

Build (optional; go run also works)
- go mod tidy
- go build -o bin\clusterd .\cmd\clusterd
- go build -o bin\nodeagent .\cmd\nodeagent
- go build -o bin\clustectl .\cmd\clustectl

Start 7 control-plane nodes (one per PowerShell window)
- Node 1 (bootstrap leader)
  go run .\cmd\clusterd --node-id=node-1 --data-dir=.\data\node-1 --raft-bind=:7000 --grpc=:18081 --ui=:18080 --serf-bind=:17946 --bootstrap --wipe-data

- Nodes 2-7 (join to node-1’s Serf)
  go run .\cmd\clusterd --node-id=node-2 --data-dir=.\data\node-2 --raft-bind=:7001 --grpc=:28081 --ui=:28080 --serf-bind=:17947 --serf-join=127.0.0.1:17946 --wipe-data
  go run .\cmd\clusterd --node-id=node-3 --data-dir=.\data\node-3 --raft-bind=:7002 --grpc=:38081 --ui=:38080 --serf-bind=:17948 --serf-join=127.0.0.1:17946 --wipe-data
  go run .\cmd\clusterd --node-id=node-4 --data-dir=.\data\node-4 --raft-bind=:7003 --grpc=:48081 --ui=:48080 --serf-bind=:17949 --serf-join=127.0.0.1:17946 --wipe-data
  go run .\cmd\clusterd --node-id=node-5 --data-dir=.\data\node-5 --raft-bind=:7004 --grpc=:58081 --ui=:58080 --serf-bind=:17950 --serf-join=127.0.0.1:17946 --wipe-data
  go run .\cmd\clusterd --node-id=node-6 --data-dir=.\data\node-6 --raft-bind=:7005 --grpc=:60081 --ui=:60080 --serf-bind=:17951 --serf-join=127.0.0.1:17946 --wipe-data
  go run .\cmd\clusterd --node-id=node-7 --data-dir=.\data\node-7 --raft-bind=:7006 --grpc=:61081 --ui=:61080 --serf-bind=:17952 --serf-join=127.0.0.1:17946 --wipe-data

Notes
- The membership controller will auto-add alive members as non-voters, then promote up to 5 voters, keeping the remainder as non-voters. You should end up with 5 voters and 2 non-voters.
- The leader’s UI shows the current raft configuration and members.

Open the UI
- Visit each node’s UI (e.g., http://localhost:18080 for node-1). The dashboard shows:
  - Control-plane status (leader, raft state)
  - Serf members
  - Nodes view (from Serf tags)
  - Cluster state (Nodes with capacity/allocated; VMs)
  - Forms to create Nodes and VMs, and buttons to delete

Create Nodes via UI (optional)
- In “Create Node” form, specify:
  - ID, Address (informational), Role (node/control-plane), Status (Alive)
  - Capacity: CPU (millicores), Memory (MiB), Disk (GiB)
- Submit to upsert the node in the cluster state (FSM). This is orthogonal to raft membership (which is driven by Serf for control-plane nodes).

Create VMs via UI
- In “Create VM” form, specify:
  - ID, Name, CPU (millicores), Memory (MiB), Disk (GiB)
  - Optional NodeID. If omitted, the scheduler will pick a node based on available capacity.

Simulate failure and migration
- Stop one control-plane process (non-leader or leader). The membership controller maintains quorum (promotes a non-voter if needed). The migration controller will re-place VMs whose nodes are not Alive.

CLI (optional)
- List nodes via HTTP proxy from UI server:
  go run .\cmd\clustectl --ui http://localhost:18080 nodes
- List VMs:
  go run .\cmd\clustectl --ui http://localhost:18080 vms

Run tests
- go test ./...

Troubleshooting
- If Raft reports “not advertisable,” ensure raft-bind is not 0.0.0.0. This project normalizes empty host to 127.0.0.1.
- If promotions fail with “not leader,” use the leader’s UI to verify leadership; only the leader will reconcile membership.
- Ports must be unique per node; adjust if conflicts occur.


