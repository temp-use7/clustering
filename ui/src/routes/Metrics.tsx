import { useEffect, useState } from 'react'

export default function Metrics() {
  const [text, setText] = useState<string>('')
  useEffect(() => {
    fetch('/metrics').then(r => r.text()).then(setText).catch(() => setText('error'))
  }, [])
  return (
    <div>
      <h2>Metrics</h2>
      <pre>{text}</pre>
    </div>
  )
}



