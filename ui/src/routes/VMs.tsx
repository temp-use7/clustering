import { FormEvent, useEffect, useState } from 'react'

type VM = { id: string; name?: string; cpu: number; memory: number; disk: number; nodeId?: string; phase?: string }

export default function VMs() {
  const [vms, setVMs] = useState<VM[]>([])
  const [form, setForm] = useState<VM>({ id: '', name: '', cpu: 200, memory: 256, disk: 1, nodeId: '' })
  const [nodes, setNodes] = useState<{ id: string }[]>([])

  const fetchVMs = () => fetch('/api/vms').then(r => r.json()).then((m: Record<string, any>) => {
    const list = Object.values(m || {}) as any[]
    setVMs(list.map(v => ({ id: v.id, name: v.name, cpu: v.resources?.cpu ?? v.cpu ?? 0, memory: v.resources?.memory ?? v.memory ?? 0, disk: v.resources?.disk ?? v.disk ?? 0, nodeId: v.nodeId, phase: v.phase })))
  })
  const fetchNodes = () => fetch('/api/nodes/list').then(r => r.json()).then((m: Record<string, any>) => {
    setNodes(Object.values(m || {}) as any)
  })

  useEffect(() => { fetchVMs(); fetchNodes() }, [])

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const payload: any = { id: form.id, name: form.name, cpu: form.cpu, memory: form.memory, disk: form.disk }
    if (form.nodeId) payload.nodeId = form.nodeId
    await fetch('/api/vms/upsert', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    setForm({ id: '', name: '', cpu: 200, memory: 256, disk: 1, nodeId: '' })
    fetchVMs()
  }

  const onDelete = async (id: string) => {
    await fetch('/api/vms/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id }) })
    fetchVMs()
  }

  return (
    <div>
      <h2>VMs</h2>
      <form className="form" onSubmit={onSubmit}>
        <div className="row">
          <label>ID</label>
          <input value={form.id} onChange={e => setForm({ ...form, id: e.target.value })} required />
          <label>Name</label>
          <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />
          <label>Node</label>
          <select value={form.nodeId} onChange={e => setForm({ ...form, nodeId: e.target.value })}>
            <option value="">Auto</option>
            {nodes.map((n, i) => <option key={i} value={(n as any).id}>{(n as any).id}</option>)}
          </select>
        </div>
        <div className="row">
          <label>CPU (m)</label>
          <input type="number" value={form.cpu} onChange={e => setForm({ ...form, cpu: Number(e.target.value) })} />
          <label>Memory (MiB)</label>
          <input type="number" value={form.memory} onChange={e => setForm({ ...form, memory: Number(e.target.value) })} />
          <label>Disk (GiB)</label>
          <input type="number" value={form.disk} onChange={e => setForm({ ...form, disk: Number(e.target.value) })} />
        </div>
        <button type="submit">Upsert VM</button>
      </form>

      <table className="table">
        <thead>
          <tr><th>ID</th><th>Name</th><th>Node</th><th>CPU</th><th>Mem</th><th>Disk</th><th>Phase</th><th>Actions</th></tr>
        </thead>
        <tbody>
          {vms.map(v => (
            <tr key={v.id}>
              <td>{v.id}</td>
              <td>{v.name}</td>
              <td>{v.nodeId || '—'}</td>
              <td>{v.cpu}</td>
              <td>{v.memory}</td>
              <td>{v.disk}</td>
              <td>{v.phase || '—'}</td>
              <td>
                <button onClick={() => onDelete(v.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}



