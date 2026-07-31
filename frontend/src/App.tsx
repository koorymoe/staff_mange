import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import RequireAdmin from './components/RequireAdmin'
import RequirePermission from './components/RequirePermission'
// TrainingPage is left as a regular eager import: Layout.tsx already imports it statically
// (rendered directly, outside the router, for trainee accounts), so lazy-loading it here
// wouldn't split it out of the main bundle anyway — it would just add a redundant chunk boundary.
import TrainingPage from './pages/TrainingPage'

// Route-level pages are code-split with React.lazy — Layout/RequireAdmin/RequirePermission
// stay eager since they render on every route.
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Employees = lazy(() => import('./pages/Employees'))
const Customers = lazy(() => import('./pages/Customers'))
const Services = lazy(() => import('./pages/Services'))
const SalesBooking = lazy(() => import('./pages/SalesBooking'))
const QualityFollowUpsPage = lazy(() => import('./pages/QualityFollowUpsPage'))
const SecurityDashboardPage = lazy(() => import('./pages/SecurityDashboardPage'))
const AssistantConversationsPage = lazy(() => import('./pages/AssistantConversationsPage'))
const Coordinator = lazy(() => import('./pages/Coordinator'))
const BookingsList = lazy(() => import('./pages/BookingsList'))
const MyTasks = lazy(() => import('./pages/MyTasks'))
const MyRanking = lazy(() => import('./pages/MyRanking'))
const MyExpenses = lazy(() => import('./pages/MyExpenses'))
const Finance = lazy(() => import('./pages/Finance'))
const ExpensesReview = lazy(() => import('./pages/ExpensesReview'))
const KpiPage = lazy(() => import('./pages/KpiPage'))
const ComplaintsPage = lazy(() => import('./pages/ComplaintsPage'))
const InventoryPage = lazy(() => import('./pages/InventoryPage'))
const MyInventory = lazy(() => import('./pages/MyInventory'))
const DeviceMaintenancePage = lazy(() => import('./pages/DeviceMaintenancePage'))
const TeamInventoryCheckPage = lazy(() => import('./pages/TeamInventoryCheckPage'))
const PermissionsPage = lazy(() => import('./pages/PermissionsPage'))
const QuotationsPage = lazy(() => import('./pages/QuotationsPage'))
const QuotationNew = lazy(() => import('./pages/QuotationNew'))
const LeaderInvoiceNew = lazy(() => import('./pages/LeaderInvoiceNew'))
const LeaderInvoicesListPage = lazy(() => import('./pages/LeaderInvoicesListPage'))
const ProductsPage = lazy(() => import('./pages/ProductsPage'))
const GpsDashboard = lazy(() => import('./pages/gps/GpsDashboard'))
const GpsCustomers = lazy(() => import('./pages/gps/GpsCustomers'))
const GpsDevices = lazy(() => import('./pages/gps/GpsDevices'))
const GpsSims = lazy(() => import('./pages/gps/GpsSims'))
const GpsRenewals = lazy(() => import('./pages/gps/GpsRenewals'))
const GpsMaintenance = lazy(() => import('./pages/gps/GpsMaintenance'))
const GpsEmployee = lazy(() => import('./pages/gps/GpsEmployee'))
const GpsPurchase = lazy(() => import('./pages/gps/GpsPurchase'))
const GpsRequestsReview = lazy(() => import('./pages/gps/GpsRequestsReview'))
const GpsDelivery = lazy(() => import('./pages/gps/GpsDelivery'))
const GpsRenewal = lazy(() => import('./pages/gps/GpsRenewal'))
const GpsRenewalsReview = lazy(() => import('./pages/gps/GpsRenewalsReview'))
const GpsMaintenanceRequestPage = lazy(() => import('./pages/gps/GpsMaintenanceRequestPage'))
const GpsMaintenanceReview = lazy(() => import('./pages/gps/GpsMaintenanceReview'))
const AttendancePage = lazy(() => import('./pages/AttendancePage'))

