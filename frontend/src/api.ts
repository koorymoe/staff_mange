const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

export interface AttendanceRecord {
  id: string
  employeeId: string
  checkIn: string
  checkOut: string | null
  date: string
  employee: { id: string; name: string } | null
}

export interface DailyAttendance {
  date: string
  sessions: AttendanceRecord[]
  firstCheckIn: string
  lastCheckOut: string | null
  stillOpen: boolean
  totalMinutes: number
}

export interface MonthlyAttendanceReport {
  employeeId: string
  month: string
  days: DailyAttendance[]
  daysPresent: number
  totalMinutes: number
}

export interface OpenSessionResponse {
  open: AttendanceRecord | null
  sessions: AttendanceRecord[]
  totalMinutes: number
  isOpen: boolean
}

export interface EmployeeDailyAttendanceSummary {
  employeeId: string
  employee: { id: string; name: string } | null
  sessionsCount: number
  firstCheckIn: string
  lastCheckOut: string | null
  currentlyActive: boolean
  totalMinutes: number
}

export interface Skill {
  id: string
  name: string
  serviceId: string
}

export interface Service {
  id: string
  name: string
  category: string | null
  skills: Skill[]
}

export interface ServiceManager {
  id: string
  createdAt: string
  employee: { id: string; name: string; position: string | null } | null
  service: Service | null
}

export interface LocationPing {
  id: string
  employeeId: string
  bookingId: string | null
  latitude: number
  longitude: number
  createdAt: string
}

export interface EmployeeSkill {
  id: string
  skillId: string
  canPerform: boolean
  skill: Skill
}

export type EmployeeRole =
  | 'ADMIN'
  | 'SALES'
  | 'HR_COORDINATOR'
  | 'TECHNICIAN'
  | 'PROJECT_MANAGER'
  | 'MONITOR'
  | 'FINANCE'
  | 'GPS_ADMIN'
  | 'QUALITY_ENGINEER'
  | 'ENGINEER'
  | 'PROCUREMENT_ADMIN'
  | 'DESIGNER'
  | 'SERVICE_MANAGER'
  | 'OWNER'

export interface Employee {
  id: string
  name: string
  certificate: string | null
  position: string | null
  phone: string | null
  status: 'ACTIVE' | 'INACTIVE' | 'ARCHIVED' | 'DELETED' | 'SUSPENDED'
  role: EmployeeRole
  // الدور الحقيقي للمالك (OWNER) — نطبّع role إلى 'ADMIN' بجلسة الواجهة حتى
  // يشتغل كل شي مبني على role === 'ADMIN' تلقائياً، ونخزن الدور الأصلي هنا
  // للعرض وللتحقق الحصري بصفحة المراقبة الخلفية.
  actualRole?: EmployeeRole
  onDuty: boolean
  username: string | null
  hasDrivingLicense: boolean
  hasSafetyCertificate: boolean
  isLeader: boolean
  isTrainee: boolean
  salary: number | null
  shift: 'MORNING' | 'EVENING' | null
  shiftStart: string | null
  shiftEnd: string | null
  monthlyLeaves: number
  jobTitle: string | null
  skills: EmployeeSkill[]
  hasRequiredSkill?: boolean
}

export interface GpsCustomer {
  id: string
  fullName: string
  fatherName: string | null
  grandfatherName: string | null
  phone: string
  address: string | null
  governorate: string | null
  idCardFrontUrl: string | null
  idCardBackUrl: string | null
  residenceCardFrontUrl: string | null
  residenceCardBackUrl: string | null
  createdAt: string
}

export interface GpsDeviceRequest {
  id: string
  customerId: string
  customer: GpsCustomer
  employeeId: string
  employee: { id: string; name: string }
  adminId: string | null
  purchaseType: 'DEVICE_SIM' | 'DEVICE_ONLY'
  subscriptionType: 'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY'
  subscriptionStart: string | null
  subscriptionEnd: string | null
  subscriptionStatus: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'DELIVERED'
  simCardId: string | null
  notes: string | null
  isChecked: boolean
  isActivated: boolean
  isDelivered: boolean
  invoicePhotoUrl: string | null
  gpsNumber: string | null
  residenceCardNumber: string | null
  activationDate: string | null
  deliveredAt: string | null
  createdAt: string
  scheduledAt: string | null
  assignedTechnician: { id: string; name: string } | null
  assignedTechnicianId?: string | null
}

export interface StaffRequest {
  id: string
  projectId: string | null
  projectName: string | null
  neededAt: string
  durationHours: number
  notes: string | null
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'FULFILLED'
  handledAt: string | null
  createdAt: string
  requester: { id: string; name: string; position: string | null } | null
  handledBy: { id: string; name: string; position: string | null } | null
  employees: { id: string; name: string; position: string | null }[]
}

export interface PerformanceReview {
  id: string
  employeeId: string
  evaluatorId: string
  rating: 'POSITIVE' | 'NEGATIVE'
  reason: string
  createdAt: string
  employee: { id: string; name: string; position: string | null } | null
  evaluator: { id: string; name: string; position: string | null } | null
}

export interface GpsRenewalRequest {
  id: string
  customerId: string
  customer: GpsCustomer
  deviceRequestId: string
  deviceRequest: GpsDeviceRequest
  employeeId: string
  adminId: string | null
  subscriptionType: 'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY'
  newEndDate: string | null
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'DELIVERED'
  createdAt: string
}

export interface GpsMaintenanceRequest {
  id: string
  customerId: string
  customer: GpsCustomer
  employeeId: string
  employee: { id: string; name: string }
  adminId: string | null
  problemDescription: string
  status: 'PENDING' | 'IN_PROGRESS' | 'COMPLETED'
  adminNotes: string | null
  createdAt: string
  resolvedAt: string | null
}

export interface TrainingMaterial {
  id: string
  serviceId: string
  service: Service
  title: string
  url: string
  type: 'VIDEO' | 'DOCUMENT'
  order: number
  createdAt: string
}

export interface BookingAssignment {
  id: string
  role: 'TECH_1' | 'TECH_2' | 'TECH_3'
  employee: Employee
}

