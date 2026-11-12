import { Suspense } from 'react'
import { lazyRetry } from './utils/staleReload'
import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import RequireAdmin from './components/RequireAdmin'
import RequirePermission from './components/RequirePermission'
import { useSession } from './session'
// TrainingPage is left as a regular eager import: Layout.tsx already imports it statically
// (rendered directly, outside the router, for trainee accounts), so lazy-loading it here
// wouldn't split it out of the main bundle anyway — it would just add a redundant chunk boundary.
import TrainingPage from './pages/TrainingPage'

// Route-level pages are code-split with React.lazy — Layout/RequireAdmin/RequirePermission
// stay eager since they render on every route.
const Dashboard = lazyRetry(() => import('./pages/Dashboard'))
const Employees = lazyRetry(() => import('./pages/Employees'))
const WorkSchedulePage = lazyRetry(() => import('./pages/WorkSchedulePage'))
const RemoteHoursPage = lazyRetry(() => import('./pages/RemoteHoursPage'))
const LeavesHub = lazyRetry(() => import('./pages/LeavesHub'))
const Customers = lazyRetry(() => import('./pages/Customers'))
const Services = lazyRetry(() => import('./pages/Services'))
const SalesBooking = lazyRetry(() => import('./pages/SalesBooking'))
const SecurityDashboardPage = lazyRetry(() => import('./pages/SecurityDashboardPage'))
const AssistantConversationsPage = lazyRetry(() => import('./pages/AssistantConversationsPage'))
const SecretaryPage = lazyRetry(() => import('./pages/SecretaryPage'))
const MatrixChatPage = lazyRetry(() => import('./components/MatrixChat'))
const AiAgentsPage = lazyRetry(() => import('./pages/AiAgentsPage'))
const WhatsAppPage = lazyRetry(() => import('./pages/WhatsAppPage'))
const BotReportsPage = lazyRetry(() => import('./pages/BotReportsPage'))
const WhatsAppSandboxPage = lazyRetry(() => import('./pages/WhatsAppSandboxPage'))
const Coordinator = lazyRetry(() => import('./pages/Coordinator'))
const SalesOpportunitiesPage = lazyRetry(() => import('./pages/SalesOpportunitiesPage'))
const WeeklyReportPage = lazyRetry(() => import('./pages/WeeklyReportPage'))
const MatrixDecisionsPage = lazyRetry(() => import('./pages/MatrixDecisionsPage'))
const MatrixMissions = lazyRetry(() => import('./components/MatrixMissions'))
const MatrixEmployeeReportPage = lazyRetry(() => import('./pages/MatrixEmployeeReportPage'))
const BookingsArchive = lazyRetry(() => import('./pages/BookingsArchive'))
const SolarPage = lazyRetry(() => import('./pages/SolarPage'))
const SolarInquiriesPage = lazyRetry(() => import('./pages/SolarInquiriesPage'))
// مراقبة النسخ الاحتياطية — الصفحة نفسها تتحقق من actualRole === 'OWNER'
// والمسار بالباك إند يرجّع 404 لأي حساب ثاني.
const OwnerBackups = lazyRetry(() => import('./pages/OwnerBackups'))
const CoreMatrixPage = lazyRetry(() => import('./pages/CoreMatrixPage'))
const CoreAdminPage = lazyRetry(() => import('./pages/CoreAdminPage'))
const OwnerGrantsPage = lazyRetry(() => import('./pages/OwnerGrantsPage'))
const OwnerIntegrity = lazyRetry(() => import('./pages/OwnerIntegrity'))
const OwnerDeleted = lazyRetry(() => import('./pages/OwnerDeleted'))
const StaffReportsPage = lazyRetry(() => import('./pages/StaffReportsPage'))
const EmployeeDossierPage = lazyRetry(() => import('./pages/EmployeeDossierPage'))
const CommandCodePage = lazyRetry(() => import('./pages/CommandCodePage'))
// حجوزات انجز منها جزء وتحتاج يوم جديد — شاشة إداري الحجوزات
const StageBucketsPage = lazyRetry(() => import('./pages/StageBucketsPage'))
const AiInsightsPage = lazyRetry(() => import('./pages/AiInsightsPage'))
const DuplicateReviewPage = lazyRetry(() => import('./pages/DuplicateReviewPage'))
const AchievementsPage = lazyRetry(() => import('./pages/AchievementsPage'))
const MyAchievementsPage = lazyRetry(() => import('./pages/MyAchievementsPage'))
const ExtraTasksPage = lazyRetry(() => import('./pages/ExtraTasksPage'))
const RolesGuidePage = lazyRetry(() => import('./pages/RolesGuidePage'))
// الطلبات — كتاب رسمي من الموظف للإدارة، متاح لكل الموظفين
const LettersPage = lazyRetry(() => import('./pages/LettersPage'))
const TrainingCenterPage = lazyRetry(() => import('./pages/TrainingCenterPage'))
const MyExamPage = lazyRetry(() => import('./pages/MyExamPage'))
const BookingsHub = lazyRetry(() => import('./pages/BookingsHub'))
const MyTasks = lazyRetry(() => import('./pages/MyTasks'))
// (دازاي 10-25) كادر الديكور: «شغلي اليوم» و«شنو مطلوب مني؟»
const DecorTodayPage = lazyRetry(() => import('./pages/DecorTodayPage'))
const DecorGuidePage = lazyRetry(() => import('./pages/DecorGuidePage'))
const MyPaperworkPage = lazyRetry(() => import('./pages/MyPaperworkPage'))
const MyHandoversPage = lazyRetry(() => import('./pages/MyHandoversPage'))
const MyRanking = lazyRetry(() => import('./pages/MyRanking'))
const MyExtraTasksPage = lazyRetry(() => import('./pages/MyExtraTasksPage'))
const MyWorkPage = lazyRetry(() => import('./pages/MyWorkPage'))
const MyExpenses = lazyRetry(() => import('./pages/MyExpenses'))
const Finance = lazyRetry(() => import('./pages/Finance'))
const ExpensesReview = lazyRetry(() => import('./pages/ExpensesReview'))
const KpiPage = lazyRetry(() => import('./pages/KpiPage'))
const ComplaintsPage = lazyRetry(() => import('./pages/ComplaintsPage'))
const InventoryPage = lazyRetry(() => import('./pages/InventoryPage'))
const ItAssetsPage = lazyRetry(() => import('./pages/ItAssetsPage'))
const MediaPage = lazyRetry(() => import('./pages/MediaPage'))
const DisciplineRecordPage = lazyRetry(() => import('./pages/DisciplineRecordPage'))
const PeerVoicePage = lazyRetry(() => import('./pages/PeerVoicePage'))
const StaffScoreHub = lazyRetry(() => import('./pages/StaffScoreHub'))
const SuppliersHub = lazyRetry(() => import('./pages/SuppliersHub'))
const ProjectPaymentsPage = lazyRetry(() => import('./pages/ProjectPaymentsPage'))
const WorkplaceIssuesPage = lazyRetry(() => import('./pages/WorkplaceIssuesPage'))
const ManagerRequestsPage = lazyRetry(() => import('./pages/ManagerRequestsPage'))
const RevisitsPage = lazyRetry(() => import('./pages/RevisitsPage'))
const ItStatsPage = lazyRetry(() => import('./pages/ItStatsPage'))
const MyInventory = lazyRetry(() => import('./pages/MyInventory'))
const DeviceMaintenancePage = lazyRetry(() => import('./pages/DeviceMaintenancePage'))
const TeamInventoryCheckPage = lazyRetry(() => import('./pages/TeamInventoryCheckPage'))
const PermissionsPage = lazyRetry(() => import('./pages/PermissionsPage'))
const AiSuggestionsPage = lazyRetry(() => import('./pages/AiSuggestionsPage'))
const VoiceTermsPage = lazyRetry(() => import('./pages/VoiceTermsPage'))
const OrgPage = lazyRetry(() => import('./pages/OrgPage'))
const OrgDeskPage = lazyRetry(() => import('./pages/OrgDeskPage'))
const OrgPlansPage = lazyRetry(() => import('./pages/OrgPlansPage'))
const MaintenanceContractsPage = lazyRetry(() => import('./pages/MaintenanceContractsPage'))
const DesignSetsPage = lazyRetry(() => import('./pages/DesignSetsPage'))
const DesignSetPage = lazyRetry(() => import('./pages/DesignSetPage'))
const PlanSetsPage = lazyRetry(() => import('./pages/PlanSetsPage'))
const PlanSetPage = lazyRetry(() => import('./pages/PlanSetPage'))
const ScoreMonitorPage = lazyRetry(() => import('./pages/ScoreMonitorPage'))
const MyOrgCard = lazyRetry(() => import('./components/MyOrgCard'))
const DepartmentsPage = lazyRetry(() => import('./pages/DepartmentsPage'))
const DesignGalleryPage = lazyRetry(() => import('./pages/DesignGalleryPage'))
const QuotationsPage = lazyRetry(() => import('./pages/QuotationsPage'))
const QuotationNew = lazyRetry(() => import('./pages/QuotationNew'))
const ProjectQuotationPage = lazyRetry(() => import('./pages/ProjectQuotationPage'))
const LeaderInvoiceNew = lazyRetry(() => import('./pages/LeaderInvoiceNew'))
const CameraCostPage = lazyRetry(() => import('./pages/CameraCostPage'))
const NetworkCostPage = lazyRetry(() => import('./pages/NetworkCostPage'))
const MonitorInboxPage = lazyRetry(() => import('./pages/MonitorInboxPage'))
const NetworkPricesPage = lazyRetry(() => import('./pages/NetworkPricesPage'))
const PermissionPreview = lazyRetry(() => import('./pages/PermissionPreview'))
const ProjectStatisticsPage = lazyRetry(() => import('./pages/ProjectStatisticsPage'))
const LeaderInvoicesListPage = lazyRetry(() => import('./pages/LeaderInvoicesListPage'))
const ProductsPage = lazyRetry(() => import('./pages/ProductsPage'))
const TechProductsHub = lazyRetry(() => import('./pages/TechProductsHub'))
const TechContentHub = lazyRetry(() => import('./pages/TechContentHub'))
const GpsDashboard = lazyRetry(() => import('./pages/gps/GpsDashboard'))
const GpsCustomers = lazyRetry(() => import('./pages/gps/GpsCustomers'))
const GpsDevices = lazyRetry(() => import('./pages/gps/GpsDevices'))
const GpsSims = lazyRetry(() => import('./pages/gps/GpsSims'))
const GpsFollowUp = lazyRetry(() => import('./pages/gps/GpsFollowUp'))
const RevolvingFundPage = lazyRetry(() => import('./pages/RevolvingFundPage'))
const GpsInstallCostsPage = lazyRetry(() => import('./pages/GpsInstallCostsPage'))
const GpsRenewals = lazyRetry(() => import('./pages/gps/GpsRenewals'))
const GpsMaintenance = lazyRetry(() => import('./pages/gps/GpsMaintenance'))
const GpsEmployee = lazyRetry(() => import('./pages/gps/GpsEmployee'))
const GpsPurchase = lazyRetry(() => import('./pages/gps/GpsPurchase'))
const GpsRequestsReview = lazyRetry(() => import('./pages/gps/GpsRequestsReview'))
const GpsRequestsHub = lazyRetry(() => import('./pages/gps/GpsRequestsHub'))
const GpsDelivery = lazyRetry(() => import('./pages/gps/GpsDelivery'))
const GpsRenewal = lazyRetry(() => import('./pages/gps/GpsRenewal'))
const GpsRenewalsReview = lazyRetry(() => import('./pages/gps/GpsRenewalsReview'))
const GpsMaintenanceRequestPage = lazyRetry(() => import('./pages/gps/GpsMaintenanceRequestPage'))
const GpsMaintenanceReview = lazyRetry(() => import('./pages/gps/GpsMaintenanceReview'))
const AttendancePage = lazyRetry(() => import('./pages/AttendancePage'))

