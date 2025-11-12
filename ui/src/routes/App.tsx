import { Link, Outlet, useLocation } from 'react-router-dom'

export default function App() {
  const { pathname } = useLocation()
  return (
    <div className="app">
      <aside className="sidebar">
        <h1>Clustering</h1>
        <nav>
          <Link className={pathname === '/' ? 'active' : ''} to="/">Dashboard</Link>
          <Link className={pathname.startsWith('/nodes') ? 'active' : ''} to="/nodes">Nodes</Link>
          <Link className={pathname.startsWith('/vms') ? 'active' : ''} to="/vms">VMs</Link>
          <Link className={pathname.startsWith('/config') ? 'active' : ''} to="/config">Config</Link>
          <Link className={pathname.startsWith('/templates') ? 'active' : ''} to="/templates">Templates</Link>
          <Link className={pathname.startsWith('/networks') ? 'active' : ''} to="/networks">Networks</Link>
          <Link className={pathname.startsWith('/storage') ? 'active' : ''} to="/storage">Storage</Link>
          <Link className={pathname.startsWith('/metrics') ? 'active' : ''} to="/metrics">Metrics</Link>
          <Link className={pathname.startsWith('/audit') ? 'active' : ''} to="/audit">Audit</Link>
          <Link className={pathname.startsWith('/members') ? 'active' : ''} to="/members">Members</Link>
          <Link className={pathname.startsWith('/health') ? 'active' : ''} to="/health">Health</Link>
        </nav>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}



