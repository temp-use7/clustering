export default function Health() {
  return (
    <div>
      <h2>Health</h2>
      <pre id="health"></pre>
      <script dangerouslySetInnerHTML={{ __html: `
        fetch('/api/health').then(r => r.json()).then(j => {
          document.getElementById('health').textContent = JSON.stringify(j, null, 2)
        }).catch(e => {
          document.getElementById('health').textContent = 'error: ' + e
        })
      `}} />
    </div>
  )
}



