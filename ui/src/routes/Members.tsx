import { useEffect, useState } from 'react'

type Member = { name: string; addr: string; status: string }

export default function Members() {
  const [members, setMembers] = useState<Member[]>([])
  useEffect(() => { fetch('/api/serf/members').then(r => r.json()).then(setMembers) }, [])
  return (
    <div>
      <h2>Serf Members</h2>
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