const WorkReportPage = lazyRetry(() => import('./pages/WorkReportPage'))
const ProjectsPage = lazyRetry(() => import('./pages/ProjectsPage'))
const VipCustomersPage = lazyRetry(() => import('./pages/VipCustomersPage'))
const PrivacyPolicyPage = lazyRetry(() => import('./pages/PrivacyPolicyPage'))
const ProjectWorkTypesSettingsPage = lazyRetry(() => import('./pages/ProjectWorkTypesSettingsPage'))
const ChecklistsPage = lazyRetry(() => import('./pages/ChecklistsPage'))
const TechShowcasePage = lazyRetry(() => import('./pages/TechShowcasePage'))
const ExhibitionsPage = lazyRetry(() => import('./pages/ExhibitionsPage'))
const ProductRequestsPage = lazyRetry(() => import('./pages/ProductRequestsPage'))
const ServiceStudiesPage = lazyRetry(() => import('./pages/ServiceStudiesPage'))
const DesignFormsListPage = lazyRetry(() => import('./pages/DesignFormsListPage'))
const DesignFormQuickAddPage = lazyRetry(() => import('./pages/DesignFormQuickAddPage'))
const DesignFormBuilderPage = lazyRetry(() => import('./pages/DesignFormBuilderPage'))
const DesignFormSubmissionsPage = lazyRetry(() => import('./pages/DesignFormSubmissionsPage'))
const PublicDesignFormPage = lazyRetry(() => import('./pages/PublicDesignFormPage'))
const StaffRequestsPage = lazyRetry(() => import('./pages/StaffRequestsPage'))
const ServiceManagersPage = lazyRetry(() => import('./pages/ServiceManagersPage'))
const PerformanceReviewPage = lazyRetry(() => import('./pages/PerformanceReviewPage'))
const MissionsPage = lazyRetry(() => import('./pages/MissionsPage'))
const ProcurementPage = lazyRetry(() => import('./pages/ProcurementPage'))
const MapPage = lazyRetry(() => import('./pages/MapPage'))
const FieldMapPage = lazyRetry(() => import('./pages/FieldMapPage'))
const LateAuditPage = lazyRetry(() => import('./pages/LateAuditPage'))
const MatrixSelfCheckPage = lazyRetry(() => import('./pages/MatrixSelfCheckPage'))
const DrillPage = lazyRetry(() => import('./pages/DrillPage'))
const StoreToolsPage = lazyRetry(() => import('./pages/StoreToolsPage'))
const AccountantDeskPage = lazyRetry(() => import('./pages/AccountantDeskPage'))
const QualityDeskPage = lazyRetry(() => import('./pages/QualityDeskPage'))
const MonitorCrewBookingsPage = lazyRetry(() => import('./pages/MonitorCrewBookingsPage'))
const TrainingManagement = lazyRetry(() => import('./pages/TrainingManagement'))
const VehiclesPage = lazyRetry(() => import('./pages/VehiclesPage'))
const FleetDashboardPage = lazyRetry(() => import('./pages/FleetDashboardPage'))
const VehicleBookingsPage = lazyRetry(() => import('./pages/VehicleBookingsPage'))
const VehicleMissionsPage = lazyRetry(() => import('./pages/VehicleMissionsPage'))
const QualityPage = lazyRetry(() => import('./pages/QualityPage'))
const WorkReportsReview = lazyRetry(() => import('./pages/WorkReportsReview'))
const ReportsDeskPage = lazyRetry(() => import('./pages/ReportsDeskPage'))
const DecorRatesPage = lazyRetry(() => import('./pages/DecorRatesPage'))
const VehicleDriversPage = lazyRetry(() => import('./pages/VehicleDriversPage'))
const StatsManagementPage = lazyRetry(() => import('./pages/StatsManagementPage'))
const CustomerAppAdminPage = lazyRetry(() => import('./pages/CustomerAppAdminPage'))
const CreditsPage = lazyRetry(() => import('./pages/CreditsPage'))
const BookingDeleteRequestsPage = lazyRetry(() => import('./pages/BookingDeleteRequestsPage'))
const AuditIssuesPage = lazyRetry(() => import('./pages/AuditIssuesPage'))
const DailyAuditPage = lazyRetry(() => import('./pages/DailyAuditPage'))
const FinanceDeskPage = lazyRetry(() => import('./pages/FinanceDeskPage'))
const CostCalcPage = lazyRetry(() => import('./pages/CostCalcPage'))
const VehicleTrackingPage = lazyRetry(() => import('./pages/VehicleTrackingPage'))
const CarWashPage = lazyRetry(() => import('./pages/CarWashPage'))
const TechBookingsPage = lazyRetry(() => import('./pages/TechBookingsPage'))
const AnnouncementsPage = lazyRetry(() => import('./pages/AnnouncementsPage'))
const MonitorDeskPage = lazyRetry(() => import('./pages/MonitorDeskPage'))
const StaffManagementDeskPage = lazyRetry(() => import('./pages/StaffManagementDeskPage'))

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
          <Route index element={<Home />} />
          <Route path="car-wash" element={<CarWashPage />} />
          <Route path="tech-bookings" element={<RequireAdmin><TechBookingsPage /></RequireAdmin>} />
          <Route path="employees" element={<Employees />} />
          <Route path="work-schedule" element={<WorkSchedulePage />} />
          <Route path="remote-hours" element={<RemoteHoursPage />} />
          <Route path="staff-leaves" element={<LeavesHub />} />
          <Route path="customers" element={<Customers />} />
          <Route path="services" element={<Services />} />
          <Route path="sales" element={<SalesBooking />} />
          {/* ٧و (11-05): نفس «مكتب الجودة» — المسار القديم يبقى بحارسه (روابط المراقب والإشعارات) */}
          <Route path="quality-follow-ups" element={<QualityDeskPage />} />
          <Route path="owner-security" element={<SecurityDashboardPage />} />
          <Route path="owner-backups" element={<OwnerBackups />} />
          {/* 🏴‍☠️ ون بيس (11-10): مصفوفة الصلاحيات ولوحة الأدمن — للمالك */}
          <Route path="core-matrix" element={<CoreMatrixPage />} />
          <Route path="core-admin" element={<CoreAdminPage />} />
          <Route path="owner-grants" element={<OwnerGrantsPage />} />
          <Route path="owner-integrity" element={<OwnerIntegrity />} />
          <Route path="owner-deleted" element={<OwnerDeleted />} />
          <Route path="staff-reports" element={<StaffReportsPage />} />
          <Route path="employee-dossier" element={<EmployeeDossierPage />} />
          <Route path="command-code" element={<CommandCodePage />} />
          {/* ⚠️ مكتب المراقب **مسار جديد** والشاشات القديمة تبقى
              بمساراتها: اكو روابط محفوظة بمتصفحات الموظفين، وحذف
              مسار يعني صفحة بيضاء بلا تفسير. */}
          <Route path="monitor-desk" element={<MonitorDeskPage />} />
          <Route path="staff-management-desk" element={<StaffManagementDeskPage />} />
          <Route path="assistant-conversations" element={<AssistantConversationsPage />} />
          {/* السكرتير تطبيق مستقل بلا القائمة؛ Layout يمرّر Outlet وحده بهذا المسار. */}
          <Route path="secretary" element={<SecretaryPage />} />
          <Route path="ai-agents" element={<AiAgentsPage />} />
          <Route path="ai-suggestions" element={<RequireAdmin><AiSuggestionsPage /></RequireAdmin>} />
          <Route path="voice-terms" element={<RequireAdmin><VoiceTermsPage /></RequireAdmin>} />
          <Route path="org" element={<RequireAdmin><OrgPage /></RequireAdmin>} />
          <Route path="org/assign" element={<RequireAdmin><OrgPage view="assign" /></RequireAdmin>} />
          <Route path="org/directors-ranking" element={<RequireAdmin><StaffScoreHub /></RequireAdmin>} />
          <Route path="org/score-monitor" element={<RequireAdmin><ScoreMonitorPage /></RequireAdmin>} />
          {/* (ع) 11-08: 🏗️ صيت التصاميم — الصلاحيات بالخادم (يشوف المصمم المكلّف صيته بس) */}
          <Route path="design-sets" element={<DesignSetsPage />} />
          <Route path="design-sets/:id" element={<DesignSetPage />} />
          {/* (ك) 11-10: 📐 سيت المخططات — الخادم يحصرها (نفس صيت التصاميم) */}
          <Route path="plan-sets" element={<PlanSetsPage />} />
          <Route path="plan-sets/:id" element={<PlanSetPage />} />
          <Route path="org/plans" element={<RequireAdmin><OrgPlansPage /></RequireAdmin>} />
          <Route path="org/desk" element={<OrgDeskPage />} />
          <Route path="org/desk/:divisionId" element={<OrgDeskPage />} />
          <Route path="my-org" element={<MyOrgCard />} />
          <Route path="whatsapp-conversations" element={<WhatsAppPage />} />
          <Route path="bot-reports" element={<BotReportsPage />} />
          <Route path="whatsapp-sandbox" element={<WhatsAppSandboxPage />} />
          <Route path="coordinator" element={<Coordinator />} />
          <Route path="maintenance-contracts" element={<RequirePermission permission="coordinator"><MaintenanceContractsPage /></RequirePermission>} />
          {/* فرص البيع: الصفحة نفسها تفحص ADMIN/OWNER أو sales_booking (نفس الخادم) */}
          <Route path="sales-opportunities" element={<SalesOpportunitiesPage />} />
          {/* التقرير الأسبوعي: ADMIN/OWNER بس (نفس الخادم) */}
          <Route path="weekly-report" element={<RequireAdmin><WeeklyReportPage /></RequireAdmin>} />
          <Route path="manager-requests" element={<RequireAdmin><ManagerRequestsPage /></RequireAdmin>} />
          <Route path="project-payments" element={<RequirePermission permission="finance" anyOf={['finance_audit', 'project_management', 'monitoring']} roles={['FINANCE', 'PROJECT_MANAGER', 'MONITOR']}><ProjectPaymentsPage /></RequirePermission>} />
          <Route path="tech-suppliers" element={<SuppliersHub />} />
          <Route path="employee-indicators" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><StaffScoreHub /></RequirePermission>} />
          <Route path="staff-score" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><StaffScoreHub /></RequirePermission>} />
          <Route path="peer-voice" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><PeerVoicePage /></RequirePermission>} />
          <Route path="revisits" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><RevisitsPage /></RequirePermission>} />
          <Route path="workplace-issues" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><WorkplaceIssuesPage /></RequirePermission>} />
          <Route path="discipline-record" element={<RequireAdmin><DisciplineRecordPage /></RequireAdmin>} />
          <Route path="matrix/decisions" element={<RequireAdmin><MatrixDecisionsPage /></RequireAdmin>} />
          {/* المراقب كمان: شغله متابعة الموظف — نفس حارس GET /api/ai/employee-report. */}
          <Route path="matrix/employee/:id" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><MatrixEmployeeReportPage /></RequirePermission>} />
          <Route path="bookings-archive" element={<BookingsArchive />} />
          {/* ═══ «الحجوزات المؤجلة» انمرجت ═══
              كانت شاشة مستقلة تعرض **نفس** حجوزات سلّة «مؤجّلة» بشاشة
              «ما وصلت للتنفيذ» — نفس الحجز بمكانين، والإداري يشتغل على
              وحدة وينسى الثانية. صارت سلّة وحدة، وميزتها (تحديد موعد
              جديد) انتقلت معها.
              ⚠️ المسار يبقى موجود ويحوّل: أكو روابط محفوظة بمتصفحات
              الموظفين، وحذفه يعني صفحة بيضاء بلا تفسير. */}
          <Route path="postponed-bookings" element={<Navigate to="/bookings" replace />} />
          {/* انمرجت بشاشة الحجوزات — المسار يحوّل حتى الروابط
              المحفوظة ما تطلع صفحة بيضاء. */}
          <Route path="partial-bookings" element={<Navigate to="/bookings" replace />} />
          <Route path="stage-buckets" element={<StageBucketsPage />} />
          <Route path="ai-insights" element={<AiInsightsPage />} />
          <Route path="duplicate-review" element={<DuplicateReviewPage />} />
          <Route path="achievements" element={<AchievementsPage />} />
          <Route path="my-achievements" element={<MyAchievementsPage />} />
          <Route path="extra-tasks" element={<ExtraTasksPage />} />
          <Route path="roles-guide" element={<RolesGuidePage />} />
          <Route path="letters" element={<LettersPage />} />
          <Route path="solar" element={<SolarPage />} />
          <Route path="solar-inquiries" element={<RequirePermission permission="solar_inquiry" anyOf={['solar_inquiry_tech', 'quality_control']} roles={['QUALITY_ENGINEER']}><SolarInquiriesPage /></RequirePermission>} />
          <Route path="training-programs" element={<TrainingCenterPage />} />
          <Route path="training-center" element={<TrainingCenterPage />} />
          <Route path="my-exam" element={<MyExamPage />} />
          {/* «الحجوزات» صارت مدخل واحد بثلاثة خيارات: حجوزات اليوم ·
              تنسيق الحجوزات · حجز جديد. والمسارات القديمة باقية شغّالة
              لمن عنده رابط محفوظ أو إشعار قديم يوديه لها. */}
          <Route path="bookings" element={<BookingsHub />} />
          <Route path="my-tasks" element={<MyTasks />} />
          <Route path="decor-today" element={<DecorTodayPage />} />
          <Route path="decor-guide" element={<DecorGuidePage />} />
          <Route path="my-paperwork" element={<MyPaperworkPage />} />
          <Route path="my-handovers" element={<MyHandoversPage />} />
          <Route path="my-ranking" element={<MyRanking />} />
          {/* مهامي الإضافية — بلا حارس: هاي مهام الموظف نفسه */}
          <Route path="my-extra-tasks" element={<MyExtraTasksPage />} />
          <Route path="my-work" element={<MyWorkPage />} />
          <Route path="my-expenses" element={<MyExpenses />} />
          <Route path="finance" element={<Finance />} />
          <Route path="expenses" element={<ExpensesReview />} />
          <Route path="kpi" element={<KpiPage />} />
          <Route path="complaints" element={<ComplaintsPage />} />
          {/* متابعة جرد كل الفنيين = شاشة إشراف. الفني يسوّي جرده من
              my-inventory؛ هاي تحتاج صلاحية «جرد الأدوات» حتى لو كتب
              رابطها بيده — نفس القيد المطبّق بالسيرفر. */}
          <Route path="inventory" element={<RequirePermission permission="inventory"><InventoryPage /></RequirePermission>} />
          <Route path="it-assets" element={<RequirePermission permission="it_assets"><ItAssetsPage /></RequirePermission>} />
          <Route path="media" element={<RequirePermission permission="media" anyOf={['project_management']} roles={['MEDIA', 'PUBLIC_RELATIONS']}><MediaPage /></RequirePermission>} />
          <Route path="it-stats" element={<RequirePermission permission="it_stats" anyOf={['it_assets']}><ItStatsPage /></RequirePermission>} />
          <Route path="my-inventory" element={<MyInventory />} />
          <Route path="permissions" element={<RequireAdmin><PermissionsPage /></RequireAdmin>} />
          <Route path="access" element={<Navigate to="/permissions" replace />} />
          <Route path="customer-app" element={<RequireAdmin><CustomerAppAdminPage /></RequireAdmin>} />
          {/* سجل الأقسام ومسؤوليها — «التعديل فقط لمالك ومدير النظام
              حالياً، وخليها صلاحية». الحارس بالسيرفر هو الفاصل؛ هذا
              يمنع فتح شاشة ما تشتغل أزرارها. */}
          <Route path="departments" element={<RequirePermission permission="departments_manage"><DepartmentsPage /></RequirePermission>} />
          <Route path="quotations" element={<QuotationsPage />} />
          <Route path="quotations/new" element={<QuotationNew />} />
          <Route path="quotations/:id/edit" element={<QuotationNew />} />
          <Route path="quotations/project/new" element={<ProjectQuotationPage />} />
          <Route path="quotations/project/:id" element={<ProjectQuotationPage />} />
          <Route path="leader-invoices" element={<LeaderInvoicesListPage />} />
          <Route path="leader-invoices/new" element={<LeaderInvoiceNew />} />
          <Route path="matrix-chat" element={<MatrixChatPage />} />
          <Route path="handover-invoices" element={<LeaderInvoiceNew initialMode="booking" handoverOnly />} />
          <Route path="camera-cost" element={<CameraCostPage />} />
          <Route path="network-cost" element={<NetworkCostPage />} />
          <Route path="monitor-inbox" element={<MonitorInboxPage />} />
          <Route path="matrix-missions" element={<RequirePermission permission="monitoring" anyOf={['auditing']} roles={['MONITOR']}><MatrixMissions /></RequirePermission>} />
          <Route path="network-prices" element={<NetworkPricesPage />} />
          <Route path="permission-preview" element={<PermissionPreview />} />
          {/* المشاريع الموجّهة للموظف — نفس واجهة إدارة المشاريع بس على مشاريعه */}
          <Route path="my-projects" element={<ProjectsPage mode="delegated" />} />
          <Route path="project-statistics" element={<ProjectStatisticsPage />} />
          {/* ⚠️ مدخلان موحّدان — والمسارات المفردة تبقى شغّالة فما ينكسر
              رابط محفوظ ولا مدزوز. */}
          <Route path="tech-products" element={<RequirePermission permission="content_technician" anyOf={['unit_technicians']}><TechProductsHub /></RequirePermission>} />
          <Route path="tech-content" element={<RequirePermission permission="content_technician" anyOf={['unit_technicians']}><TechContentHub /></RequirePermission>} />
          <Route path="products" element={<ProductsPage />} />
          <Route path="attendance" element={<AttendancePage />} />
          <Route path="work-reports" element={<WorkReportPage />} />
          <Route path="suppliers" element={<SuppliersHub />} />
          <Route path="stats" element={<RequirePermission permission="employee_stats"><StatsManagementPage /></RequirePermission>} />
          <Route path="projects" element={<ProjectsPage />} />
          {/* الشخصيات المهمة: مدير النظام، أو أي إداري ينمنح صلاحية الإضافة اليدوية */}
          <Route path="vip-customers" element={<RequirePermission permission="vip_manual_add"><VipCustomersPage /></RequirePermission>} />
          {/* سياسة الخصوصية مفتوحة لكل موظف — يقراها من الإعدادات */}
          <Route path="privacy-policy" element={<PrivacyPolicyPage />} />
          <Route path="project-work-types" element={<ProjectWorkTypesSettingsPage />} />
          <Route path="checklists" element={<ChecklistsPage />} />
          <Route path="tech-showcase" element={<TechShowcasePage />} />
          <Route path="exhibitions" element={<RequirePermission permission="content_technician" anyOf={['unit_technicians']}><ExhibitionsPage /></RequirePermission>} />
          <Route path="product-requests" element={<RequirePermission permission="content_technician" anyOf={['unit_technicians']}><ProductRequestsPage /></RequirePermission>} />
          <Route path="service-studies" element={<RequirePermission permission="content_technician" anyOf={['unit_technicians']} roles={['SERVICE_DEVELOPER']}><ServiceStudiesPage /></RequirePermission>} />
          {/* ⚠️ چانن كلهن `RequireAdmin` — يعني حتى بعد ما ينفتح
              الخادم للمصممة، الواجهة تحجبها قبل ما يوصل النداء.
              صارت الصلاحية نفسها الي يفحصها الخادم. */}
          <Route path="unit-design" element={<Navigate to="/design-forms" replace />} />
          <Route path="design-forms" element={<RequirePermission permission="design_forms"><DesignFormsListPage /></RequirePermission>} />
          <Route path="design-forms/quick-add" element={<RequirePermission permission="design_forms"><DesignFormQuickAddPage /></RequirePermission>} />
          <Route path="design-forms/:formId" element={<RequirePermission permission="design_forms"><DesignFormBuilderPage /></RequirePermission>} />
          <Route path="design-forms/:formId/submissions" element={<RequirePermission permission="design_forms"><DesignFormSubmissionsPage /></RequirePermission>} />
          <Route path="design-gallery" element={<RequirePermission permission="design_gallery" anyOf={['unit_design']}><DesignGalleryPage /></RequirePermission>} />
          <Route path="unit-pr" element={<Navigate to="/media" replace />} />
          <Route path="staff-requests" element={<StaffRequestsPage />} />
          <Route path="service-managers" element={<RequirePermission permission="service_managers"><ServiceManagersPage /></RequirePermission>} />
          <Route path="employee-stats" element={<RequirePermission permission="employee_stats"><StaffScoreHub /></RequirePermission>} />
          <Route path="staff-numbers" element={<RequirePermission permission="employee_stats"><StaffScoreHub /></RequirePermission>} />
          <Route path="stats-management" element={<RequirePermission permission="employee_stats"><StatsManagementPage /></RequirePermission>} />
          <Route path="performance-review" element={<PerformanceReviewPage />} />
          <Route path="missions" element={<MissionsPage />} />
          <Route path="procurement" element={<ProcurementPage />} />
          <Route path="map" element={<MapPage />} />
          <Route path="monitor" element={<MonitorDeskPage />} />
          <Route path="field-map" element={<FieldMapPage />} />
          <Route path="late-audit" element={<LateAuditPage />} />
          <Route path="matrix-selfcheck" element={<MatrixSelfCheckPage />} />
          <Route path="drill/:key" element={<DrillPage />} />
          <Route path="store-tools" element={<StoreToolsPage />} />
          <Route path="accountant-desk" element={<AccountantDeskPage />} />
          <Route path="quality-desk" element={<QualityDeskPage />} />
          <Route path="supplier-relations" element={<SuppliersHub />} />
          <Route path="gps" element={<GpsDashboard />} />
          <Route path="gps/customers" element={<GpsCustomers />} />
          <Route path="gps/devices" element={<GpsDevices />} />
          <Route path="gps/sims" element={<GpsSims />} />
          <Route path="gps/follow-up" element={<GpsFollowUp />} />
          <Route path="revolving-fund" element={<RevolvingFundPage />} />
          <Route path="leaves" element={<LeavesHub />} />
          <Route path="booking-delete-requests" element={<BookingDeleteRequestsPage />} />
          <Route path="audit-issues" element={<AuditIssuesPage />} />
          <Route path="daily-audit" element={<DailyAuditPage />} />
          <Route path="credits" element={<CreditsPage />} />
          <Route path="finance-desk" element={<FinanceDeskPage />} />
          <Route path="cost-calc" element={<CostCalcPage />} />
          <Route path="vehicle-tracking" element={<RequirePermission permission="vehicle_management"><VehicleTrackingPage /></RequirePermission>} />
          <Route path="announcements" element={<RequireAdmin><AnnouncementsPage /></RequireAdmin>} />
          <Route path="gps-install-costs" element={<GpsInstallCostsPage />} />
          <Route path="gps/renewals" element={<GpsRenewals />} />
          <Route path="gps/maintenance" element={<GpsMaintenance />} />
          <Route path="gps/employee" element={<GpsEmployee />} />
          <Route path="gps/purchase" element={<GpsPurchase />} />
          {/* ⚠️ المسارات القديمة الثلاثة تبقى شغّالة: أي رابط محفوظ أو
              مدزوز بالتلغرام ما ينكسر. الجديد يلمّهن بمدخل واحد. */}
          <Route path="gps/requests-hub" element={<GpsRequestsHub />} />
          <Route path="gps/requests" element={<GpsRequestsReview />} />
          <Route path="gps/delivery" element={<GpsDelivery />} />
          <Route path="gps/renewal" element={<GpsRenewal />} />
          <Route path="gps/renewals-review" element={<GpsRenewalsReview />} />
          <Route path="gps/maintenance-request" element={<GpsMaintenanceRequestPage />} />
          <Route path="gps/maintenance-review" element={<GpsMaintenanceReview />} />
          <Route path="training" element={<TrainingPage />} />
          <Route path="training-management" element={<RequirePermission permission="content_technician" anyOf={['unit_technicians']}><TrainingManagement /></RequirePermission>} />
          <Route path="vehicles" element={<RequirePermission permission="vehicle_management"><VehiclesPage /></RequirePermission>} />
          <Route path="fleet-dashboard" element={<RequirePermission permission="vehicle_management"><FleetDashboardPage /></RequirePermission>} />
          {/* الحجز والمهمات: كل كتاباتهم ورا requireVehicleMgmt بالباك */}
          <Route path="vehicle-bookings" element={<RequirePermission permission="vehicle_management"><VehicleBookingsPage /></RequirePermission>} />
          <Route path="vehicle-missions" element={<RequirePermission permission="vehicle_management"><VehicleMissionsPage /></RequirePermission>} />
          <Route path="quality" element={<RequirePermission permission="quality_control"><QualityPage /></RequirePermission>} />
          <Route path="crew-bookings-audit" element={<RequirePermission permission="crew_management"><MonitorCrewBookingsPage /></RequirePermission>} />
          <Route path="work-reports-review" element={<WorkReportsReview />} />
          <Route path="reports-desk" element={<ReportsDeskPage />} />
          <Route path="vehicle-drivers" element={<VehicleDriversPage />} />
          <Route path="decor-rates" element={<RequireAdmin><DecorRatesPage /></RequireAdmin>} />
          <Route path="device-maintenance" element={<DeviceMaintenancePage />} />
          <Route path="team-inventory" element={<TeamInventoryCheckPage />} />
        </Route>
      </Routes>
    </Suspense>
  )
}

export default App

// عامل الغسل صفحته الرئيسية غسل السيارات — اللوحة العامة تنادي مسارات مو إله.
function Home() {
  const { employee } = useSession()
  return employee?.role === 'CAR_WASHER' ? <Navigate to="/car-wash" replace /> : <Dashboard />
}