export interface Booking {
  id: string
  code: string
  sequenceNumber: number | null
  scheduledAt: string | null
  scheduleLogs: { id: string; changedById: string; changedBy: { id: string; name: string; role: string }; oldTime: string | null; newTime: string; createdAt: string }[]
  materialsReadyAt: string | null
  materialsReadyBy: { id: string; name: string } | null
  responseMinutes: number | null
  customer: Customer
  service: Service | null
  transferEmployee: Employee | null
  projectSupervisor: Employee | null
  expenseResponsible: Employee | null
  expenseResponsibleId: string | null
  confirmedByEmployee: Employee | null
  notes: string | null
  vehicleType: string | null
  priority: 'NORMAL' | 'URGENT'
  status: 'PENDING' | 'CONFIRMED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED'
  transferToProjects: boolean
  confirmedByName: string | null
  adminNotes: string | null
  assignedVehicle: string | null
  quotedPrice: number | null
  address: string | null
  mapLocation: string | null
  mapLatitude: number | null
  mapLongitude: number | null
  completedAt: string | null
  completionNotes: string | null
  amountCollected: number | null
  advancePaid: number | null
  amountVerified: boolean
  assignments: BookingAssignment[]
  cartItems: CartItem[]
  createdAt: string
}

export interface Stats {
  totals: {
    totalCustomers: number
    totalBookings: number
    pendingBookings: number
    confirmedBookings: number
    completedBookings: number
    cancelledBookings: number
    urgentPending: number
    totalRevenue: number
    unverifiedRevenue: number
  }
  salesStats: { employeeId: string; name: string; totalTransferred: number; confirmed: number; today: number; thisMonth: number }[]
  coordinatorStats: { employeeId: string; name: string; totalConfirmed: number; today: number; thisMonth: number }[]
  technicianStats: {
    employeeId: string
    name: string
    onDuty: boolean
    totalAssigned: number
    completed: number
    revenueHandled: number
  }[]
  serviceBreakdown: { serviceId: string | null; name: string; count: number }[]
  roleCounts: { role: EmployeeRole; count: number }[]
  recentBookings: {
    id: string
    code: string
    status: string
    priority: string
    customerName: string
    serviceName: string | null
    createdAt: string
  }[]
}

export interface Expense {
  id: string
  employeeId: string
  employee: { id: string; name: string; position: string | null }
  amount: number
  description: string | null
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  createdAt: string
}

export interface Customer {
  id: string
  customerCode: number
  code: string
  name: string
  phone: string
  location: string | null
  mapLatitude: number | null
  mapLongitude: number | null
  services: string[]
  previousBookingsCount?: number
}

export interface GpsCustomerListItem extends Customer {
  gpsNumber: string | null
  deviceId: string | null
  subscriptionEnd: string | null
}

export interface ComplaintCustomerStat {
  customerId: string
  customerName: string
  customerPhone: string
  complaintCount: number
  openCount: number
}

export interface QualityFollowUp {
  id: string
  status: 'PENDING' | 'CONTACTED_OK' | 'CONTACTED_ISSUE' | 'CONVERTED' | 'CLOSED'
  contactNotes: string | null
  contactedAt: string | null
  createdAt: string
  booking: Booking
  customer: Customer
  contactedByEmployee: { id: string; name: string } | null
}

