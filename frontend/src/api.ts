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

export type Division = 'ENGINEERING' | 'DECOR'

export interface Service {
  id: string
  name: string
  category: string | null
  division: Division
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
  | 'TECHNICAL'
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
  division: Division
  attendanceIcon?: string | null

  // ═══ ملف الموارد البشرية ═══
  // careerStatus تنحسب بالسيرفر من الخبرة والمستوى والتقييم — قاعدة
  // وحدة للكل، ما تنغيّر لو أحد فتح أدوات المطوّر.
  department: string | null
  hireDate: string | null
  experienceYears: number | null
  lastReview: string | null
  careerStatus: string
  jobLevel: number
  nextRole: string | null
  trainingNeeds: string | null

  skills: EmployeeSkill[]
  hasRequiredSkill?: boolean
}

export interface AttendanceIconRequest {
  id: string
  requestedIcon: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  createdAt: string
  employee: { id: string; name: string } | null
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

export interface GpsStats {
  totalCustomers?: number
  totalSims?: number
  availableSims?: number
  inUseSims?: number
}


/** الدوار: مبلغ دوّار للعمل يصرفه المحاسب ويرجع لما يتسوّى */
export interface RevolvingFund {
  id: string
  name: string
  balance: number
  isActive: boolean
  outstandingTotal: number
  /** المبلغ الي الدوار ناطره من المحاسب: مصروف مدقّق وما انخرّج بعد */
  awaitingDischargeTotal: number
}

export type FundTxnKind = 'DISBURSE' | 'SETTLEMENT' | 'TOPUP' | 'DISCHARGE'
export type FundTxnStatus = 'PENDING' | 'APPROVED' | 'REJECTED'

export interface RevolvingFundTxn {
  id: string
  fundId: string
  fundName: string
  employeeId?: string | null
  employeeName?: string | null
  kind: FundTxnKind
  kindLabel: string
  amount: number
  spentAmount: number
  returnedAmount: number
  bookingId?: string | null
  receiptImage?: string | null
  notes?: string | null
  status: FundTxnStatus
  statusLabel: string
  reviewedByName?: string | null
  reviewedAt?: string | null
  reviewNote?: string | null
  createdAt: string
  /** التخريج — الخطوة الي ترجّع المبلغ المصروف للدوار */
  requestId?: string | null
  dischargedAt?: string | null
  dischargedByName?: string | null
  dischargeAccount?: string | null
  dischargeNote?: string | null
  /** تسوية مدققة، بيها مصروف، وما انخرجت بعد → الدوار ناطرها */
  awaitingDischarge: boolean
}

export interface EmployeeFundBalance {
  employeeId: string
  employeeName: string
  jobTitle?: string | null
  totalTaken: number
  totalSettled: number
  outstanding: number
  pendingSettlements: number
}


/** تفاصيل المشروع والمبالغ بمتابعة الجودة — الفارق محسوب بالسيرفر */
export interface QualityFollowUpFinancials {
  bookingCode: string
  serviceName?: string | null
  location?: string | null
  workDetails?: string | null
  projectCode?: string | null
  projectName?: string | null
  projectStage?: string | null
  quotedPrice?: number | null
  projectPrice?: number | null
  advancePaid?: number | null
  amountCollected?: number | null
  agreedTotal: number
  receivedTotal: number
  difference: number
}

/** حساب تكاليف الشد — تفصيلي لكل الكوادر */
export interface GpsInstallCostRow { month: string; employeeName: string; installs: number; total: number }
export interface GpsInstallCostSummary {
  rows: GpsInstallCostRow[]
  byEmployee: { employeeName: string; total: number }[]
  grandTotal: number
  totalInstalls: number
  monthCount: number
}

/** سجل إضافة كمية للمخزون */
export interface StockIntake {
  id: string
  toolId: string
  toolName: string
  quantity: number
  unitPrice?: number | null
  supplier?: string | null
  notes?: string | null
  createdName?: string | null
  createdAt: string
}


/** طلب إجازة — الموافقة تروح للمخوّل حسب نوع كادر الموظف */
// مسار الموافقة = شفت الموظف. إداري الكوادر يوافق على شفته هو بس.
export type LeaveRoute = 'MORNING' | 'EVENING'
export type LeaveStatus = 'PENDING' | 'PRELIMINARY' | 'APPROVED' | 'REJECTED' | 'CANCELLED'

export interface LeaveRequest {
  id: string
  employeeId: string
  employeeName: string
  employeeRole: string
  employeeShift?: string | null
  jobTitle?: string | null
  startDate: string
  endDate: string
  days: number
  reason?: string | null
  route: LeaveRoute
  routeLabel: string
  status: LeaveStatus
  statusLabel: string
  decidedByName?: string | null
  decidedAt?: string | null
  decisionNote?: string | null
  // الموافقة الأولية — مرحلة وسط، الطلب يبقى مفتوح عند المدير بعدها
  preliminaryByName?: string | null
  preliminaryAt?: string | null
  preliminaryNote?: string | null
  createdAt: string
}

export interface GpsSimCard {
  id: string
  simNumber: string
  iccid: string
  operator: string
  // Some legacy records use 'AVAILABLE'/'IN_USE' instead of 'ACTIVE'/'INACTIVE'.
  status: string
  statusLabel?: string
  customerId?: string | null
  customerName?: string | null
  customer?: GpsCustomer | null
  assignedAt?: string | null
  releasedAt?: string | null
  burnedAt?: string | null
  createdAt?: string
}

/** نتيجة اتصال مهندس الجودة بزبون انتهى اشتراكه */
export type GpsFollowUpOutcome = 'WILL_RENEW' | 'WILL_MOVE' | 'REFUSED' | 'NO_ANSWER'

/** مرحلة الاشتراك المنتهي بدورة المتابعة — تجي محسوبة من السيرفر */
export type GpsFollowUpStage = 'GRACE' | 'CALL_DUE' | 'WAITING' | 'BURN_DUE' | 'RESOLVED'

export interface GpsSubscriptionFollowUp {
  deviceRequestId: string
  customerId: string
  customerName: string
  customerPhone: string
  subscriptionEnd: string | null
  daysSinceExpiry: number
  simCardId?: string | null
  simNumber?: string | null
  simStatus?: string | null
  gpsNumber?: string | null
  lastOutcome?: GpsFollowUpOutcome | null
  lastOutcomeLabel?: string
  lastCalledAt?: string | null
  stage: GpsFollowUpStage
  stageLabel: string
  daysUntilNextStep: number
}

export interface GpsRenewalFollowUp {
  id: string
  deviceRequestId: string
  outcome: GpsFollowUpOutcome
  outcomeLabel: string
  notes?: string | null
  daysSinceExpiry?: number | null
  calledByName?: string | null
  calledAt: string
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
  simCard?: GpsSimCard | null
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
  price?: number | null
  credentialsMessage: string | null
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
  subscriptionType: string
  newEndDate: string | null
  currentEnd?: string | null
  status: string
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

// حالة اكتمال الحجز بعد الإنجاز — تنحسب بالسيرفر من وجود فاتورة
// التكاليف المربوطة بالحجز وتقرير العمل.
export type CompletionState =
  | 'CONFIRMED_ONLY'   // مثبت — انتأكد بس ماكو أحد منكلّف بيه بعد
  | 'ASSIGNED'         // في حالة التكليف — منترحّل لليدر وصار عليه مسؤول
  | 'STOPPED'          // توقف العمل
  | 'DONE_NO_BOTH'     // تم الإنجاز بدون فاتورة وتقرير
  | 'DONE_NO_INVOICE'  // تم الإنجاز بدون فاتورة
  | 'DONE_NO_REPORT'   // تم الإنجاز بدون تقرير
  | 'DONE_FULL'        // تم الإنجاز بشكل كامل

export interface Booking {
  /** كم مرة انأجّل الحجز لليوم الجاي بسبب إنجاز جزئي */
  partialCount: number
  lastPartialAt: string | null
  id: string
  code: string

  // ═══ الأرشفة ═══ الحجز «المحذوف» ما يروح: يختفي من الشاشات ويضل
  // بالأرشيف بسبب حذفه ومنو حذفه.
  archivedAt: string | null
  archivedById: string | null
  archiveReason: string | null

  // ═══ في الانتظار ═══ عدد المحاولات يفرّق بين زبون ما رد مرة وزبون
  // ما رد خمس مرات — الثاني قرار مو انتظار.
  waitingSince: string | null
  waitingNote: string | null
  contactAttempts: number
  lastContactAttemptAt: string | null

  // ═══ التأجيل ═══ حجز تأجل أربع مرات علامة على شي غلط ولازم ينشاف.
  postponeCount: number
  lastPostponedAt: string | null
  postponeReason: string | null

