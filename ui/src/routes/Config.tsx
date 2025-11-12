import { useEffect, useState } from 'react'

export default function Config() {
  const [cfg, setCfg] = useState<any>({ desiredVoters: 5, desiredNonVoters: 2 })
  const [version, setVersion] = useState<number>(0)
  const load = async () => {
    const c = await fetch('/api/config').then(r => r.json())
    setCfg(c)
    const v = await fetch('/api/config/version').then(r => r.json()).catch(() => ({} as any))
    if (v && typeof v.version === 'number') setVersion(v.version)
  }
  useEffect(() => { load() }, [])
  const onSave = async () => {
    await fetch('/api/config', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(cfg) })
    load()
  }
  const onRollback = async () => {
    await fetch('/api/config/rollback', { method: 'POST' })
    load()
  }
  return (
    <div>
      <h2>Config</h2>
      <div className="row">
        <label>Desired Voters</label>
        <input type="number" value={cfg.desiredVoters||0} onChange={e => setCfg({ ...cfg, desiredVoters: Number(e.target.value) })} />
        <label>Desired NonVoters</label>
        <input type="number" value={cfg.desiredNonVoters||0} onChange={e => setCfg({ ...cfg, desiredNonVoters: Number(e.target.value) })} />
      </div>
      <div className="row">
        <button onClick={onSave}>Save</button>
        <button onClick={onRollback}>Rollback</button>
        <span style={{ marginLeft: 16 }}>Version: {version}</span>
      </div>
    </div>
  )
}



