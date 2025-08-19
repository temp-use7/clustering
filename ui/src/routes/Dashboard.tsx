import { useEffect, useState } from 'react'

type Member = { name: string; addr: string; status: string }
type ClusterStatus = { leader: string; voterCount: number; nonvoterCount: number }

export default function Dashboard() {
  const [status, setStatus] = useState<ClusterStatus | null>(null)
  const [members, setMembers] = useState<Member[]>([])

  useEffect(() => {
    // Best-effort: use HTTP endpoints exposed by clusterd
    fetch('/api/serf/members').then(r => r.json()).then(setMembers).catch(() => {})
    // Cluster status via minimal HTTP: we don't have an HTTP endpoint, only gRPC.
    // Fall back to raft config for visibility.
    fetch('/api/raft/config').then(r => r.json()).then(cfg => {
      const leader = cfg.leader || ''
      const voterCount = (cfg.servers || []).filter((s: any) => s.suffrage === 'Voter' || s.suffrage === 0).length
      const nonvoterCount = (cfg.servers || []).length - voterCount
      setStatus({ leader, voterCount, nonvoterCount })
    }).catch(() => {})
  }, [])

  return (
    <div>
      <h2>Dashboard</h2>
      {status && (
        <div className="cards">
          <div className="card"><div className="label">Leader</div><div className="value">{status.leader || '—'}</div></div>
          <div className="card"><div className="label">Voters</div><div className="value">{status.voterCount}</div></div>
          <div className="card"><div className="label">Non-voters</div><div className="value">{status.nonvoterCount}</div></div>
        </div>
      )}

      <h3>Members</h3>
      <table className="table">
        <thead>
          <tr><th>Name</th><th>Address</th><th>Status</th></tr>
        </thead>
        <tbody>
          {members.map((m, i) => (
            <tr key={i}><td>{m.name}</td><td>{m.addr}</td><td>{m.status}</td></tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}