  workStoppedAt: string | null
  workStopReason: string | null
  hasInvoice: boolean
  hasReport: boolean
  completionState: CompletionState
  sequenceNumber: number | null
  scheduledAt: string | null
  // نهاية المدى المتفق عليه مع الزبون (ساعة بعد البداية) — الموعد
  // ينحكى للزبون «من ... إلى ...» مو وقت واحد.
  scheduledEndAt: string | null
  scheduleLogs: { id: string; changedById: string; changedBy: { id: string; name: string; role: string }; oldTime: string | null; newTime: string; kind: string; reason: string | null; createdAt: string }[]
  materialsReadyAt: string | null
  materialsReadyBy: { id: string; name: string } | null
  responseMinutes: number | null
  arrivedAt: string | null
  startedAt: string | null
  confirmationContactedAt: string | null
  // وقت تحويل الحجز لتنسيق الحجوزات (التثبيت)
  confirmedAt: string | null
  // الموقع: عنوان كلامي + نقطة على الخريطة + رابط (الرابط يغني عن التحديد)
  locationUrl: string | null
  confirmationContactedBy: { id: string; name: string } | null
  systemType: string | null
  systemCount: number | null
  deviceCount: number | null
  bookingType: string
  customer: Customer | null
  service: Service | null
  // كل الخدمات المطلوبة بنفس الحجز (الزبون ممكن يطلب أكثر من منظومة)
  services?: Service[]
  transferEmployee: Employee | null
  projectSupervisor: Employee | null
  expenseResponsible: Employee | null
  expenseResponsibleId: string | null
  confirmedByEmployee: Employee | null
  lastEditedBy: { id: string; name: string } | null
  lastEditedAt: string | null
  notes: string | null
  vehicleType: string | null
  priority: 'NORMAL' | 'URGENT'
  // WAITING: اتصلنا بالزبون بعد التثبيت وما رد — الحجز ينزاح من طابور
  // الشغل ويضل محفوظ لحد ما يرد.
  status: 'PENDING' | 'CONFIRMED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED' | 'WAITING' | 'PARTIAL'
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

export interface JobDurationEstimate {
  expectedMinutes: number | null
  sampleCount: number
  minSamples: number
}

/** إحصائية الأعمال المنجزة داخل الشركة خلال شهر */
export interface InternalWorksReport {
  month: string
  inHouseCount: number
  onSiteCount: number
  inHouseAmount: number
  services: { name: string; count: number; amount: number }[]
  crew: { employeeName: string; count: number }[]
  works: { code: string; completedAt: string | null; serviceName: string; amount: number }[]
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
  serviceBreakdown: { serviceId: string | null; name: string; count: number; firstAt: string | null; lastAt: string | null; revenue: number; profit: number }[]
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
  locationUrl?: string | null
  position?: string | null
  /** كم مرة انكشف إن شكواه ما كانت صحيحة — المبيعات والإداري يشوفونها
   *  قبل ما يوعدون الزبون بشي */
  falseClaimCount?: number
  lastFalseClaimAt?: string | null
  falseClaimNote?: string | null
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
  notContactedCount: number
}

export interface BookingCrewMember {
  employeeId: string
  name: string
  role: string
  isLeader: boolean
}

/** «شنو صار بهذا الحجز فعلاً» — يقراها مهندس الجودة قبل ما يتصل */
export interface BookingExecutionDetail {
  startedAt: string | null
  completedAt: string | null
  durationMinutes: number | null
  completionNotes: string | null
  workStoppedAt: string | null
  workStopReason: string | null
  crew: BookingCrewMember[]
  progressReports: BookingProgressReport[]
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
  /** حكم الجودة */
  reportType: 'POSITIVE' | 'NEGATIVE' | null
  inspectionStatus: 'NONE' | 'PENDING' | 'DONE'
  inspectionResult: 'CUSTOMER_RIGHT' | 'CUSTOMER_WRONG' | null
  inspectionNotes: string | null
  inspectedAt: string | null
  inspectedBy: { id: string; name: string } | null
  penalizedEmployee: { id: string; name: string } | null
  execution: BookingExecutionDetail | null
  // تفاصيل المشروع والمبالغ — مهندس الجودة يحتاجها وهو يتصل بالزبون
  financials?: QualityFollowUpFinancials | null
}

function currentToken(): string | null {
  return localStorage.getItem('authToken')
}

// أي صفحة تطلق عدة طلبات بنفس الوقت. لما تنتهي الجلسة كانت كلها ترجع 401
// وكل وحدة تفتح alert وتعمل reload — فتتراكم رسائل فوق بعض والمستخدم يحس
// النظام "معلّق". هذا القفل يخلي أول 401 بس هو الي يتصرّف، والباقي ينهمل.
let sessionExpiredHandled = false

export function handleSessionExpired(reason?: string) {
  if (sessionExpiredHandled) return
  // ما اكو توكن أصلاً؟ يعني إحنا بشاشة الدخول وطلب خلفي انطلق ورجع 401.
  // لو حدّثنا الصفحة هنا ندخل بحلقة لا تنتهي: تحديث ← طلب ← 401 ← تحديث،
  // وهذا الي كان يخلي النظام يبين "معلّق" وما ينفع معه تحديث.
  if (!currentToken()) return
  sessionExpiredHandled = true
  localStorage.removeItem('authToken')
  localStorage.removeItem('currentEmployee')
  // نمرر السبب لشاشة الدخول بدل alert الي يوقف الصفحة وينتظر ضغطة
  sessionStorage.setItem('sessionEndedReason', reason || 'انتهت صلاحية الجلسة، الرجاء تسجيل الدخول مجدداً')
  window.location.reload()
}

// tokenExpiresAt يقرأ وقت انتهاء التوكن من داخله (حقل exp) بدون ما نسأل
// السيرفر — يخلينا نعرف إن الجلسة منتهية قبل ما نطلق أي طلب فاشل أصلاً.
export function tokenExpiresAt(): number | null {
  const t = currentToken()
  if (!t) return null
  try {
    const payload = JSON.parse(atob(t.split('.')[1]))
    return typeof payload.exp === 'number' ? payload.exp * 1000 : null
  } catch {
    return null
  }
}

export function isTokenExpired(): boolean {
  const exp = tokenExpiresAt()
  return exp !== null && Date.now() >= exp
}

// وسم الملفات: ينجاب مرة وحدة ويتجدد قبل ما ينتهي. وسم <img> ما يرسل
// ترويسة Authorization، فالرابط لازم يحمل الوسم بنفسه.
let fileToken = ''
let fileTokenAt = 0
export async function ensureFileToken(): Promise<string> {
  // نجدده كل 10 دقائق — صلاحيته 15، فما نوصل لحافة الانتهاء
  if (fileToken && Date.now() - fileTokenAt < 10 * 60 * 1000) return fileToken
  try {
    const res = await request<{ token: string }>('/files/token')
    fileToken = res.token
    fileTokenAt = Date.now()
  } catch {
    // فشل جلب الوسم ما يكسر الشاشة — الصور القديمة (data:) تضل تشتغل
  }
  return fileToken
}

/** fileUrl يحوّل القيمة المخزّنة لرابط يعرضه المتصفح.
 *
 *  يتعامل مع الثلاثة: data: URL قديمة، مسار /api/files/ جديد، ورابط
 *  خارجي كامل. الشاشات تنادي هذي بدل ما تحط القيمة مباشرة بـsrc. */
export function fileUrl(value?: string | null): string {
  if (!value) return ''
  if (value.startsWith('data:') || value.startsWith('http')) return value
  if (value.startsWith('/api/files/')) {
    const base = API_URL.replace(/\/api$/, '')
    return `${base}${value}${fileToken ? `?ft=${encodeURIComponent(fileToken)}` : ''}`
  }
  return value
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
    if (!path.startsWith('/auth/login')) {
      handleSessionExpired(body.error)
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

// توفر المنتج: بمخزن الشركة، أو ينطلب من المجهز وقت الحاجة
export type ProductAvailability = 'IN_STOCK' | 'ON_DEMAND'

export interface Product {
  id: string
  name: string
  unit: string | null
  defaultPrice: number | null
  wholesalePrice?: number | null
  imageBase64?: string
  availability: ProductAvailability
  availabilityLabel: string
  /** تصنيف الموظف — هو المعتمد */
  serviceId: string | null
  serviceName: string | null
  specs: string | null
  source: string | null
  modelName: string | null
  /** الخدمة كما كتبها التقني بنفسه — مو اقتراح نظام */
  serviceText: string | null
  createdById: string | null
  /** اسم التقني الي ضاف المنتج */
  createdByName: string | null
}

/** الحقول الي ينرسلها المستخدم — بدون المحسوبة بالسيرفر */
export interface ProductInput {
  name: string
  unit?: string | null
  defaultPrice?: number | null
  wholesalePrice?: number | null
  imageBase64?: string
  availability?: ProductAvailability
  serviceId?: string | null
  specs?: string | null
  source?: string | null
  modelName?: string | null
  serviceText?: string | null
}

// EmployeeMonthlyStats صف واحد بصفحة إحصائيات الموظفين الشهرية (OWNER/ADMIN فقط)
// — نفس الصف يُستخدم بالإحصائية الأسبوعية (from/to بدل month).
export interface EmployeeMonthlyStats {
  employeeId: string
  employeeName: string
  role: string
  month: string
  from?: string
  to?: string
  kpiPoints: number
  kpiPointsValue: number
  // نسبة زمن الموظف للمتوسط العام بنفس المنظومة ونوع الشغل:
  // فوق ١ أسرع، تحت ١ أبطأ. null يعني ماكو عيّنات كافية.
  workSpeedScore: number | null
  workSpeedSamples: number
  vehicleCleanlinessScore: number | null
  vehicleRatingsCount: number
  complaintsCount: number
  salesCount: number
  completedBookingsCount: number
  totalCommission: number
  totalBookingsCount: number
  maintenanceBookingsCount: number
  freeMaintenanceCount: number
  servicesKnownCount: number
  inHouseWorksCount: number
  inHouseWorkTypes: string[]
}

export interface MonthlyPointsBucket {
  month: string
  points: number
}

export interface MonthlyCommissionBucket {
  month: string
  amount: number
}

export interface EmployeePerformanceCurve {
  employeeId: string
  employeeName: string
  points: MonthlyPointsBucket[]
  commission: MonthlyCommissionBucket[]
}

export type DesignFormQuestionType = 'TEXT' | 'TEXTAREA' | 'NUMBER' | 'DATE' | 'SELECT' | 'CHECKBOX' | 'FILE'

export interface DesignForm {
  id: string
  name: string
  publicToken: string
  createdAt: string
}

export interface DesignFormQuestion {
  id: string
  formId: string
  label: string
  type: DesignFormQuestionType
  options: string[]
  required: boolean
  order: number
  createdAt: string
}

export interface DesignFormSubmission {
  id: string
  formId: string
  answers: Record<string, string | number | string[] | null>
  submittedAt: string
}

export interface DailyStats {
  date: string
  totalBookings: number
  morningBookings: number
  eveningBookings: number
  crewOutCount: number
  vehiclesOutCount: number
  totalEmployeesCount: number
  totalSalesAmount: number
  totalProfitAmount: number
  employees: {
    employeeId: string; employeeName: string; role: string
    bookingsAssigned: number; bookingsCompleted: number; checkedIn: boolean
  }[]
}

export interface WeeklyStats {
  from: string
  to: string
  morningSalesAmount: number
  eveningSalesAmount: number
  totalSalesAmount: number
  employees: EmployeeMonthlyStats[]
}

export interface ProjectStageStats {
  stage: string
  count: number
}

export interface SystemPriceCatalog {
  id: string
  systemName: string
  itemName: string
  category: 'install' | 'wiring' | 'programming'
  value: number
  createdAt: string
}

export interface Material {
  id: string
  name: string
  code: string
  sellPrice: number
  profitPerUnit: number
  createdAt: string
}

export interface ExecutionCostItem {
  systemName: string
  itemName: string
  count: number
  heightMeters: number
  wiringItemName?: string
  // ارتفاع التسليك — قاعدته ثنائية بالاكسل (٥م فما فوق = ضعف السعر)، غير
  // متدرجة مثل ارتفاع التركيب
  wiringHeightMeters?: number
  cableLengthMeters?: number
  programmingItem?: string
}

export interface LeaderInvoiceMaterialItem {
  id: string
  leaderInvoiceId: string
  materialId: string | null
  name: string
  quantity: number
  unitPrice: number
  profitPerUnit: number
  lineTotal: number
  createdAt: string
}

export interface LeaderInvoice {
  // رقم الفاتورة المحاسبية الصادرة من نظام المحاسب الخارجي — إجباري
  // وقت الاعتماد، ومؤرشف حتى يلكاها بيه لمن يحتاجها.
  externalInvoiceNumber: string | null
  externalInvoiceAt: string | null
  adjustedReason: string | null
  adjustedAt: string | null
  employeeName?: string
  employeeRole?: string
  employeePhone?: string | null
  approvedByName?: string | null
  bookingCode?: string | null
  booking?: Booking | null
  id: string
  bookingId: string | null
  employeeId: string
  customerName: string | null
  customerPhone: string | null
  customerAddress: string | null
  totalDeviceCount: number
  executionCost: number
  materialsTotal: number
  discountValue: number
  netTotal: number
  accountingCode: string
  status: string
  createdAt: string
  approvedByEmployeeId: string | null
  approvedAt: string | null
  systems: string[]
  items: ExecutionCostItem[]
  materials: LeaderInvoiceMaterialItem[]
}

export interface BookingDeleteRequest {
  id: string
  bookingId: string
  reason: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  statusLabel: string
  decisionNote: string | null
  createdAt: string
  bookingCode: string
  customerName: string
  bookingStatus: string
  requestedByName: string
  decidedByName: string | null
}

/** التدقيق اليومي: حجوزات يوم واحد بمجاميعه الأربعة */
export interface DailyAuditRow {
  id: string
  code: string
  status: string
  customerName: string
  customerPhone: string
  serviceName: string
  amountVerified: boolean
  collected: number
  quotedPrice: number
  invoiceTotal: number | null
  invoiceCode: string | null
  expectedAmount: number
  openIssues: number
}

export interface DailyAuditReport {
  date: string
  rows: DailyAuditRow[]
  completedCount: number
  pendingCount: number
  issuesCount: number
  collectedTotal: number
  notVerifiedTotal: number
  verifiedTotal: number
  allAmountsTotal: number
  expectedTotal: number
}

/** بلاغ خطأ سجّله المحاسب أثناء التدقيق */
export interface AuditIssue {
  id: string
  bookingId: string
  kind: 'MISMATCH' | 'PRICE_ERROR'
  kindLabel: string
  routedTo: string
  note: string | null
  expectedAmount: number | null
  actualAmount: number | null
  status: 'OPEN' | 'RESOLVED'
  createdAt: string
  bookingCode: string
  customerName: string
  raisedByName: string
}

export interface Announcement {
  id: string
  body: string
  active: boolean
  createdAt: string
  createdByName: string
}

export interface VipCustomer {
  id: string
  customerId: string
  bookingId: string | null
  requestSummary: string | null
  customerPosition: string | null
  boughtFromUs?: boolean
  sourceLabel?: string
  note: string | null
  markedByEmployeeId: string
  createdAt: string
  customerName: string
  customerPhone: string
  bookingCode: string | null
  markedByName: string
}

export interface ProjectWorkType {
  id: string
  name: string
  createdAt: string
}

// تفصيل حساب بند واحد — يُعرض للّيدر حتى يشوف من وين طلع كل رقم
export interface ExecutionCostBreakdownLine {
  systemName: string
  itemName: string
  count: number
  unitInstallPrice: number
  heightMeters: number
  heightMultiplier: number
  installTotal: number
  wiringItemName: string
  wiringMultiplier: number
  wiringHeightMeters: number
  wiringHeightWeight: number
  cableLengthMeters: number
  wiringPricePerMeter: number
  wiringByDeviceCount: number
  wiringByCableLength: number
  wiringBasis: string
  wiringTotal: number
  programmingItem: string
  programmingTotal: number
  lineTotal: number
}

// تفصيل تطبيق الحدود الدنيا لكل منظومة (صفَّي G59 و R59 بالاكسل)
export interface ExecutionCostSystemMinimum {
  systemName: string
  deviceCount: number
  installWiringCalculated: number
  installMinimumPerDevice: number
  installMinimumTotal: number
  installApplied: number
  installFloorUsed: boolean
  programmingCount: number
  programmingCalculated: number
  programmingMinimum: number
  programmingApplied: number
  programmingFloorUsed: boolean
}

export interface EstimateExecutionCostResponse {
  executionCost: number
  totalDeviceCount: number
  breakdown: ExecutionCostBreakdownLine[]
  systemMinimums: ExecutionCostSystemMinimum[]
}

// ── استمارة حساب تكلفة كاميرات المراقبة (شيت مستقل بمعادلة مختلفة) ──
// مشروع موجّه للموظف — يستخدمه الليدر لاختيار الشغل الي راح يسويله فاتورة
export interface DirectedProject {
  id: string
  code: string
  name: string
  stage: string
  phone: string | null
  location: string | null
  bookingId: string | null
  delegatedToName: string | null
}

export interface CameraCostRow {
  normalCableMeters: number
  vipCableMeters: number
  heightAbove3m: boolean
}

export interface CameraCostExtras {
  screenLarge43Count: number
  screenSmall43Count: number
  rackCount: number
  boardCount: number
  vipInternetMeters: number
  normalInternetMeters: number
  programmingAmount: number
  otherAmount: number
}

export interface CameraCostRequest {
  placeType: string
  systemType: string
  rows: CameraCostRow[]
  extras: CameraCostExtras
  discount: number
}

export interface CameraCostRowResult {
  index: number
  basePrice: number
  placeMultiplier: number
  afterPlace: number
  systemMultiplier: number
  afterSystem: number
  heightMultiplier: number
  total: number
  countsAsCamera: boolean
}

export interface CameraCostResponse {
  rows: CameraCostRowResult[]
  cameraCount: number
  camerasTotal: number
  extrasTotal: number
  discount: number
  finalAmount: number
  note: string
}

export interface CreateMaterialLineRequest {
  materialCode?: string
  name?: string
  quantity: number
  unitPrice?: number
  profitPerUnit?: number
}

export interface CreateLeaderInvoiceRequest {
  bookingId?: string
  customerName?: string
  customerPhone?: string
  customerAddress?: string
  systems: string[]
  items: ExecutionCostItem[]
  materials: CreateMaterialLineRequest[]
  discountValue: number
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

/** سيارة بالقائمة المنسدلة — الاسم والرقم واللون بس */
export interface VehicleOption {
  id: string
  name: string
  plateNumber: string
  color: string | null
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

export interface VehicleMissionRating {
  id: string
  missionId: string
  ratedById: string
  commitment: number
  vehicleCare: number
  driving: number
  cleanliness: number
  notes: string | null
  createdAt: string
  ratedBy: { id: string; name: string } | null
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
  rating?: VehicleMissionRating | null
  // فقط عند إرجاع استجابة بدء مهمة جديدة (startVehicleMission) — يوجّه الواجهة
  // تعرض خطوة فحص أدوات المركبة العامة، ومقصور على الليدر (isLeader) فقط.
  requiresToolCheck?: boolean
  bookingWarning?: string | null
}

export interface DriverRatingSummary {
  employeeId: string
  ratingsCount: number
  avgCommitment: number
  avgVehicleCare: number
  avgDriving: number
  avgCleanliness: number
  avgOverall: number
}

export interface VehicleBooking {
  id: string
  vehicleId: string
  requestedById: string
  purpose: string
  startAt: string
  endAt: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'CANCELLED'
  approvedById: string | null
  rejectionReason: string | null
  createdAt: string
  decidedAt: string | null
  vehicle: Vehicle | null
  requestedBy: { id: string; name: string } | null
  approvedBy: { id: string; name: string } | null
}

export interface VehicleLog {
  id: string
  vehicleId: string
  type: 'FUEL' | 'CLEANING' | 'OIL_CHANGE' | 'MAINTENANCE'
  performedAt: string
  nextDueAt: string | null
  nextDueOdometer: number | null
  odometer: number | null
  cost: number | null
  notes: string | null
  createdAt: string
  recordedBy: { id: string; name: string } | null
  // تفاصيل تعبئة الوقود
  liters: number | null
  filledByEmployeeId: string | null
  filledByName: string | null
  receiptNumber: string | null
  stationName: string | null
  // القائمة ترجع العلم فقط — الصورة تنجلب بـgetVehicleLogReceiptPhoto عند الطلب
  hasReceiptPhoto: boolean
}

export interface EmployeeFuelStat {
  employeeId: string
  employeeName: string
  fillCount: number
  totalLiters: number
  totalCost: number
}

export interface VehicleIncidentAttachment {
  id: string
  incidentId: string
  url: string
  mediaType: 'IMAGE' | 'VIDEO'
  createdAt: string
}

export interface VehiclePart {
  id: string
  vehicleId: string
  partType: 'TIRE' | 'BATTERY'
  installedAt: string
  installedOdometer: number
  expectedLifespanKm: number | null
  expectedLifespanMonths: number | null
  notes: string | null
  replacedAt: string | null
  cost: number | null
  createdAt: string
  dueSoon: boolean
}

export interface VehicleAlert {
  vehicleId: string
  vehicleName: string
  alertType: 'MAINTENANCE' | 'PART' | 'DOCUMENT' | 'FUEL_ANOMALY'
  message: string
  severity: 'warning' | 'danger'
}

export interface FuelAnomalyResult {
  isAnomaly: boolean
  averageCost: number
  newCost: number
  percentAboveAvg: number
}

export interface VehicleExpenseSummary {
  vehicleId: string
  period: string
  fuelCost: number
  maintenanceCost: number
  partsCost: number
  incidentCost: number
  cleaningCost: number
  totalCost: number
  distanceKm: number | null
  avgCostPerKm: number | null
}

export interface VehicleExpenseRow {
  vehicleId: string
  vehicleName: string
  plateNumber: string
  totalCost: number
  fuelCost: number
}

export interface VehicleUsageRankRow {
  vehicleId: string
  vehicleName: string
  plateNumber: string
  missionCount: number
  distanceKm: number
  totalCost: number
}

export interface FleetDashboardSummary {
  period: string
  totalVehicles: number
  activeVehiclesCount: number
  inMaintenanceCount: number
  onMissionCount: number
  needsServiceCount: number
  expiringDocsCount: number
  alerts: VehicleAlert[]
  fleetFuelCostThisMonth: number
  fleetTotalCostThisMonth: number
  vehicleExpenses: VehicleExpenseRow[]
  topByUsage: VehicleUsageRankRow[]
  topByCost: VehicleUsageRankRow[]
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
  type: 'FAULT' | 'DAMAGE' | 'ACCIDENT'
  description: string
  cost: number | null
  status: 'OPEN' | 'RESOLVED'
  createdAt: string
  resolvedAt: string | null
  responsibleEmployee: { id: string; name: string } | null
  reportedBy: { id: string; name: string } | null
  location: string | null
  peoplePresent: string | null
  policeReportNumber: string | null
  repairCost: number | null
  driver: { id: string; name: string } | null
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
  booking: { id: string; code: string; customerName: string; customerPhone?: string | null } | null
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

export interface BookingToolCheck {
  id: string
  bookingId: string
  employeeId: string
  employee?: { id: string; name: string } | null
  missingItems: string | null
  checkedAt: string
}

export interface VehicleTool {
  id: string
  name: string
  // الباركود صار اختياري — الكمية أخذت محله لأن نفس الأداة تتكرر بنفس السيارة
  barcode: string | null
  quantity: number
  vehicleId: string
  vehicleName: string
  vehiclePlate: string
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
  createdAt: string
}

export interface OnDemandTool {
  id: string
  name: string
  barcode: string
  totalQuantity: number
  availableQuantity: number
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
}

// العدة القياسية — قائمة رئيسية بأسماء الأدوات الشخصية الي كل موظف لازم يكون
// عنده إياها. إضافة عنصر جديد تطبّق فوراً على كل الموظفين الحاليين، وأي موظف
// جديد ياخذها تلقائياً وقت إنشاء حسابه.
export interface PersonalToolTemplateItem {
  id: string
  name: string
  createdAt: string
}

// لقطة أدوات المركبة العامة الناقصة عند بدء مهمة من قبل ليدر (نفس فكرة
// BookingToolCheck بس لأدوات المركبة، ومقصورة على الليدر فقط).
export interface VehicleToolCheck {
  id: string
  vehicleId: string
  missionId: string
  employeeId: string
  missingToolNames: string | null
  createdAt: string
  employee?: { id: string; name: string; position: string | null }
}

// صيانة الأجهزة العامة (شيت "صيانة الاجهزة") — منفصلة عن صيانة اشتراكات GPS
export interface DeviceMaintenanceTicket {
  id: string
  appointmentDate: string | null
  customerId: string
  customer?: Customer
  deviceTypeName: string
  problem: string
  deviceSerial: string | null
  receivedAt: string | null
  deliveredAt: string | null
  invoiceNumber: string
  employeeId: string
  employee?: { id: string; name: string }
  createdAt: string
  status: 'NEW' | 'IN_PROGRESS' | 'DELIVERED'
}

// جرد الفريق ("جرد العدد")
export interface TeamInventoryToolCatalogItem {
  id: string
  name: string
  createdAt: string
}

export type TeamInventoryPersonRole = 'LEADER' | 'EMPLOYEE1' | 'EMPLOYEE2'
export type TeamInventoryShortageReason = 'FORGOTTEN' | 'DAMAGED' | 'UNKNOWN'

export interface TeamInventoryCheckItem {
  id: string
  checkId: string
  toolName: string
  personRole: TeamInventoryPersonRole
  present: boolean
  reason: TeamInventoryShortageReason | null
}

export interface TeamInventoryCheck {
  id: string
  leaderId: string
  employee1Id: string | null
  employee2Id: string | null
  createdAt: string
  leader?: { id: string; name: string }
  employee1?: { id: string; name: string } | null
  employee2?: { id: string; name: string } | null
  items: TeamInventoryCheckItem[]
}

export type ToolRequest = ToolRequestItem
export interface ToolRequestItem {
  id: string
  employeeId: string
  employee?: { id: string; name: string }
  toolId: string
  tool?: OnDemandTool
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'RETURNED'
  reason: ToolRequestReason | null
  reasonLabel: string
  requestKind?: string | null
  kindLabel?: string
  description: string | null
  purchasePrice: number | null
  procurementRequestId: string | null
  approvedById: string | null
  requestedAt: string
  approvedAt: string | null
  returnedAt: string | null
}

export interface PrivacyPolicyPoint {
  id: string
  content: string
  order: number
  isActive: boolean
  createdByEmployeeId: string | null
  // اسم الي أضاف النقطة — يرجع للمالك ومدير النظام فقط
  createdByName: string | null
  createdAt: string
  updatedAt: string
}

export interface PrivacyPolicyStatus {
  needsAcceptance: boolean
  points: PrivacyPolicyPoint[]
  acceptedAt: string | null
}

export type PersonalToolStatus = 'AVAILABLE' | 'LOST' | 'DAMAGED' | 'REPAIRING' | 'RETIRED' | 'CHECKED_OUT'

// نفس النصوص الموجودة بالباك إند (model/inventory.go)
export const personalToolStatusLabels: Record<PersonalToolStatus, string> = {
  AVAILABLE: 'موجودة',
  LOST: 'مفقودة',
  DAMAGED: 'تالفة',
  REPAIRING: 'بالتصليح',
  RETIRED: 'خارج الخدمة',
  CHECKED_OUT: 'مصروفة',
}

export const personalToolStatusColors: Record<PersonalToolStatus, string> = {
  AVAILABLE: 'bg-emerald-50 text-emerald-700',
  LOST: 'bg-red-100 text-red-700',
  DAMAGED: 'bg-amber-100 text-amber-700',
  REPAIRING: 'bg-blue-50 text-blue-700',
  RETIRED: 'bg-slate-100 text-slate-500',
  CHECKED_OUT: 'bg-indigo-50 text-indigo-700',
}

export interface PersonalToolEvent {
  id: string
  toolId: string
  toolName: string
  employeeId: string
  employeeName: string | null
  eventType: 'CREATED' | 'STATUS_CHANGED' | 'RENAMED' | 'CHECKED_OUT' | 'RETURNED' | 'DELETED'
  eventLabel: string
  fromStatus: string | null
  toStatus: string | null
  fromStatusText: string
  toStatusText: string
  note: string | null
  actorId: string | null
  actorName: string | null
  createdAt: string
}

export type ToolRequestReason =
  | 'DAMAGED' | 'LOST' | 'WORN' | 'STOLEN' | 'NEVER_HAD' | 'EXTRA' | 'OTHER'

// نفس النصوص الموجودة بالباك إند (model/inventory.go) — أهم سببين أول القائمة
// لأنهم الأكثر استعمالاً: الأداة تالفة أو ضايعة.
export const toolRequestReasonLabels: Record<ToolRequestReason, string> = {
  DAMAGED: 'الأداة الي عندي تالفة',
  LOST: 'الأداة الي عندي ضايعة',
  WORN: 'الأداة مستهلكة من كثر الاستعمال',
  STOLEN: 'الأداة مسروقة',
  NEVER_HAD: 'ما عندي هذي الأداة أصلاً',
  EXTRA: 'أحتاج نسخة إضافية لطبيعة الشغل',
  OTHER: 'سبب آخر',
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
  /** حالة الاتصال بالزبون — منو اتصل ومتى */
  contactedAt: string | null
  contactedByName: string | null
  notes: string | null
  createdAt: string
  resolvedAt: string | null
}

export type ProcurementRequestType = 'PERSONAL_SUPPLY' | 'CUSTOMER_PRODUCT' | 'MANUAL_SUPPLY'

export const procurementTypeLabels: Record<ProcurementRequestType, string> = {
  PERSONAL_SUPPLY: 'احتياجات شخصية',
  CUSTOMER_PRODUCT: 'منتج للزبون',
  MANUAL_SUPPLY: 'طلب مادة يدوي',
}

export interface ProcurementItem {
  id: string
  requestId: string
  productName: string
  quantity: number
  unitPrice: number | null
  totalPrice: number | null
  fulfilled: boolean
  /** تفاصيل الطلب اليدوي */
  model?: string | null
  spec?: string | null
  itemNotes?: string | null
}

export interface ProcurementRequest {
  id: string
  code: string
  requestedBy: { id: string; name: string; role: string }
  requestedById: string
  booking: { id: string; code: string; customer: { id: string; name: string; phone: string } | null } | null
  bookingId: string | null
  requestType: ProcurementRequestType
  typeLabel: string
  supplierId: string | null
  supplier: { id: string; name: string } | null
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

export interface ProjectChecklist {
  id: string
  projectId: string | null
  project: { id: string; name: string; code: string } | null
  title: string
  createdBy: { id: string; name: string } | null
  photoUrls: string[]
  createdAt: string
}

export interface TechShowcaseItem {
  id: string
  title: string
  description: string | null
  employee: { id: string; name: string } | null
  mediaUrls: string[]
  createdAt: string
}

export interface Exhibition {
  id: string
  title: string
  location: string
  startDate: string
  endDate: string
  companies: string[]
  productsToShow: string[]
  nominatedEmployeeIds: string[]
  nominatedEmployees: { id: string; name: string }[]
  businessCardPhotos: string[]
  keyFindings: string | null
  visitReport: string | null
  archived: boolean
  createdBy: { id: string; name: string } | null
  createdAt: string
}

export interface ProductRequest {
  id: string
  productName: string
  specs: string | null
  source: string | null
  model: string | null
  category: string | null
  price: number | null
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  requestedBy: { id: string; name: string } | null
  resolvedBy: { id: string; name: string } | null
  createdAt: string
  resolvedAt: string | null
}

/** منو يعوّض الدوار عن المنتج المشترى */
export type PayerKind = 'COMPANY' | 'CUSTOMER'

/** تجهيز طلب منتج من الدوار — أبو الحسابات يشتري، والمبلغ يبقى ناقص
 *  من الدوار لحد ما المحاسب يرجّعه. */
export interface ProductProcurement {
  id: string
  requestId: string
  fundId: string
  supplierId: string
  productName: string
  spentAmount: number
  reason: string
  receiptImage: string | null
  payerKind: PayerKind
  customerNote: string | null
  bookingId: string | null
  status: 'PENDING' | 'SETTLED'
  purchasedById: string
  settledById: string | null
  settledAt: string | null
  settleNote: string | null
  createdAt: string
  fundName: string
  supplierName: string
  purchasedByName: string
  settledByName: string | null
  statusLabel: string
  payerLabel: string
}

export interface FulfillProductInput {
  fundId: string
  supplierId: string
  spentAmount: number
  reason: string
  receiptImage: string
  payerKind: PayerKind
  customerNote?: string
  bookingId?: string
}

export interface ServiceStudyReport {
  id: string
  serviceStudyId: string
  content: string
  employee: { id: string; name: string } | null
  createdAt: string
}

export interface ServiceStudy {
  id: string
  name: string
  archived: boolean
  createdBy: { id: string; name: string } | null
  assignedEmployees: { id: string; name: string }[]
  reports: ServiceStudyReport[]
  createdAt: string
}

// ═══ نقاط الانضباط ═══
export interface DisciplinePoints {
  employeeId: string
  employeeName: string
  points: number
  deductedDinar: number
  lastRestoredAt: string | null
  updatedAt: string
}
export interface DisciplineEvent {
  byEmployeeId?: string | null
  id: string
  employeeId: string
  employeeName: string
  bookingId: string | null
  kind: 'LATE_PAPERWORK' | 'UNBALANCED_ASSIGNMENT' | 'RESTORE' | 'MANUAL'
  delta: number
  reason: string
  createdAt: string
}

/** مجاميع لوحة المحاسب والمراقب — كلها محسوبة على الأرشيف الكامل بالسيرفر */
export interface FinanceSummary {
  completedCount: number
  unverifiedCount: number
  verifiedCount: number
  todayCompleted: number
  pendingCount: number
  confirmedCount: number
  inProgressCount: number
  activeCrewCount: number
  totalCollected: number
  totalQuoted: number
  totalCartValue: number
  pendingExpenses: number
  approvedExpenses: number
  totalExpenseValue: number
}


// ═══════════ نظام الطاقة الشمسية ═══════════
// منقول من نظام Solar Expert الي جان على Google Sheets، ومبني من جديد
// بالـGo داخل نظامنا.

export type SolarCategory = 'PANEL' | 'INVERTER' | 'BATTERY' | 'BOARD' | 'IRON'

export const SOLAR_CATEGORY_LABELS: Record<SolarCategory, string> = {
  PANEL: 'لوح شمسي',
  INVERTER: 'إنفيرتر',
  BATTERY: 'بطارية',
  BOARD: 'بورد حماية',
  IRON: 'حديد وحدادة',
}

/** مواصفات كل تصنيف — نفس حقول النظام القديم بس بمفاتيح واضحة بدل specA..specK */
export const SOLAR_SPEC_FIELDS: Record<SolarCategory, string[]> = {
  PANEL: ['القدرة (W)', 'النوع (Mono/Poly)', 'Vmp (V)', 'Imp (A)', 'Voc (V)', 'Isc (A)', 'الكفاءة %', 'الضمان (سنوات)', 'الأبعاد (mm)', 'الوزن (kg)', 'ضمان الأداء (سنوات)'],
  INVERTER: ['القدرة (KW)', 'النوع (On/Off/Hybrid)', 'جهد البطارية (V)', 'أقصى شحن (A)', 'عدد MPPT', 'نطاق PV (V)', 'الكفاءة %', 'الضمان (سنوات)', 'جهد الخرج (VAC)', 'درجة الحماية (IP)', 'الاتصال (WiFi/RS485)'],
  BATTERY: ['السعة (Ah)', 'النوع (LiFePO4/Lead)', 'الجهد (V)', 'أقصى شحن (A)', 'عدد الدورات', 'DOD %', 'الكفاءة %', 'الضمان (سنوات)', 'الأبعاد (mm)', 'الوزن (kg)', 'أقصى تفريغ (A)'],
  BOARD: ['الأمبير (A)', 'النوع (AC/DC Combo)', 'عدد الفيوزات', 'نوع القاطع (MCB/SPD)', 'الضمان (سنوات)', 'الأبعاد (mm)', 'درجة الحماية (IP)'],
  IRON: ['نوع الحديد', 'المقاس', 'السماكة (mm)', 'الطول القياسي (متر)', 'الوزن (kg/متر)', 'درجة الجودة', 'نوع الطلاء', 'الاستخدام المقترح'],
}

export const SOLAR_BRANDS = ['Deye', 'Felicity', 'Growatt', 'Must', 'SRNE', 'منظومات اخرى']

export const SOLAR_WIRING_TYPES = ['DC 4mm²', 'DC 6mm²', 'DC 10mm²', 'AC 2.5mm²', 'AC 4mm²', 'AC 6mm²', 'كابل اتصال']

export const SOLAR_IRON_TYPES = ['زاوية حديد 4x4', 'زاوية حديد 5x5', 'قضبان تسليح 12mm', 'قضبان تسليح 16mm', 'حديد مسطح 5cm', 'حديد مسطح 10cm', 'صاج مجلفن 2mm', 'صاج مجلفن 3mm', 'مسامير وبراغي وصواميل', 'دهانات حماية ضد الصدأ']

export const SOLAR_IRON_UNITS = ['متر', 'قطعة', 'كغ', 'مجموعة']

export interface SolarComponent {
  id: string
  name: string
  category: SolarCategory
  quantity: number
  price: number
  minStock: number
  specs: Record<string, string>
  notes: string | null
  createdAt: string
  updatedAt: string
}

export interface SolarWiringLine { type: string; length: number; price: number }
export interface SolarIronLine { type: string; qty: number; unit: string; price: number }

/** المكوّن الي ما يكفي بالمخزن — missing يعني مو موجود أصلاً، وهذا أخطر */
export interface SolarShortage {
  componentId: string
  name: string
  required: number
  available: number
  missing: boolean
}

export interface SolarPrice {
  panels: number; inverters: number; batteries: number; board: number
  wiring: number; iron: number; install: number; program: number; warranty: number
  components: number; total: number
}

export interface SolarSystem {
  id: string
  brand: string
  model: string
  capacity: string
  panelId: string | null;    panelQty: number
  inverterId: string | null; inverterQty: number
  batteryId: string | null;  batteryQty: number
  boardId: string | null
  wiringDetails: SolarWiringLine[] | null
  wiringTotalCost: number
  ironDetails: SolarIronLine[] | null
  ironTotalCost: number
  installPrice: number
  programPrice: number
  warrantyPrice: number
  notes: string | null
  createdAt: string
  updatedAt: string
  panel: SolarComponent | null
  inverter: SolarComponent | null
  battery: SolarComponent | null
  board: SolarComponent | null
  /** ينحسب من أسعار المخزن الحالية — الكتالوك يعكس سعر اليوم مو سعر يوم كتابته */
  price: SolarPrice
  shortages: SolarShortage[]
}

export interface SolarInstallation {
  id: string
  systemId: string
  customerId: string
  installDate: string
  followUpAt: string
  contactedAt: string | null
  contactNotes: string | null
  status: 'PENDING' | 'CONTACTED'
  totalPrice: number
  priceBreakdown: SolarPrice | null
  notes: string | null
  createdAt: string
  customer: { id: string; name: string; phone: string | null; location: string | null } | null
  system: SolarSystem | null
  contactedByName: string | null
  dueForFollowUp: boolean
  daysOverdue: number
}

export interface SolarStats {
  systemCount: number
  componentCount: number
  inventoryValue: number
  lowStockCount: number
  outOfStockCount: number
  processedCount: number
  customerCount: number
  dueFollowUpCount: number
  contactedCount: number
  installedThisMonth: number
  totalWiring: number
  totalIron: number
  totalInstall: number
  totalProgram: number
}

export interface SaveSolarSystemInput {
  brand: string; model: string; capacity: string
  panelId?: string | null; panelQty: number
  inverterId?: string | null; inverterQty: number
  batteryId?: string | null; batteryQty: number
  boardId?: string | null
  wiringDetails: SolarWiringLine[]
  ironDetails: SolarIronLine[]
  installPrice: number; programPrice: number; warrantyPrice: number
  notes?: string | null
}


// ═══ برامج التدريب ═══
export interface SkillWithService {
  id: string
  name: string
  category: string
  description: string | null
  serviceId: string | null
  serviceName: string | null
}

export interface TrainingParticipant {
  employeeId: string
  name: string
  department: string | null
  jobTitle: string | null
  passed: boolean | null
}

export interface TrainingProgram {
  id: string
  name: string
  level: 'مبتدئ' | 'متوسط' | 'متقدم'
  durationDays: number
  startDate: string | null
  endDate: string | null
  targetDepartment: string | null
  instructorId: string | null
  instructorName: string | null
  objectives: string | null
  content: string | null
  passRate: number
  cost: number
  status: 'قيد التخطيط' | 'جاري التنفيذ' | 'مكتمل'
  progress: number
  createdAt: string
  participants: TrainingParticipant[]
  skills: { skillId: string; name: string; category: string }[]
}

export interface SaveTrainingProgramInput {
  name: string
  level: string
  durationDays: number
  startDate: string
  targetDepartment: string
  instructorId: string
  objectives: string
  content: string
  passRate: number
  cost: number
  status: string
  progress: number
  participantIds: string[]
  skillIds: string[]
}

export const DEPARTMENTS = ['الفنية', 'المبيعات', 'الصيانة', 'الإدارة', 'المستودع']
export const REVIEW_GRADES = ['ممتاز', 'جيد جداً', 'جيد', 'يحتاج تحسين']

export const api = {
  getMe: () => request<Employee>('/auth/me'),
  // تغيير كلمة السر يبطل كل الجلسات القديمة — بضمنها توكن الجهاز الحالي.
  // السيرفر يرجّع توكن جديد ولازم نخزنه فوراً، وإلا كل طلب بعدها يطلع
  // 401 والمستخدم يشوف الشاشة واقفة كأن التغيير ما خلص.
  changePassword: async (currentPassword: string, newPassword: string) => {
    const res = await request<{ token: string }>('/auth/change-password', {
      method: 'PUT',
      body: JSON.stringify({ currentPassword, newPassword }),
    })
    if (res?.token) localStorage.setItem('authToken', res.token)
  },
  createAttendanceIconRequest: (requestedIcon: string) =>
    request<AttendanceIconRequest>('/attendance-icon-requests', { method: 'POST', body: JSON.stringify({ requestedIcon }) }),
  getPendingAttendanceIconRequests: () => request<AttendanceIconRequest[]>('/attendance-icon-requests'),
  approveAttendanceIconRequest: (id: string) => request<void>(`/attendance-icon-requests/${id}/approve`, { method: 'PUT' }),
  rejectAttendanceIconRequest: (id: string) => request<void>(`/attendance-icon-requests/${id}/reject`, { method: 'PUT' }),
  getChecklists: () => request<ProjectChecklist[]>('/checklists'),
  getTechShowcase: () => request<TechShowcaseItem[]>('/tech-showcase'),
  createTechShowcaseItem: (data: { title: string; description?: string }) =>
    request<TechShowcaseItem>('/tech-showcase', { method: 'POST', body: JSON.stringify(data) }),
  addTechShowcaseMedia: (id: string, mediaUrls: string[]) =>
    request<TechShowcaseItem>(`/tech-showcase/${id}/media`, { method: 'PUT', body: JSON.stringify({ mediaUrls }) }),

  // وحدة التقنيين — إدارة المعارض
  getExhibitions: () => request<Exhibition[]>('/exhibitions'),
  createExhibition: (data: { title: string; location: string; startDate: string; endDate: string; companies: string[]; productsToShow: string[] }) =>
    request<Exhibition>('/exhibitions', { method: 'POST', body: JSON.stringify(data) }),
  nominateExhibition: (id: string, employeeIds: string[]) =>
    request<Exhibition>(`/exhibitions/${id}/nominate`, { method: 'PUT', body: JSON.stringify({ employeeIds }) }),
  addExhibitionPhotos: (id: string, photoUrls: string[]) =>
    request<Exhibition>(`/exhibitions/${id}/photos`, { method: 'PUT', body: JSON.stringify({ photoUrls }) }),
  setExhibitionFindings: (id: string, keyFindings: string) =>
    request<Exhibition>(`/exhibitions/${id}/findings`, { method: 'PUT', body: JSON.stringify({ keyFindings }) }),
  generateExhibitionReport: (id: string) => request<Exhibition>(`/exhibitions/${id}/report`, { method: 'POST' }),
  archiveExhibition: (id: string) => request<Exhibition>(`/exhibitions/${id}/archive`, { method: 'PUT' }),

  // وحدة التقنيين — إدارة المنتجات
  getProductRequests: () => request<ProductRequest[]>('/product-requests'),
  createProductRequest: (data: { productName: string; specs?: string; source?: string; model?: string; category?: string; price?: number }) =>
    request<ProductRequest>('/product-requests', { method: 'POST', body: JSON.stringify(data) }),
  approveProductRequest: (id: string) => request<ProductRequest>(`/product-requests/${id}/approve`, { method: 'PUT' }),
  rejectProductRequest: (id: string) => request<ProductRequest>(`/product-requests/${id}/reject`, { method: 'PUT' }),

  // الموردين المضافين مسبقاً — لاختيار المورد وقت تجهيز الطلب
  /** بحث المناطق — يمر من سيرفرنا لأن خدمة الخرائط تحجب المتصفحات */
  searchPlaces: (q: string, limit = 5, proximity?: { viewbox: string; bounded: string }) => {
    const params = new URLSearchParams({ q, limit: String(limit) })
    if (proximity) { params.set('viewbox', proximity.viewbox); params.set('bounded', proximity.bounded) }
    return request<{ lat: string; lon: string; display_name: string }[]>(`/geo/search?${params}`)
  },

  getSupplierOptions: () => request<{ id: string; companyName: string; ownerName: string }[]>('/suppliers'),

  // تجهيز طلب المنتج من الدوار — أبو الحسابات (نفسه أبو الكميات)
  getProductProcurements: (status?: string) =>
    request<ProductProcurement[]>(`/product-procurements${status ? `?status=${status}` : ''}`),
  fulfillProductRequest: (id: string, data: FulfillProductInput) =>
    request<ProductProcurement>(`/product-requests/${id}/fulfill`, { method: 'POST', body: JSON.stringify(data) }),
  settleProductProcurement: (id: string, note?: string) =>
    request<ProductProcurement>(`/product-procurements/${id}/settle`, { method: 'PUT', body: JSON.stringify({ note }) }),

  // وحدة التقنيين — إدارة الخدمات
  getServiceStudies: () => request<ServiceStudy[]>('/service-studies'),
  createServiceStudy: (name: string) => request<ServiceStudy>('/service-studies', { method: 'POST', body: JSON.stringify({ name }) }),
  assignServiceStudy: (id: string, employeeIds: string[]) =>
    request<ServiceStudy>(`/service-studies/${id}/assign`, { method: 'PUT', body: JSON.stringify({ employeeIds }) }),
  addServiceStudyReport: (id: string, content: string) =>
    request<ServiceStudyReport>(`/service-studies/${id}/reports`, { method: 'POST', body: JSON.stringify({ content }) }),
  archiveServiceStudy: (id: string) => request<ServiceStudy>(`/service-studies/${id}/archive`, { method: 'PUT' }),
  getProjectsBrief: () => request<{ projects: { id: string; name: string; code: string }[] }>('/projects').then((d) => d.projects),
  createChecklist: (data: { projectId?: string | null; title: string }) =>
    request<ProjectChecklist>('/checklists', { method: 'POST', body: JSON.stringify(data) }),
  addChecklistPhotos: (id: string, photoUrls: string[]) =>
    request<ProjectChecklist>(`/checklists/${id}/photos`, { method: 'PUT', body: JSON.stringify({ photoUrls }) }),
  getServices: () => request<Service[]>('/services'),
  // توقف العمل: الليدر بدأ وما كدر يكمّل. السبب إجباري.
  stopBookingWork: (id: string, reason: string) =>
    request<Booking>(`/bookings/${id}/stop-work`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  resumeBookingWork: (id: string) =>
    request<Booking>(`/bookings/${id}/resume-work`, { method: 'PUT' }),
  getDisciplinePoints: () => request<DisciplinePoints[]>('/discipline'),
  getDisciplineEvents: (employeeId?: string) =>
    request<DisciplineEvent[]>(`/discipline/events${employeeId ? '?employeeId=' + employeeId : ''}`),
  runDisciplineSweep: () => request<DisciplinePoints[]>('/discipline/run', { method: 'POST' }),
  // تعديل يدوي على رصيد موظف — للمالك ومدير النظام. delta موجب يزيد وسالب ينقص.
  adjustDisciplinePoints: (employeeId: string, delta: number, reason: string) =>
    request<DisciplinePoints>('/discipline/adjust', {
      method: 'POST',
      body: JSON.stringify({ employeeId, delta, reason }),
    }),
  createService: (data: { name: string; category?: string; division?: 'ENGINEERING' | 'DECOR' | '' }) =>
    request<Service>('/services', { method: 'POST', body: JSON.stringify(data) }),
  createSkill: (serviceId: string, name: string) =>
    request<Skill>(`/services/${serviceId}/skills`, { method: 'POST', body: JSON.stringify({ name }) }),
  deleteService: (id: string) => request<void>(`/services/${id}`, { method: 'DELETE' }),

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
      division?: Division
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

  /** بلا وسائط يجيب الكل (متل ما كان). مع بحث/حد يفلتر بالسيرفر — بـ٥٠٠٠ زبون
   *  كان الجلب الكامل ١٫٤ ميغا بكل فتحة صفحة. */
  getCustomers: (opts?: { search?: string; limit?: number }) => {
    const p = new URLSearchParams()
    if (opts?.search) p.set('search', opts.search)
    if (opts?.limit) p.set('limit', String(opts.limit))
    const qs = p.toString()
    return request<Customer[]>(`/customers${qs ? '?' + qs : ''}`)
  },
  getCustomersByGpsService: () => request<GpsCustomerListItem[]>('/customers/gps'),
  lookupCustomer: (phone: string) =>
    request<Customer | null>(`/customers/lookup?phone=${phone}`).catch(() => null),
  createCustomer: (data: { name: string; phone: string; location?: string }) =>
    request<Customer & { existed: boolean }>('/customers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateCustomer: (id: string, data: { name: string; phone: string; location?: string | null; mapLatitude?: number | null; mapLongitude?: number | null; locationUrl?: string | null; position?: string | null }) =>
    request<Customer>(`/customers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  setComplaintContacted: (id: string, contacted: boolean) =>
    request<Complaint>(`/complaints/${id}/contact`, { method: 'PUT', body: JSON.stringify({ contacted }) }),
  setComplaintNotes: (id: string, notes: string) =>
    request<Complaint>(`/complaints/${id}/notes`, { method: 'PUT', body: JSON.stringify({ notes }) }),
  getDailyAudit: (date?: string) =>
    request<DailyAuditReport>(`/finance/daily-audit${date ? `?date=${date}` : ''}`),

  // التدقيق: قرار المحاسب — مطابق أو بلاغ خطأ ينوجّه للمعني
  auditBooking: (id: string, data: { action: 'VERIFY' | 'MISMATCH' | 'PRICE_ERROR'; amountCollected?: number; advancePaid?: number; note?: string }) =>
    request<Booking | AuditIssue>(`/bookings/${id}/audit`, { method: 'PUT', body: JSON.stringify(data) }),
  getAuditIssues: (status?: string) =>
    request<AuditIssue[]>(`/audit-issues${status ? `?status=${status}` : ''}`),
  resolveAuditIssue: (id: string) =>
    request<{ ok: boolean }>(`/audit-issues/${id}/resolve`, { method: 'PUT' }),

  // شريط الإعلانات
  getAnnouncements: (all?: boolean) =>
    request<Announcement[]>(`/announcements${all ? '?all=1' : ''}`),
  createAnnouncement: (body: string) =>
    request<Announcement>('/announcements', { method: 'POST', body: JSON.stringify({ body }) }),
  setAnnouncementActive: (id: string, active: boolean) =>
    request<{ ok: boolean }>(`/announcements/${id}/active`, { method: 'PUT', body: JSON.stringify({ active }) }),
  deleteAnnouncement: (id: string) =>
    request<void>(`/announcements/${id}`, { method: 'DELETE' }),

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
      // الحسابات المحظورة تلقائياً — المالك وحده يفك حظرها
      lockedEmployees: {
        id: string
        name: string
        username: string | null
        role: string
        lockedAt: string | null
        lockedReason: string | null
        lockedDetail: string | null
        failedLoginStreak: number
        authzViolations: number
      }[] | null
      // حسابات المالك/مدير النظام الي تتعرض لتخمين — مستثناة من الحظر
      // التلقائي فما تظهر بقائمة المحظورين، بس لازم المالك يشوفها
      employeesUnderAttack: {
        id: string
        name: string
        username: string | null
        role: string
        failedLoginStreak: number
        authzViolations: number
      }[] | null
      // سجل الأحداث الأمنية (محاولات دخول فاشلة، وصول غير مخوّل، حظر، فك حظر،
      // تغيير صلاحيات)
      securityEvents: {
        id: string
        employeeId: string | null
        employeeName: string | null
        kind: string
        detail: string | null
        ip: string | null
        userAgent: string | null
        createdAt: string
      }[] | null
    }>('/security/dashboard'),
  unlockEmployee: (id: string) =>
    request<{ ok: boolean }>(`/security/unlock/${id}`, { method: 'POST' }),
  resetLoginAttempts: (id: string) =>
    request<{ ok: boolean }>(`/security/reset-attempts/${id}`, { method: 'POST' }),
  freeServerMemory: () => request<{ memoryUsedMB: number }>('/security/free-memory', { method: 'POST' }),

  askAssistant: (message: string) =>
    request<{ reply: string }>('/assistant/ask', { method: 'POST', body: JSON.stringify({ message }) }),
  managerChatAssistant: (message: string, history: { role: 'user' | 'assistant'; text: string }[]) =>
    request<{ reply: string }>('/assistant/manager-chat', { method: 'POST', body: JSON.stringify({ message, history }) }),

  getAssistantConversations: (params: { employeeId?: string; from?: string; to?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams()
    if (params.employeeId) q.set('employeeId', params.employeeId)
    if (params.from) q.set('from', params.from)
    if (params.to) q.set('to', params.to)
    if (params.limit) q.set('limit', String(params.limit))
    if (params.offset) q.set('offset', String(params.offset))
    return request<{
      conversations: {
        id: string
        employeeId: string
        message: string
        reply: string
        createdAt: string
        employee: { id: string; name: string; position: string | null } | null
      }[]
      total: number
    }>(`/assistant/conversations?${q.toString()}`)
  },
  getAssistantConversationEmployees: () =>
    request<{ employees: { id: string; name: string; position: string | null }[] }>('/assistant/conversations/employees'),

  getQualityFollowUps: () => request<QualityFollowUp[]>('/quality-follow-ups'),
  /** حكم الجودة: إيجابي (ما يترتب عليه شي) أو سلبي (يخصم نقطة من الليدر،
   *  إلا إذا طلبنا كشف — وقتها الخصم ينتظر النتيجة) */
  qualityVerdict: (id: string, body: { reportType: 'POSITIVE' | 'NEGATIVE'; notes: string; needsInspection: boolean }) =>
    request<QualityFollowUp[]>(`/quality-follow-ups/${id}/verdict`, { method: 'POST', body: JSON.stringify(body) }),
  /** نتيجة الكشف الميداني: كلام الزبون صح (يتغرّم الليدر) أو كذب (علامة على الزبون) */
  qualityInspect: (id: string, body: { result: 'CUSTOMER_RIGHT' | 'CUSTOMER_WRONG'; notes: string }) =>
    request<QualityFollowUp[]>(`/quality-follow-ups/${id}/inspect`, { method: 'POST', body: JSON.stringify(body) }),

  updateQualityFollowUp: (
    id: string,
    data: { status: QualityFollowUp['status']; contactNotes?: string },
  ) =>
    request<QualityFollowUp>(`/quality-follow-ups/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  // assignedTo: 'me' يرجّع مهام الموظف الحالي بس — الفلترة بقاعدة
  // البيانات بدل ما ننزّل كل حجوزات الشركة ونفلترها بالمتصفح.
  // ═══ أرشيف الحجوزات ═══
  // «الحذف» ما يمحي: الحجز يختفي من الحجوزات ومن تنسيق الحجوزات ويجي
  // هنا بسبب حذفه ومنو حذفه — حتى نكدر نجاوب «شكد حجز انلغى وليش؟».
  // ═══ برامج التدريب ═══
  getSkills: () => request<SkillWithService[]>('/skills'),
  getTrainingPrograms: (status?: string) =>
    request<TrainingProgram[]>(`/training-programs${status ? `?status=${encodeURIComponent(status)}` : ''}`),
  createTrainingProgram: (data: SaveTrainingProgramInput) =>
    request<TrainingProgram>('/training-programs', { method: 'POST', body: JSON.stringify(data) }),
  updateTrainingProgram: (id: string, data: SaveTrainingProgramInput) =>
    request<TrainingProgram>(`/training-programs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  /** الإكمال يمنح مهارات البرنامج لكل المشاركين فعلاً */
  completeTrainingProgram: (id: string) =>
    request<TrainingProgram>(`/training-programs/${id}/complete`, { method: 'PUT' }),
  deleteTrainingProgram: (id: string) =>
    request<{ ok: boolean }>(`/training-programs/${id}`, { method: 'DELETE' }),

  // ═══ الطاقة الشمسية ═══
  getSolarStats: () => request<SolarStats>('/solar/stats'),
  getSolarLowStock: () => request<SolarComponent[]>('/solar/low-stock'),

  getSolarComponents: (category?: SolarCategory) =>
    request<SolarComponent[]>(`/solar/components${category ? `?category=${category}` : ''}`),
  createSolarComponent: (data: Omit<SolarComponent, 'id' | 'createdAt' | 'updatedAt'>) =>
    request<SolarComponent>('/solar/components', { method: 'POST', body: JSON.stringify(data) }),
  updateSolarComponent: (id: string, data: Omit<SolarComponent, 'id' | 'createdAt' | 'updatedAt'>) =>
    request<SolarComponent>(`/solar/components/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteSolarComponent: (id: string) =>
    request<{ ok: boolean }>(`/solar/components/${id}`, { method: 'DELETE' }),

  getSolarSystems: (brand?: string) =>
    request<SolarSystem[]>(`/solar/systems${brand ? `?brand=${encodeURIComponent(brand)}` : ''}`),
  getSolarSystem: (id: string) => request<SolarSystem>(`/solar/systems/${id}`),
  createSolarSystem: (data: SaveSolarSystemInput) =>
    request<SolarSystem>('/solar/systems', { method: 'POST', body: JSON.stringify(data) }),
  updateSolarSystem: (id: string, data: SaveSolarSystemInput) =>
    request<SolarSystem>(`/solar/systems/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteSolarSystem: (id: string) =>
    request<{ ok: boolean }>(`/solar/systems/${id}`, { method: 'DELETE' }),

  /** تجهيز منظومة لزبون — يخصم مكوّناتها من المخزن. يرمي خطأ لو المخزن ما يكفي. */
  processSolarSystem: (id: string, data: { customerId?: string; customerName: string; customerPhone: string; customerAddress: string; installDate: string; notes?: string }) =>
    request<SolarInstallation>(`/solar/systems/${id}/process`, { method: 'POST', body: JSON.stringify(data) }),

  getSolarInstallations: (status?: 'PENDING' | 'CONTACTED') =>
    request<SolarInstallation[]>(`/solar/installations${status ? `?status=${status}` : ''}`),
  markSolarContacted: (id: string, notes?: string) =>
    request<SolarInstallation>(`/solar/installations/${id}/contacted`, { method: 'PUT', body: JSON.stringify({ notes: notes ?? '' }) }),

  // ── النسخ الاحتياطية: للمالك وحده ──
  // المسار محمي بـRequireOwner ويرجّع 404 لأي حساب ثاني (حتى ADMIN) —
  // مقصود، حتى ما ينكشف وجود الميزة أصلاً.
  getOwnerBackups: (limit?: number) =>
    request<BackupOverview>(`/owner/backups${limit ? `?limit=${limit}` : ''}`),

  getArchivedBookings: (limit?: number) =>
    request<Booking[]>(`/bookings/archived${limit ? `?limit=${limit}` : ''}`),
  archiveBooking: (id: string, reason: string) =>
    request<Booking>(`/bookings/${id}`, { method: 'DELETE', body: JSON.stringify({ reason }) }),
  restoreBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/restore`, { method: 'PUT' }),

  /** تأجيل الموعد بطلب الزبون — السبب إجباري وينعد بعدد التأجيلات */
  postponeBooking: (id: string, scheduledAt: string, reason: string) =>
    request<Booking>(`/bookings/${id}/postpone`, { method: 'PUT', body: JSON.stringify({ scheduledAt, reason }) }),

  /** اتصلنا بالزبون وما رد — الحجز ينزاح من طابور الشغل ويضل محفوظ */
  markBookingWaiting: (id: string, note?: string) =>
    request<Booking>(`/bookings/${id}/waiting`, { method: 'PUT', body: JSON.stringify({ note: note ?? '' }) }),
  /** الزبون رد — يرجع للطابور، وعدد المحاولات يبقى مسجّل */
  resumeBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/resume`, { method: 'PUT' }),

  getBookings: (params?: { status?: Booking['status'] | Booking['status'][]; customerId?: string; date?: string; assignedTo?: 'me'; limit?: number }) => {
    const query = new URLSearchParams()
    if (params?.assignedTo) query.set('assignedTo', params.assignedTo)
    if (params?.status) query.set('status', Array.isArray(params.status) ? params.status.join(',') : params.status)
    if (params?.customerId) query.set('customerId', params.customerId)
    if (params?.date) query.set('date', params.date)
    if (params?.limit) query.set('limit', String(params.limit))
    const qs = query.toString()
    return request<Booking[]>(`/bookings${qs ? `?${qs}` : ''}`)
  },
  // bookingType: 'INTERNAL' يخلي الحجز شغل داخل الشركة (للإداري فما فوق)
  createBooking: (data: {
    customerId: string
    serviceId?: string
    // خدمات متعددة بنفس الحجز — الأولى تنعتبر الرئيسية
    serviceIds?: string[]
    locationUrl?: string
    notes?: string
    vehicleType?: string
    priority?: 'NORMAL' | 'URGENT'
    transferEmployeeId?: string
    address?: string
    mapLatitude?: number
    mapLongitude?: number
    bookingType?: 'REGULAR' | 'MAINTENANCE' | 'INTERNAL' | 'SOLAR'
    // حجز طاقة شمسية: المنظومة من الكتالوك (اختيارية) واستهلاك الزبون
    // الشهري. السعر المقدّر ينحسب بالسيرفر من مكوّنات المنظومة.
    solarSystemId?: string
    solarMonthlyKwh?: number
    internalEmployeeName?: string
    internalEmployeePhone?: string
    internalDepartment?: string
    internalApproved?: boolean
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
    data: { quotedPrice?: number | null; address?: string; assignedVehicle?: string; mapLocation?: string; mapLatitude?: number | null; mapLongitude?: number | null; expenseResponsibleId?: string | null; serviceIds?: string[] },
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
  // ── الإنجاز الجزئي: الحجز الي ياخذ أكثر من يوم ──
  partialCompleteBooking: (id: string, body: PartialCompleteBody) =>
    request<BookingProgressReport>(`/bookings/${id}/partial-complete`, {
      method: 'POST', body: JSON.stringify(body),
    }),
  getBookingProgress: (id: string) =>
    request<BookingProgressReport[]>(`/bookings/${id}/progress`),
  getSuggestedCrew: (id: string) =>
    request<SuggestedCrewMember[]>(`/bookings/${id}/suggested-crew`),
  /** الإداري يحدد موعد إكمال حجز منجز جزئياً — مو «تأجيل»، فما ينعد بعداد التأجيلات */
  scheduleContinuation: (id: string, scheduledAt: string) =>
    request<Booking>(`/bookings/${id}/schedule-continuation`, {
      method: 'POST', body: JSON.stringify({ scheduledAt }),
    }),

  completeBooking: (
    id: string,
    data: { completionNotes?: string; amountCollected?: number; advancePaid?: number },
  ) => request<Booking>(`/bookings/${id}/complete`, { method: 'PUT', body: JSON.stringify(data) }),
  startBooking: (id: string, missingToolIds?: string[]) =>
    request<Booking>(`/bookings/${id}/start`, { method: 'PUT', body: JSON.stringify({ missingToolIds: missingToolIds || [] }) }),
  markArrived: (id: string) =>
    request<Booking>(`/bookings/${id}/arrived`, { method: 'PUT', body: JSON.stringify({}) }),
  setMaterialsReady: (id: string) =>
    request<Booking>(`/bookings/${id}/materials-ready`, { method: 'PUT', body: JSON.stringify({}) }),
  verifyAmount: (id: string) =>
    request<Booking>(`/bookings/${id}/verify`, { method: 'PUT', body: JSON.stringify({}) }),
  // إرجاع الحجز للتدقيق بعد ما انتأشر مدقّق — مدير النظام حصراً
  unverifyBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/unverify`, { method: 'PUT' }),
  markConfirmationContacted: (id: string) =>
    request<Booking>(`/bookings/${id}/confirmation-contacted`, { method: 'PUT', body: JSON.stringify({}) }),
  getPendingAudit: () => request<Booking[]>('/bookings/pending-audit'),
  getBookingToolChecks: (id: string) => request<BookingToolCheck[]>(`/bookings/${id}/tool-checks`),
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
        // ملف الموارد البشرية — careerStatus ما تنرسل: تنحسب بالسيرفر
        | 'department' | 'hireDate' | 'experienceYears' | 'lastReview' | 'jobLevel' | 'nextRole' | 'trainingNeeds'
      >
    > & {
      username?: string
      password?: string
    },
  ) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  linkHistoricalRecords: (id: string) =>
    request<{ bookingsLinked: number; complaintsLinked: number }>(`/employees/${id}/link-historical`, { method: 'POST' }),

  // GPS
  getGpsStats: () => request<GpsStats>('/gps/stats'),
  getGpsCustomers: () => request<GpsCustomer[]>('/gps/customers'),
  createGpsCustomer: (data: Partial<GpsCustomer>) =>
    request<GpsCustomer>('/gps/customers', { method: 'POST', body: JSON.stringify(data) }),
  getGpsDevices: () => request<GpsDeviceRequest[]>('/gps/devices'),
  createGpsDevice: (data: Partial<GpsDeviceRequest>) =>
    request<GpsDeviceRequest>('/gps/devices', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsDevice: (id: string, data: Partial<GpsDeviceRequest>) =>
    request<GpsDeviceRequest>(`/gps/devices/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  // ── الدوار ──
  getGpsInstallCosts: () => request<GpsInstallCostSummary>('/finance/gps-install-costs'),
  addStockIntake: (data: { toolId: string; quantity: number; unitPrice?: number | null; supplier?: string | null; notes?: string | null }) =>
    request<StockIntake>('/inventory/stock-intake', { method: 'POST', body: JSON.stringify(data) }),
  getStockIntakes: (toolId?: string) =>
    request<StockIntake[]>(`/inventory/stock-intake${toolId ? '?toolId=' + toolId : ''}`),
  /** أرقام اللوحة الرئيسية — بدون سحب أرشيف الشركة كامل للمتصفح */
  getDashboardSummary: () => request<{ employeeCount: number; customerCount: number; bookingCount: number; gpsDeviceCount: number }>('/dashboard/summary'),
  /** مجاميع المحاسب والمراقب — تنحسب بقاعدة البيانات بدل تنزيل كل الأرشيف المنجز */
  getFinanceSummary: () => request<FinanceSummary>('/dashboard/finance-summary'),

  // ── الإجازات ──
  createLeave: (data: { startDate: string; endDate?: string; reason?: string | null }) =>
    request<LeaveRequest>('/leaves', { method: 'POST', body: JSON.stringify(data) }),
  getMyLeaves: () => request<LeaveRequest[]>('/leaves/mine'),
  cancelLeave: (id: string) => request<{ ok: boolean }>(`/leaves/${id}`, { method: 'DELETE' }),
  getLeaveInbox: (status?: string) => request<LeaveRequest[]>(`/leaves/inbox${status ? '?status=' + status : ''}`),
  getLeavePendingCount: () => request<{ count: number; canApprove: boolean }>('/leaves/pending-count'),
  preliminaryLeave: (id: string, note?: string) =>
    request<LeaveRequest>(`/leaves/${id}/preliminary`, { method: 'PUT', body: JSON.stringify({ note }) }),
  decideLeave: (id: string, approve: boolean, note?: string) =>
    request<LeaveRequest>(`/leaves/${id}/decide`, { method: 'PUT', body: JSON.stringify({ approve, note: note || null }) }),

  getFunds: () => request<RevolvingFund[]>('/funds'),
  updateFund: (id: string, data: { name?: string; balance?: number; isActive?: boolean }) =>
    request<RevolvingFund>(`/funds/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  topupFund: (id: string, amount: number, notes?: string) =>
    request<{ ok: boolean }>(`/funds/${id}/topup`, { method: 'POST', body: JSON.stringify({ amount, notes: notes || null }) }),
  disburseFund: (data: { fundId: string; employeeId: string; amount: number; bookingId?: string | null; requestId?: string | null; notes?: string | null }) =>
    request<RevolvingFundTxn>('/funds/disburse', { method: 'POST', body: JSON.stringify(data) }),
  getFundBalances: () => request<EmployeeFundBalance[]>('/funds/balances'),
  getFundTransactions: (params?: { employeeId?: string; status?: string }) => {
    const q = new URLSearchParams()
    if (params?.employeeId) q.set('employeeId', params.employeeId)
    if (params?.status) q.set('status', params.status)
    const qs = q.toString()
    return request<RevolvingFundTxn[]>(`/funds/transactions${qs ? '?' + qs : ''}`)
  },
  reviewFundSettlement: (id: string, approve: boolean, reviewNote?: string) =>
    request<RevolvingFundTxn>(`/funds/settlements/${id}/review`, {
      method: 'PUT', body: JSON.stringify({ approve, reviewNote: reviewNote || null }),
    }),
  getMyFundBalance: () => request<EmployeeFundBalance>('/funds/my-balance'),
  getMyFundTransactions: () => request<RevolvingFundTxn[]>('/funds/my-transactions'),

  /** المحاسب يأشر «تم التخريج» ويحدد الحساب — هنا يرجع المصروف للدوار */
  dischargeSettlement: (id: string, data: { account: string; note?: string | null }) =>
    request<RevolvingFundTxn>(`/funds/settlements/${id}/discharge`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  getDischargeAccounts: () => request<string[]>('/funds/discharge-accounts'),
  submitFundSettlement: (data: { fundId: string; spentAmount: number; returnedAmount: number; receiptImage?: string | null; bookingId?: string | null; notes?: string | null }) =>
    request<RevolvingFundTxn>('/funds/settlements', { method: 'POST', body: JSON.stringify(data) }),

  getSimCards: () => request<GpsSimCard[]>('/gps/sims'),
  createSimCard: (data: Partial<GpsSimCard>) =>
    request<GpsSimCard>('/gps/sims', { method: 'POST', body: JSON.stringify(data) }),
  updateSimCard: (id: string, data: Partial<GpsSimCard>) =>
    request<GpsSimCard>(`/gps/sims/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // ── دورة حياة الشريحة ومتابعة التجديد ──
  getAvailableSimCards: () => request<GpsSimCard[]>('/gps/sims/available'),
  assignSimCard: (id: string, customerId: string) =>
    request<GpsSimCard>(`/gps/sims/${id}/assign`, { method: 'POST', body: JSON.stringify({ customerId }) }),
  releaseSimCard: (id: string) =>
    request<GpsSimCard>(`/gps/sims/${id}/release`, { method: 'POST' }),
  burnSimCard: (id: string) =>
    request<GpsSimCard>(`/gps/sims/${id}/burn`, { method: 'POST' }),
  getSubscriptionFollowUps: () =>
    request<GpsSubscriptionFollowUp[]>('/gps/subscriptions/follow-up'),
  getDeviceFollowUps: (deviceId: string) =>
    request<GpsRenewalFollowUp[]>(`/gps/devices/${deviceId}/follow-up`),
  createDeviceFollowUp: (deviceId: string, outcome: GpsFollowUpOutcome, notes?: string) =>
    request<GpsRenewalFollowUp>(`/gps/devices/${deviceId}/follow-up`, {
      method: 'POST', body: JSON.stringify({ outcome, notes: notes || null }),
    }),
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
  getRatableEmployees: () => request<{ id: string; name: string }[]>('/performance-reviews/ratable'),
  getPerformanceReviewsForEmployee: (employeeId: string) => request<PerformanceReview[]>(`/performance-reviews/employee/${employeeId}`),

  // Products
  getProducts: () => request<Product[]>('/products'),
  // ── الملفات المخزّنة برّا قاعدة البيانات ──
  //
  // الصور القديمة لسّه data: URLs والجديدة صارت /api/files/... —
  // fileUrl() يتعامل مع الاثنين، فما نحتاج نعدّل كل شاشة.
  getFileToken: () => request<{ token: string }>('/files/token'),
  uploadFile: async (file: File, folder?: string) => {
    const form = new FormData()
    form.append('file', file)
    const token = currentToken()
    // ماكو Content-Type: المتصفح لازم يحطه بنفسه مع حدود multipart
    const res = await fetch(
      `${API_URL}/files${folder ? `?folder=${encodeURIComponent(folder)}` : ''}`,
      { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form },
    )
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'تعذر رفع الملف')
    return res.json() as Promise<{ key: string; url: string; size: number; type: string }>
  },

  // المضاف من شاشة «إضافة منتج» بس — بدون كتالوج عروض الأسعار القديم
  getAddedProducts: () => request<Product[]>('/products?added=1'),
  // الحقول المحسوبة (العناوين واقتراح النظام) ما تنرسل — السيرفر يحسبها
  createProduct: (data: ProductInput) =>
    request<Product>('/products', { method: 'POST', body: JSON.stringify(data) }),
  updateProduct: (id: string, data: Partial<ProductInput> & { clearService?: boolean }) =>
    request<Product>(`/products/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteProduct: (id: string) =>
    request<void>(`/products/${id}`, { method: 'DELETE' }),

  // Quotations
  getQuotations: (search?: string) => request<Quotation[]>(`/quotations${search ? `?search=${encodeURIComponent(search)}` : ''}`),
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

  // Leader invoices (تحل محل شيت جوجل الليدر)
  getSystemPriceCatalog: (systemName?: string) =>
    request<SystemPriceCatalog[]>(`/system-price-catalog${systemName ? `?systemName=${encodeURIComponent(systemName)}` : ''}`),
  getMaterials: (code?: string) =>
    request<Material[]>(`/materials${code ? `?code=${encodeURIComponent(code)}` : ''}`),
  getLeaderInvoices: (employeeId?: string) =>
    request<LeaderInvoice[]>(`/leader-invoices${employeeId ? `?employeeId=${encodeURIComponent(employeeId)}` : ''}`),
  getLeaderInvoice: (id: string) => request<LeaderInvoice>(`/leader-invoices/${id}`),
  createLeaderInvoice: (data: CreateLeaderInvoiceRequest) =>
    request<LeaderInvoice>('/leader-invoices', { method: 'POST', body: JSON.stringify(data) }),
  estimateLeaderInvoiceCost: (items: ExecutionCostItem[]) =>
    request<EstimateExecutionCostResponse>('/leader-invoices/estimate', { method: 'POST', body: JSON.stringify({ items }) }),
  approveLeaderInvoice: (id: string, externalInvoiceNumber: string) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/approve`, {
      method: 'PUT',
      body: JSON.stringify({ externalInvoiceNumber }),
    }),
  // ربط رقم فاتورة محاسبية بفاتورة معتمدة أصلاً (الفواتير القديمة)
  setInvoiceExternalNumber: (id: string, externalInvoiceNumber: string) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/external-number`, {
      method: 'PUT', body: JSON.stringify({ externalInvoiceNumber }),
    }),
  // تعديل المحاسب على مبالغ الفاتورة — السبب إجباري
  adjustLeaderInvoice: (id: string, data: { executionCost: number; materialsTotal: number; discountValue: number; reason: string }) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/adjust`, { method: 'PUT', body: JSON.stringify(data) }),
  // البحث برقم فاتورة المحاسب — سبب أرشفة الرقم أصلاً
  findInvoiceByNumber: (number: string) =>
    request<LeaderInvoice>(`/leader-invoices/by-number?number=${encodeURIComponent(number)}`),
  // المشاريع الموجّهة للموظف الحالي — الليدر يسوي فاتورة للشغل الموجّه له
  getProjectsDirectedToMe: () =>
    request<{ projects: DirectedProject[] }>('/projects/delegated-to-me'),
  calculateCameraCost: (data: CameraCostRequest) =>
    request<CameraCostResponse>('/leader-invoices/camera-cost', { method: 'POST', body: JSON.stringify(data) }),
  getCameraCostOptions: () =>
    request<{ placeTypes: string[]; systemTypes: string[]; note: string }>('/leader-invoices/camera-cost/options'),

  // إعدادات وحدة إدارة المشاريع — أنواع الأعمال قابلة للإضافة/الحذف
  // الشخصيات المهمة (VIP)
  getVipCustomers: () => request<VipCustomer[]>('/vip-customers'),
  getVipCustomerIds: () => request<string[]>('/vip-customers/ids'),
  // phone بديل عن customerId بالإضافة اليدوية — السيرفر يطلع الزبون من رقمه
  // طلب حذف حجز: الإداري يطلب، والمراقب/المدير يبت
  requestBookingDelete: (bookingId: string, reason: string) =>
    request<BookingDeleteRequest>(`/bookings/${bookingId}/delete-request`, { method: 'POST', body: JSON.stringify({ reason }) }),
  getBookingDeleteRequests: (status?: string) =>
    request<BookingDeleteRequest[]>(`/booking-delete-requests${status ? `?status=${status}` : ''}`),
  decideBookingDelete: (id: string, approve: boolean, note?: string) =>
    request<BookingDeleteRequest>(`/booking-delete-requests/${id}/decide`, { method: 'PUT', body: JSON.stringify({ approve, note: note || null }) }),

  markVipCustomer: (data: { customerId?: string; phone?: string; name?: string; location?: string | null; locationUrl?: string | null; boughtFromUs?: boolean; bookingId?: string; requestSummary?: string; customerPosition?: string; note?: string }) =>
    request<VipCustomer>('/vip-customers', { method: 'POST', body: JSON.stringify(data) }),
  unmarkVipCustomer: (customerId: string) => request<void>(`/vip-customers/${customerId}`, { method: 'DELETE' }),

  getProjectWorkTypes: () => request<ProjectWorkType[]>('/project-work-types'),
  createProjectWorkType: (name: string) =>
    request<ProjectWorkType>('/project-work-types', { method: 'POST', body: JSON.stringify({ name }) }),
  deleteProjectWorkType: (id: string) => request<void>(`/project-work-types/${id}`, { method: 'DELETE' }),

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
  createKpiEvaluation: (data: { employeeId: string; evaluatorId: string; points: number; reason: string; announce?: boolean }) =>
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

  /**
   * قائمة مبسطة لاختيار سيارة لحجز — بلا صلاحية إدارة الأسطول.
   * /vehicles محمية بصلاحية vehicle_management وما عدها موظف التنسيق،
   * فالقائمة المنسدلة جانت تطلع فاضية وما يكدر يحدد سيارة.
   */
  getVehicleOptions: () => request<VehicleOption[]>('/vehicles/options'),
  createVehicle: (data: { name: string; plateNumber: string; color?: string; type?: string }) =>
    request<Vehicle>('/vehicles', { method: 'POST', body: JSON.stringify(data) }),
  getVehicleLogs: (vehicleId: string) => request<VehicleLog[]>(`/vehicles/${vehicleId}/logs`),
  createVehicleLog: (vehicleId: string, data: { type: 'FUEL' | 'CLEANING' | 'OIL_CHANGE' | 'MAINTENANCE'; performedAt?: string; nextDueAt?: string; nextDueOdometer?: number; odometer?: number; cost?: number; notes?: string; liters?: number; filledByEmployeeId?: string; receiptNumber?: string; stationName?: string; receiptPhotoBase64?: string }) =>
    request<VehicleLog & { fuelAnomaly?: FuelAnomalyResult }>(`/vehicles/${vehicleId}/logs`, { method: 'POST', body: JSON.stringify(data) }),
  updateVehicleLog: (vehicleId: string, logId: string, data: { performedAt?: string; odometer?: number; cost?: number; notes?: string; nextDueAt?: string; nextDueOdometer?: number; liters?: number; filledByEmployeeId?: string; receiptNumber?: string; stationName?: string; receiptPhotoBase64?: string; clearReceiptPhoto?: boolean }) =>
    request<VehicleLog>(`/vehicles/${vehicleId}/logs/${logId}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteVehicleLog: (vehicleId: string, logId: string) =>
    request<void>(`/vehicles/${vehicleId}/logs/${logId}`, { method: 'DELETE' }),
  getVehicleLogReceiptPhoto: (vehicleId: string, logId: string) =>
    request<{ receiptPhotoBase64: string | null }>(`/vehicles/${vehicleId}/logs/${logId}/receipt-photo`),
  getEmployeeFuelStats: (params?: { vehicleId?: string; month?: string }) => {
    const qs = new URLSearchParams()
    if (params?.vehicleId) qs.set('vehicleId', params.vehicleId)
    if (params?.month) qs.set('month', params.month)
    const s = qs.toString()
    return request<EmployeeFuelStat[]>(`/vehicles/fuel-stats/by-employee${s ? `?${s}` : ''}`)
  },
  getVehicleExpenseSummary: (vehicleId: string, params?: { month?: string; year?: string }) => {
    const qs = params?.year ? `?year=${params.year}` : params?.month ? `?month=${params.month}` : ''
    return request<VehicleExpenseSummary>(`/vehicles/${vehicleId}/expense-summary${qs}`)
  },
  getVehicleIncidents: (vehicleId: string) => request<VehicleIncident[]>(`/vehicles/${vehicleId}/incidents`),
  createVehicleIncident: (vehicleId: string, data: { type: 'FAULT' | 'DAMAGE' | 'ACCIDENT'; description: string; responsibleEmployeeId?: string; cost?: number; location?: string; driverId?: string; peoplePresent?: string; policeReportNumber?: string; repairCost?: number }) =>
    request<VehicleIncident>(`/vehicles/${vehicleId}/incidents`, { method: 'POST', body: JSON.stringify(data) }),
  updateVehicleIncident: (id: string, data: { status?: 'OPEN' | 'RESOLVED'; cost?: number }) =>
    request<VehicleIncident>(`/vehicle-incidents/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getVehicleIncidentAttachments: (incidentId: string) => request<VehicleIncidentAttachment[]>(`/vehicle-incidents/${incidentId}/attachments`),
  createVehicleIncidentAttachment: (incidentId: string, data: { url: string; mediaType: 'IMAGE' | 'VIDEO' }) =>
    request<VehicleIncidentAttachment>(`/vehicle-incidents/${incidentId}/attachments`, { method: 'POST', body: JSON.stringify(data) }),
  deleteVehicleIncidentAttachment: (incidentId: string, attachmentId: string) =>
    request<{ ok: boolean }>(`/vehicle-incidents/${incidentId}/attachments/${attachmentId}`, { method: 'DELETE' }),
  getVehicleParts: (vehicleId: string) => request<VehiclePart[]>(`/vehicles/${vehicleId}/parts`),
  createVehiclePart: (vehicleId: string, data: { partType: 'TIRE' | 'BATTERY'; installedAt?: string; installedOdometer: number; expectedLifespanKm?: number; expectedLifespanMonths?: number; notes?: string; cost?: number }) =>
    request<VehiclePart>(`/vehicles/${vehicleId}/parts`, { method: 'POST', body: JSON.stringify(data) }),
  replaceVehiclePart: (partId: string) => request<VehiclePart>(`/vehicle-parts/${partId}/replace`, { method: 'PUT' }),
  getVehicleAlerts: () => request<VehicleAlert[]>('/vehicles/alerts'),
  getFleetDashboard: () => request<FleetDashboardSummary>('/vehicles/dashboard'),
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
  deleteVehicle: (id: string) => request<void>(`/vehicles/${id}`, { method: 'DELETE' }),

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
  }) => request<VehicleMission & { bookingWarning?: string }>('/vehicle-missions', { method: 'POST', body: JSON.stringify(data) }),
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
  createVehicleMissionRating: (missionId: string, data: {
    commitment: number; vehicleCare: number; driving: number; cleanliness: number; notes?: string
  }) => request<VehicleMissionRating>(`/vehicle-missions/${missionId}/rating`, { method: 'POST', body: JSON.stringify(data) }),
  getDriverRatingSummary: (employeeId: string) =>
    request<DriverRatingSummary>(`/employees/${employeeId}/driver-rating-summary`),

  createVehicleBooking: (data: { vehicleId: string; purpose: string; startAt: string; endAt: string }) =>
    request<VehicleBooking>('/vehicle-bookings', { method: 'POST', body: JSON.stringify(data) }),
  decideVehicleBooking: (id: string, data: { approve: boolean; rejectionReason?: string }) =>
    request<VehicleBooking>(`/vehicle-bookings/${id}/decide`, { method: 'PUT', body: JSON.stringify(data) }),
  cancelVehicleBooking: (id: string) =>
    request<VehicleBooking>(`/vehicle-bookings/${id}/cancel`, { method: 'PUT' }),
  getVehicleBookings: (filters?: { vehicleId?: string; requestedById?: string; status?: string; from?: string; to?: string }) => {
    const params = new URLSearchParams()
    if (filters?.vehicleId) params.set('vehicleId', filters.vehicleId)
    if (filters?.requestedById) params.set('requestedById', filters.requestedById)
    if (filters?.status) params.set('status', filters.status)
    if (filters?.from) params.set('from', filters.from)
    if (filters?.to) params.set('to', filters.to)
    const qs = params.toString()
    return request<VehicleBooking[]>(`/vehicle-bookings${qs ? `?${qs}` : ''}`)
  },
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
  // آخر جرد للفني نفسه — يرجّع null لو ما جرد أبداً
  getMyLastInventoryCheck: () => request<InventoryCheck | null>('/inventory/checks/mine'),
  resolveInventoryCheck: (id: string) => request<InventoryCheck>(`/inventory/checks/${id}/resolve`, { method: 'POST' }),
  getPersonalTools: (employeeId?: string) =>
    request<PersonalTool[]>(`/inventory/personal${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createPersonalTool: (data: { employeeId: string; name: string; barcode: string }) =>
    request<PersonalTool>('/inventory/personal', { method: 'POST', body: JSON.stringify(data) }),
  // سجل حركة الأدوات: toolId لأداة وحدة، employeeId لكل عدة موظف، بلا شي = الكل
  // سياسة الخصوصية
  getPrivacyPolicy: (all?: boolean) =>
    request<PrivacyPolicyPoint[]>(`/privacy-policy${all ? '?all=true' : ''}`),
  getPrivacyPolicyStatus: () => request<PrivacyPolicyStatus>('/privacy-policy/status'),
  acceptPrivacyPolicy: () =>
    request<{ accepted: boolean }>('/privacy-policy/accept', { method: 'POST' }),
  createPrivacyPolicyPoint: (content: string, order?: number) =>
    request<PrivacyPolicyPoint>('/privacy-policy', { method: 'POST', body: JSON.stringify({ content, order }) }),
  updatePrivacyPolicyPoint: (id: string, data: { content?: string; order?: number; isActive?: boolean }) =>
    request<PrivacyPolicyPoint>(`/privacy-policy/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePrivacyPolicyPoint: (id: string) =>
    request<void>(`/privacy-policy/${id}`, { method: 'DELETE' }),
  getToolEvents: (params?: { toolId?: string; employeeId?: string }) => {
    const qs = new URLSearchParams()
    if (params?.toolId) qs.set('toolId', params.toolId)
    if (params?.employeeId) qs.set('employeeId', params.employeeId)
    const s = qs.toString()
    return request<PersonalToolEvent[]>(`/inventory/tool-events${s ? `?${s}` : ''}`)
  },
  updatePersonalTool: (id: string, data: { name?: string; barcode?: string; status?: PersonalToolStatus; checkedOut?: boolean; note?: string }) =>
    request<PersonalTool>(`/inventory/personal/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePersonalTool: (id: string) => request<void>(`/inventory/personal/${id}`, { method: 'DELETE' }),
  getVehicleTools: (vehicleId?: string) =>
    request<VehicleTool[]>(`/inventory/vehicle${vehicleId ? `?vehicleId=${vehicleId}` : ''}`),
  createVehicleTool: (data: { name: string; quantity: number; vehicleId: string; barcode?: string }) =>
    request<VehicleTool>('/inventory/vehicle', { method: 'POST', body: JSON.stringify(data) }),
  updateVehicleTool: (id: string, data: Partial<VehicleTool>) =>
    request<VehicleTool>(`/inventory/vehicle/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteVehicleTool: (id: string) => request<void>(`/inventory/vehicle/${id}`, { method: 'DELETE' }),
  getOnDemandTools: () => request<OnDemandTool[]>('/inventory/ondemand'),
  createOnDemandTool: (data: { name: string; barcode: string; totalQuantity: number }) =>
    request<OnDemandTool>('/inventory/ondemand', { method: 'POST', body: JSON.stringify(data) }),
  updateOnDemandTool: (id: string, data: Partial<OnDemandTool>) =>
    request<OnDemandTool>(`/inventory/ondemand/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  // يفتح رابط الخريطة بالسيرفر ويرجع النقطة — ضروري للروابط المختصرة
  // (maps.app.goo.gl) الي ما تحتوي إحداثيات إلا بعد ما تنفتح.
  resolveMapLink: (url: string) =>
    request<{ lat: number; lng: number }>(`/geo/resolve-map-link?url=${encodeURIComponent(url)}`),
  getToolRequests: (employeeId?: string) =>
    request<ToolRequestItem[]>(`/inventory/requests${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createToolRequest: (data: { employeeId: string; toolId: string; reason: ToolRequestReason; description?: string }) =>
    request<ToolRequestItem>('/inventory/requests', { method: 'POST', body: JSON.stringify(data) }),
  // purchasePrice يُرسل فقط لما الأداة مو متوفرة بالمخزن — السيرفر يرفض
  // الموافقة بدونه بهذي الحالة وينشئ طلب مشتريات للمحاسب لما ينوصل.
  approveToolRequest: (id: string, approvedById: string, purchasePrice?: number) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/approve`, { method: 'PUT', body: JSON.stringify({ approvedById, purchasePrice }) }),
  rejectToolRequest: (id: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/reject`, { method: 'PUT', body: JSON.stringify({}) }),
  // العدة القياسية (PersonalToolTemplateItem)
  getPersonalToolTemplate: () => request<PersonalToolTemplateItem[]>('/inventory/personal-template'),
  createPersonalToolTemplateItem: (name: string) =>
    request<PersonalToolTemplateItem>('/inventory/personal-template', { method: 'POST', body: JSON.stringify({ name }) }),
  deletePersonalToolTemplateItem: (id: string) =>
    request<void>(`/inventory/personal-template/${id}`, { method: 'DELETE' }),
  // فحوصات أدوات المركبات (VehicleToolCheck) + كل فحوصات الحجوزات (BookingToolCheck)
  getVehicleToolChecks: () => request<VehicleToolCheck[]>('/inventory/vehicle-tool-checks'),
  getAllBookingToolChecks: () => request<BookingToolCheck[]>('/inventory/booking-tool-checks'),
  createVehicleMissionToolCheck: (missionId: string, missingToolNames: string[]) =>
    request<VehicleToolCheck>(`/vehicle-missions/${missionId}/tool-check`, { method: 'POST', body: JSON.stringify({ missingToolNames }) }),
  returnToolRequest: (id: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/return`, { method: 'PUT', body: JSON.stringify({}) }),
  deleteToolRequest: (id: string) =>
    request<void>(`/inventory/requests/${id}`, { method: 'DELETE' }),

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
  getGpsSettings: () => request<Record<string, unknown>[]>('/gps/settings'),
  updateGpsSettings: (data: Record<string, unknown>) =>
    request<Record<string, unknown>>('/gps/settings', { method: 'PUT', body: JSON.stringify(data) }),
  updateGpsCustomer: (id: string, data: Partial<GpsCustomer>) =>
    request<GpsCustomer>(`/gps/customers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // Procurement
  getProcurementRequests: () => request<ProcurementRequest[]>('/procurement'),
  getProcurementStats: () => request<ProcurementStats>('/procurement/stats'),
  createProcurementRequest: (data: { requestedById: string; bookingId?: string; requestType: ProcurementRequestType; notes?: string; items: { productName: string; quantity: number; unitPrice?: number | null; model?: string | null; spec?: string | null; itemNotes?: string | null }[] }) =>
    request<ProcurementRequest>('/procurement', { method: 'POST', body: JSON.stringify(data) }),
  updateProcurementStatus: (id: string, status: string) =>
    request<ProcurementRequest>(`/procurement/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
  fulfillProcurementRequest: (id: string, data: { fulfilledById: string; supplierId: string; totalCost?: number; fulfillmentNotes?: string; items?: { id: string; unitPrice?: number; totalPrice?: number; fulfilled?: boolean }[] }) =>
    request<ProcurementRequest>(`/procurement/${id}/fulfill`, { method: 'PUT', body: JSON.stringify(data) }),

  // صيانة الأجهزة العامة (شيت "صيانة الاجهزة") — ليدر فقط
  getDeviceMaintenanceTickets: () => request<DeviceMaintenanceTicket[]>('/device-maintenance'),
  createDeviceMaintenanceTicket: (data: { appointmentDate?: string | null; customerCode: number; deviceTypeName: string; problem: string; deviceSerial?: string | null }) =>
    request<DeviceMaintenanceTicket>('/device-maintenance', { method: 'POST', body: JSON.stringify(data) }),
  updateDeviceMaintenanceTicket: (id: string, data: { appointmentDate?: string | null; deviceTypeName?: string; problem?: string; deviceSerial?: string | null; markReceived?: boolean; markDelivered?: boolean }) =>
    request<DeviceMaintenanceTicket>(`/device-maintenance/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // جرد الفريق ("جرد العدد") — ليدر فقط
  getTeamInventoryTools: () => request<TeamInventoryToolCatalogItem[]>('/team-inventory/tools'),
  createTeamInventoryTool: (name: string) =>
    request<TeamInventoryToolCatalogItem>('/team-inventory/tools', { method: 'POST', body: JSON.stringify({ name }) }),
  getTeamInventoryChecks: () => request<TeamInventoryCheck[]>('/team-inventory/checks'),
  createTeamInventoryCheck: (data: { employee1Id?: string | null; employee2Id?: string | null; items: { toolName: string; personRole: TeamInventoryPersonRole; present: boolean; reason?: TeamInventoryShortageReason | null }[] }) =>
    request<TeamInventoryCheck>('/team-inventory/checks', { method: 'POST', body: JSON.stringify(data) }),

  // تقدير مدة العمل المتعلَّم تلقائياً (بدون رقم مفروض يدوياً) — يرجع expectedMinutes:
  // null لو البيانات التاريخية غير كافية بعد (sampleCount < minSamples).
  getJobDurationEstimate: (params: { systemName: string; jobType: 'INSTALL' | 'MAINTENANCE'; itemCount: number; crewSize: number }) =>
    request<JobDurationEstimate>(
      `/job-duration-estimate?systemName=${encodeURIComponent(params.systemName)}&jobType=${params.jobType}&itemCount=${params.itemCount}&crewSize=${params.crewSize}`,
    ),

  // إحصائيات الموظفين الشهرية — OWNER/ADMIN فقط
  getEmployeeMonthlyStats: (month: string) =>
    request<EmployeeMonthlyStats[]>(`/employee-stats/monthly?month=${encodeURIComponent(month)}`),
  exportEmployeeMonthlyStats: (month: string) =>
    downloadFile(`/employee-stats/monthly/export?month=${encodeURIComponent(month)}`, `employee-stats-${month}.xlsx`),
  getDailyStats: (date?: string) => request<DailyStats>(`/stats-management/daily${date ? `?date=${encodeURIComponent(date)}` : ''}`),
  getInternalWorks: (month?: string) =>
    request<InternalWorksReport>(`/stats-management/internal-works${month ? `?month=${encodeURIComponent(month)}` : ''}`),
  getWeeklyStats: (from: string, to: string) =>
    request<WeeklyStats>(`/stats-management/weekly?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),
  getProjectStageStats: () => request<ProjectStageStats[]>('/stats-management/projects'),
  getEmployeePerformanceCurve: (employeeId: string, month?: string, months = 6) =>
    request<EmployeePerformanceCurve>(`/employee-stats/curve/${employeeId}?months=${months}${month ? `&month=${encodeURIComponent(month)}` : ''}`),

  // وحدة التصميم — عدة استمارات مستقلة، كل وحدة بأسئلتها وبرابطها العام الخاص
  getDesignForms: () => request<DesignForm[]>('/design-forms'),
  createDesignForm: (name: string) => request<DesignForm>('/design-forms', { method: 'POST', body: JSON.stringify({ name }) }),
  deleteDesignForm: (id: string) => request<void>(`/design-forms/${id}`, { method: 'DELETE' }),
  getDesignFormSubmissions: (formId: string) => request<DesignFormSubmission[]>(`/design-forms/${formId}/submissions`),

  getDesignFormQuestions: (formId: string) => request<DesignFormQuestion[]>(`/design-forms/${formId}/questions`),
  createDesignFormQuestion: (formId: string, data: { label: string; type: DesignFormQuestionType; options?: string[]; required?: boolean }) =>
    request<DesignFormQuestion>(`/design-forms/${formId}/questions`, { method: 'POST', body: JSON.stringify(data) }),
  updateDesignFormQuestion: (id: string, data: { label?: string; type?: DesignFormQuestionType; options?: string[]; required?: boolean }) =>
    request<DesignFormQuestion>(`/design-form/questions/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteDesignFormQuestion: (id: string) => request<void>(`/design-form/questions/${id}`, { method: 'DELETE' }),
  reorderDesignFormQuestions: (questionIds: string[]) =>
    request<void>('/design-form/questions/reorder', { method: 'PUT', body: JSON.stringify({ questionIds }) }),

  // رابط عام للزبون (بدون تسجيل دخول)
  getPublicDesignForm: (token: string) => request<{ name: string; questions: DesignFormQuestion[] }>(`/public/design-forms/${token}`),
  submitPublicDesignForm: (token: string, answers: Record<string, unknown>) =>
    request<DesignFormSubmission>(`/public/design-forms/${token}/submit`, { method: 'POST', body: JSON.stringify({ answers }) }),
}

export interface BackupRun {
  id: string
  startedAt: string
  finishedAt: string | null
  ok: boolean
  fileName: string | null
  sizeBytes: number
  tableCount: number
  encrypted: boolean
  offsite: boolean
  offsiteTarget: string | null
  hasUploads: boolean
  hasEnv: boolean
  warnings: string | null
  error: string | null
  keptCount: number
}

export interface BackupHealth {
  last: BackupRun | null
  lastSuccess: BackupRun | null
  hoursSinceSuccess: number | null
  consecutiveFails: number
  success7d: number
  failed7d: number
  offsiteOK: boolean
  encrypted: boolean
  status: 'OK' | 'WARN' | 'CRITICAL' | 'UNKNOWN'
  message: string
  recommendations: string[]
  totalRunsRecorded: number
}

export interface BackupOverview {
  health: BackupHealth
  runs: BackupRun[]
}

export interface PartialCompleteBody {
  workDone: string
  remainingWork: string
  percentDone: number
  blockers?: string
  materialsUsed?: string
  amountCollected?: number
}

export interface BookingProgressReport {
  id: string
  bookingId: string
  dayNumber: number
  reportedById: string | null
  reportedBy: { id: string; name: string } | null
  workDone: string
  remainingWork: string
  percentDone: number
  blockers: string | null
  materialsUsed: string | null
  crewSnapshot: string | null
  createdAt: string
}

export interface SuggestedCrewMember {
  employeeId: string
  name: string
  role: string
  daysWorked: number
  available: boolean
  note: string
}
