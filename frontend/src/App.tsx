import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Employees from './pages/Employees'
import Customers from './pages/Customers'
import Services from './pages/Services'
import SalesBooking from './pages/SalesBooking'
import Coordinator from './pages/Coordinator'
import BookingsList from './pages/BookingsList'
import MyTasks from './pages/MyTasks'
import MyRanking from './pages/MyRanking'
import MyExpenses from './pages/MyExpenses'
import Finance from './pages/Finance'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="employees" element={<Employees />} />
        <Route path="customers" element={<Customers />} />
        <Route path="services" element={<Services />} />
        <Route path="sales" element={<SalesBooking />} />
        <Route path="coordinator" element={<Coordinator />} />
        <Route path="bookings" element={<BookingsList />} />
        <Route path="my-tasks" element={<MyTasks />} />
        <Route path="my-ranking" element={<MyRanking />} />
        <Route path="my-expenses" element={<MyExpenses />} />
        <Route path="finance" element={<Finance />} />
      </Route>
    </Routes>
  )
}

export default App
