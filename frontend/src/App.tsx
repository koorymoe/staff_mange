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
import KpiPage from './pages/KpiPage'
import ComplaintsPage from './pages/ComplaintsPage'
import InventoryPage from './pages/InventoryPage'
import MyInventory from './pages/MyInventory'
import PermissionsPage from './pages/PermissionsPage'
import QuotationsPage from './pages/QuotationsPage'
import QuotationNew from './pages/QuotationNew'
import ProductsPage from './pages/ProductsPage'
import GpsDashboard from './pages/gps/GpsDashboard'
import GpsCustomers from './pages/gps/GpsCustomers'
import GpsDevices from './pages/gps/GpsDevices'
import GpsSims from './pages/gps/GpsSims'
import GpsRenewals from './pages/gps/GpsRenewals'
import GpsMaintenance from './pages/gps/GpsMaintenance'
import GpsEmployee from './pages/gps/GpsEmployee'
import AttendancePage from './pages/AttendancePage'
import WorkReportPage from './pages/WorkReportPage'

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
        <Route path="kpi" element={<KpiPage />} />
        <Route path="complaints" element={<ComplaintsPage />} />
        <Route path="inventory" element={<InventoryPage />} />
        <Route path="my-inventory" element={<MyInventory />} />
        <Route path="permissions" element={<PermissionsPage />} />
        <Route path="quotations" element={<QuotationsPage />} />
        <Route path="quotations/new" element={<QuotationNew />} />
        <Route path="products" element={<ProductsPage />} />
        <Route path="attendance" element={<AttendancePage />} />
        <Route path="work-reports" element={<WorkReportPage />} />
        <Route path="gps" element={<GpsDashboard />} />
        <Route path="gps/customers" element={<GpsCustomers />} />
        <Route path="gps/devices" element={<GpsDevices />} />
        <Route path="gps/sims" element={<GpsSims />} />
        <Route path="gps/renewals" element={<GpsRenewals />} />
        <Route path="gps/maintenance" element={<GpsMaintenance />} />
        <Route path="gps/employee" element={<GpsEmployee />} />
      </Route>
    </Routes>
  )
}

export default App
