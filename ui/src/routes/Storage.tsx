import { FormEvent, useEffect, useState } from 'react'

type Pool = { id: string; type?: string; size: number }
type Vol = { id: string; size: number; node?: string }

export default function Storage() {
  const [pools, setPools] = useState<Pool[]>([])
  const [vols, setVols] = useState<Vol[]>([])
  const [pForm, setPForm] = useState<Pool>({ id: '', type: 'local', size: 100 })
  const [vForm, setVForm] = useState<Vol>({ id: '', size: 10, node: '' })
  const fetchPools = () => fetch('/api/storagepools').then(r => r.json()).then((m: Record<string, any>) => {
    const list = Object.values(m || {}) as any[]
    setPools(list.map(v => ({ id: (v as any).id || '', type: (v as any).type || '', size: (v as any).size || 0 })))
  })
  const fetchVols = () => fetch('/api/volumes').then(r => r.json()).then((m: Record<string, any>) => {
    const list = Object.values(m || {}) as any[]
    setVols(list.map(v => ({ id: (v as any).id || '', size: (v as any).size || 0, node: (v as any).node || '' })))
  })
  useEffect(() => { fetchPools(); fetchVols() }, [])
  const onPoolSubmit = async (e: FormEvent) => {
    e.preventDefault()
    await fetch('/api/storagepools', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(pForm) })
    setPForm({ id: '', type: 'local', size: 100 })
    fetchPools()
  }
  const onPoolDelete = async (id: string) => {
    await fetch('/api/storagepools/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id }) })
    fetchPools()
  }
  const onVolSubmit = async (e: FormEvent) => {
    e.preventDefault()
    await fetch('/api/volumes', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(vForm) })
    setVForm({ id: '', size: 10, node: '' })
    fetchVols()
  }
  const onVolDelete = async (id: string) => {
    await fetch('/api/volumes/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id }) })
    fetchVols()
  }
  return (
    <div>
      <h2>Storage</h2>
      <h3>Pools</h3>
      <form className="form" onSubmit={onPoolSubmit}>
        <div className="row">
          <label>ID</label><input value={pForm.id} onChange={e => setPForm({ ...pForm, id: e.target.value })} required />
          <label>Type</label><input value={pForm.type} onChange={e => setPForm({ ...pForm, type: e.target.value })} />
          <label>Size</label><input type="number" value={pForm.size} onChange={e => setPForm({ ...pForm, size: Number(e.target.value) })} />
        </div>
        <button type="submit">Upsert Pool</button>
      </form>
      <table className="table"><thead><tr><th>ID</th><th>Type</th><th>Size</th><th>Actions</th></tr></thead><tbody>
        {pools.map(p => (<tr key={p.id}><td>{p.id}</td><td>{p.type}</td><td>{p.size}</td><td><button onClick={() => onPoolDelete(p.id)}>Delete</button></td></tr>))}
      </tbody></table>
      <h3>Volumes</h3>
      <form className="form" onSubmit={onVolSubmit}>
        <div className="row">
          <label>ID</label><input value={vForm.id} onChange={e => setVForm({ ...vForm, id: e.target.value })} required />
          <label>Size</label><input type="number" value={vForm.size} onChange={e => setVForm({ ...vForm, size: Number(e.target.value) })} />
          <label>Node</label><input value={vForm.node} onChange={e => setVForm({ ...vForm, node: e.target.value })} />
        </div>
        <button type="submit">Upsert Volume</button>
      </form>
      <table className="table"><thead><tr><th>ID</th><th>Size</th><th>Node</th><th>Actions</th></tr></thead><tbody>
        {vols.map(v => (<tr key={v.id}><td>{v.id}</td><td>{v.size}</td><td>{v.node}</td><td><button onClick={() => onVolDelete(v.id)}>Delete</button></td></tr>))}
      </tbody></table>
    </div>
  )
}