function currentToken(): string | null {
  return localStorage.getItem('authToken')
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = currentToken()
  const res = await fetch(`${API_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...options,
  })
  if (res.status === 401) {
    const body = await res.json().catch(() => ({}))
    localStorage.removeItem('authToken')
    localStorage.removeItem('currentEmployee')
    if (!path.startsWith('/auth/login')) {
      // نوريه سبب رجوعه لتسجيل الدخول قبل ما نحدّث الصفحة — بدون هذا كانت
      // الشاشة تطلع بيضاء فجأة بدون أي تفسير (يحس المستخدم إنه "خطأ بالنظام").
      alert(body.error || 'انتهت صلاحية الجلسة، الرجاء تسجيل الدخول مجدداً')
      window.location.reload()
    }
    throw new Error(body.error || 'يجب تسجيل الدخول')
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

// downloadFile يجيب ملف (إكسل مثلاً) من الـ API ويحفزّ تنزيله بالمتصفح عبر
// رابط <a download> مصطنع — يستخدم توكن التوثيق متل باقي الطلبات.
async function downloadFile(path: string, filename: string): Promise<void> {
  const token = currentToken()
  const res = await fetch(`${API_URL}${path}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  const blob = await res.blob()
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  window.URL.revokeObjectURL(url)
}

export interface Product {
  id: string
  name: string
  unit: string
  defaultPrice: number
  imageBase64?: string
}

export interface QuotationItem {
  id?: string
  productName: string
  unit: string
  quantity: number
  unitPrice: number
  totalPrice: number
}

export interface Quotation {
  id: string
  quotationNumber: string
  customerName: string
  customerPhone: string | null
  customerAddress: string | null
  projectName: string | null
  items: QuotationItem[]
  grandTotal: number
  discountPercent: number
  discountValue: number
  netTotal: number
  duration: string | null
  notes: string | null
  status: 'NEW' | 'SENT' | 'ACCEPTED' | 'REJECTED'
  createdAt: string
}

export interface Permission {
  id: string
  name: string
  label: string
}

export interface Vehicle {
  id: string
  name: string
  plateNumber: string
  color: string | null
  type: string | null
  model: string | null
  year: number | null
  chassisNumber: string | null
  engineNumber: string | null
  fuelType: string | null
  currentOdometer: number
  condition: string | null
  isActive: boolean
  createdAt: string
}

export interface VehicleDocument {
  id: string
  vehicleId: string
  documentType: 'INSURANCE' | 'ANNUAL_LICENSE' | 'INSPECTION' | 'OTHER'
  documentNumber: string | null
  issueDate: string | null
  expiryDate: string | null
  fileUrl: string | null
  notes: string | null
  createdAt: string
}

export interface VehiclePhoto {
  id: string
  vehicleId: string
  url: string
  caption: string | null
  createdAt: string
}

export interface VehicleMissionPassenger {
  id: string
  missionId: string
  employeeId: string
  employee: { id: string; name: string } | null
}

export interface VehicleMission {
  id: string
  vehicleId: string
  driverId: string
  purpose: string
  destination: string
  startedAt: string
  endedAt: string | null
  startOdometer: number
  endOdometer: number | null
  distanceKm: number | null
  notes: string | null
  status: 'IN_PROGRESS' | 'COMPLETED'
  createdAt: string
  vehicle: Vehicle | null
  driver: { id: string; name: string } | null
  passengers: VehicleMissionPassenger[]
}

export interface VehicleLog {
  id: string
  vehicleId: string
  type: 'FUEL' | 'CLEANING' | 'OIL_CHANGE'
  performedAt: string
  nextDueAt: string | null
  odometer: number | null
  cost: number | null
  notes: string | null
  createdAt: string
  recordedBy: { id: string; name: string } | null
}

export interface VehicleWashRating {
  id: string
  dailyRatingId: string
  employeeId: string
  score: number
  employee: { id: string; name: string } | null
}

export interface VehicleDailyRating {
  id: string
  vehicleId: string
  ratedDate: string
  wash: number | null
  exteriorClean: number | null
  exteriorCondition: number | null
  tireCondition: number | null
  glassClean: number | null
  lightsCondition: number | null
  technicalFaults: number | null
  faultDescription: string | null
  interiorClean: number | null
  seatsCondition: number | null
  interiorDirt: number | null
  smell: number | null
  notes: string | null
  weightedScore: number | null
  recordedBy: { id: string; name: string } | null
  washRatings: VehicleWashRating[]
}

export interface VehicleScoreSummary {
  vehicleId: string
  vehicleName: string
  ratingsCount: number
  averageScore: number
}

export interface TechnicianWashSummary {
  employeeId: string
  employeeName: string
  vehiclesWashed: number
  totalPoints: number
  suggestedWage: number
  monthlyCap: number
}

export interface VehicleIncident {
  id: string
  vehicleId: string
  type: 'FAULT' | 'DAMAGE'
  description: string
  cost: number | null
  status: 'OPEN' | 'RESOLVED'
  createdAt: string
  resolvedAt: string | null
  responsibleEmployee: { id: string; name: string } | null
  reportedBy: { id: string; name: string } | null
}

export interface VehicleMonthlyStatus {
  id: string
  vehicleId: string
  month: string
  hasIssue: boolean
  issueDescription: string | null
  resolved: boolean
  notes: string | null
  createdAt: string
}

export interface QualityIssue {
  id: string
  category: 'EXECUTION' | 'OVERSIGHT'
  title: string
  description: string | null
  bookingId: string | null
  status: 'OPEN' | 'IN_PROGRESS' | 'RESOLVED'
  createdAt: string
  resolvedAt: string | null
  responsibleEmployee: { id: string; name: string } | null
  reportedBy: { id: string; name: string } | null
}

export interface KpiLeaderboardEntry {
  employeeId: string
  employeeName: string
  points: number
  evaluationCount: number
  completedBookings: number
}

export interface RoleKpiLeaderboard {
  role: string
  weekly: KpiLeaderboardEntry[]
  monthly: KpiLeaderboardEntry[]
}

export interface KpiEvaluation {
  id: string
  employeeId: string
  employee: { id: string; name: string }
  evaluatorId: string
  evaluator: { id: string; name: string }
  points: number
  reason: string
  deductionAmount: number
  cancelled: boolean
  cancelledAt: string | null
  cancelledByEmployee: { id: string; name: string } | null
  createdAt: string
}

export interface KpiCriterion {
  id: string
  label: string
  createdAt: string
}

export interface WorkReport {
  id: string
  bookingId: string
  employeeId: string
  workStatus: 'COMPLETED' | 'STOPPED'
  events: string | null
  extraRequests: string | null
  cleanedSite: boolean
  gaveInfo: boolean
  tookPhotos: boolean
  stopReason: string | null
  notes: string | null
  createdAt: string
  employee: { id: string; name: string } | null
  booking: { id: string; code: string; customerName: string } | null
}

export interface TechnicianKpi {
  employeeId: string
  employeeName: string
  period: string
  breakdown: {
    completedBookings: { count: number; points: number }
    completionSpeed: { avgMinutes: number; points: number }
    workReports: { count: number; fullReports: number; points: number }
    attendance: { daysPresent: number; totalDays: number; points: number }
    complaints: { count: number; points: number }
    manualDeductions: { count: number; points: number }
  }
  totalPoints: number
}

export interface CartItem {
  id: string
  bookingId: string
  productName: string
  quantity: number
  unitPrice: number
  totalPrice: number
  notes: string | null
  createdAt: string
}

export interface InventoryCheck {
  id: string
  employeeId: string
  complete: boolean
  missingItems: string | null
  checkedAt: string
  resolved: boolean
  resolvedById: string | null
  resolvedAt: string | null
  employee: { id: string; name: string } | null
  resolvedBy: { id: string; name: string } | null
}

export interface Notification {
  id: string
  employeeId: string
  type: string
  message: string
  read: boolean
  createdAt: string
}

export interface PersonalTool {
  id: string
  employeeId: string
  employee?: { id: string; name: string }
  name: string
  barcode: string
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
  checkedOut: boolean
}

export interface VehicleTool {
  id: string
  name: string
  barcode: string
  vehicleId: string
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
}

export interface OnDemandTool {
  id: string
  name: string
  barcode: string
  totalQuantity: number
  availableQuantity: number
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
}

export type ToolRequest = ToolRequestItem
export interface ToolRequestItem {
  id: string
  employeeId: string
  employee?: { id: string; name: string }
  toolId: string
  tool?: OnDemandTool
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'RETURNED'
  approvedById: string | null
  requestedAt: string
  approvedAt: string | null
  returnedAt: string | null
}

export type ComplaintType = 'DELAY' | 'DISORGANIZED' | 'TECHNICAL' | 'EXECUTION_ERROR' | 'INCOMPLETE' | 'OTHER'

export const complaintTypeLabels: Record<ComplaintType, string> = {
  DELAY: 'تأخير بالتنفيذ',
  DISORGANIZED: 'عمل غير منظم',
  TECHNICAL: 'مشكلة فنية',
  EXECUTION_ERROR: 'خطأ تنفيذي',
  INCOMPLETE: 'لم يتم إكمال العمل',
  OTHER: 'أخرى',
}

export interface Complaint {
  id: string
  customerId: string
  customer: { id: string; name: string; phone: string }
  bookingId: string | null
  type: ComplaintType
  description: string
  status: 'NEW' | 'IN_PROGRESS' | 'RESOLVED' | 'CLOSED'
  createdByEmployeeId: string
  createdByEmployee: { id: string; name: string }
  assignedToEmployeeId: string | null
  assignedToEmployee: { id: string; name: string } | null
  relatedEmployee: { id: string; name: string } | null
  resolution: string | null
  createdAt: string
  resolvedAt: string | null
}

export interface ProcurementItem {
  id: string
  requestId: string
  productName: string
  quantity: number
  unitPrice: number | null
  totalPrice: number | null
  fulfilled: boolean
}

export interface ProcurementRequest {
  id: string
  code: string
  requestedBy: { id: string; name: string; role: string }
  requestedById: string
  booking: (Booking & { customer: Customer }) | null
  bookingId: string | null
  requestType: 'PERSONAL_SUPPLY' | 'CUSTOMER_PRODUCT'
  notes: string | null
  status: 'PENDING' | 'IN_PROGRESS' | 'FULFILLED' | 'REJECTED'
  fulfilledBy: { id: string; name: string } | null
  fulfilledById: string | null
  totalCost: number | null
  fulfillmentNotes: string | null
  createdAt: string
  fulfilledAt: string | null
  items: ProcurementItem[]
}

export interface ProcurementStats {
  totalSpent: number
  totalItems: number
  pendingCount: number
  monthlySpent: number
  fulfilledCount: number
  byMonth: Record<string, number>
}

export const api = {
  getMe: () => request<Employee>('/auth/me'),
  getServices: () => request<Service[]>('/services'),
  createService: (data: { name: string; category?: string }) =>
    request<Service>('/services', { method: 'POST', body: JSON.stringify(data) }),
  createSkill: (serviceId: string, name: string) =>
    request<Skill>(`/services/${serviceId}/skills`, { method: 'POST', body: JSON.stringify({ name }) }),

  // مسؤول خدمة عام (تعميم فكرة أبو الجي بي اس لأي مجموعة خدمات)
  getServiceManagers: () => request<ServiceManager[]>('/service-managers'),
  setServiceManagers: (employeeId: string, serviceIds: string[]) =>
    request<ServiceManager[]>('/service-managers', { method: 'PUT', body: JSON.stringify({ employeeId, serviceIds }) }),

  // تتبع الموقع الحي
  createLocationPing: (data: { latitude: number; longitude: number; bookingId?: string | null }) =>
    request<LocationPing>('/location-pings', { method: 'POST', body: JSON.stringify(data) }),
  getLatestLocations: () => request<LocationPing[]>('/location-pings/latest'),
  getLocationPath: (employeeId: string, bookingId?: string) =>
    request<LocationPing[]>(`/location-pings/path?employeeId=${employeeId}${bookingId ? `&bookingId=${bookingId}` : ''}`),

  getEmployees: () => request<Employee[]>('/employees'),
  getArchivedEmployees: () => request<Employee[]>('/employees/archived'),
  createEmployee: (
    data: Pick<Employee, 'name' | 'certificate' | 'position' | 'phone'> & {
      username?: string
      password?: string
      jobTitle?: string
      salary?: number
      shift?: 'MORNING' | 'EVENING'
      shiftStart?: string
      shiftEnd?: string
      role?: EmployeeRole
    },
  ) => request<Employee>('/employees', { method: 'POST', body: JSON.stringify(data) }),
  login: async (username: string, password: string) => {
    const result = await request<Employee & { token: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    const { token, ...employee } = result
    localStorage.setItem('authToken', token)
    return employee
  },
  updateEmployeeSkills: (id: string, skills: { skillId: string; canPerform: boolean }[]) =>
    request<Employee>(`/employees/${id}/skills`, {
      method: 'PUT',
      body: JSON.stringify({ skills }),
    }),
  matchEmployees: (serviceId: string) =>
    request<Employee[]>(`/employees/match?serviceId=${serviceId}`),
  getSupervisors: () => request<Employee[]>('/employees/supervisors'),

  getCustomers: () => request<Customer[]>('/customers'),
  getCustomersByGpsService: () => request<GpsCustomerListItem[]>('/customers/gps'),
  lookupCustomer: (phone: string) =>
    request<Customer | null>(`/customers/lookup?phone=${phone}`).catch(() => null),
  createCustomer: (data: { name: string; phone: string; location?: string }) =>
    request<Customer & { existed: boolean }>('/customers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateCustomer: (id: string, data: { name: string; phone: string }) =>
    request<Customer>(`/customers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  getComplaintStats: () => request<ComplaintCustomerStat[]>('/complaints/stats'),

  getSecurityDashboard: () =>
    request<{
      serverUptimeSeconds: number
      goroutineCount: number
      memoryUsedMB: number
      failedLoginsLastHour: number
      totalRequests: number
      requestsLastMinute: number
      cpuCount: number
      diskTotalGB: number
      diskUsedGB: number
      diskFreeGB: number
      dbSizeMB: number
      dbConnectionsOpen: number
      dbConnectionsInUse: number
      onlineEmployees: number
      recentLogins: {
        id: string
        username: string
        success: boolean
        ipAddress: string | null
        userAgent: string | null
        createdAt: string
        employee: { id: string; name: string } | null
      }[]
    }>('/security/dashboard'),
  freeServerMemory: () => request<{ memoryUsedMB: number }>('/security/free-memory', { method: 'POST' }),

  askAssistant: (message: string) =>
    request<{ reply: string }>('/assistant/ask', { method: 'POST', body: JSON.stringify({ message }) }),
  managerChatAssistant: (message: string, history: { role: 'user' | 'assistant'; text: string }[]) =>
    request<{ reply: string }>('/assistant/manager-chat', { method: 'POST', body: JSON.stringify({ message, history }) }),

  getQualityFollowUps: () => request<QualityFollowUp[]>('/quality-follow-ups'),
  updateQualityFollowUp: (
    id: string,
    data: { status: QualityFollowUp['status']; contactNotes?: string },
  ) =>
    request<QualityFollowUp>(`/quality-follow-ups/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  getBookings: (params?: { status?: Booking['status']; customerId?: string }) => {
    const query = new URLSearchParams()
    if (params?.status) query.set('status', params.status)
    if (params?.customerId) query.set('customerId', params.customerId)
    const qs = query.toString()
    return request<Booking[]>(`/bookings${qs ? `?${qs}` : ''}`)
  },
  createBooking: (data: {
    customerId: string
    serviceId?: string
    notes?: string
    vehicleType?: string
    priority?: 'NORMAL' | 'URGENT'
    transferEmployeeId?: string
    address?: string
    mapLatitude?: number
    mapLongitude?: number
  }) => request<Booking>('/bookings', { method: 'POST', body: JSON.stringify(data) }),
  confirmBooking: (
    id: string,
    data: {
      confirmedByName: string
      confirmedByEmployeeId?: string
      adminNotes?: string
      transferToProjects: boolean
      quotedPrice?: number
      address?: string
      scheduledAt?: string
    },
  ) => request<Booking>(`/bookings/${id}/confirm`, { method: 'PUT', body: JSON.stringify(data) }),
  scheduleBooking: (id: string, scheduledAt: string, changedById?: string) =>
    request<Booking>(`/bookings/${id}/schedule`, { method: 'PUT', body: JSON.stringify({ scheduledAt, changedById }) }),
  updateBookingDetails: (
    id: string,
    data: { quotedPrice?: number | null; address?: string; assignedVehicle?: string; mapLocation?: string; mapLatitude?: number | null; mapLongitude?: number | null; expenseResponsibleId?: string | null },
  ) => request<Booking>(`/bookings/${id}/details`, { method: 'PUT', body: JSON.stringify(data) }),
  assignTechnician: (
    id: string,
    data: { employeeId: string; role: 'TECH_1' | 'TECH_2' | 'TECH_3'; assignedVehicle?: string },
  ) => request<Booking>(`/bookings/${id}/assign`, { method: 'PUT', body: JSON.stringify(data) }),
  assignSupervisor: (id: string, employeeId: string | null) =>
    request<Booking>(`/bookings/${id}/supervisor`, {
      method: 'PUT',
      body: JSON.stringify({ employeeId }),
    }),
  completeBooking: (
    id: string,
    data: { completionNotes?: string; amountCollected?: number; advancePaid?: number },
  ) => request<Booking>(`/bookings/${id}/complete`, { method: 'PUT', body: JSON.stringify(data) }),
  startBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/start`, { method: 'PUT', body: JSON.stringify({}) }),
  setMaterialsReady: (id: string) =>
    request<Booking>(`/bookings/${id}/materials-ready`, { method: 'PUT', body: JSON.stringify({}) }),
  verifyAmount: (id: string) =>
    request<Booking>(`/bookings/${id}/verify`, { method: 'PUT', body: JSON.stringify({}) }),
  getStats: () => request<Stats>('/stats'),

  getExpenses: (employeeId?: string) =>
    request<Expense[]>(`/expenses${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createExpense: (data: { employeeId: string; amount: number; description?: string }) =>
    request<Expense>('/expenses', { method: 'POST', body: JSON.stringify(data) }),
  updateExpenseStatus: (id: string, status: 'APPROVED' | 'REJECTED') =>
    request<Expense>(`/expenses/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),

  updateEmployee: (
    id: string,
    data: Partial<
      Pick<
        Employee,
        'role' | 'onDuty' | 'status' | 'name' | 'position' | 'phone' | 'certificate' | 'hasDrivingLicense' | 'hasSafetyCertificate' | 'isTrainee'
        | 'isLeader' | 'salary' | 'shift' | 'shiftStart' | 'shiftEnd' | 'monthlyLeaves' | 'jobTitle'
      >
    > & {
      username?: string
      password?: string
    },
  ) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  linkHistoricalRecords: (id: string) =>
    request<{ bookingsLinked: number; complaintsLinked: number }>(`/employees/${id}/link-historical`, { method: 'POST' }),

  // GPS
  getGpsStats: () => request<any>('/gps/stats'),
  getGpsCustomers: () => request<GpsCustomer[]>('/gps/customers'),
  createGpsCustomer: (data: Partial<GpsCustomer>) =>
    request<GpsCustomer>('/gps/customers', { method: 'POST', body: JSON.stringify(data) }),
  getGpsDevices: () => request<GpsDeviceRequest[]>('/gps/devices'),
  createGpsDevice: (data: Partial<GpsDeviceRequest>) =>
    request<GpsDeviceRequest>('/gps/devices', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsDevice: (id: string, data: Partial<GpsDeviceRequest>) =>
    request<GpsDeviceRequest>(`/gps/devices/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getSimCards: () => request<any[]>('/gps/sims'),
  createSimCard: (data: any) =>
    request<any>('/gps/sims', { method: 'POST', body: JSON.stringify(data) }),
  updateSimCard: (id: string, data: any) =>
    request<any>(`/gps/sims/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getGpsRenewals: () => request<GpsRenewalRequest[]>('/gps/renewals'),
  createGpsRenewal: (data: { customerId: string; deviceRequestId: string; employeeId: string; subscriptionType: string }) =>
    request<GpsRenewalRequest>('/gps/renewals', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsRenewal: (id: string, data: Partial<GpsRenewalRequest>) =>
    request<GpsRenewalRequest>(`/gps/renewals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getGpsMaintenance: () => request<GpsMaintenanceRequest[]>('/gps/maintenance'),
  createGpsMaintenance: (data: { customerId: string; employeeId: string; problemDescription: string }) =>
    request<GpsMaintenanceRequest>('/gps/maintenance', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsMaintenance: (id: string, data: Partial<GpsMaintenanceRequest>) =>
    request<GpsMaintenanceRequest>(`/gps/maintenance/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // طلبات الكادر (مدير المشاريع يطلب، إدارة الكوادر تلبي)
  getStaffRequests: (mine?: boolean) => request<StaffRequest[]>(`/staff-requests${mine ? '?mine=1' : ''}`),
  createStaffRequest: (data: { projectId?: string | null; neededAt: string; durationHours: number; notes?: string | null; employeeIds: string[] }) =>
    request<StaffRequest>('/staff-requests', { method: 'POST', body: JSON.stringify(data) }),
  updateStaffRequestStatus: (id: string, status: 'APPROVED' | 'REJECTED' | 'FULFILLED') =>
    request<StaffRequest>(`/staff-requests/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),

  // تقييم الأداء (منفصل عن KPI مال الغرامات) — يحدد استحقاق التدريب فقط
  createPerformanceReview: (data: { employeeId: string; rating: 'POSITIVE' | 'NEGATIVE'; reason: string }) =>
    request<PerformanceReview>('/performance-reviews', { method: 'POST', body: JSON.stringify(data) }),
  getPerformanceReviews: () => request<PerformanceReview[]>('/performance-reviews'),
  getPerformanceReviewsForEmployee: (employeeId: string) => request<PerformanceReview[]>(`/performance-reviews/employee/${employeeId}`),

  // Products
  getProducts: () => request<Product[]>('/products'),
  createProduct: (data: Omit<Product, 'id'> & { imageBase64?: string }) =>
    request<Product>('/products', { method: 'POST', body: JSON.stringify(data) }),
  updateProduct: (id: string, data: Partial<Omit<Product, 'id'>> & { imageBase64?: string }) =>
    request<Product>(`/products/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteProduct: (id: string) =>
    request<void>(`/products/${id}`, { method: 'DELETE' }),

  // Quotations
  getQuotations: () => request<Quotation[]>('/quotations'),
  getQuotation: (id: string) => request<Quotation>(`/quotations/${id}`),
  createQuotation: (data: {
    customerName: string
    customerPhone?: string
    customerAddress?: string
    projectName?: string
    items: Omit<QuotationItem, 'id'>[]
    grandTotal: number
    discountPercent: number
    discountValue: number
    netTotal: number
    duration?: string
    notes?: string
    createdByEmployeeId?: string
  }) => request<Quotation>('/quotations', { method: 'POST', body: JSON.stringify(data) }),
  updateQuotation: (id: string, data: {
    customerName?: string
    customerPhone?: string
    customerAddress?: string
    projectName?: string
    items?: Omit<QuotationItem, 'id'>[]
    discountPercent?: number
    duration?: string
    notes?: string
    status?: Quotation['status']
  }) => request<Quotation>(`/quotations/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteQuotation: (id: string) =>
    request<void>(`/quotations/${id}`, { method: 'DELETE' }),

  // Permissions
  getPermissions: () => request<Permission[]>('/permissions'),
  getEmployeePermissions: (employeeId: string) =>
    request<Permission[]>(`/permissions/employee/${employeeId}`),
  setEmployeePermissions: (employeeId: string, permissionIds: string[]) =>
    request<Permission[]>(`/permissions/employee/${employeeId}`, {
      method: 'PUT',
      body: JSON.stringify({ permissionIds }),
    }),
  applyDefaultPermissions: (employeeId: string) =>
    request<Permission[]>(`/permissions/employee/${employeeId}/apply-defaults`, { method: 'POST' }),
  getRoleDefaults: () =>
    request<Record<string, string[]>>('/permissions/role-defaults'),

  // Attendance
  checkIn: () => request<AttendanceRecord>('/attendance/checkin', { method: 'POST' }),
  checkOut: () => request<AttendanceRecord>('/attendance/checkout', { method: 'POST' }),
  getMyAttendanceToday: () => request<AttendanceRecord | null>('/attendance/mine'),
  getMyOpenSession: () => request<OpenSessionResponse>('/attendance/open'),
  getTodayAttendance: () => request<AttendanceRecord[]>('/attendance/today'),
  getTodaySummary: () => request<EmployeeDailyAttendanceSummary[]>('/attendance/today-summary'),
  getMonthlyAttendance: (employeeId: string, month: string) =>
    request<MonthlyAttendanceReport>(`/attendance/employee/${employeeId}?month=${month}`),
  correctAttendance: (id: string, data: { checkIn?: string; checkOut?: string }) =>
    request<AttendanceRecord>(`/attendance/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  exportEmployeeAttendance: (employeeId: string, month: string) =>
    downloadFile(`/attendance/export/employee/${employeeId}?month=${month}`, `attendance-${employeeId}-${month}.xlsx`),
  exportTodayAttendance: (date?: string) =>
    downloadFile(`/attendance/export/today${date ? `?date=${date}` : ''}`, `attendance-today-${date || 'now'}.xlsx`),

  // KPI
  getKpiEvaluations: () => request<KpiEvaluation[]>('/kpi'),
  getEmployeeKpi: (employeeId: string) => request<KpiEvaluation[]>(`/kpi/employee/${employeeId}`),
  createKpiEvaluation: (data: { employeeId: string; evaluatorId: string; points: number; reason: string }) =>
    request<KpiEvaluation>('/kpi', { method: 'POST', body: JSON.stringify(data) }),
  deleteKpiEvaluation: (id: string) => request<void>(`/kpi/${id}`, { method: 'DELETE' }),
  cancelKpiEvaluation: (id: string) => request<KpiEvaluation>(`/kpi/${id}/cancel`, { method: 'PUT' }),

  getKpiCriteria: () => request<KpiCriterion[]>('/kpi-criteria'),
  createKpiCriterion: (label: string) =>
    request<KpiCriterion>('/kpi-criteria', { method: 'POST', body: JSON.stringify({ label }) }),
  deleteKpiCriterion: (id: string) => request<void>(`/kpi-criteria/${id}`, { method: 'DELETE' }),
  completeTraining: (employeeId: string) =>
    request<KpiEvaluation>(`/employees/${employeeId}/complete-training`, { method: 'POST' }),

  // Vehicles
  getVehicles: () => request<Vehicle[]>('/vehicles'),
  createVehicle: (data: { name: string; plateNumber: string; color?: string; type?: string }) =>
    request<Vehicle>('/vehicles', { method: 'POST', body: JSON.stringify(data) }),
  getVehicleLogs: (vehicleId: string) => request<VehicleLog[]>(`/vehicles/${vehicleId}/logs`),
  createVehicleLog: (vehicleId: string, data: { type: 'FUEL' | 'CLEANING' | 'OIL_CHANGE'; performedAt?: string; nextDueAt?: string; odometer?: number; cost?: number; notes?: string }) =>
    request<VehicleLog>(`/vehicles/${vehicleId}/logs`, { method: 'POST', body: JSON.stringify(data) }),
  getVehicleIncidents: (vehicleId: string) => request<VehicleIncident[]>(`/vehicles/${vehicleId}/incidents`),
  createVehicleIncident: (vehicleId: string, data: { type: 'FAULT' | 'DAMAGE'; description: string; responsibleEmployeeId?: string; cost?: number }) =>
    request<VehicleIncident>(`/vehicles/${vehicleId}/incidents`, { method: 'POST', body: JSON.stringify(data) }),
  updateVehicleIncident: (id: string, data: { status?: 'OPEN' | 'RESOLVED'; cost?: number }) =>
    request<VehicleIncident>(`/vehicle-incidents/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getVehicleMonthlyStatus: (vehicleId: string) => request<VehicleMonthlyStatus[]>(`/vehicles/${vehicleId}/monthly-status`),
  setVehicleMonthlyStatus: (vehicleId: string, data: { month: string; hasIssue: boolean; issueDescription?: string; resolved: boolean; notes?: string }) =>
    request<VehicleMonthlyStatus>(`/vehicles/${vehicleId}/monthly-status`, { method: 'POST', body: JSON.stringify(data) }),
  createVehicleDailyRating: (vehicleId: string, data: {
    ratedDate?: string; wash?: number; exteriorClean?: number; exteriorCondition?: number; tireCondition?: number
    glassClean?: number; lightsCondition?: number; technicalFaults?: number; faultDescription?: string
    interiorClean?: number; seatsCondition?: number; interiorDirt?: number; smell?: number; notes?: string
    technicianRatings?: { employeeId: string; score: number }[]
  }) => request<VehicleDailyRating>(`/vehicles/${vehicleId}/ratings`, { method: 'POST', body: JSON.stringify({ vehicleId, ...data }) }),
  getVehicleDailyRatings: (vehicleId: string, since?: string) =>
    request<VehicleDailyRating[]>(`/vehicles/${vehicleId}/ratings${since ? `?since=${since}` : ''}`),
  getVehicleScoreSummaries: (since?: string) =>
    request<VehicleScoreSummary[]>(`/vehicles/ratings/vehicle-summary${since ? `?since=${since}` : ''}`),
  updateVehicle: (id: string, data: Partial<{
    name: string; plateNumber: string; color: string; type: string; model: string; year: number
    chassisNumber: string; engineNumber: string; fuelType: string; currentOdometer: number
    condition: string; isActive: boolean
  }>) => request<Vehicle>(`/vehicles/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  getVehicleDocuments: (vehicleId: string) => request<VehicleDocument[]>(`/vehicles/${vehicleId}/documents`),
  createVehicleDocument: (vehicleId: string, data: {
    documentType: 'INSURANCE' | 'ANNUAL_LICENSE' | 'INSPECTION' | 'OTHER'
    documentNumber?: string; issueDate?: string; expiryDate?: string; fileUrl?: string; notes?: string
  }) => request<VehicleDocument>(`/vehicles/${vehicleId}/documents`, { method: 'POST', body: JSON.stringify(data) }),
  deleteVehicleDocument: (vehicleId: string, docId: string) =>
    request<{ ok: boolean }>(`/vehicles/${vehicleId}/documents/${docId}`, { method: 'DELETE' }),

  getVehiclePhotos: (vehicleId: string) => request<VehiclePhoto[]>(`/vehicles/${vehicleId}/photos`),
  createVehiclePhoto: (vehicleId: string, data: { url: string; caption?: string }) =>
    request<VehiclePhoto>(`/vehicles/${vehicleId}/photos`, { method: 'POST', body: JSON.stringify(data) }),
  deleteVehiclePhoto: (vehicleId: string, photoId: string) =>
    request<{ ok: boolean }>(`/vehicles/${vehicleId}/photos/${photoId}`, { method: 'DELETE' }),

  startVehicleMission: (data: {
    vehicleId: string; driverId?: string; purpose: string; destination: string
    startOdometer: number; passengerIds?: string[]
  }) => request<VehicleMission>('/vehicle-missions', { method: 'POST', body: JSON.stringify(data) }),
  endVehicleMission: (id: string, data: { endOdometer: number; notes?: string }) =>
    request<VehicleMission>(`/vehicle-missions/${id}/end`, { method: 'PUT', body: JSON.stringify(data) }),
  getVehicleMissions: (filters?: { vehicleId?: string; driverId?: string; status?: 'IN_PROGRESS' | 'COMPLETED'; from?: string; to?: string }) => {
    const params = new URLSearchParams()
    if (filters?.vehicleId) params.set('vehicleId', filters.vehicleId)
    if (filters?.driverId) params.set('driverId', filters.driverId)
    if (filters?.status) params.set('status', filters.status)
    if (filters?.from) params.set('from', filters.from)
    if (filters?.to) params.set('to', filters.to)
    const qs = params.toString()
    return request<VehicleMission[]>(`/vehicle-missions${qs ? `?${qs}` : ''}`)
  },
  getVehicleMission: (id: string) => request<VehicleMission>(`/vehicle-missions/${id}`),
  getTechnicianWashSummaries: (since?: string) =>
    request<TechnicianWashSummary[]>(`/vehicles/ratings/technician-summary${since ? `?since=${since}` : ''}`),

  // Quality
  getQualityIssues: (category?: 'EXECUTION' | 'OVERSIGHT') =>
    request<QualityIssue[]>(`/quality/issues${category ? `?category=${category}` : ''}`),
  createQualityIssue: (data: { category: 'EXECUTION' | 'OVERSIGHT'; title: string; description?: string; responsibleEmployeeId?: string; bookingId?: string }) =>
    request<QualityIssue>('/quality/issues', { method: 'POST', body: JSON.stringify(data) }),
  updateQualityIssue: (id: string, data: { status: 'OPEN' | 'IN_PROGRESS' | 'RESOLVED' }) =>
    request<QualityIssue>(`/quality/issues/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // Training
  getMyTraining: (employeeId: string) =>
    request<{ services: Service[]; materials: TrainingMaterial[] }>(`/training/materials/mine?employeeId=${employeeId}`),
  getTrainingAssignments: (employeeId: string) =>
    request<Service[]>(`/training/assignments/${employeeId}`),
  setTrainingAssignments: (employeeId: string, serviceIds: string[]) =>
    request<Service[]>(`/training/assignments/${employeeId}`, {
      method: 'PUT',
      body: JSON.stringify({ serviceIds }),
    }),
  getTrainingMaterials: (serviceId?: string) =>
    request<TrainingMaterial[]>(`/training/materials${serviceId ? `?serviceId=${serviceId}` : ''}`),
  createTrainingMaterial: (data: { serviceId: string; title: string; url: string; type?: string; order?: number }) =>
    request<TrainingMaterial>('/training/materials', { method: 'POST', body: JSON.stringify(data) }),
  updateTrainingMaterial: (id: string, data: { title: string; url: string; type?: string; order?: number }) =>
    request<TrainingMaterial>(`/training/materials/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteTrainingMaterial: (id: string) =>
    request<void>(`/training/materials/${id}`, { method: 'DELETE' }),

  // Work Reports
  createWorkReport: (data: {
    bookingId: string
    workStatus: 'COMPLETED' | 'STOPPED'
    events?: string
    extraRequests?: string
    cleanedSite: boolean
    gaveInfo: boolean
    tookPhotos: boolean
    stopReason?: string
    notes?: string
  }) => request<WorkReport>('/work-reports', { method: 'POST', body: JSON.stringify(data) }),
  getWorkReports: (employeeId?: string) =>
    request<WorkReport[]>(`/work-reports${employeeId ? `?employeeId=${employeeId}` : ''}`),

  // Smart KPI
  getTechnicianKpi: (employeeId: string, month?: string) =>
    request<TechnicianKpi>(`/smart-kpi/technician/${employeeId}${month ? `?month=${month}` : ''}`),
  getKpiLeaderboard: (month?: string) =>
    request<TechnicianKpi[]>(`/smart-kpi/leaderboard${month ? `?month=${month}` : ''}`),
  getRoleKpiLeaderboard: (role: string) => request<RoleKpiLeaderboard>(`/kpi/leaderboard/${role}`),

  // Cart
  getCartItems: (bookingId: string) => request<CartItem[]>(`/cart/booking/${bookingId}`),
  addCartItem: (bookingId: string, data: { productName: string; quantity: number; unitPrice: number; notes?: string }) =>
    request<CartItem>(`/cart/booking/${bookingId}`, { method: 'POST', body: JSON.stringify(data) }),
  updateCartItem: (id: string, data: Partial<CartItem>) =>
    request<CartItem>(`/cart/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteCartItem: (id: string) => request<void>(`/cart/${id}`, { method: 'DELETE' }),

  // Notifications
  getNotifications: () => request<{ notifications: Notification[]; unreadCount: number }>('/notifications'),
  markNotificationRead: (id: string) => request<{ success: boolean }>(`/notifications/${id}/read`, { method: 'POST' }),
  markAllNotificationsRead: () => request<{ success: boolean }>('/notifications/read-all', { method: 'POST' }),

  // Inventory
  createInventoryCheck: (data: { complete: boolean; missingItems?: string }) =>
    request<InventoryCheck>('/inventory/checks', { method: 'POST', body: JSON.stringify(data) }),
  getTodaysInventoryChecks: () => request<InventoryCheck[]>('/inventory/checks/today'),
  resolveInventoryCheck: (id: string) => request<InventoryCheck>(`/inventory/checks/${id}/resolve`, { method: 'POST' }),
  getPersonalTools: (employeeId?: string) =>
    request<PersonalTool[]>(`/inventory/personal${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createPersonalTool: (data: { employeeId: string; name: string; barcode: string }) =>
    request<PersonalTool>('/inventory/personal', { method: 'POST', body: JSON.stringify(data) }),
  updatePersonalTool: (id: string, data: Partial<PersonalTool>) =>
    request<PersonalTool>(`/inventory/personal/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePersonalTool: (id: string) => request<void>(`/inventory/personal/${id}`, { method: 'DELETE' }),
  getVehicleTools: (vehicleId?: string) =>
    request<VehicleTool[]>(`/inventory/vehicle${vehicleId ? `?vehicleId=${vehicleId}` : ''}`),
  createVehicleTool: (data: { name: string; barcode: string; vehicleId: string }) =>
    request<VehicleTool>('/inventory/vehicle', { method: 'POST', body: JSON.stringify(data) }),
  updateVehicleTool: (id: string, data: Partial<VehicleTool>) =>
    request<VehicleTool>(`/inventory/vehicle/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteVehicleTool: (id: string) => request<void>(`/inventory/vehicle/${id}`, { method: 'DELETE' }),
  getOnDemandTools: () => request<OnDemandTool[]>('/inventory/ondemand'),
  createOnDemandTool: (data: { name: string; barcode: string; totalQuantity: number }) =>
    request<OnDemandTool>('/inventory/ondemand', { method: 'POST', body: JSON.stringify(data) }),
  updateOnDemandTool: (id: string, data: Partial<OnDemandTool>) =>
    request<OnDemandTool>(`/inventory/ondemand/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getToolRequests: (employeeId?: string) =>
    request<ToolRequestItem[]>(`/inventory/requests${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createToolRequest: (data: { employeeId: string; toolId: string }) =>
    request<ToolRequestItem>('/inventory/requests', { method: 'POST', body: JSON.stringify(data) }),
  approveToolRequest: (id: string, approvedById: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/approve`, { method: 'PUT', body: JSON.stringify({ approvedById }) }),
  rejectToolRequest: (id: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/reject`, { method: 'PUT', body: JSON.stringify({}) }),
  returnToolRequest: (id: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/return`, { method: 'PUT', body: JSON.stringify({}) }),

  // Complaints
  getComplaints: () => request<Complaint[]>('/complaints'),
  createComplaint: (data: {
    customerId: string
    bookingId?: string
    type: ComplaintType
    description?: string
    relatedEmployeeId?: string
    createdByEmployeeId: string
  }) => request<Complaint>('/complaints', { method: 'POST', body: JSON.stringify(data) }),
  updateComplaint: (id: string, data: { status?: string; assignedToEmployeeId?: string; resolution?: string }) =>
    request<Complaint>(`/complaints/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  resolveComplaint: (id: string, resolution: string) =>
    request<Complaint>(`/complaints/${id}/resolve`, { method: 'PUT', body: JSON.stringify({ resolution }) }),

  // GPS Settings
  getGpsSettings: () => request<any[]>('/gps/settings'),
  updateGpsSettings: (data: any) =>
    request<any>('/gps/settings', { method: 'PUT', body: JSON.stringify(data) }),
  updateGpsCustomer: (id: string, data: any) =>
    request<any>(`/gps/customers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // Procurement
  getProcurementRequests: () => request<ProcurementRequest[]>('/procurement'),
  getProcurementStats: () => request<ProcurementStats>('/procurement/stats'),
  createProcurementRequest: (data: { requestedById: string; bookingId?: string; requestType: 'PERSONAL_SUPPLY' | 'CUSTOMER_PRODUCT'; notes?: string; items: { productName: string; quantity: number }[] }) =>
    request<ProcurementRequest>('/procurement', { method: 'POST', body: JSON.stringify(data) }),
  updateProcurementStatus: (id: string, status: string) =>
    request<ProcurementRequest>(`/procurement/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
  fulfillProcurementRequest: (id: string, data: { fulfilledById: string; totalCost?: number; fulfillmentNotes?: string; items?: { id: string; unitPrice?: number; totalPrice?: number; fulfilled?: boolean }[] }) =>
    request<ProcurementRequest>(`/procurement/${id}/fulfill`, { method: 'PUT', body: JSON.stringify(data) }),
}