const WorkReportPage = lazy(() => import('./pages/WorkReportPage'))
const SuppliersPage = lazy(() => import('./pages/SuppliersPage'))
const StatsPage = lazy(() => import('./pages/StatsPage'))
const ProjectsPage = lazy(() => import('./pages/ProjectsPage'))
const VipCustomersPage = lazy(() => import('./pages/VipCustomersPage'))
const ProjectWorkTypesSettingsPage = lazy(() => import('./pages/ProjectWorkTypesSettingsPage'))
const ChecklistsPage = lazy(() => import('./pages/ChecklistsPage'))
const TechShowcasePage = lazy(() => import('./pages/TechShowcasePage'))
const ExhibitionsPage = lazy(() => import('./pages/ExhibitionsPage'))
const ProductRequestsPage = lazy(() => import('./pages/ProductRequestsPage'))
const ServiceStudiesPage = lazy(() => import('./pages/ServiceStudiesPage'))
const ComingSoonUnit = lazy(() => import('./pages/ComingSoonUnit'))
const DesignFormsListPage = lazy(() => import('./pages/DesignFormsListPage'))
const DesignFormQuickAddPage = lazy(() => import('./pages/DesignFormQuickAddPage'))
const DesignFormBuilderPage = lazy(() => import('./pages/DesignFormBuilderPage'))
const DesignFormSubmissionsPage = lazy(() => import('./pages/DesignFormSubmissionsPage'))
const PublicDesignFormPage = lazy(() => import('./pages/PublicDesignFormPage'))
const StaffRequestsPage = lazy(() => import('./pages/StaffRequestsPage'))
const ServiceManagersPage = lazy(() => import('./pages/ServiceManagersPage'))
const PerformanceReviewPage = lazy(() => import('./pages/PerformanceReviewPage'))
const MissionsPage = lazy(() => import('./pages/MissionsPage'))
const ProcurementPage = lazy(() => import('./pages/ProcurementPage'))
const MapPage = lazy(() => import('./pages/MapPage'))
const MonitorDashboard = lazy(() => import('./pages/MonitorDashboard'))
const MonitorCrewBookingsPage = lazy(() => import('./pages/MonitorCrewBookingsPage'))
const TrainingManagement = lazy(() => import('./pages/TrainingManagement'))
const VehiclesPage = lazy(() => import('./pages/VehiclesPage'))
const FleetDashboardPage = lazy(() => import('./pages/FleetDashboardPage'))
const QualityPage = lazy(() => import('./pages/QualityPage'))
const WorkReportsReview = lazy(() => import('./pages/WorkReportsReview'))
const EmployeeMonthlyStatsPage = lazy(() => import('./pages/EmployeeMonthlyStatsPage'))
const StatsManagementPage = lazy(() => import('./pages/StatsManagementPage'))

function RouteLoading() {
  return (
    <div className="flex items-center justify-center p-12">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" />
    </div>
  )
}

function App() {
  return (
    <Suspense fallback={<RouteLoading />}>
      <Routes>
        <Route path="design-forms/public/:token" element={<PublicDesignFormPage />} />
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="employees" element={<Employees />} />
          <Route path="customers" element={<Customers />} />
          <Route path="services" element={<Services />} />
          <Route path="sales" element={<SalesBooking />} />
          <Route path="quality-follow-ups" element={<QualityFollowUpsPage />} />
          <Route path="owner-security" element={<SecurityDashboardPage />} />
          <Route path="assistant-conversations" element={<AssistantConversationsPage />} />
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
          <Route path="leader-invoices" element={<LeaderInvoicesListPage />} />
          <Route path="leader-invoices/new" element={<LeaderInvoiceNew />} />
          <Route path="products" element={<ProductsPage />} />
          <Route path="attendance" element={<AttendancePage />} />
          <Route path="work-reports" element={<WorkReportPage />} />
          <Route path="suppliers" element={<SuppliersPage />} />
          <Route path="stats" element={<StatsPage />} />
          <Route path="projects" element={<ProjectsPage />} />
          <Route path="vip-customers" element={<RequireAdmin><VipCustomersPage /></RequireAdmin>} />
          <Route path="project-work-types" element={<ProjectWorkTypesSettingsPage />} />
          <Route path="checklists" element={<ChecklistsPage />} />
          <Route path="tech-showcase" element={<TechShowcasePage />} />
          <Route path="exhibitions" element={<RequirePermission permission="content_technician"><ExhibitionsPage /></RequirePermission>} />
          <Route path="product-requests" element={<RequirePermission permission="content_technician"><ProductRequestsPage /></RequirePermission>} />
          <Route path="service-studies" element={<RequirePermission permission="content_technician"><ServiceStudiesPage /></RequirePermission>} />
          <Route path="unit-design" element={<RequireAdmin><DesignFormsListPage /></RequireAdmin>} />
          <Route path="design-forms" element={<RequireAdmin><DesignFormsListPage /></RequireAdmin>} />
          <Route path="design-forms/quick-add" element={<RequireAdmin><DesignFormQuickAddPage /></RequireAdmin>} />
          <Route path="design-forms/:formId" element={<RequireAdmin><DesignFormBuilderPage /></RequireAdmin>} />
          <Route path="design-forms/:formId/submissions" element={<RequireAdmin><DesignFormSubmissionsPage /></RequireAdmin>} />
          <Route path="unit-pr" element={<ComingSoonUnit title="وحدة الإعلام والعلاقات العامة" />} />
          <Route path="staff-requests" element={<StaffRequestsPage />} />
          <Route path="service-managers" element={<RequireAdmin><ServiceManagersPage /></RequireAdmin>} />
          <Route path="employee-stats" element={<RequireAdmin><EmployeeMonthlyStatsPage /></RequireAdmin>} />
          <Route path="stats-management" element={<RequireAdmin><StatsManagementPage /></RequireAdmin>} />
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
          <Route path="fleet-dashboard" element={<RequirePermission permission="vehicle_management"><FleetDashboardPage /></RequirePermission>} />
          <Route path="quality" element={<RequirePermission permission="quality_control"><QualityPage /></RequirePermission>} />
          <Route path="crew-bookings-audit" element={<RequirePermission permission="crew_management"><MonitorCrewBookingsPage /></RequirePermission>} />
          <Route path="work-reports-review" element={<WorkReportsReview />} />
          <Route path="device-maintenance" element={<DeviceMaintenancePage />} />
          <Route path="team-inventory" element={<TeamInventoryCheckPage />} />
        </Route>
      </Routes>
    </Suspense>
  )
}

export default App
