import { FormEvent, useEffect, useMemo, useState } from 'react'

type Resources = { cpu: number; memory: number; disk: number }
type Node = {
  id: string
  address?: string
  role?: string
  status?: string
  capacity?: Resources
  allocated?: Resources
  labels?: Record<string, string>
}

export default function Nodes() {
  const [list, setList] = useState<any[]>([])
  const [nodesState, setNodesState] = useState<Record<string, Node>>({})
  const [form, setForm] = useState<Node>({ id: '', address: '', role: 'node', capacity: { cpu: 1000, memory: 1024, disk: 10 }, labels: {} })

  const fetchSerfNodes = () => fetch('/api/nodes').then(r => r.json()).then(setList)
  const fetchState = () => fetch('/api/state').then(r => r.json()).then((s) => setNodesState(s.nodes || {}))

  useEffect(() => { fetchSerfNodes(); fetchState() }, [])

  const merged = useMemo(() => {
    const map: Record<string, Node> = { ...nodesState }
    for (const m of list) {
      const id = m.name
      map[id] = {
        id,
        address: m.addr,
        role: (m.tags && m.tags.role) || map[id]?.role,
        status: m.status,
        capacity: map[id]?.capacity,
        allocated: map[id]?.allocated,
        labels: map[id]?.labels
      }
    }
    return Object.values(map)
  }, [list, nodesState])

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    await fetch('/api/nodes/upsert', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(form) })
    setForm({ id: '', address: '', role: 'node', capacity: { cpu: 1000, memory: 1024, disk: 10 }, labels: {} })
    fetchState()
  }

  const onDelete = async (id: string) => {
    await fetch('/api/nodes/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id }) })
    fetchState()
  }

  return (
    <div>
      <h2>Nodes</h2>

      <form className="form" onSubmit={onSubmit}>
        <div className="row">
          <label>ID</label>
          <input value={form.id} onChange={e => setForm({ ...form, id: e.target.value })} required />
          <label>Address</label>
          <input value={form.address} onChange={e => setForm({ ...form, address: e.target.value })} />
          <label>Role</label>
          <select value={form.role} onChange={e => setForm({ ...form, role: e.target.value })}>
            <option value="node">node</option>
            <option value="control-plane">control-plane</option>
          </select>
        </div>
        <div className="row">
          <label>CPU (m)</label>
          <input type="number" value={form.capacity?.cpu || 0} onChange={e => setForm({ ...form, capacity: { ...form.capacity!, cpu: Number(e.target.value), memory: form.capacity!.memory, disk: form.capacity!.disk } })} />
          <label>Memory (MiB)</label>
          <input type="number" value={form.capacity?.memory || 0} onChange={e => setForm({ ...form, capacity: { ...form.capacity!, cpu: form.capacity!.cpu, memory: Number(e.target.value), disk: form.capacity!.disk } })} />
          <label>Disk (GiB)</label>
          <input type="number" value={form.capacity?.disk || 0} onChange={e => setForm({ ...form, capacity: { ...form.capacity!, cpu: form.capacity!.cpu, memory: form.capacity!.memory, disk: Number(e.target.value) } })} />
        </div>
        <button type="submit">Upsert Node</button>
      </form>

      <table className="table">
        <thead>
          <tr><th>ID</th><th>Address</th><th>Role</th><th>Status</th><th>Capacity</th><th>Allocated</th><th>Actions</th></tr>
        </thead>
        <tbody>
          {merged.map(n => (
            <tr key={n.id}>
              <td>{n.id}</td>
              <td>{n.address || '—'}</td>
              <td>{n.role || '—'}</td>
              <td>{n.status || '—'}</td>
              <td>{n.capacity ? `${n.capacity.cpu}m / ${n.capacity.memory}Mi / ${n.capacity.disk}Gi` : '—'}</td>
              <td>{n.allocated ? `${n.allocated.cpu}m / ${n.allocated.memory}Mi / ${n.allocated.disk}Gi` : '—'}</td>
              <td>
                <button onClick={() => onDelete(n.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}



