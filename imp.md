## Implementation Plan (imp.md)

This document enumerates the goals from `goal.md`, maps current status, identifies gaps, and defines a test-driven plan to deliver each item. We work in incremental milestones with unit and integration tests.

### Legend
- Status: Implemented | Partial | Planned
- Tests: UT (unit tests), IT (integration tests), e2e (manual/automated multi-process)

---

## 1) Cluster Management

1.1 Distributed Cluster Formation (Raft)
- Status: Implemented (Raft boot with Bolt stores, snapshots; bootstrap flag)
- Gaps: Documented joint consensus usage, graceful shutdown hooks
- Plan: 
  - Add graceful shutdown and snapshot on exit (Planned)
  - IT: multi-node raft config and leader election stability

1.2 Leader Election (Raft)
- Status: Implemented (native Raft)
- Gaps: None immediate
- Tests: IT/e2e: observe leadership transitions under churn

1.3 Cluster Membership (Serf)
- Status: Partial (Serf running; membership controller adds nonvoters, plans promotions/demotions, executes simple promotion/demotion)
- Gaps: Robust demotion via joint consensus; remove departed nodes; backoff/retry
- Plan:
  - Implement remove on Serf Left/Failed (demote/remove) with retries
  - Respect desired voters from distributed config (see 4.x)
  - UT: planning logic
  - IT: add/remove nodes dynamically

1.4 Node Failure Detection
- Status: Partial (Serf events; nodesync writes status; migration controller re-places VMs)
- Gaps: Health heartbeat to agent HTTP; thresholds; state reconciliation
- Plan:
  - Add health controller to probe `nodeagent` `/healthz`; persist into FSM
  - UT: health aggregation logic; IT: kill agent and observe status change

1.5 Dynamic Node Add/Remove
- Status: Partial (AddNonvoter on alive; basic demote/remove)
- Gaps: Automated cleanup on Left; leader-only reconciliation
- Plan: Covered in 1.3

1.6 State Consistency (FSM)
- Status: Implemented (FSM apply/snapshot/restore tested)
- Gaps: Validation on commands; schema evolution strategy
- Plan: Add validation paths per command

---

## 2) Node Discovery & Registration

2.1 Auto-discovery (Serf)
- Status: Implemented

2.2 Node Credential Validation
- Status: Planned
- Plan:
  - Introduce shared join token (env/flag) advertised by `nodeagent` and verified by control-plane on registration
  - Optional mTLS (future)
  - UT: token validation; IT: agent with/without token

2.3 Node Inventory/Capabilities
- Status: Partial (default capacity via nodesync; UI/API upsert)
- Plan:
  - Extend `nodeagent` to report inventory (CPU/Memory/Disk) via Serf tags or HTTP
  - Persist into FSM
  - UT: parsing/merge rules

2.4 Node Health/Heartbeat
- Status: Planned
- Plan:
  - Implement health controller that periodically probes agent HTTP and updates FSM
  - UT/IT as in 1.4

---

## 3) Resource Management

3.1 APIs for Allocation/Scheduling
- Status: Partial (HTTP+gRPC VM upsert with simple spread scheduler)
- Plan: Add placement policies hook (5.x), input validation

3.2 Resource Metrics (OpenTelemetry)
- Status: Planned
- Plan: Instrument scheduler/VM lifecycle and expose metrics endpoint; add OTEL exporter toggle
  - UT: metrics labels; IT: scrape endpoint

3.3 Placement Policies (models)
- Status: Planned
- Plan: Policy structs in state; scheduler reads them
- UT: policy evaluation

3.4 Affinity
- Status: Planned
- Plan: VM/node label-based affinity/anti-affinity in scheduler
- UT: scenarios

3.5 Load Balancing
- Status: Planned
- Plan: Balance by multiple resources; rebalancing controller (low priority)

---

## 4) Configuration Management

4.1 Config Distribution
- Status: Planned
- Plan (Milestone M1):
  - Add `ClusterConfig` to `ClusterState` (replicated via Raft)
  - FSM command `SetConfig` with validation
  - HTTP: `GET/POST /api/config`
  - Wire membership controller to read `DesiredVoters` from state
  - UT: FSM set/restore; handler validation
  - IT: update config and observe membership reconciliation

4.2 Config Versioning/Rollback
- Status: Planned
- Plan: Maintain version and history log in state; rollback command

4.3 Validation
- Status: Planned
- Plan: Schema and constraints (voters >= 1), endpoint-level validation

---

## 5) VM Lifecycle Management

5.1 VM CRUD
- Status: Implemented (HTTP+gRPC; FSM tested)
- Plan: Input validation; richer phases

5.2 VM Clone
- Status: Planned
- Plan: FSM command to clone metadata; runtime hook (mock)
  - UT: state changes; IT: API roundtrip

5.3 VM Migrate
- Status: Partial (gRPC/API; migration controller re-places VMs)
- Plan: Runtime simulation hook; migration completion

5.4 VM Snapshot
- Status: Planned
- Plan: FSM snapshot metadata; runtime hook (mock)

5.5 VM Template
- Status: Planned

---

## 6) Networking & Storage
- Status: Planned
- Plan: Define minimal types (Networks, Pools, Disks) and CRUD; no real dataplane yet.
- UT: store commands; IT: API endpoints

---

## 7) High Availability & DR
- Status: Planned
- Plan: HA policy in state; controller to auto-restart; backup hooks later.

---

## 8) Monitoring & Observability

8.1 Metrics
- Status: Planned
- Plan: Prometheus metrics endpoint; OTEL optional

8.2 Health
- Status: Partial (HTTP `/api/health`, gRPC health)
- Plan: Node/cluster health surfaces in UI

8.3 Audit Logs
- Status: Planned
- Plan: Structured audit events on state-changing commands

---

## 9) Interfaces

9.1 gRPC API
- Status: Partial (servers exist; codegen stubs referenced)
- Plan: Define/commit proto schema and full handlers; add e2e tests

9.2 CLI
- Status: Partial (nodes/vms list & vm ops)
- Plan: Add config and node ops

9.3 Web Dashboard
- Status: Partial (React UI exists; served under `/ui/`)
- Plan: Surfaces for state, nodes, VMs, config, health

---

## Milestone M1 (Foundations)
Scope: Config Distribution + Validation; Health Controller; Membership reads desired voters; tests.

- [x] Add `ClusterConfig` to state and FSM `SetConfig` command
- [x] HTTP `GET/POST /api/config` with validation
- [x] Membership controller reads `DesiredVoters` from state
- [ ] Health controller probes agents and updates node health
- [x] UT: FSM config
- [ ] UT: handlers; membership planning with dynamic desired voters
- [ ] IT: single-node demo — change desired voters; multi-node membership adjusts

## Milestone M2 (Scheduling/Policies)
- [ ] Policy structs; scheduler reads affinity/priority
- [ ] UT: policy evaluation; IT: placement behavior

## Milestone M3 (Security & Credentials)
- [ ] Join token validation
- [ ] UT/IT for token flows

## Milestone M4 (Observability)
- [ ] Metrics endpoint + basic counters
- [ ] UT for metrics labels; IT scrape

## Milestone M5 (VM Clone/Snapshot)
- [ ] FSM + mock runtime; UT/IT for transitions


