import React from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import App from './routes/App'
import Dashboard from './routes/Dashboard'
import Nodes from './routes/Nodes'
import VMs from './routes/VMs'
import Members from './routes/Members'
import Health from './routes/Health'
import './styles.css'

const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'nodes', element: <Nodes /> },
      { path: 'vms', element: <VMs /> },
      { path: 'members', element: <Members /> },
      { path: 'health', element: <Health /> }
    ]
  }
], { basename: '/ui' })

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
)


