import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Employees from './pages/Employees'
import Customers from './pages/Customers'
import Services from './pages/Services'
import SalesBooking from './pages/SalesBooking'
import QualityFollowUpsPage from './pages/QualityFollowUpsPage'
import SecurityDashboardPage from './pages/SecurityDashboardPage'
import Coordinator from './pages/Coordinator'
import BookingsList from './pages/BookingsList'
import MyTasks from './pages/MyTasks'
import MyRanking from './pages/MyRanking'
import MyExpenses from './pages/MyExpenses'
import Finance from './pages/Finance'
import ExpensesReview from './pages/ExpensesReview'
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
import GpsPurchase from './pages/gps/GpsPurchase'
import GpsRequestsReview from './pages/gps/GpsRequestsReview'
import GpsDelivery from './pages/gps/GpsDelivery'
import GpsRenewal from './pages/gps/GpsRenewal'
import GpsRenewalsReview from './pages/gps/GpsRenewalsReview'
import GpsMaintenanceRequestPage from './pages/gps/GpsMaintenanceRequestPage'
import GpsMaintenanceReview from './pages/gps/GpsMaintenanceReview'
import AttendancePage from './pages/AttendancePage'
import WorkReportPage from './pages/WorkReportPage'
import SuppliersPage from './pages/SuppliersPage'
import StatsPage from './pages/StatsPage'
import ProjectsPage from './pages/ProjectsPage'
import StaffRequestsPage from './pages/StaffRequestsPage'
import ServiceManagersPage from './pages/ServiceManagersPage'
import PerformanceReviewPage from './pages/PerformanceReviewPage'
import MissionsPage from './pages/MissionsPage'
import ProcurementPage from './pages/ProcurementPage'
import MapPage from './pages/MapPage'
import MonitorDashboard from './pages/MonitorDashboard'
import TrainingPage from './pages/TrainingPage'
import TrainingManagement from './pages/TrainingManagement'
import VehiclesPage from './pages/VehiclesPage'
import VehicleMissionsPage from './pages/VehicleMissionsPage'
import QualityPage from './pages/QualityPage'
import WorkReportsReview from './pages/WorkReportsReview'
import RequireAdmin from './components/RequireAdmin'
import RequirePermission from './components/RequirePermission'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="employees" element={<Employees />} />
        <Route path="customers" element={<Customers />} />
        <Route path="services" element={<Services />} />
        <Route path="sales" element={<SalesBooking />} />
        <Route path="quality-follow-ups" element={<QualityFollowUpsPage />} />
        <Route path="owner-security" element={<SecurityDashboardPage />} />
        <Route path="coordinator" element={<Coordinator />} />
        <Route path="bookings" element={<BookingsList />} />
        <Route path="my-tasks" element={<MyTasks />} />
        <Route path="my-ranking" element={<MyRanking />} />
        <Route path="my-expenses" element={<MyExpenses />} />
        <Route path="finance" element={<Finance />} />
        <Route path="expenses" element={<ExpensesReview />} />
        <Route path="kpi" element={<KpiPage />} />
        <Route path="complaints" element={<ComplaintsPage />} />
        <Route path="inventory" element={<InventoryPage />} />
        <Route path="my-inventory" element={<MyInventory />} />
        <Route path="permissions" element={<RequireAdmin><PermissionsPage /></RequireAdmin>} />
        <Route path="quotations" element={<QuotationsPage />} />
        <Route path="quotations/new" element={<QuotationNew />} />
        <Route path="quotations/:id/edit" element={<QuotationNew />} />
        <Route path="products" element={<ProductsPage />} />
        <Route path="attendance" element={<AttendancePage />} />
        <Route path="work-reports" element={<WorkReportPage />} />
        <Route path="suppliers" element={<SuppliersPage />} />
        <Route path="stats" element={<StatsPage />} />
        <Route path="projects" element={<ProjectsPage />} />
        <Route path="staff-requests" element={<StaffRequestsPage />} />
        <Route path="service-managers" element={<RequireAdmin><ServiceManagersPage /></RequireAdmin>} />
        <Route path="performance-review" element={<PerformanceReviewPage />} />
        <Route path="missions" element={<MissionsPage />} />
        <Route path="procurement" element={<ProcurementPage />} />
        <Route path="map" element={<MapPage />} />
        <Route path="monitor" element={<MonitorDashboard />} />
        <Route path="gps" element={<GpsDashboard />} />
        <Route path="gps/customers" element={<GpsCustomers />} />
        <Route path="gps/devices" element={<GpsDevices />} />
        <Route path="gps/sims" element={<GpsSims />} />
        <Route path="gps/renewals" element={<GpsRenewals />} />
        <Route path="gps/maintenance" element={<GpsMaintenance />} />
        <Route path="gps/employee" element={<GpsEmployee />} />
        <Route path="gps/purchase" element={<GpsPurchase />} />
        <Route path="gps/requests" element={<GpsRequestsReview />} />
        <Route path="gps/delivery" element={<GpsDelivery />} />
        <Route path="gps/renewal" element={<GpsRenewal />} />
        <Route path="gps/renewals-review" element={<GpsRenewalsReview />} />
        <Route path="gps/maintenance-request" element={<GpsMaintenanceRequestPage />} />
        <Route path="gps/maintenance-review" element={<GpsMaintenanceReview />} />
        <Route path="training" element={<TrainingPage />} />
        <Route path="training-management" element={<RequirePermission permission="content_technician"><TrainingManagement /></RequirePermission>} />
        <Route path="vehicles" element={<RequirePermission permission="vehicle_management"><VehiclesPage /></RequirePermission>} />
        <Route path="vehicle-missions" element={<RequirePermission permission="vehicle_management"><VehicleMissionsPage /></RequirePermission>} />
        <Route path="quality" element={<RequirePermission permission="quality_control"><QualityPage /></RequirePermission>} />
        <Route path="work-reports-review" element={<WorkReportsReview />} />
      </Route>
    </Routes>
  )
}

export default App
