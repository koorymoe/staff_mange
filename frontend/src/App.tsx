import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Employees from './pages/Employees'
import Customers from './pages/Customers'
import Services from './pages/Services'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="employees" element={<Employees />} />
        <Route path="customers" element={<Customers />} />
        <Route path="services" element={<Services />} />
      </Route>
    </Routes>
  )
}

export default App
