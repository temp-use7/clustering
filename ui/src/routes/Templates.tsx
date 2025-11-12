import { FormEvent, useEffect, useState } from 'react'

type Tpl = { id: string; name?: string; baseImage?: string; cpu: number; memory: number; disk: number }

export default function Templates() {
  const [tpls, setTpls] = useState<Tpl[]>([])
  const [form, setForm] = useState<Tpl>({ id: '', name: '', baseImage: '', cpu: 200, memory: 256, disk: 5 })
  const [inst, setInst] = useState<{ templateId: string; newId: string }>({ templateId: '', newId: '' })

  const fetchTpls = () => fetch('/api/templates').then(r => r.json()).then((m: Record<string, any>) => {
    const list = Object.values(m || {}) as any[]
    setTpls(list.map(v => ({ id: v.id, name: v.name, baseImage: v.baseImage, cpu: v.resources?.cpu ?? 0, memory: v.resources?.memory ?? 0, disk: v.resources?.disk ?? 0 })))
  })

  useEffect(() => { fetchTpls() }, [])

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const payload: any = { id: form.id, name: form.name, baseImage: form.baseImage, resources: { cpu: form.cpu, memory: form.memory, disk: form.disk } }
    await fetch('/api/templates', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    setForm({ id: '', name: '', baseImage: '', cpu: 200, memory: 256, disk: 5 })
    fetchTpls()
  }

  const onDelete = async (id: string) => {
    await fetch('/api/templates/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id }) })
    fetchTpls()
  }

  const onInstantiate = async (e: FormEvent) => {
    e.preventDefault()
    await fetch('/api/vms/cloneFromTemplate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(inst) })
    setInst({ templateId: '', newId: '' })
  }

  return (
    <div>
      <h2>Templates</h2>
      <form className="form" onSubmit={onSubmit}>
        <div className="row">
          <label>ID</label>
          <input value={form.id} onChange={e => setForm({ ...form, id: e.target.value })} required />
          <label>Name</label>
          <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />
          <label>Base Image</label>
          <input value={form.baseImage} onChange={e => setForm({ ...form, baseImage: e.target.value })} />
        </div>
        <div className="row">
          <label>CPU (m)</label>
          <input type="number" value={form.cpu} onChange={e => setForm({ ...form, cpu: Number(e.target.value) })} />
          <label>Memory (MiB)</label>
          <input type="number" value={form.memory} onChange={e => setForm({ ...form, memory: Number(e.target.value) })} />
          <label>Disk (GiB)</label>
          <input type="number" value={form.disk} onChange={e => setForm({ ...form, disk: Number(e.target.value) })} />
        </div>
        <button type="submit">Upsert Template</button>
      </form>

      <h3>Instantiate</h3>
      <form className="form" onSubmit={onInstantiate}>
        <div className="row">
          <label>Template</label>
          <select value={inst.templateId} onChange={e => setInst({ ...inst, templateId: e.target.value })}>
            <option value="">Select...</option>
            {tpls.map(t => <option key={t.id} value={t.id}>{t.id}</option>)}
          </select>
          <label>New VM ID</label>
          <input value={inst.newId} onChange={e => setInst({ ...inst, newId: e.target.value })} />
        </div>
        <button type="submit">Instantiate VM</button>
      </form>

      <table className="table">
        <thead><tr><th>ID</th><th>Name</th><th>Base</th><th>CPU</th><th>Mem</th><th>Disk</th><th>Actions</th></tr></thead>
        <tbody>
          {tpls.map(t => (
            <tr key={t.id}>
              <td>{t.id}</td><td>{t.name}</td><td>{t.baseImage}</td><td>{t.cpu}</td><td>{t.memory}</td><td>{t.disk}</td>
              <td><button onClick={() => onDelete(t.id)}>Delete</button></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}



