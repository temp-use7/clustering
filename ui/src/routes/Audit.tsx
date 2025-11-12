import { useEffect, useState } from 'react'

export default function Audit() {
  const [events, setEvents] = useState<any[]>([])
  useEffect(() => { fetch('/api/audit').then(r => r.json()).then(setEvents).catch(() => {}) }, [])
  return (
    <div>
      <h2>Audit</h2>
      <table className="table">
        <thead><tr><th>Type</th><th>Info</th></tr></thead>
        <tbody>
          {events.map((e, i) => (<tr key={i}><td>{e.type}</td><td>{e.info}</td></tr>))}
        </tbody>
      </table>
    </div>
  )
}



