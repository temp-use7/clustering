import { FormEvent, useEffect, useState } from 'react'

type Net = { id: string; cidr: string }

export default function Networks() {
  const [nets, setNets] = useState<Net[]>([])
  const [form, setForm] = useState<Net>({ id: '', cidr: '' })
  const fetchNets = () => fetch('/api/networks').then(r => r.json()).then((m: Record<string, any>) => {
    const list = Object.values(m || {}) as any[]
    setNets(list.map(v => ({ id: (v as any).id || '', cidr: (v as any).cidr || '' })))
  })
  useEffect(() => { fetchNets() }, [])
  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    await fetch('/api/networks', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(form) })
    setForm({ id: '', cidr: '' })
    fetchNets()
  }
  const onDelete = async (id: string) => {
    await fetch('/api/networks/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id }) })
    fetchNets()
  }
  return (
    <div>
      <h2>Networks</h2>
      <form className="form" onSubmit={onSubmit}>
        <div className="row">
          <label>ID</label>
          <input value={form.id} onChange={e => setForm({ ...form, id: e.target.value })} required />
          <label>CIDR</label>
          <input value={form.cidr} onChange={e => setForm({ ...form, cidr: e.target.value })} required />
        </div>
        <button type="submit">Upsert Network</button>
      </form>
      <table className="table">
        <thead><tr><th>ID</th><th>CIDR</th><th>Actions</th></tr></thead>
        <tbody>
          {nets.map(n => (
            <tr key={n.id}><td>{n.id}</td><td>{n.cidr}</td><td><button onClick={() => onDelete(n.id)}>Delete</button></td></tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}



