import { readAIStream, readMatrixStream, type AIStreamEvent, type MatrixStreamEvent } from './utils/readMatrixStream'
/** عنوان السيرفر — مكان واحد (DRY 11-04) */
import type { AttendanceRecord, AttendanceBoardRow, MonthlyAttendanceReport, OpenSessionResponse, EmployeeDailyAttendanceSummary, Skill, Division, Service, ServiceManager, EmployeeRole, Employee, AttendanceIconRequest, GpsCustomer, GpsStats, RevolvingFund, RevolvingFundTxn, EmployeeFundBalance, GpsInstallCostSummary, StockIntake, LeaveKind, LeaveRequest, ReturnedFromProjects, GpsSimCard, GpsFollowUpOutcome, GpsSubscriptionFollowUp, GpsRenewalFollowUp, GpsDeviceRequest, StaffRequest, BookingAwaitingReview, ReviewRating, PerformanceReview, GpsRenewalRequest, GpsMaintenanceRequest, TrainingMaterial, Booking, BookingVisit, TodayBoardData, JobDurationEstimate, DelayRisk, SalesOpportunity, CrewRecommendation, NewEmployeeCurve, StockForecast, ReplacementSuggestions, WeeklyReport, InternalWorksReport, Stats, Expense, Customer, GpsCustomerListItem, AuditVerdict, ComplaintEvent, LeaderSkillRating, AuditCloseAction, MissingRoleDefault, ComplaintCustomerStat, QualityFollowUp, SecretaryCard, SecretaryProposal, SecretaryConversation, SecretaryReply, SecretaryChat, SecretaryToday, Product, ProductInput, EmployeeMonthlyStats, EmployeePerformanceCurve, DesignFormQuestionType, DesignForm, DesignFormQuestion, DesignFormSubmission, DailyStats, WeeklyStats, ProjectStageStats, SystemPriceCatalog, ExecutionCostItem, LocateResult, LeaderInvoice, CreditSale, BookingDeleteChannel, BookingDeleteRequestType, BookingDeleteRequest, BookingDeleteRequestCounts, CoordinationAlert, CoordinationAlertSummary, ItAssetKind, ItAssetStatus, ItAsset, ItAssetInput, ItAssetLog, ItStats, SurveyPhotoOwner, SurveyPhoto, DailyAuditReport, AuditIssue, Announcement, VipCustomer, ProjectWorkType, EstimateExecutionCostResponse, DirectedProject, MonitorStage, StageBucket, AiSignal, AiCatalog, AiMetric, ExtraTask, BookingChain, RoleChainReport, ChainLearning, ProjectsReport, ProjectChain, MatrixBookingSuggestion, MatrixSuggestAccuracy, MatrixAutonomyStatus, AIUsageReport, MatrixEyesReport, MediaBrief, DisciplineRecord, PeerNote, WorkplaceIssue, WorkplaceIssueDetail, MyPeerState, PeerVoiceReport, PendingCrewRating, BookingTimeline, MonitorDeskCounts, DesignAsset, DepartmentHead, Department, MonitorReview, NetworkPriceItem, NetworkCostResponse, CameraCostRequest, CameraCostResponse, DecorRate, CreateLeaderInvoiceRequest, ServicePriceStats, ServicePriceSuggestion, QuotationItem, Quotation, QuotationVersion, LeaderInvoiceAdjustment, Permission, VehicleOption, Vehicle, VehicleDocument, VehiclePhoto, VehicleMissionRating, VehicleMission, DriverRatingSummary, VehicleBooking, VehicleLog, EmployeeFuelStat, VehicleIncidentAttachment, VehiclePart, VehicleAlert, FuelAnomalyResult, VehicleExpenseSummary, FleetDashboardSummary, VehicleDailyRating, VehicleScoreSummary, TechnicianWashSummary, VehicleWashMonthly, VehicleIncident, VehicleMonthlyStatus, QualityIssue, RoleKpiLeaderboard, KpiEvaluation, KpiCriterion, EmployeeLedger, WorkReport, TechnicianKpi, CartItem, BookingCrewInventoryState, InventoryCheck, Notification, PersonalTool, BookingToolCheck, VehicleTool, OnDemandTool, PersonalToolTemplateItem, PersonalToolExemption, VehicleToolCheck, DeviceMaintenanceTicket, TeamInventoryCheck, ToolRequestItem, PrivacyPolicyPoint, PrivacyPolicyStatus, PersonalToolStatus, PersonalToolEvent, ToolRequestReason, ComplaintType, Complaint, ProcurementRequestType, ProcurementRequest, ProcurementStats, ProjectChecklist, TechShowcaseItem, Exhibition, ProductRequest, ProductProcurement, FulfillProductInput, ServiceStudyReport, ServiceStudy, DisciplinePoints, DisciplineEvent, TodayPulse, FinanceSummary, SolarCategory, SolarComponent, SolarSystem, SolarInstallation, SolarStats, SaveSolarSystemInput, SkillWithService, TrainingProgram, SaveTrainingProgramInput, CustomerPage, LeaderInvoicePage, BackupOverview, PartialCompleteBody, ProgressRecentRow, BookingProgressReport, SurveyReportBody, BookingSurveyReport, DuplicateCandidate, Achievement, VoiceAnalysis, SuggestedCrewMember, FreeWorkReason, EmployeeLetter, GpsMonitorSnapshot, MatrixDecisions, TcInsight, TcCurriculum, TcQuestion, TcExamReportRow, TcPlan, TcMyExam, Justification, PaymentAuditRow, CourierMessage, CourierState, MatrixWatch, MatrixGuideRule, MatrixProposal, EmployeeReport, MatrixBusiness, WorkScheduleRow, GroupPerformance, MatrixFeedItem, MatrixTrendDay, LateFocus, RemoteEntry, RemoteTimer, RemoteSummary, RemoteHistory, StaffRatingStage, PendingStaffRating, StaffScoreBoard, StaffScoreDetail, AttendanceGateState, AttendanceClaim, ProjectMoney, ProjectPaymentsState, TechSupplierIn, TechSupplier, VehicleTracking, CarWashVehicle, MonthStats, ProjectDelay, StageLimitRow, ProcurementWatch, MatrixChatRow, MatrixChatMessage, MatrixMission, StaffPointsRow, MatrixLesson, DepartOrigin, CompanyLocation, BookingTrack, FieldMap, PurchaseControlRow, PurchasePriceRow, PurchaseAlert, SupplierScore, SupplierReviewInput, StandardPrice, StandardSuggestion, PurchaseMonth, PurchaseBias, DrillResult, DutiesReport, WorkPhoto, InvoiceDeleteRequest, CarryoverItem, AccessGroup, EmployeeAccess, RoleDriftRow, PermissionAuditRow, RemotePlan, ManagerRequest, BotReport, RevisitReport, CustomerSignature, BookingHistoryLine } from './apiTypes'
export type * from './apiTypes'

export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

/** ٧ي (11-04): الخروج يمسح كوكي الجلسة بالسيرفر — بلاها تضل الجلسة بالمتصفح بعد الخروج. */
export function serverLogout(): void {
  try {
    void fetch(`${API_URL}/auth/logout`, { method: 'POST', credentials: 'include', keepalive: true, headers: { 'X-Requested-With': 'fetch' } }).catch(() => {})
  } catch { /* الخروج المحلي يكمل حتى لو السيرفر ما رد */ }
}

export const LEAVE_KINDS: { value: LeaveKind; label: string; note?: string }[] = [
  { value: 'REGULAR', label: 'إجازة قبل يومين', note: 'تنطلب قبل يومين على الأقل — حتى يرتبون الشفت' },
  { value: 'URGENT',  label: 'إجازة طارئة',    note: 'بلا مهلة — الطارئ ما ينتظر' },
]

export const REVIEW_RATINGS: {
  value: ReviewRating; label: string; icon: string; tone: 'emerald' | 'amber' | 'red' | 'orange'; hint: string
}[] = [
  { value: 'POSITIVE',       label: 'إيجابي',       icon: '👍', tone: 'emerald', hint: 'شغل زين — ينتسجّل بسجله' },
  { value: 'NEEDS_TRAINING', label: 'يحتاج تدريب',  icon: '🎓', tone: 'amber',   hint: 'نقص مهارة — يتحوّل متدرب تلقائياً' },
  { value: 'MISCONDUCT',     label: 'مخالفة سلوك',  icon: '⚠️', tone: 'red',     hint: 'أسلوب أو تصرّف — بلاغ للإدارة، وهي تقرر' },
  { value: 'COMMITMENT',     label: 'خلل بالالتزام', icon: '⏰', tone: 'orange',  hint: 'تأخر أو غياب عن الشغلة — بلاغ للإدارة' },
]

export const auditVerdictLabels: Record<AuditVerdict, string> = {
  NEEDS_FOLLOWUP: 'مطلوب متابعة',
  RECHECK_RATING: 'مراجعة التقييم',
  APPROVED: 'معتمد',
}

export const eventKindLabels: Record<ComplaintEvent['kind'], string> = {
  CREATED: 'انفتحت الشكوى',
  CONTACTED: 'تواصل مع الزبون',
  RATED: 'سجّل تقييم الزبون',
  NOTED: 'كتب ملاحظات',
  ASSIGNED: 'انكلّف مهندس',
  AUDITED: 'تدقيق',
  RESOLVED: 'انحلّت',
}

/** مهارات القيادة الخمس — نفس القائمة بالخادم (model.LeaderSkills). */
export const LEADER_SKILLS = ['القيادة', 'إدارة الفريق', 'حل المشكلات', 'التواصل', 'اتخاذ القرار'] as const

export const auditCloseLabels: Record<AuditCloseAction, string> = {
  PENALIZE: 'سجّل مخالفة انضباط',
  NO_FAULT: 'تأكدت — ماكو خطأ',
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
  serverLogout()
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

// tokenClaim حقل من داخل التوكن الحالي (بلا سؤال السيرفر).
function tokenClaim(name: string): unknown {
  const t = currentToken()
  if (!t) return null
  try {
    return JSON.parse(atob(t.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))[name] ?? null
  } catch {
    return null
  }
}

/** loginSessionKey مفتاح ثابت لجلسة الدخول الحالية (وقت صدور التوكن) —
 *  يتبدّل بكل تسجيل دخول جديد. «بعدين» بالتبريرات تنربط بيه (ع 10-23). */
export function loginSessionKey(): string {
  const iat = tokenClaim('iat')
  return iat == null ? '' : String(iat)
}

/** 🕶️ الجلسة الحالية دخول متخفي من المالك؟ (ع 10-23) */
export function isStealthSession(): boolean {
  return !!tokenClaim('stealthBy')
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

/** storeFile (٧ل): يرفع الملف لمخزن الملفات ويرجّع رابطه (/api/files/…) حتى
 *  ما ينحفظ base64 بالقاعدة. لو فشل الرفع يرجّع data URL مثل قبل — ما نضيّع الملف. */
export async function storeFile(file: File, folder: string): Promise<string> {
  try {
    return (await api.uploadFile(file, folder)).url
  } catch {
    return await new Promise<string>((resolve, reject) => {
      const r = new FileReader()
      r.onload = () => resolve(String(r.result || ''))
      r.onerror = reject
      r.readAsDataURL(file)
    })
  }
}

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) { super(message); this.name = 'ApiError'; this.status = status }
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
    throw new ApiError(body.error || describeHttpStatus(res.status), res.status)
  }
  // عين ماتركس: كل إجراء ناجح (حفظ/تعديل/حذف) تنتبهله العين لحظة.
  const method = (options?.method || 'GET').toUpperCase()
  if (method !== 'GET' && !path.startsWith('/auth') && typeof window !== 'undefined') {
    window.dispatchEvent(new Event('matrix-saw'))
    // وبعد ثانيتين (الخادم يفحص التذكيرات لحظة الحفظ): العين والجرس يتحدّثون،
    // فلمن الموظف ينفّذ التذكير (يسجّل حضور مثلاً) العين ترجع هادئة فوراً.
    window.clearTimeout(refreshTimer)
    refreshTimer = window.setTimeout(() => window.dispatchEvent(new Event('matrix-refresh')), 2000)
  }
  return res.json()
}
let refreshTimer = 0

// apiRequest نفس طريقة الاتصال الموحدة (الجلسة المنتهية، رسائل الأخطاء بالعراقي،
// تحديث عين ماتركس) — للشاشات الي چانت عندها طريقة خاصة بيها.
export const apiRequest = request

export async function streamMatrixChat(id: string, text: string, image: string | undefined, emit: (event: MatrixStreamEvent) => void, signal: AbortSignal) {
  const token = currentToken()
  const res = await fetch(`${API_URL}/matrix/chats/${encodeURIComponent(id)}/stream`, {
    method: 'POST', signal,
    headers: { 'Content-Type': 'application/json', Accept: 'text/event-stream', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: JSON.stringify({ text, image }),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    if (res.status === 401) handleSessionExpired(body.error)
    throw new ApiError(body.error || describeHttpStatus(res.status), res.status)
  }
  if (!res.body || !res.headers.get('Content-Type')?.includes('text/event-stream')) throw new Error('الخادم ما رجّع بث محادثة صالح')
  await readMatrixStream(res.body, emit, signal)
}

export async function streamSecretaryMessage(input: { text?: string; audio?: Blob; conversationId?: string }, emit: (event: AIStreamEvent<SecretaryReply>) => void, signal: AbortSignal) {
  const token = currentToken()
  let body: FormData | string
  if (input.audio) {
    body = new FormData()
    body.append('file', input.audio, 'voice')
    if (input.conversationId) body.append('conversationId', input.conversationId)
  } else body = JSON.stringify({ text: input.text, conversationId: input.conversationId })
  const res = await fetch(`${API_URL}/secretary/message?stream=1`, {
    method: 'POST', signal, body,
    headers: { Accept: 'text/event-stream', ...(!input.audio ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) },
  })
  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    if (res.status === 401) handleSessionExpired(error.error)
    throw new ApiError(error.error || describeHttpStatus(res.status), res.status)
  }
  if (!res.body || !res.headers.get('Content-Type')?.includes('text/event-stream')) throw new Error('الخادم ما رجّع بث ايريس صالح')
  await readAIStream<SecretaryReply>(res.body, emit, signal, (message) => {
    const value = message as SecretaryReply | null
    return Boolean(value?.messageId && typeof value.reply === 'string' && Array.isArray(value.cards) && typeof value.conversationId === 'string')
  })
}

// ═══ ليش نترجم رقم الحالة ═══
//
// «Request failed: 413» ما يگول للموظف ولا شي. هو يعرف إنه ضغط زر
// وما اشتغل — بس ما يعرف **ليش** ولا شنو يسوي.
//
// وأخطر شي: الأرقام الي تجي من nginx أو من الوسيط — هذني ما يوصلهن
// نظامنا أصلاً، فما بيهن رسالة عربية. الموظف يشوف رقم ويتصل بالدعم،
// والدعم يدور ساعة على شي حله سطر واحد.
//
// ⚠️ 413 بالذات: صار فعلاً بالإنتاج. الموظف يرفع صورة وصل الدوار،
// وnginx يرفضها قبل ما توصل السيرفر، والموظف يشوف رقم غامض ويعيد
// المحاولة عشر مرات بلا فايدة.
function describeHttpStatus(status: number): string {
  switch (status) {
    case 413:
      return 'الملف كبير جداً — جرّب صورة أصغر أو صوّرها بجودة أقل.'
    case 403:
      return 'ما عندك صلاحية لهاي العملية — راجع المدير.'
    case 404:
      return 'الشي الي تدور عليه مو موجود — يمكن انحذف.'
    case 408:
    case 504:
      return 'الطلب طوّل أكثر من اللازم — جرّب مرة ثانية.'
    case 429:
      return 'محاولات كثيرة بوقت قصير — استنى شوي وجرّب.'
    case 500:
    case 502:
    case 503:
      return 'عطل بالسيرفر — جرّب بعد دقيقة، وإذا تكرر خبّر الدعم.'
    default:
      return `تعذر تنفيذ الطلب (رمز ${status}).`
  }
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

export const AUDIT_VERDICTS: { key: 'MATCHED' | 'MISMATCH' | 'PRICE_ERROR' | 'FREE' | 'CREDIT'; label: string; cls: string }[] = [
  { key: 'MATCHED', label: '✅ مطابق', cls: 'bg-emerald-100 text-emerald-800' },
  { key: 'MISMATCH', label: '⚠️ غير مطابق', cls: 'bg-amber-100 text-amber-800' },
  { key: 'PRICE_ERROR', label: '❌ خطأ بالسعر', cls: 'bg-red-100 text-red-800' },
  /** الصافي صفر وهو **الصح** — مو فرق بالمبلغ */
  { key: 'FREE', label: '🎁 صيانة مجانية', cls: 'bg-teal-100 text-teal-800' },
  /** (ع) 10-14: الفلوس مؤجلة — بسجل الآجل، وينعتمد بلا رقم */
  { key: 'CREDIT', label: '⏳ آجل', cls: 'bg-violet-100 text-violet-800' },
]

export const bookingDeleteChannelLabels: Record<BookingDeleteChannel, string> = {
  WEBSITE: 'موقع ويب',
  MOBILE_APP: 'تطبيق الجوال',
  CALL_CENTER: 'مركز الاتصال',
}

export const bookingDeleteTypeLabels: Record<BookingDeleteRequestType, string> = {
  RECURRING_DUPLICATE: 'حجز متكرر',
  CUSTOMER_CANCEL: 'إلغاء من الزبون',
  DATA_CORRECTION: 'تصحيح بيانات',
  CUSTOMER_INQUIRY: 'استفسار زبون (ما نفّذ)',
}

/**
 * ═══ «الزبون ما رد» — خيار بنفس قائمة نوع الطلب، بس مو طلب حذف ═══
 *
 * «اريد منك خيار الزبون مارد حته مناك ننطي مارد… زبون مارد ماريده
 * يضهرلي بمكانين، اريده يضهر فقط بالمكان مال الزبون مارد».
 *
 * ⚠️ ما ينفتح إله طلب حذف ولا يمرّ على المراقب — قراره صريح: **ينشال
 * فوراً**. اختياره ينقل الحجز لحالة WAITING الموجودة أصلاً
 * (`markBookingWaiting`)، فيختفي من طابور الشغل ويظهر بمكانه الواحد،
 * وصاحب النظام يرجّعه يدوياً وقت ما يريد. **ما ينحذف ولا ينأرشف.**
 *
 * لهذا هو **مو** قيمة بـBookingDeleteRequestType ولا يوصل الخادم
 * كنوع طلب — لو انضاف هناك، صار طلب حذف وهو مو حذف.
 */
export const BOOKING_NO_ANSWER_CHOICE = 'CUSTOMER_NO_ANSWER'
export const bookingNoAnswerLabel = 'الزبون ما رد — ينزاح للانتظار فوراً'

export const COORDINATION_ALERT_THRESHOLD = 10
export const IT_KIND_LABELS: Record<ItAssetKind, string> = {
  COMPUTER: '🖥️ حاسبة', LAPTOP: '💻 لابتوب', SERVER: '🗄️ مخدّم', NETWORK: '📡 شبكة',
  PRINTER: '🖨️ طابعة', CAMERA: '📷 كاميرا', PHONE: '📱 هاتف', OTHER: '🔧 أخرى',
}
export const IT_STATUS_LABELS: Record<ItAssetStatus, string> = {
  ACTIVE: 'شغّال', REPAIR: 'بالتصليح', SPARE: 'احتياط', RETIRED: 'مستبعد',
}
export const IT_LOG_LABELS: Record<string, string> = {
  NOTE: 'ملاحظة', REPAIR: 'تصليح', MAINTENANCE: 'صيانة', MOVE: 'نقل', STATUS: 'تغيير حالة',
}

export const STAGE_BUCKETS: { key: StageBucket; label: string; icon: string }[] = [
  // ⚠️ ماكو «مؤجّل قبل التثبيت»: «ما عدي هيج شي، شلون أأجّل موعد
  // وأني أصلاً ما محددله موعد؟» — والتأجيل بالكود ينقل من موعد لموعد.
  { key: 'POSTPONED_AFTER_CONFIRM', label: 'مؤجّلة', icon: '⏳' },
  { key: 'NO_ANSWER_BEFORE_CONFIRM', label: 'ما رد — قبل التثبيت', icon: '📵' },
  { key: 'NO_ANSWER_AFTER_CONFIRM', label: 'ما رد — بعد التثبيت', icon: '📵' },
  { key: 'CANCELLED_BEFORE_CONFIRM', label: 'ملغى قبل التثبيت', icon: '✖️' },
  { key: 'CANCELLED_AFTER_CONFIRM', label: 'ملغى بعد التثبيت', icon: '✖️' },
  { key: 'AWAITING_CUSTOMER_DECISION', label: 'بانتظار موافقة الزبون', icon: '🤔' },
]

export const AI_METRIC_LABELS: Record<string, string> = {
  STOP_RATE: 'نسبة توقف العمل',
  STOP_MINUTES_AVG: 'متوسط الوقت الضايع بالتوقف',
  MATERIAL_MISS_RATE: 'توقف بسبب مادة ما انطلبت',
  SCOPE_CREEP_RATE: 'زيادة طلبات الزبون بالموقع',
  PROCUREMENT_DELAY: 'تأخر توفير المواد',
  LATE_START_RATE: 'نسبة التأخر بالخروج للزبون',
}

export const EXTRA_TASK_STATUS: Record<ExtraTask['status'], string> = {
  NEW: 'جديدة',
  IN_PROGRESS: 'قيد التنفيذ',
  DONE: 'منجزة',
  CANCELLED: 'ملغاة',
}

/** يفكّ لقطة النسخة المؤرشفة. يرجّع null لو انكسرت — الوثيقة
 *  التاريخية ما تسقّط الشاشة، تنعرض «تعذّر قراءتها». */
export function parseQuotationSnapshot(snapshot: string): Quotation | null {
  try {
    const json = typeof atob === 'function' ? atob(snapshot) : snapshot
    return JSON.parse(decodeURIComponent(escape(json))) as Quotation
  } catch {
    try { return JSON.parse(snapshot) as Quotation } catch { return null }
  }
}

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

export const complaintTypeLabels: Record<ComplaintType, string> = {
  DELAY: 'تأخير بالتنفيذ',
  DISORGANIZED: 'عمل غير منظم',
  TECHNICAL: 'مشكلة فنية',
  EXECUTION_ERROR: 'خطأ تنفيذي',
  INCOMPLETE: 'لم يتم إكمال العمل',
  OTHER: 'أخرى',
}

export const procurementTypeLabels: Record<ProcurementRequestType, string> = {
  PERSONAL_SUPPLY: 'احتياجات شخصية',
  CUSTOMER_PRODUCT: 'منتج للزبون',
  MANUAL_SUPPLY: 'طلب مادة يدوي',
}

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

export const DEPARTMENTS = ['الفنية', 'المبيعات', 'الصيانة', 'الإدارة', 'المستودع']
export const REVIEW_GRADES = ['ممتاز', 'جيد جداً', 'جيد', 'يحتاج تحسين']

export const api = {
  // 📊 ملف سوّاه كمال (دازاي 10-28) — الرابط من بطاقة السكرتير
  downloadSecretaryFile: (link: string, name: string) => downloadFile(link, name),
  secretaryVoiceCompare: (messageId: string) =>
    request<{ primary: string; results: { name: string; text: string; seconds: number; error?: string }[] }>(
      `/voice/compare/${encodeURIComponent(messageId)}`,
      { method: 'POST' },
    ),
  secretaryAudio: async (messageId: string): Promise<Blob> => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/secretary/audio/${encodeURIComponent(messageId)}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    if (!res.ok) throw new Error('ما گدرت أجيب البصمة')
    return res.blob()
  },
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
  suggestProductSpecs: (name: string, modelName: string) =>
    request<{ specs: string }>('/products/specs-suggest', { method: 'POST', body: JSON.stringify({ name, modelName }) }),
  // توقف العمل: الليدر بدأ وما كدر يكمّل. السبب إجباري.
  stopBookingWork: (id: string, reason: string) =>
    request<Booking>(`/bookings/${id}/stop-work`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  resumeBookingWork: (id: string) =>
    request<Booking>(`/bookings/${id}/resume-work`, { method: 'PUT' }),
  // ── الطلبات: كتاب رسمي من الموظف للإدارة ──
  getLetterAddressees: () => request<string[]>('/letters/addressees'),
  createLetter: (body: { addressedTo: string; subject: string; body: string }) =>
    request<EmployeeLetter>('/letters', { method: 'POST', body: JSON.stringify(body) }),
  getMyLetters: () => request<EmployeeLetter[]>('/letters/mine'),
  /** صندوق الإدارة — للمالك ومدير النظام */
  getLetters: (status?: string) =>
    request<EmployeeLetter[]>(`/letters${status ? '?status=' + status : ''}`),
  decideLetter: (id: string, body: { approve: boolean; note: string }) =>
    request<EmployeeLetter>(`/letters/${id}/decide`, { method: 'PUT', body: JSON.stringify(body) }),

  getDisciplinePoints: () => request<DisciplinePoints[]>('/discipline'),
  getDisciplineEvents: (employeeId?: string) =>
    request<DisciplineEvent[]>(`/discipline/events${employeeId ? '?employeeId=' + employeeId : ''}`),
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
  /** حجوزات خدماتي المؤشّرة الي ورقها عليّ كمسؤول خدمة. */
  getManagerPaperwork: () => request<Booking[]>('/bookings/manager-paperwork'),
  /** حجوزات خدمتي المنجزة مرشّحة بالنوع — منتقي الربط بفاتورة الخدمة.
   *  ⚠️ الموظف من التوكن بالخادم، فما ينمرّر هنا. */
  getServicePaperwork: (kind: 'GPS' | 'DASHCAM') =>
    request<Booking[]>(`/bookings/service-paperwork?kind=${kind}`),
  /** الورق (تقرير + فاتورة) على مسؤول الخدمة مو على الفني. */
  setServiceManagerPaperwork: (serviceId: string, enabled: boolean) =>
    request<{ enabled: boolean }>(`/services/${serviceId}/manager-paperwork`, { method: 'PUT', body: JSON.stringify({ enabled }) }),
  /** نوع خدمة الفاتورة اليدوية — `''` يمسح التأشير. */
  setServiceKind: (serviceId: string, kind: 'GPS' | 'DASHCAM' | '') =>
    request<{ kind: string }>(`/services/${serviceId}/kind`, { method: 'PUT', body: JSON.stringify({ kind }) }),

  // تتبع الموقع الحي

  getEmployees: () => request<Employee[]>('/employees'),
  // ساعات العمل من البيت — صلاحية remote_hours_manage
  // الفترة المفتوحة (من آخر تصفير)
  getRemoteHours: (employeeId = '') =>
    request<{ entries: RemoteEntry[]; totalSeconds: number; plan?: RemotePlan | null }>(`/remote-hours?employeeId=${encodeURIComponent(employeeId)}`),
  getRemoteSummary: () => request<RemoteSummary[]>('/remote-hours/summary'),
  getRemoteHistory: (month = '', employeeId = '') => request<RemoteHistory>(`/remote-hours/history?month=${month}&employeeId=${encodeURIComponent(employeeId)}`),
  addRemoteHours: (b: { employeeId: string; date: string; hours: number; minutes: number; seconds: number; note?: string }) =>
    request<{ id: string; seconds: number }>('/remote-hours', { method: 'POST', body: JSON.stringify(b) }),
  deleteRemoteHours: (id: string) => request<{ ok: boolean }>(`/remote-hours/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  updateRemoteHours: (id: string, b: { hours: number; minutes: number; seconds: number; note?: string }) =>
    request<{ ok: boolean; seconds: number }>(`/remote-hours/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(b) }),
  resetRemoteHours: () => request<{ closed: number }>('/remote-hours/reset', { method: 'POST' }),
  // الموظف لنفسه — صلاحية remote_hours_self
  getMyRemote: () => request<{ entries: RemoteEntry[]; totalSeconds: number; timer: RemoteTimer; plan: RemotePlan | null }>('/remote-hours/mine'),
  /** (دازاي 11-01) المطلوب بالشهر وسعر الساعة — الموظف مرة وحدة، والإدارة تعدّل */
  setRemotePlan: (b: { targetHours: number; hourlyRateIQD: number; employeeId?: string }) =>
    request<{ plan: RemotePlan }>('/remote-hours/plan', { method: 'PUT', body: JSON.stringify(b) }),
  addMyRemote: (b: { date: string; hours: number; minutes: number; seconds: number; note?: string }) =>
    request<{ id: string; seconds: number }>('/remote-hours/mine', { method: 'POST', body: JSON.stringify(b) }),
  remoteTimer: (action: 'start' | 'pause' | 'finish', note?: string) =>
    request<{ timer: RemoteTimer; saved?: number }>(`/remote-hours/timer/${action}`, { method: 'POST', body: JSON.stringify({ note }) }),
  exportRemoteHours: (month: string, employeeId = '', label = '') =>
    downloadFile(`/remote-hours/export?month=${month}&employeeId=${encodeURIComponent(employeeId)}`, `ساعات-البيت-${label ? label + '-' : ''}${month}.xlsx`),
  getArchivedEmployees: () => request<Employee[]>('/employees/archived'),
  getSuspendedEmployees: () => request<{ id: string; name: string; role: Employee['role']; status: string; reason: string; suspendedAt: string | null; suspendedBy: string }[]>('/employees/suspended'),
  suspendEmployee: (id: string, reason: string) => request<{ ok: boolean }>(`/employees/${encodeURIComponent(id)}/suspend`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  reactivateEmployee: (id: string) => request<{ ok: boolean }>(`/employees/${encodeURIComponent(id)}/reactivate`, { method: 'PUT' }),
  updateEmployeeAll: (id: string, data: Record<string, unknown>) => request<Employee>(`/employees/${encodeURIComponent(id)}/all`, { method: 'PUT', body: JSON.stringify(data) }),
  getWorkSchedule: () => request<WorkScheduleRow[]>('/work-schedule'),
  setEmployeeSchedule: (id: string, data: { shift?: string; shiftStart?: string; shiftEnd?: string }) =>
    request<Employee>(`/employees/${encodeURIComponent(id)}/schedule`, { method: 'PUT', body: JSON.stringify(data) }),
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
      secondaryRoles?: EmployeeRole[]
    },
  ) => request<Employee>('/employees', { method: 'POST', body: JSON.stringify(data) }),
  /** 🔍 وضع الفحص (ع 10-23): المالك يشوف حساب موظف مثل ما يشوفه هو — قراءة بس،
   *  السيرفر يرفض أي تعديل. جلسة المالك تنحفظ بالتبويب حتى «رجوع» يرجّعه. */
  inspectEnter: async (employeeId: string) => {
    const r = await request<Employee & { token: string }>('/auth/stealth', { method: 'POST', body: JSON.stringify({ employeeId }) })
    try {
      sessionStorage.setItem('inspectReturn', JSON.stringify({ token: localStorage.getItem('authToken'), employee: localStorage.getItem('currentEmployee') }))
      sessionStorage.setItem('inspectName', r.name)
    } catch { /* بلا تخزين: «رجوع» يطلّعه لشاشة الدخول */ }
    const { token, ...emp } = r
    localStorage.setItem('authToken', token)
    // الواجهة تقرا الجلسة من currentEmployee — بلاه تطلع شاشة الدخول
    localStorage.setItem('currentEmployee', JSON.stringify(emp))
    fileToken = ''
    fileTokenAt = 0
    window.location.assign(import.meta.env.BASE_URL)
  },
  /** يرجّع المالك لحسابه. */
  inspectExit: () => {
    let saved: { token: string | null; employee: string | null } | null = null
    try { saved = JSON.parse(sessionStorage.getItem('inspectReturn') || 'null'); sessionStorage.removeItem('inspectReturn'); sessionStorage.removeItem('inspectName') } catch { /* */ }
    if (saved?.token) {
      localStorage.setItem('authToken', saved.token)
      if (saved.employee) localStorage.setItem('currentEmployee', saved.employee); else localStorage.removeItem('currentEmployee')
    } else {
      localStorage.removeItem('authToken'); localStorage.removeItem('currentEmployee')
    }
    fileToken = ''
    fileTokenAt = 0
    window.location.assign(import.meta.env.BASE_URL)
  },
  login: async (username: string, password: string) => {
    const result = await request<Employee & { token: string; realm?: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    const { token, realm, ...employee } = result
    localStorage.setItem('authToken', token)
    // وسم الملفات صار باسم الموظف — وسم الحساب السابق بنفس الجهاز ما ينفع.
    fileToken = ''
    fileTokenAt = 0
    // الطبقة تنخزن حتى إعادة فتح الصفحة ترجّعك لنفس النظام.
    // نفس اليوزر وباسورد ثاني = نظام ثاني (فكرة PPSK).
    localStorage.setItem('authRealm', realm || 'staff')
    return employee
  },
  /** الطبقة الحالية — staff = نظام الشركة، command = مركز القيادة */
  currentRealm: () => localStorage.getItem('authRealm') || 'staff',
  setCommandPassword: (newPassword: string, employeeId?: string) =>
    request<{ ok: boolean }>('/auth/command-password', {
      method: 'PUT', body: JSON.stringify({ newPassword, employeeId }),
    }),
  getMyProjectDelays: () => request<ProjectDelay[]>('/projects/my-delays'),
  getProjectDelays: () => request<ProjectDelay[]>('/projects/delays'),
  // (ع) 10-19: حذف نهائي للمرفوض بسبب (نسخته تبقى بأرشيف المحذوفات)
  purgeProject: (id: string, reason: string) =>
    request<{ name: string }>(`/projects/${encodeURIComponent(id)}/purge`, { method: 'DELETE', body: JSON.stringify({ reason }) }),
  // (ع) 10-19: المشروع المرفوض يرجع لمرحلته قبل الرفض
  restoreProject: (id: string, reason: string) =>
    request<{ stage: string }>(`/projects/${encodeURIComponent(id)}/restore`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  addProjectDelayReason: (id: string, reason: string) =>
    request<{ ok: boolean }>(`/projects/${id}/delay-reason`, { method: 'POST', body: JSON.stringify({ reason }) }),
  getStageLimits: () => request<StageLimitRow[]>('/projects/stage-limits'),
  setStageLimit: (stage: string, days: number | null) =>
    request<StageLimitRow[]>('/projects/stage-limits', { method: 'PUT', body: JSON.stringify({ stage, days }) }),
  getProcurementWatch: () => request<ProcurementWatch>('/monitor/procurement-watch'),
  getMatrixChats: () => request<{ chats: MatrixChatRow[]; enabled: boolean }>('/matrix/chats'),
  createMatrixChat: (signal?: AbortSignal) => request<MatrixChatRow>('/matrix/chats', { method: 'POST', signal }),
  getMatrixChat: (id: string) => request<MatrixChatMessage[]>(`/matrix/chats/${id}`),
  deleteMatrixChat: (id: string) => request<{ ok: boolean }>(`/matrix/chats/${id}`, { method: 'DELETE' }),
  sendMatrixChat: (id: string, text: string, image?: string) =>
    request<MatrixChatMessage>(`/matrix/chats/${id}/messages`, { method: 'POST', body: JSON.stringify({ text, image }) }),
  correctMatrixChat: (id: string, msgId: string, correction: string) =>
    request<{ ok: boolean }>(`/matrix/chats/${id}/messages/${msgId}/correct`, { method: 'POST', body: JSON.stringify({ correction }) }),
  getMatrixLessons: () => request<MatrixLesson[]>('/matrix/lessons'),
  getEmployeeAdminScope: (id: string) => request<{ roles: string[] }>(`/employees/${id}/admin-scope`),
  setEmployeeAdminScope: (id: string, roles: string[]) =>
    request<{ roles: string[] }>(`/employees/${id}/admin-scope`, { method: 'PUT', body: JSON.stringify({ roles }) }),
  getMyAdminScope: () => request<{ all: boolean; roles: string[] }>('/employees/my-admin-scope'),
  getMatrixMissions: () => request<MatrixMission[]>('/matrix/missions'),
  getMatrixMission: (id: string) => request<{ mission: MatrixMission; logs: { day: string; text: string }[] }>(`/matrix/missions/${id}`),
  createMatrixMission: (body: { employeeId: string; note?: string; days: number }) =>
    request<MatrixMission>('/matrix/missions', { method: 'POST', body: JSON.stringify(body) }),
  stopMatrixMission: (id: string) => request<{ ok: boolean }>(`/matrix/missions/${id}`, { method: 'DELETE' }),
  disableMatrixLesson: (id: string) => request<{ ok: boolean }>(`/matrix/lessons/${id}`, { method: 'DELETE' }),
  getMonthCompare: (months: string[]) => request<MonthStats[]>(`/stats-management/month-compare?months=${months.join(',')}`),
  getCommandAccounts: () =>
    request<{ id: string; name: string; username: string | null; setAt: string | null }[]>('/auth/command-accounts'),
  clearCommandPassword: (employeeId: string) =>
    request<{ ok: boolean }>(`/auth/command-password/${employeeId}`, { method: 'DELETE' }),
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
  getCustomer: (id: string) => request<Customer>(`/customers/${encodeURIComponent(id)}`),
  /** بند ٧م: صفحة زبائن وحدة ويا أرقام الرأس من الخادم. */
  getCustomersPage: (q: { q?: string; location?: string; kind?: string; claim?: string; gps?: string; page: number; perPage: number }) =>
    request<CustomerPage>(`/customers/page?${new URLSearchParams(Object.entries(q).filter(([, v]) => v !== '' && v !== undefined).map(([k, v]) => [k, String(v)])).toString()}`),
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

  // التقييم يمشي **بنفس** طلب التواصل: مهندس الجودة يسأل الزبون وهو
  // بالمكالمة، فتقسيمهن لخطوتين يعني نص التقييمات ما تنسجّل.
  complaintNoAnswer: (id: string) => request<Complaint>(`/complaints/${id}/no-answer`, { method: 'PUT' }),
  /** ⛔ سدّها فوراً — الزبون ما يرد (ع 10-23) */
  closeComplaintNoAnswer: (id: string) => request<Complaint>(`/complaints/${id}/no-answer-close`, { method: 'PUT' }),
  setComplaintContacted: (id: string, contacted: boolean, rating?: number | null) =>
    request<Complaint>(`/complaints/${id}/contact`, {
      method: 'PUT',
      body: JSON.stringify({ contacted, rating: rating ?? null }),
    }),
  auditComplaint: (id: string, verdict: AuditVerdict, note?: string) =>
    request<Complaint>(`/complaints/${id}/audit`, {
      method: 'PUT', body: JSON.stringify({ verdict, note: note || null }),
    }),
  getComplaintEvents: (id: string) =>
    request<ComplaintEvent[]>(`/complaints/${id}/events`),
  setComplaintNotes: (id: string, notes: string) =>
    request<Complaint>(`/complaints/${id}/notes`, { method: 'PUT', body: JSON.stringify({ notes }) }),
  getDailyAudit: (date?: string) =>
    request<DailyAuditReport>(`/finance/daily-audit${date ? `?date=${date}` : ''}`),
  /**
   * بحث بحجوزات التدقيق **بلا تاريخ** — برقم الحجز أو اسم الزبون أو
   * هاتفه، عبر كل الأيام. المحاسب يدوّر على حجز ما يعرف تاريخه.
   */
  searchDailyAudit: (q: string) =>
    request<DailyAuditReport>(`/finance/daily-audit?q=${encodeURIComponent(q)}`),

  // التدقيق: قرار المحاسب — مطابق أو بلاغ خطأ ينوجّه للمعني
  /** FREE: صيانة مجانية — الصافي صفر **بقصد** (ضمان/إعادة عمل/تعويض).
   *  بلاها المحاسب لازم يكتب مبلغاً ما انستلم أو يأشّر «غير مطابق»
   *  فتنفتح مخالفة على شغل سليم. والسبب اختياري. */
  auditBooking: (id: string, data: { action: 'VERIFY' | 'MISMATCH' | 'PRICE_ERROR' | 'FREE' | 'CREDIT'; amountCollected?: number; advancePaid?: number; note?: string; freeReasonId?: string }) =>
    request<Booking | AuditIssue>(`/bookings/${id}/audit`, { method: 'PUT', body: JSON.stringify(data) }),
  // ═══ الآجل (ع) 10-14 ═══
  getCredits: (status: 'open' | 'settled' | 'all', q = '') => request<CreditSale[]>(`/credits?status=${status}&q=${encodeURIComponent(q)}`),
  getCreditPayments: (id: string) => request<{ id: string; amount: number; paidAt: string; note: string | null; byName: string | null }[]>(`/credits/${id}/payments`),
  addCreditPayment: (id: string, data: { amount: number; paidAt?: string; note?: string }) =>
    request<{ ok: boolean }>(`/credits/${id}/payments`, { method: 'POST', body: JSON.stringify(data) }),
  createCredit: (data: { projectCode?: string; bookingCode?: string; title?: string; customerName?: string; phone?: string; amount: number; note?: string; dueDate?: string }) =>
    request<{ ok: boolean }>('/credits', { method: 'POST', body: JSON.stringify(data) }),
  exportCredits: async (month: string) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/credits/export?month=${month}`, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (!res.ok) throw new Error('تعذر تنزيل الإكسل')
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `الآجل-${month}.xlsx`
    a.click()
    URL.revokeObjectURL(a.href)
  },
  getAuditIssues: (status?: string) =>
    request<AuditIssue[]>(`/audit-issues${status ? `?status=${status}` : ''}`),
  // ⚠️ الإغلاق يتطلب إجراءً وسبباً — الخادم يرفض بدونهما. قبلها
  // چان الزر يسكّر البلاغ بلا أثر ولا أحد يدري شنو صار بيه.
  resolveAuditIssue: (id: string, data: { action: AuditCloseAction; reason: string; points?: number }) =>
    request<{ ok: boolean }>(`/audit-issues/${id}/resolve`, {
      method: 'PUT', body: JSON.stringify(data),
    }),

  // أخبار اليوم للشريط — أرقام مجرّدة بلا أسماء
  getDisciplineTodayHeadline: () =>
    request<{ penalized: number; restored: number }>('/discipline/today-headline'),

  // شريط الإعلانات
  /**
   * مفاتيح إطفاء الميزات — القراءة لكل موظف، والكتابة للمالك حصراً
   * (الخادم يرد ٤٠٣ لغيره).
   */
  getSystemSwitches: () => request<Record<string, boolean>>('/system/switches'),
  setSystemSwitch: (key: string, enabled: boolean) =>
    request<Record<string, boolean>>(`/system/switches/${key}`, {
      method: 'PUT', body: JSON.stringify({ enabled }),
    }),

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
      requestsInFlight?: number
      backgroundJobs?: { name: string; lane: 'heavy' | 'light'; runs: number; failures: number; running: boolean; queued: boolean; lastStart: string; lastMs: number; lastWaitMs: number; maxMs: number; everySeconds: number }[]
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
  setQualityMonitorVerdict: (id: string, body: { verdict: 'OK' | 'NOT_OK' | 'SHORTCOMING'; note?: string; targetEmployeeId?: string }) =>
    request<QualityFollowUp>(`/quality-follow-ups/${id}/monitor-verdict`, { method: 'PUT', body: JSON.stringify(body) }),
  /** حكم الجودة: إيجابي (ما يترتب عليه شي)، سلبي (يخصم نقطة من الليدر
   *  إلا إذا طلبنا كشف — وقتها الخصم ينتظر النتيجة)، أو الزبون مارد
   *  (ما نگدر نطلّع حكم — تنسكّر بلا خصم ولا تبرئة) */
  qualityVerdict: (id: string, body: { reportType: 'POSITIVE' | 'NEGATIVE' | 'STUBBORN'; notes: string; needsInspection: boolean }) =>
    request<QualityFollowUp[]>(`/quality-follow-ups/${id}/verdict`, { method: 'POST', body: JSON.stringify(body) }),
  /** نتيجة الكشف الميداني: كلام الزبون صح (يتغرّم الليدر) أو كذب (علامة على الزبون) */
  qualityInspect: (id: string, body: { result: 'CUSTOMER_RIGHT' | 'CUSTOMER_WRONG'; notes: string }) =>
    request<QualityFollowUp[]>(`/quality-follow-ups/${id}/inspect`, { method: 'POST', body: JSON.stringify(body) }),
  /** بعد ما يفتح حجز صيانة حقيقي من شاشة المتابعة — يربطها بيه بدل
   *  ما تروح بالهوا. تبقى معلّقة لحد ما الحجز ينجز. */
  getDecorRates: (all = false) => request<DecorRate[]>(`/decor-rates${all ? '?all=1' : ''}`),
  saveDecorRate: (x: { id?: string; serviceId: string; name: string; unit: string; price: number; active: boolean }) =>
    request<DecorRate[]>('/decor-rates', { method: 'PUT', body: JSON.stringify(x) }),
  getBookingDecor: (id: string) => request<Booking[]>(`/bookings/${id}/decor`),
  createBookingDecor: (id: string, body: { serviceId: string; timing: 'BEFORE' | 'WITH'; scheduledAt?: string; notes?: string }) =>
    request<Booking>(`/bookings/${id}/decor`, { method: 'POST', body: JSON.stringify(body) }),
  getBookingWorkWindow: (id: string) =>
    request<{ startAt: string | null; endAt: string | null; startSrc: string | null; endSrc: string | null }>(`/bookings/${id}/work-window`),
  getRecentProgressReports: () => request<ProgressRecentRow[]>('/progress-reports/recent'),
  linkComplaintBooking: (id: string, bookingId: string) =>
    request<Complaint>(`/complaints/${id}/link-booking`, { method: 'PUT', body: JSON.stringify({ bookingId }) }),
  linkQualityFollowUpBooking: (id: string, bookingId: string) =>
    request<QualityFollowUp>(`/quality-follow-ups/${id}/link-booking`, {
      method: 'PUT',
      body: JSON.stringify({ bookingId }),
    }),

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
  /** ربط/فك ربط هذا الحجز بحجز آخر كإنجاز جزئي لنفس الشغلة — code
   *  فاضي يفك الربط. حراستها صلاحية `booking_partial_link` بالخادم. */
  setPartialJobLink: (id: string, code: string) =>
    request<Booking>(`/bookings/${id}/partial-job-link`, { method: 'PUT', body: JSON.stringify({ code }) }),
  /** تأشير حجز قديم ككشف بأثر رجعي من شاشة تدقيق الحسابات + ربطه
   *  بالحجز الحقيقي الي نتج عنه (code فاضي = بلا ربط). حراستها نفس
   *  صلاحية `booking_partial_link`. */
  markBookingAsSurvey: (id: string, code: string) =>
    request<Booking>(`/bookings/${id}/mark-survey`, { method: 'PUT', body: JSON.stringify({ code }) }),

  // ── تدقيق التكرار (ماتركس) — حجوزات وزبائن مكررون بالغلط ──
  getDuplicateCandidates: (kind?: 'BOOKING' | 'CUSTOMER', status?: 'PENDING' | 'DISMISSED' | 'RESOLVED') => {
    const params = new URLSearchParams()
    if (kind) params.set('kind', kind)
    if (status) params.set('status', status)
    const qs = params.toString()
    return request<DuplicateCandidate[]>(`/duplicate-candidates${qs ? `?${qs}` : ''}`)
  },
  dismissDuplicateCandidate: (id: string) =>
    request<{ ok: boolean }>(`/duplicate-candidates/${id}/dismiss`, { method: 'PUT' }),
  /** شنو راح ينتقل لو دمجنا الزبون الثاني بـkeep. */
  getDuplicateReport: () =>
    request<{ pendingBookings: number; pendingCustomers: number; resolved30: number; detected30: number; byCreator: { name: string; count: number; quick: number }[] }>('/duplicate-candidates/report'),
  duplicateMergePreview: (id: string, keep: string) =>
    request<{ moves: { label: string; count: number }[] }>(`/duplicate-candidates/${id}/merge-preview?keep=${encodeURIComponent(keep)}`),
  mergeDuplicateCustomers: (id: string, keepId: string) =>
    request<{ ok: boolean; note: string }>(`/duplicate-candidates/${id}/merge`, { method: 'POST', body: JSON.stringify({ keepId }) }),
  /** طلب حذف للحجز المكرر — ينتظر موافقة المراقب أو المدير. */
  requestDuplicateBookingDelete: (id: string, bookingId: string) =>
    request<{ ok: boolean; bookingCode: string }>(`/duplicate-candidates/${id}/request-delete`, { method: 'POST', body: JSON.stringify({ bookingId }) }),

  // ── الإنجازات (ماتركس) — تقرير يومي حر من أي موظف بأي دور ──
  createAchievement: (bookingId: string | null, reportText: string) =>
    request<Achievement>('/achievements', { method: 'POST', body: JSON.stringify({ bookingId, reportText }) }),
  getMyAchievements: (limit = 30) => request<Achievement[]>(`/achievements/mine?limit=${limit}`),
  getAchievements: (params?: { employeeId?: string; day?: string; limit?: number }) => {
    const qs = new URLSearchParams()
    if (params?.employeeId) qs.set('employeeId', params.employeeId)
    if (params?.day) qs.set('day', params.day)
    if (params?.limit) qs.set('limit', String(params.limit))
    const s = qs.toString()
    return request<Achievement[]>(`/achievements${s ? `?${s}` : ''}`)
  },
  /** فويس مسج للإنجاز — يبقى بسيرفرنا (multipart) */
  uploadAchievementVoice: async (id: string, blob: Blob, seconds: number) => {
    const fd = new FormData()
    fd.append('file', blob, 'voice')
    fd.append('seconds', String(Math.round(seconds)))
    const token = currentToken()
    const res = await fetch(`${API_URL}/achievements/${id}/voice`, { method: 'POST', body: fd, headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (!res.ok) { const b = await res.json().catch(() => ({})); throw new Error(b.error || describeHttpStatus(res.status)) }
    return res.json() as Promise<{ ok: boolean; seconds: number }>
  },
  /** يجيب الفويس كـBlob (المسار محمي بالترويسة، فما يصير src مباشر) */
  getAchievementVoice: async (id: string) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/achievements/${id}/voice`, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (!res.ok) throw new Error(describeHttpStatus(res.status))
    return res.blob()
  },
  getAchievementVoiceAnalysis: (id: string) => request<{ analysis: VoiceAnalysis; enabled: boolean }>(`/achievements/${id}/voice/analysis`),
  getMatrixDialect: () => request<{ phrases: { phrase: string; meaning: string; uses: number; lastSeen: string }[]; analyzed: number; pending: number; enabled: boolean }>('/ai/dialect'),
  reviewAchievement: (id: string, status: 'GOOD' | 'NEEDS_REVIEW', note: string) =>
    request<Achievement>(`/achievements/${id}/review`, { method: 'PUT', body: JSON.stringify({ status, note }) }),

  /** تأجيل الموعد بطلب الزبون — السبب إجباري وينعد بعدد التأجيلات */
  /** الحجوزات المؤجلة بلا موعد — طابور قرارات الإداري */
  /** كل طلعة صارت على الحجز — بتاريخها وكادرها */
  getBookingVisits: (id: string) => request<BookingVisit[]>(`/bookings/${id}/visits`),

  /** scheduledAt فارغ = تأجيل بلا موعد */
  postponeBooking: (id: string, scheduledAt: string, reason: string) =>
    request<Booking>(`/bookings/${id}/postpone`, { method: 'PUT', body: JSON.stringify({ scheduledAt, reason }) }),

  /** اتصلنا بالزبون وما رد — الحجز ينزاح من طابور الشغل ويضل محفوظ */
  /** ينزيح الحجز لطابور الانتظار. `kind`:
   *  - `NO_ANSWER` (افتراضي) اتصلنا وما رد → «ما وصلت للتنفيذ»
   *  - `CUSTOMER_DECISION` استفسر ورايح يرجعلنا خبر → «بانتظار موافقة الزبون»
   *  ⚠️ الافتراضي «ما رد» حتى أي نداء قديم يبقى بمعناه. */
  // حجز «منجز» بالغلط يرجع للكادر حتى يسجّل إنجاز جزئي — صلاحية booking_reopen
  // (ع) 10-19: «بدء العمل» انضغط بالغلط — المالك بس
  unstartBooking: (id: string, reason: string) =>
    request<{ ok: boolean; code: string; notified: number }>(`/bookings/${encodeURIComponent(id)}/unstart`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  reopenBooking: (id: string, reason: string) =>
    request<{ ok: boolean; code: string; notified: number; warning: string }>(`/bookings/${encodeURIComponent(id)}/reopen`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  markBookingWaiting: (id: string, note?: string, kind?: 'NO_ANSWER' | 'CUSTOMER_DECISION') =>
    request<Booking>(`/bookings/${id}/waiting`, {
      method: 'PUT',
      body: JSON.stringify({ note: note ?? '', kind: kind ?? 'NO_ANSWER' }),
    }),
  /** ═══ تسوية إدارية لحجز قديم — «تم الإنجاز بدون تفاصيل» ═══
   *  للمالك وحده (السيرفر يفرضها). الحجز ينقفل منجزاً بلا كادر ولا
   *  مبالغ، وينعلّم حتى ينستثنى من غرامات الفاتورة والتقرير. */
  settleLegacyBooking: (id: string, note?: string) =>
    request<Booking>(`/bookings/${id}/settle-legacy`, { method: 'POST', body: JSON.stringify({ note: note ?? '' }) }),
  /** الزبون رد — يرجع للطابور، وعدد المحاولات يبقى مسجّل */
  resumeBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/resume`, { method: 'PUT' }),

  /** لوحة اليوم — كل أرقامها بنداء واحد */
  getTodayBoard: () => request<TodayBoardData>('/dashboard/today'),
  /** «الرقم = القائمة» (ع 10-16): الحجوزات الي عدّتها البطاقة بالضبط */
  getDrill: (key: string) => request<DrillResult>(`/drill/${encodeURIComponent(key)}`),

  /** «وين هذا الحجز؟» — يرجّع المحطة الي بيها الحجز فعلاً */
  locateBookings: (q: string) =>
    request<{ id: string; code: string; customerName: string | null; station: string }[]>(
      `/bookings/locate?q=${encodeURIComponent(q)}`),

  /** عدّاد كل محطة — استعلام واحد بالسيرفر */
  getBookingStationCounts: () => request<Record<string, number>>('/bookings/station-counts'),

  // ═══ الحجوزات صفحة صفحة ═══
  //
  // «حتى لا يضل يحمّل السيرفر بتحميل كل الحجوزات — يحمّل جزء جزء».
  //
  // ⚠️ مسار مستقل عن `getBookings`: هذاك يرجّع مصفوفة وتناديه عشرات
  // الشاشات، وتغيير شكل جوابه يكسرهن كلهن. وهنا الفرز والعدّ يصيرون
  // بالسيرفر — الواجهة ما تشوف إلا صفحتها.
  getBookingsPaged: (params: {
    bucket?: string; search?: string; date?: string; month?: string; page?: number; pageSize?: number
  }) => {
    const q = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== null && v !== '') q.set(k, String(v)) })
    return request<{ items: Booking[]; total: number; page: number }>(`/bookings/paged?${q.toString()}`)
  },

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
  // ═══ تتبّع المراحل ═══
  // ═══ الذكاء الاصطناعي ═══
  // ⚠️ كلها للمالك ومدير النظام حصراً — السيرفر يفرضها، مو الواجهة.
  getAiSignals: (kind?: string) =>
    request<AiSignal[]>(`/ai/signals${kind ? `?kind=${encodeURIComponent(kind)}` : ''}`),
  runAiProcess: () => request<{ analyzed: number }>('/ai/process', { method: 'POST' }),
  getAiCatalog: () => request<AiCatalog>('/ai/catalog'),
  getAiMetrics: () => request<AiMetric[]>('/ai/metrics'),
  recomputeAiMetrics: () => request<{ metrics: number }>('/ai/metrics/recompute', { method: 'POST' }),

  /** حكم التدقيق قبل الاعتماد: مطابق / غير مطابق / خطأ بالسعر */
  setInvoiceAuditVerdict: (id: string, data: { verdict: string; note: string; auditedAmount?: number | null }) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/audit`, { method: 'PUT', body: JSON.stringify(data) }),
  /** سحب اعتماد فاتورة انعتمدت بالغلط — السبب إجباري */
  /** المحاسب يرسلها للمراقب حتى يراجعها ويدققها. */
  requestInvoiceMonitorReview: (id: string, note: string) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/monitor-request`, { method: 'PUT', body: JSON.stringify({ note }) }),
  /** المراقب يبتّ: OK أو FLAGGED (والملاحظة إجبارية بالثانية). */
  decideInvoiceMonitorReview: (id: string, verdict: 'OK' | 'FLAGGED', note: string) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/monitor-decide`, { method: 'PUT', body: JSON.stringify({ verdict, note }) }),
  /** ⚠️ المالك وحده: يرجّعها للمحاسب حتى يرتّبها من جديد. */
  returnInvoiceToAccountant: (id: string, reason: string) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/return`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  revokeInvoiceApproval: (id: string, reason: string) =>
    request<LeaderInvoice>(`/leader-invoices/${id}/revoke`, { method: 'PUT', body: JSON.stringify({ reason }) }),
  /** الفواتير المعتمدة بلا رقم فاتورة محاسبية — الفجوة الي كلّفت */

  /** قصة الحجز كاملة + التأخيرات */
  getBookingTimeline: (id: string) => request<BookingTimeline>(`/bookings/${id}/timeline`),
  /** ماتركس ٢٠٥٠: سلسلة الحجز ١٤ محطة — المدير والمراقب */
  getBookingChain: (id: string) => request<BookingChain>(`/ai/chain/${id}`),
  getRoleChainRoles: () => request<{ key: string; title: string }[]>('/ai/role-chain/roles'),
  getRoleChain: (role: string, days: number) => request<RoleChainReport>(`/ai/role-chain?role=${role}&days=${days}`),
  getChainLearning: () => request<ChainLearning>('/ai/chain-learning'),
  getMatrixProjects: () => request<ProjectsReport>('/ai/projects'),
  getMatrixProject: (id: string) => request<ProjectChain>(`/ai/projects/${id}`),
  getAfterInventoryPending: () => request<{ bookingId: string; bookingCode: string; completedAt: string }[]>('/inventory/after-checks/pending'),
  saveAfterInventory: (bookingId: string, complete: boolean, missingItems?: string) =>
    request<{ ok: boolean }>('/inventory/after-checks', { method: 'POST', body: JSON.stringify({ bookingId, complete, missingItems }) }),
  getMatrixSuggestion: (bookingId: string) => request<MatrixBookingSuggestion>(`/ai/suggest/${bookingId}`),
  getMatrixSuggestAccuracy: (days = 30) => request<MatrixSuggestAccuracy>(`/ai/suggest-accuracy?days=${days}`),
  getMatrixAutonomy: () => request<MatrixAutonomyStatus>('/ai/autonomy'),
  getMatrixAutoCrew: () => request<{ actionId: string; bookingId: string; summary: string }[]>('/ai/autonomy/crew'),
  undoMatrixAutoCrew: (actionId: string) => request<{ ok: boolean }>(`/ai/autonomy/crew/${actionId}/undo`, { method: 'POST' }),
  getAIUsage: () => request<AIUsageReport>('/ai/usage'),
  getMatrixEyes: () => request<MatrixEyesReport>('/ai/watch'),
  getQuotationTerms: () => request<{ id: string; text: string; sortOrder: number }[]>('/quotation-terms'),
  saveQuotationTerms: (terms: string[]) =>
    request<{ id: string; text: string; sortOrder: number }[]>('/quotation-terms', { method: 'PUT', body: JSON.stringify({ terms }) }),
  getMediaBriefs: () => request<MediaBrief[]>('/media/briefs'),
  updateMediaBrief: (id: string, body: { status: string; shootAt?: string | null; publishedUrl?: string | null; mediaNotes?: string | null }) =>
    request<MediaBrief>(`/media/briefs/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  getDisciplineRecord: () => request<DisciplineRecord>('/discipline-record'),
  getMyPeer: () => request<MyPeerState>('/peer/me'),
  submitPeer: (body: { mood: number | null; worry: string; suggestion: string; nothing: boolean; notes: { targetId: string; score: number; kind: string; text: string }[] }) =>
    request<MyPeerState>('/peer/me', { method: 'POST', body: JSON.stringify(body) }),
  peerFollowup: (id: string, what: string, why: string, wish: string) =>
    request<PeerNote>(`/peer/notes/${id}/followup`, { method: 'POST', body: JSON.stringify({ What: what, Why: why, Wish: wish }) }),
  peerAccept: (id: string, accepted: boolean, note: string) =>
    request<PeerNote>(`/peer/notes/${id}/accept`, { method: 'POST', body: JSON.stringify({ accepted, note }) }),
  getPeerVoice: () => request<PeerVoiceReport>('/peer/voice'),
  getWorkplaceIssues: () => request<WorkplaceIssue[]>('/workplace-issues'),
  getWorkplaceIssue: (id: string) => request<WorkplaceIssueDetail>(`/workplace-issues/${id}`),
  openWorkplaceIssue: (body: { title: string; partyAId: string; partyBId: string | null; description: string; severity: string; sourceNoteId?: string | null }) =>
    request<{ id: string }>('/workplace-issues', { method: 'POST', body: JSON.stringify(body) }),
  addIssueEntry: (id: string, kind: string, text: string) =>
    request<WorkplaceIssueDetail>(`/workplace-issues/${id}/entries`, { method: 'POST', body: JSON.stringify({ Kind: kind, Text: text }) }),
  analyzeIssue: (id: string) => request<WorkplaceIssueDetail>(`/workplace-issues/${id}/analyze`, { method: 'POST' }),
  decideIssue: (id: string, decision: string, status: string, followUpAt: string | null) =>
    request<WorkplaceIssueDetail>(`/workplace-issues/${id}/decide`, { method: 'POST', body: JSON.stringify({ decision, status, followUpAt }) }),
  issuePartyFollowup: (id: string, better: boolean, text: string) =>
    request<{ ok: boolean }>(`/workplace-issues/${id}/party-followup`, { method: 'POST', body: JSON.stringify({ better, text }) }),
  // ═══ تقييم الموظفين: ماتركس ٦٠٪ + البشر ٤٠٪ ═══
  getStaffScoreBoard: (month = '') => request<StaffScoreBoard>(`/staff-score${month ? `?month=${month}` : ''}`),
  getStaffScore: (id: string, month = '') => request<StaffScoreDetail>(`/staff-score/${id}${month ? `?month=${month}` : ''}`),
  getStaffPoints: (month = '') => request<{ month: string; staff: StaffPointsRow[] }>(`/staff-points${month ? `?month=${month}` : ''}`),
  getStaffScorePeers: () => request<{ id: string; final: number | null }[]>('/staff-score/peers'),
  getMyScore: (month = '') => request<{ score: StaffScoreDetail; on: boolean } | null>(`/staff-score/me${month ? `?month=${month}` : ''}`),
  runStaffScore: () => request<{ added: number }>('/staff-score/run', { method: 'POST' }),
  cancelMatrixScore: (id: string, note: string) => request<{ ok: boolean }>(`/matrix-score/${id}/cancel`, { method: 'POST', body: JSON.stringify({ note }) }),
  getCoordPendingRatings: () => request<PendingStaffRating[]>('/staff-ratings/coord-pending'),
  getAuditPendingRatings: () => request<PendingStaffRating[]>('/staff-ratings/audit-pending'),
  getBookingRatingState: (bookingId: string, stage: StaffRatingStage) => request<{ parties: PendingStaffRating[]; mine: Record<string, number> }>(`/staff-ratings/booking/${bookingId}?stage=${stage}`),
  getPeriodicRatings: () => request<{ period: string; people: { id: string; name: string; role: string }[]; mine: Record<string, number> }>('/staff-ratings/periodic'),
  rateStaff: (stage: 'coord' | 'audit' | 'quality' | 'periodic', body: { bookingId?: string; rateeId: string; score: number; note?: string }) =>
    request<{ ok: boolean }>(`/staff-ratings/${stage}`, { method: 'POST', body: JSON.stringify(body) }),
  getPendingCrewRatings: () => request<PendingCrewRating[]>('/crew-ratings/pending'),
  rateCrew: (bookingId: string, ratings: { technicianId: string; score: number; note?: string }[]) =>
    request<{ ok: boolean }>('/crew-ratings', { method: 'POST', body: JSON.stringify({ bookingId, ratings }) }),

  // ═══ المهام الإضافية ═══
  /** المدير يوجّه مهمة لموظف */
  createExtraTask: (data: {
    title: string; description?: string; assignedToId: string
    priority: 'NORMAL' | 'URGENT'; dueAt?: string | null
  }) => request<ExtraTask>('/extra-tasks', { method: 'POST', body: JSON.stringify(data) }),
  // ═══ 🗂️ السكرتير (docs/secretary.md) — المالك والمدير بس ═══
  secretaryToday: () => request<SecretaryToday>('/secretary/today'),
  secretaryConversations: () => request<SecretaryConversation[]>('/secretary/conversations'),
  secretaryNewConversation: () => request<SecretaryConversation>('/secretary/conversations', { method: 'POST', body: JSON.stringify({}) }),
  secretaryRenameConversation: (id: string, title: string) => request<{ ok: boolean }>(`/secretary/conversations/${encodeURIComponent(id)}`, { method: 'PATCH', body: JSON.stringify({ title }) }),
  secretaryDeleteConversation: (id: string) => request<{ ok: boolean }>(`/secretary/conversations/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  secretaryHistory: (limit = 50, conversationId?: string) => request<SecretaryChat[]>(`/secretary/history?limit=${limit}${conversationId ? `&conversationId=${encodeURIComponent(conversationId)}` : ''}`),
  secretaryMessage: (text: string, conversationId?: string) => request<SecretaryReply>('/secretary/message', { method: 'POST', body: JSON.stringify({ text, conversationId }) }),
  secretaryVoice: async (audio: Blob, conversationId?: string) => {
    const form = new FormData()
    form.append('file', audio, 'voice')
    if (conversationId) form.append('conversationId', conversationId)
    const token = currentToken()
    const res = await fetch(`${API_URL}/secretary/message`, { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'تعذر إرسال البصمة')
    return res.json() as Promise<SecretaryReply>
  },
  secretaryConfirm: (id: string, fields: SecretaryProposal['fields']) =>
    request<{ ok: boolean; card: SecretaryCard }>(`/secretary/actions/${encodeURIComponent(id)}/confirm`, { method: 'POST', body: JSON.stringify({ fields }) }),
  secretaryCancel: (id: string) => request<{ ok: boolean }>(`/secretary/actions/${encodeURIComponent(id)}/cancel`, { method: 'POST' }),
  /** مهمة بالفويس: مسودة بس (عنوان، تفاصيل، أولوية، موعد، موظف مقترح) — التوجيه يبقى بـcreateExtraTask */
  voiceTaskDraft: async (audio: Blob) => {
    const form = new FormData()
    form.append('file', audio, 'voice')
    const token = currentToken()
    const res = await fetch(`${API_URL}/extra-tasks/voice-draft`, { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'تعذر تحويل الفويس')
    return res.json() as Promise<{
      transcript: string; title: string; description: string; priority: 'NORMAL' | 'URGENT'
      dueAt: string; suggestedEmployeeIds: string[]; understood: boolean
    }>
  },
  /** كل المهام — للمدير */
  getExtraTasks: (params: { status?: string; assigneeId?: string } = {}) => {
    const q = new URLSearchParams()
    if (params.status) q.set('status', params.status)
    if (params.assigneeId) q.set('assigneeId', params.assigneeId)
    const qs = q.toString()
    return request<ExtraTask[]>(`/extra-tasks${qs ? `?${qs}` : ''}`)
  },
  /** مهامي أنا — المعرّف من التوكن، ما ينرسل بالطلب */
  getMyExtraTasks: (includeDone = false) =>
    request<ExtraTask[]>(`/extra-tasks/mine${includeDone ? '?includeDone=1' : ''}`),
  markExtraTaskSeen: (id: string) =>
    request<{ ok: boolean }>(`/extra-tasks/${id}/seen`, { method: 'PUT' }),
  startExtraTask: (id: string) =>
    request<ExtraTask>(`/extra-tasks/${id}/start`, { method: 'PUT' }),
  completeExtraTask: (id: string, doneNote: string) =>
    request<ExtraTask>(`/extra-tasks/${id}/complete`, { method: 'PUT', body: JSON.stringify({ doneNote }) }),
  cancelExtraTask: (id: string, reason: string) =>
    request<ExtraTask>(`/extra-tasks/${id}/cancel`, { method: 'PUT', body: JSON.stringify({ reason }) }),

  /** ملاحظة الإداري للكادر المنفّذ — يقراها الفريق بشاشة «مهامي». */
  setBookingCrewNotes: (id: string, note: string) =>
    request<Booking>(`/bookings/${id}/crew-notes`, { method: 'PUT', body: JSON.stringify({ note }) }),
  /** ملاحظة الإداري لمدير المشاريع. */
  setBookingProjectNotes: (id: string, note: string) =>
    request<Booking>(`/bookings/${id}/project-notes`, { method: 'PUT', body: JSON.stringify({ note }) }),
  /** إلغاء الحجز — السبب إجباري. */
  /** حجوزات سلّة مرحلة وحدة (مؤجّل/ما رد/ملغى × قبل وبعد التثبيت). */
  getBookingsByStageBucket: (bucket: StageBucket) =>
    request<Booking[]>(`/bookings/stage-bucket?bucket=${encodeURIComponent(bucket)}`),
  /** أعداد السلال الست — للأرقام فوق التبويبات. */
  getStageBucketCounts: () => request<Record<StageBucket, number>>('/bookings/stage-bucket-counts'),

  // bookingType: 'INTERNAL' يخلي الحجز شغل داخل الشركة (للإداري فما فوق)
  createBooking: (data: {
    customerId: string
    serviceId?: string
    // خدمات متعددة بنفس الحجز — الأولى تنعتبر الرئيسية
    serviceIds?: string[]
    locationUrl?: string
    notes?: string
    vehicleType?: string
    /** عدد الأجهزة — إجباري للخدمات المؤشّرة requiresDeviceInfo */
    deviceCount?: number
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
    /** القسم من السجل — الاسم النصي فوق يبقى منسوخاً للتقارير القديمة. */
    internalDepartmentId?: string
    /** مسؤول القسم الطالب — اسمه وهاتفه ينتنسخان بالسيرفر من السجل. */
    internalHeadId?: string
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
    /** `shift: ''` يمسح الفترة ويرجّعها «غير محدَّدة»؛ وعدم إرسالها يخليها مثل ما هي. */
    data: { quotedPrice?: number | null; address?: string; assignedVehicle?: string; mapLocation?: string; mapLatitude?: number | null; mapLongitude?: number | null; expenseResponsibleId?: string | null; serviceIds?: string[]; internalHrNote?: string; shift?: 'MORNING' | 'EVENING' | '' },
  ) => request<Booking>(`/bookings/${id}/details`, { method: 'PUT', body: JSON.stringify(data) }),
  /** إلغاء تكليف موظف من خانة كادر — الحجز يبقى مثبّت بلا كادر. */
  unassignTechnician: (id: string, role: string) =>
    request<Booking>(`/bookings/${id}/assign?role=${encodeURIComponent(role)}`, { method: 'DELETE' }),
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
  // ── تقرير زيارة معاينة («كشف») ──
  submitSurveyReport: (bookingId: string, body: SurveyReportBody) =>
    request<BookingSurveyReport>('/booking-survey-reports', {
      method: 'POST', body: JSON.stringify({ bookingId, ...body }),
    }),

  /** تغيير نوع الحجز — للمالك ومدير النظام بس (السيرفر يرفض غيرهم) */
  /** تسعير الأعمال الداخلية القديمة (مؤقت — قرار (ع) 10-07). */
  getInternalBackfill: () => request<{ id: string; code: string; department: string | null; notes: string | null }[]>('/internal-backfill'),
  priceInternalBackfill: (id: string, work: string, price: number) =>
    request<unknown>(`/internal-backfill/${id}`, { method: 'POST', body: JSON.stringify({ work, price }) }),
  /** طابور الإداري: أعمال داخلية خلصت وناقصها فاتورة أو تقرير. */
  getInternalPaperwork: () => request<{ id: string; code: string; department: string | null; hasInvoice: boolean; hasReport: boolean; doneAt: string | null }[]>('/bookings/internal-paperwork'),
  changeBookingType: (id: string, bookingType: 'REGULAR' | 'MAINTENANCE' | 'INTERNAL' | 'SOLAR' | 'SURVEY' | 'COMPLAINT') =>
    request<Booking>(`/bookings/${id}/type`, { method: 'PUT', body: JSON.stringify({ bookingType }) }),

  completeBooking: (
    id: string,
    data: { completionNotes?: string; amountCollected?: number; advancePaid?: number },
  ) => request<Booking>(`/bookings/${id}/complete`, { method: 'PUT', body: JSON.stringify(data) }),
  startBooking: (id: string, missingToolIds?: string[]) =>
    request<Booking>(`/bookings/${id}/start`, { method: 'PUT', body: JSON.stringify({ missingToolIds: missingToolIds || [] }) }),
  // (ع) 10-14: وقت المرحلة الحقيقي + موقعها — يشتغل حتى لو انضغطت بلا نت
  stageProof: (id: string, data: { stage: 'ACCEPTED' | 'MATERIALS' | 'DEPARTED' | 'ARRIVED' | 'STARTED' | 'COMPLETED'; pressedAt: string; offline: boolean; lat?: number; lng?: number; accuracy?: number }) =>
    request<{ ok: boolean; verdict: 'OK' | 'FAR' | 'BATCH'; distanceM: number | null }>(`/bookings/${id}/stage-proof`, { method: 'POST', body: JSON.stringify(data) }),
  // (دازاي 10-25) صور «قبل/بعد» لشغلة الحجز — كادر الديكور يصوّر بدل ما يكتب
  // (دازاي 10-25) طلب حذف فاتورة — المحاسب يطلب، والمالك يحسم
  requestInvoiceDelete: (id: string, reason: string) =>
    request<{ id: string }>(`/leader-invoices/${id}/delete-request`, { method: 'POST', body: JSON.stringify({ reason }) }),
  getInvoiceDeleteRequests: (status = 'PENDING') => request<InvoiceDeleteRequest[]>(`/leader-invoices/delete-requests?status=${status}`),
  decideInvoiceDelete: (id: string, approve: boolean, note = '') =>
    request<{ ok: boolean }>(`/leader-invoices/delete-requests/${id}/decide`, { method: 'POST', body: JSON.stringify({ approve, note }) }),
  // (دازاي 10-26) ورق الحجوزات القديمة لليدر السابق
  getCarryover: () => request<CarryoverItem[]>('/paperwork/carryover'),
  getCarryoverBookings: () => request<Booking[]>('/paperwork/carryover/bookings'),
  // (دازاي 10-28) حجوزاتي المرحّلة لإدارة التقنيين الي بلا فاتورة
  getHandoverBookings: () => request<Booking[]>('/paperwork/handover-bookings'),
  // (دازاي 10-26) «الصلاحيات والوصول»
  getPermissionCatalog: () => request<AccessGroup[]>('/permissions/catalog'),
  getEmployeeAccess: (id: string) => request<EmployeeAccess>(`/employees/${id}/access`),
  saveEmployeeAccess: (id: string, body: { add: string[]; remove: string[]; reason: string; version: number }) =>
    request<EmployeeAccess>(`/employees/${id}/access`, { method: 'PUT', body: JSON.stringify(body) }),
  getRoleDrift: () => request<RoleDriftRow[]>('/permissions/role-drift'),
  alignRole: (id: string, reason: string, version: number) =>
    request<RoleDriftRow>(`/permissions/role-drift/${id}/align`, { method: 'POST', body: JSON.stringify({ reason, version }) }),
  addPermissionException: (body: { employeeId: string; permission: string; reason: string; expiresAt?: string; version: number }) =>
    request<{ ok: boolean }>('/permissions/exceptions', { method: 'POST', body: JSON.stringify(body) }),
  getPermissionAuditLog: (employeeId?: string) =>
    request<PermissionAuditRow[]>(`/permissions/audit-log${employeeId ? `?employeeId=${encodeURIComponent(employeeId)}` : ''}`),
  getWorkPhotos: (id: string) => request<WorkPhoto[]>(`/bookings/${id}/work-photos`),
  uploadWorkPhoto: async (id: string, phase: 'BEFORE' | 'AFTER', file: Blob) => {
    const form = new FormData()
    form.append('phase', phase)
    form.append('file', file, 'photo')
    const token = currentToken()
    const res = await fetch(`${API_URL}/bookings/${id}/work-photos`, { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'تعذر رفع الصورة')
    return (await res.json()) as WorkPhoto
  },
  markArrived: (id: string) =>
    request<Booking>(`/bookings/${id}/arrived`, { method: 'PUT', body: JSON.stringify({}) }),
  setMaterialsReady: (id: string) =>
    request<Booking>(`/bookings/${id}/materials-ready`, { method: 'PUT', body: JSON.stringify({}) }),
  // (ع) 10-14: استلام المهمة · الانطلاق (منفصل عن الوصول) · نقاط المسار · خريطة الميدان
  acceptBooking: (id: string) =>
    request<{ ok: boolean }>(`/bookings/${id}/accept`, { method: 'PUT', body: JSON.stringify({}) }),
  departBooking: (id: string, pos?: { lat?: number; lng?: number }) =>
    request<{ ok: boolean; origin: DepartOrigin }>(`/bookings/${id}/depart`, { method: 'PUT', body: JSON.stringify(pos ?? {}) }),
  pingBatch: (points: { lat: number; lng: number; accuracy?: number; at: string; bookingId?: string }[]) =>
    request<{ saved: number }>('/location-pings/batch', { method: 'POST', body: JSON.stringify({ points }) }),
  getCompanyLocation: () => request<{ company: CompanyLocation | null }>('/company-location'),
  setCompanyLocation: (c: CompanyLocation) =>
    request<{ company: CompanyLocation }>('/company-location', { method: 'PUT', body: JSON.stringify(c) }),
  getBookingTrack: (id: string) => request<BookingTrack>(`/bookings/${id}/track`),
  lateAudit: (from: string, to: string) => request<{ total: number; rows: unknown[]; causes: Record<string, string> }>(`/late-audit?from=${from}&to=${to}`),
  lateAuditExcel: async (from: string, to: string) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/late-audit?from=${from}&to=${to}&format=xlsx`, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (!res.ok) throw new Error('تعذر تنزيل الإكسل')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(await res.blob())
    a.download = `المتأخرات-${from}-${to}.xlsx`
    a.click()
  },
  matrixSelfCheck: () => request<{ checks: unknown[]; at: string }>('/matrix-selfcheck'),
  getFieldMap: (from: string, to: string) => request<FieldMap>(`/field-map?from=${from}&to=${to}`),
  // إرجاع الحجز للتدقيق بعد ما انتأشر مدقّق — مدير النظام حصراً
  unverifyBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/unverify`, { method: 'PUT' }),
  markConfirmationContacted: (id: string) =>
    request<Booking>(`/bookings/${id}/confirmation-contacted`, { method: 'PUT', body: JSON.stringify({}) }),
  getPendingAudit: () => request<Booking[]>('/bookings/pending-audit'),
  addCoordinationAlert: (bookingId: string, reason: string) =>
    request<{ openCount: number }>(`/bookings/${bookingId}/coordination-alerts`, { method: 'POST', body: JSON.stringify({ reason }) }),
  resolveCoordinationAlerts: (bookingId: string, note: string) =>
    request<{ ok: boolean }>(`/bookings/${bookingId}/coordination-alerts/resolve`, { method: 'PUT', body: JSON.stringify({ note }) }),
  getCoordinationAlertsForBooking: (bookingId: string) =>
    request<CoordinationAlert[]>(`/bookings/${bookingId}/coordination-alerts`),
  getCoordinationAlertSummaries: () =>
    request<CoordinationAlertSummary[]>('/coordination-alerts/summaries'),
  getStats: () => request<Stats>('/stats'),

  getExpenses: (employeeId?: string) =>
    request<Expense[]>(`/expenses${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createExpense: (data: { employeeId: string; amount: number; description?: string; bookingId?: string; purchase?: { techSupplierId?: string; productName: string; quantity: number; unitPrice: number } }) =>
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
        'role' | 'onDuty' | 'status' | 'name' | 'position' | 'phone' | 'certificate' | 'photoUrl' | 'hasDrivingLicense' | 'hasSafetyCertificate' | 'isTrainee'
        | 'isLeader' | 'salary' | 'shift' | 'shiftStart' | 'shiftEnd' | 'monthlyLeaves' | 'jobTitle'
        // ملف الموارد البشرية — careerStatus ما تنرسل: تنحسب بالسيرفر
        | 'department' | 'hireDate' | 'experienceYears' | 'lastReview' | 'jobLevel' | 'nextRole' | 'trainingNeeds'
        | 'secondaryRoles' | 'language'
      >
    > & {
      username?: string
      password?: string
    },
  ) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  /** «تعديل اسم الموظف ودوره» — لصاحب الصلاحية بدون ما يكون مدير. */
  updateEmployeeIdentity: (id: string, data: { name?: string; role?: Employee['role'] }) =>
    request<Employee>(`/employees/${id}/identity`, { method: 'PUT', body: JSON.stringify(data) }),
  /** حقول الملف (راتب، دوام، إجازات…) لصاحب «تعديل ملف الموظف». الخادم
   *  يشيل الاسم والدور وبيانات الدخول لو انرسلت. */
  updateEmployeeProfile: (id: string, data: Record<string, unknown>) =>
    request<Employee>(`/employees/${id}/profile`, { method: 'PUT', body: JSON.stringify(data) }),
  linkHistoricalRecords: (id: string) =>
    request<{ bookingsLinked: number; complaintsLinked: number }>(`/employees/${id}/link-historical`, { method: 'POST' }),

  // GPS
  /**
   * نتائج الجي بي اس للمراقب: قربت تنتهي · انتهت وما انجدّدت ·
   * مشاكل مفتوحة. **قراءة فقط** — التنفيذ بشاشات مسؤول الجي بي اس.
   */
  getGpsMonitorSnapshot: (windowDays?: number) =>
    request<GpsMonitorSnapshot>(`/gps/monitor-snapshot${windowDays ? `?windowDays=${windowDays}` : ''}`),
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
  /** أرقام اللوحة الرئيسية — بدون سحب أرشيف الشركة كامل للمتصفح */
  /** آخر ظهور بآخر ساعة — «منو فاتح النظام هسه» (المدير/المالك والمراقب). */
  getOnlineEmployees: () => request<{ id: string; name: string; role: string; photoUrl: string | null; lastSeenAt: string }[]>('/presence/online'),
  getDashboardSummary: () => request<{ employeeCount: number; customerCount: number; bookingCount: number; gpsDeviceCount: number }>('/dashboard/summary'),
  /** أرقام يوم الإداري — كلها محسوبة بقاعدة البيانات. */
  getTodayPulse: () => request<TodayPulse>('/dashboard/today-pulse'),
  /** مجاميع المحاسب والمراقب — تنحسب بقاعدة البيانات بدل تنزيل كل الأرشيف المنجز */
  getFinanceSummary: () => request<FinanceSummary>('/dashboard/finance-summary'),

  // ── الإجازات ──
  createLeave: (data: { startDate: string; endDate?: string; reason?: string | null; kind?: LeaveKind }) =>
    request<LeaveRequest>('/leaves', { method: 'POST', body: JSON.stringify(data) }),
  getMyLeaves: () => request<LeaveRequest[]>('/leaves/mine'),
  cancelLeave: (id: string) => request<{ ok: boolean }>(`/leaves/${id}`, { method: 'DELETE' }),
  getLeaveInbox: (status?: string) => request<LeaveRequest[]>(`/leaves/inbox${status ? '?status=' + status : ''}`),
  /** 🗓️ إجازات الموظفين (وحدة إدارة الكوادر) — المجازين اليوم وكل طلبات الشهر */
  getLeaveRegister: (month: string) => request<{ month: string; onLeaveToday: LeaveRequest[]; items: LeaveRequest[] }>(`/leaves/register?month=${month}`),
  /** ↩️ الحجوزات الراجعة من إدارة المشاريع — تنتظر استلام المنسّق */
  getReturnedFromProjects: () => request<ReturnedFromProjects[]>('/bookings/returned-from-projects'),
  getBookingHistory: (id: string) => request<BookingHistoryLine[]>(`/bookings/${encodeURIComponent(id)}/history`),
  markReturnHandled: (id: string) => request<{ ok: boolean }>(`/bookings/${encodeURIComponent(id)}/return-handled`, { method: 'PUT' }),
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
  getVehicleTracking: (period: string) => request<VehicleTracking>(`/vehicles/tracking?period=${period}`),
  mergeTrackingVehicle: (fromId: string, toId: string) =>
    request<{ ok: boolean }>('/vehicles/tracking/merge', { method: 'POST', body: JSON.stringify({ fromId, toId }) }),
  exportVehicleTracking: (period: string) => downloadFile(`/vehicles/tracking/export?period=${period}`, `متابعة-السيارات-${period}.xlsx`),
  getCarWash: (date?: string) => request<{ date: string; vehicles: CarWashVehicle[] }>(`/car-wash${date ? `?date=${date}` : ''}`),
  markCarWashed: (vehicleId: string) => request<{ ok: boolean }>(`/car-wash/${vehicleId}`, { method: 'POST' }),
  unmarkCarWashed: (vehicleId: string) => request<{ ok: boolean }>(`/car-wash/${vehicleId}`, { method: 'DELETE' }),
  saveVehicleTrackingDay: (date: string, rows: Record<string, unknown>[]) =>
    request<{ ok: boolean }>('/vehicles/tracking/day', { method: 'PUT', body: JSON.stringify({ date, rows }) }),
  getFundTxnReceipt: (id: string) => request<{ receiptImage: string }>(`/funds/transactions/${id}/receipt`),
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

  /** المحاسب يأشر «تم التخريج» ويحدد الحساب — هنا يرجع المصروف للدوار */
  dischargeSettlement: (id: string, data: { account: string; note?: string | null }) =>
    request<RevolvingFundTxn>(`/funds/settlements/${id}/discharge`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  getDischargeAccounts: () => request<string[]>('/funds/discharge-accounts'),
  submitFundSettlement: (data: { fundId: string; spentAmount: number; returnedAmount: number; receiptImage?: string | null; bookingId?: string | null; notes?: string | null }) =>
    request<RevolvingFundTxn>('/funds/settlements', { method: 'POST', body: JSON.stringify(data) }),
  /** (ع) 10-16: صاحب صلاحية الدوار يسوّي بدل موظف نسى — تروح للتدقيق مثل أي تسوية */
  submitFundSettlementFor: (employeeId: string, data: { fundId: string; spentAmount: number; returnedAmount: number; receiptImage?: string | null; bookingId?: string | null; notes?: string | null }) =>
    request<RevolvingFundTxn>(`/funds/settlements/for/${employeeId}`, { method: 'POST', body: JSON.stringify(data) }),

  getSimCards: () => request<GpsSimCard[]>('/gps/sims'),
  createSimCard: (data: Partial<GpsSimCard>) =>
    request<GpsSimCard>('/gps/sims', { method: 'POST', body: JSON.stringify(data) }),

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
  createPerformanceReview: (data: {
    employeeId: string; rating: ReviewRating; reason: string; bookingId?: string | null
    commitmentScore?: number | null; speedScore?: number | null; qualityScore?: number | null
  }) =>
    request<PerformanceReview>('/performance-reviews', { method: 'POST', body: JSON.stringify(data) }),
  /** حجوزات الليدر المنجزة وكادر كل وحدة وحالة تقييمهم. */
  // النطاق يمرّ للخادم مو ينترشح محلياً: الخادم چان يقصّ على آخر ٣٠
  // يوم قبل ما توصل البيانات للواجهة، فاختيار تاريخ أقدم ما يجيب ولا شي.
  getMyBookingsForReview: (from?: string, to?: string) => {
    const q = new URLSearchParams()
    if (from) q.set('from', from)
    if (to) q.set('to', to)
    const qs = q.toString()
    return request<BookingAwaitingReview[]>(
      `/performance-reviews/my-bookings${qs ? `?${qs}` : ''}`)
  },

  // Products
  getProducts: () => request<Product[]>('/products'),
  // ── الملفات المخزّنة برّا قاعدة البيانات ──
  //
  // الصور القديمة لسّه data: URLs والجديدة صارت /api/files/... —
  // fileUrl() يتعامل مع الاثنين، فما نحتاج نعدّل كل شاشة.
  // (ع) 10-14: صورة عرض/عمل بتطبيق الزبائن
  uploadCustomerPromoImage: async (id: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    const token = currentToken()
    const res = await fetch(`${API_URL}/customer-app/promos/${id}/image`, { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'تعذر رفع الصورة')
  },
  customerPromoImageUrl: (id: string) => `${API_URL}/c/img/${id}`,
  customerServiceImageUrl: (id: string) => `${API_URL}/c/img/s/${id}`,
  uploadCustomerServiceImage: async (id: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    const token = currentToken()
    const res = await fetch(`${API_URL}/customer-app/services/${id}/image`, { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'تعذر رفع الصورة')
  },
  // ═══ بوت بلاغات الموظفين — تلغرام (دازاي 11-05) ═══
  getMyTelegram: () => request<{ enabled: boolean; linked: boolean; tgUsername?: string | null; linkedAt?: string }>('/me/telegram'),
  createTelegramLink: () => request<{ url: string; expiresInMinutes: number }>('/me/telegram/link', { method: 'POST' }),
  // الربط مرة وحدة — الفك بس من المالك
  getTelegramLinks: () => request<{ employeeId: string; name: string; role: string; status: string; tgUsername: string | null; linkedAt: string }[]>('/telegram-links'),
  ownerUnlinkTelegram: (employeeId: string) => request<{ ok: boolean }>(`/telegram-links/${employeeId}`, { method: 'DELETE' }),
  // البلاغ من داخل النظام
  submitBotReport: (body: { body: string; photos: string[]; page: string }) => request<BotReport>('/me/bot-reports', { method: 'POST', body: JSON.stringify(body) }),
  getMyBotReports: () => request<{ items: BotReport[]; statuses: Record<string, string> }>('/me/bot-reports'),
  getBotReports: (status = '') =>
    request<{ items: BotReport[]; statuses: Record<string, string> }>(`/bot-reports${status ? `?status=${encodeURIComponent(status)}` : ''}`),
  updateBotReport: (id: string, body: { status?: string; reply?: string }) =>
    request<BotReport>(`/bot-reports/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
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

  // ═══ وحدة تقنية المعلومات ═══
  getItAssets: () => request<ItAsset[]>('/it/assets'),
  createItAsset: (data: ItAssetInput) => request<ItAsset>('/it/assets', { method: 'POST', body: JSON.stringify(data) }),
  updateItAsset: (id: string, data: ItAssetInput) => request<ItAsset>(`/it/assets/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteItAsset: async (id: string) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/it/assets/${id}`, { method: 'DELETE', headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || describeHttpStatus(res.status))
  },
  getItAssetLogs: (id: string) => request<ItAssetLog[]>(`/it/assets/${id}/logs`),
  addItAssetLog: (id: string, data: { kind: string; note: string; cost?: number | null }) =>
    request<ItAssetLog>(`/it/assets/${id}/logs`, { method: 'POST', body: JSON.stringify(data) }),
  getItStats: () => request<ItStats>('/it/stats'),

  // ═══ صور الكشف (حجز أو مشروع) ═══
  // ⚠️ الصورة تنجلب بترويسة Authorization (blob) مو برابط <img> عام:
  // الخادم يفحص الدور بكل طلب، فرابط مسرّب ما يفتح لغير المصرّح.
  getSurveyPhotos: (owner: SurveyPhotoOwner) =>
    request<SurveyPhoto[]>(`/survey-photos?${owner.bookingId ? `bookingId=${owner.bookingId}` : `projectId=${owner.projectId}`}`),
  uploadSurveyPhoto: async (owner: SurveyPhotoOwner, file: File) => {
    const form = new FormData()
    form.append('file', file)
    if (owner.bookingId) form.append('bookingId', owner.bookingId)
    if (owner.projectId) form.append('projectId', owner.projectId)
    const token = currentToken()
    const res = await fetch(`${API_URL}/survey-photos`, {
      method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form,
    })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || describeHttpStatus(res.status))
    return res.json() as Promise<SurveyPhoto>
  },
  surveyPhotoBlob: async (id: string, download = false) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/survey-photos/${id}/file${download ? '?download=1' : ''}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || describeHttpStatus(res.status))
    return res.blob()
  },
  deleteSurveyPhoto: async (id: string) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/survey-photos/${id}`, {
      method: 'DELETE', headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || describeHttpStatus(res.status))
  },
  /** «كشف»: الإداري يطلّع ليدر يعاين الموقع — الحجز يصير كشف وينكلّف الليدر. */
  sendBookingToSurvey: (id: string, leaderId: string) =>
    request<Booking>(`/bookings/${id}/send-to-survey`, { method: 'PUT', body: JSON.stringify({ leaderId }) }),
  /** ترحيل الحجز لتقني أو مسؤول خدمة — الحجز يصير برقبته. */
  getTechHandovers: () => request<Booking[]>('/bookings?handover=all'),
  getHandoverCandidates: () => request<{ id: string; name: string; position: string }[]>('/bookings/handover-candidates'),
  handoverBooking: (id: string, employeeId: string, reason: string) =>
    request<Booking>(`/bookings/${id}/handover`, { method: 'PUT', body: JSON.stringify({ employeeId, reason }) }),
  /** خطوات التقني على الحجز المرحّل إله. */
  techStep: (id: string, step: 'contacted' | 'diagnosis' | 'crew' | 'resolve' | 'visited' | 'no-answer' | 'await-customer' | 'postpone' | 'resume', text = '', visitAt = '') =>
    request<Booking>(`/bookings/${id}/tech/${step}`, { method: 'PUT', body: JSON.stringify({ text, visitAt }) }),
  /** قرار التقني بعد التواصل: انحلّت بالتلفون، لو زيارة بموعد (بتوقيت بغداد «YYYY-MM-DDTHH:mm»). */
  techDecide: (id: string, decision: 'PHONE' | 'VISIT', text: string, visitAt = '') =>
    request<Booking>(`/bookings/${id}/tech/decide`, { method: 'PUT', body: JSON.stringify({ decision, text, visitAt }) }),

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
  /** ═══ تعلّم أسعار الخدمات ═══ */
  getServicePriceSamples: () => request<ServicePriceStats[]>('/service-prices/samples'),
  getServicePriceSuggestions: (status?: string) =>
    request<ServicePriceSuggestion[]>(`/service-prices/suggestions${status ? `?status=${status}` : ''}`),
  /** المحاسب يراجع الاقتراح — خطوة إلزامية قبل المالك */
  reviewServicePrice: (id: string, note?: string) =>
    request<{ success: boolean }>(`/service-prices/suggestions/${id}/review`, {
      method: 'PUT', body: JSON.stringify({ note: note ?? '' }),
    }),
  /** المالك يعتمد أو يرفض */
  decideServicePrice: (id: string, approve: boolean) =>
    request<{ success: boolean }>(`/service-prices/suggestions/${id}/decide`, {
      method: 'PUT', body: JSON.stringify({ approve }),
    }),

  getQuotation: (id: string) => request<Quotation>(`/quotations/${id}`),
  // القائمة المقسّمة الخفيفة — بلا بنود ولا صور (تنزل بس من ينفتح العرض).
  getQuotationsPage: (page: number, search = '') =>
    request<{ items: Quotation[]; total: number; page: number; pages: number }>(`/quotations?page=${page}&limit=50${search ? `&search=${encodeURIComponent(search)}` : ''}`),
  /** النسخ القديمة المؤرشفة — الأحدث أولاً */
  getQuotationVersions: (id: string) =>
    request<QuotationVersion[]>(`/quotations/${id}/versions`),
  getQuotationByProject: (projectId: string) => request<Quotation | null>(`/quotation-by-project/${projectId}`),
  createQuotation: (data: {
    projectId?: string
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
    terms?: string[]
    createdByEmployeeId?: string
    validUntil?: string
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
    terms?: string[]
    status?: Quotation['status']
    validUntil?: string
  }) => request<Quotation>(`/quotations/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteQuotation: (id: string) =>
    request<void>(`/quotations/${id}`, { method: 'DELETE' }),

  // Leader invoices (تحل محل شيت جوجل الليدر)
  /** أسباب الشغل المجاني — فاتورة بصفر لازم تنربط بسبب من القائمة */
  getFreeWorkReasons: () => request<FreeWorkReason[]>('/free-work-reasons'),

  getSystemPriceCatalog: (systemName?: string) =>
    request<SystemPriceCatalog[]>(`/system-price-catalog${systemName ? `?systemName=${encodeURIComponent(systemName)}` : ''}`),
  getInvoiceAdjustments: (id: string) =>
    request<LeaderInvoiceAdjustment[]>(`/leader-invoices/${id}/adjustments`),
  /** يدوّر على الحجز بكل الأماكن ويقول وين هو. */
  locate: (q: string) => request<LocateResult>(`/locate?q=${encodeURIComponent(q)}`),
  /** بند ٧م: صفحة وحدة من فواتير الليدرية ويا العدّادات والمجموع من الخادم. */
  getLeaderInvoicesPage: (q: { tab: string; q?: string; leader?: string; customer?: string; system?: string; month?: string; page: number; perPage: number }) =>
    request<LeaderInvoicePage>(`/leader-invoices/page?${new URLSearchParams(Object.entries(q).filter(([, v]) => v !== '' && v !== undefined).map(([k, v]) => [k, String(v)])).toString()}`),
  getLeaderInvoices: (employeeId?: string) =>
    request<LeaderInvoice[]>(`/leader-invoices${employeeId ? `?employeeId=${encodeURIComponent(employeeId)}` : ''}`),
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
  // المشاريع الموجّهة للموظف الحالي — الليدر يسوي فاتورة للشغل الموجّه له
  getProjectsDirectedToMe: () =>
    request<{ projects: DirectedProject[] }>('/projects/delegated-to-me'),
  // ── صندوق المراقب ──
  getMonitorReviews: (params: { stage?: string; status?: string; ownerRole?: string; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (params.stage) q.set('stage', params.stage)
    if (params.status) q.set('status', params.status)
    if (params.ownerRole) q.set('ownerRole', params.ownerRole)
    if (params.limit) q.set('limit', String(params.limit))
    const qs = q.toString()
    return request<MonitorReview[]>(`/monitor-reviews${qs ? `?${qs}` : ''}`)
  },
  getMonitorReviewCounts: () => request<{ stage: MonitorStage; count: number }[]>('/monitor-reviews/counts'),
  getMonitorReviewCards: () => request<{ invoices: number; bookings: number }>('/monitor-reviews/cards'),
  getMonitorDeskCounts: () => request<MonitorDeskCounts>('/monitor-desk/counts'),
  // «اتخاذ إجراء» على صف بصندوق المراقب
  monitorReviewAction: (id: string, action: 'NOTIFY' | 'JUSTIFY' | 'FOLLOW_UP' | 'ESCALATE', note: string) =>
    request<{ done: string }>(`/monitor-reviews/${encodeURIComponent(id)}/action`, { method: 'POST', body: JSON.stringify({ action, note }) }),
  decideMonitorReview: (id: string, data: { flag: boolean; note: string }) =>
    request<MonitorReview>(`/monitor-reviews/${id}/decide`, { method: 'POST', body: JSON.stringify(data) }),

  // ── تكلفة الشبكات ──
  getNetworkCostItems: () => request<NetworkPriceItem[]>('/network-cost/items'),
  calculateNetworkCost: (data: { lines: { itemId: string; quantity: number }[]; discount: number }) =>
    request<NetworkCostResponse>('/network-cost/calculate', { method: 'POST', body: JSON.stringify(data) }),
  getNetworkPrices: () => request<NetworkPriceItem[]>('/network-cost/prices'),
  createNetworkPrice: (data: Partial<NetworkPriceItem>) =>
    request<NetworkPriceItem>('/network-cost/prices', { method: 'POST', body: JSON.stringify(data) }),
  updateNetworkPrice: (id: string, data: Partial<NetworkPriceItem>) =>
    request<NetworkPriceItem>(`/network-cost/prices/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deactivateNetworkPrice: (id: string) =>
    request<{ ok: boolean }>(`/network-cost/prices/${id}`, { method: 'DELETE' }),
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
  requestBookingDelete: (bookingId: string, reason: string, channel: BookingDeleteChannel, requestType: BookingDeleteRequestType) =>
    request<BookingDeleteRequest>(`/bookings/${bookingId}/delete-request`, { method: 'POST', body: JSON.stringify({ reason, channel, requestType }) }),
  getBookingDeleteRequests: (status?: string) =>
    request<BookingDeleteRequest[]>(`/booking-delete-requests${status ? `?status=${status}` : ''}`),
  getBookingDeleteRequestCounts: () =>
    request<BookingDeleteRequestCounts>('/booking-delete-requests/counts'),
  decideBookingDelete: (id: string, approve: boolean, note?: string) =>
    request<BookingDeleteRequest>(`/booking-delete-requests/${id}/decide`, { method: 'PUT', body: JSON.stringify({ approve, note: note || null }) }),
  setBookingDeleteNeedsInfo: (id: string, note: string) =>
    request<BookingDeleteRequest>(`/booking-delete-requests/${id}/needs-info`, { method: 'PUT', body: JSON.stringify({ note }) }),

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
  // فحص: منو ناقصه صلاحيات دوره. قراءة بس — ما يمنح ولا صلاحية.
  auditRoleDefaults: () => request<MissingRoleDefault[]>('/permissions/audit-defaults'),
  getHandoverPreview: (id: string) =>
    request<{ position: string; permissions: number; bookings: number; missions: number; projects: number; complaints: number; extraTasks: number; deptHead: number }>(`/employees/${id}/handover`),
  handoverEmployee: (id: string, toEmployeeId: string) =>
    request<unknown>(`/employees/${id}/handover`, { method: 'POST', body: JSON.stringify({ toEmployeeId }) }),
  applyDefaultsAll: () => request<MissingRoleDefault[]>('/permissions/apply-defaults-all', { method: 'POST' }),
  // ⚠️ `all=1` للشاشة الإدارية (يشمل المعطّل)، ومنتقي الحجز ياخذ
  // الفعّال بس حتى ما ينحجز لقسم انلغى.
  /** فاتورة جي بي اس / داش كام بسعر يكتبه مسؤول الخدمة — مسار منفصل
   *  عن فاتورة الليدر (الي تلزم منظومات وبنود وسعرها ينحسب بالسيرفر). */
  createServiceInvoice: (data: {
    kind: 'GPS' | 'DASHCAM'
    price: number
    bookingId?: string
    customerName?: string
    customerPhone?: string
    customerAddress?: string
    note?: string
    extras?: { name: string; count: number; unitPrice: number }[]
    isFree?: boolean
    freeReasonId?: string
    freeReasonNote?: string
  }) => request<LeaderInvoice>('/leader-invoices/service', { method: 'POST', body: JSON.stringify(data) }),
  /** فاتورة بكلفة يدوية: الليدر يكتب شنو اشتغل ويحطّ السعر بنفسه —
   *  للشغل الي ماكو إله بند بجدول الكلفة. مسار منفصل بصلاحيته
   *  (`invoice_manual`)، حتى ما ينفكّ الإلزام عن الفاتورة العادية. */
  createManualInvoice: (data: {
    work: string
    price: number
    bookingId?: string
    customerName?: string
    customerPhone?: string
    customerAddress?: string
    systems?: string[]
    /** الخدمة الي تخص هاي الكلفة — منها يبني النظام معدّل السعر */
    serviceId?: string
    note?: string
    /** مجانية: السعر يقبل صفر، والسبب إجباري */
    isFree?: boolean
    freeReasonId?: string
    freeReasonNote?: string
  }) => request<LeaderInvoice>('/leader-invoices/manual', { method: 'POST', body: JSON.stringify(data) }),
  /** حجوزات الشغل داخل الشركة — مسار مستقل لأن قائمة الحجوزات
   *  العامة تنضيّق لحجوزات الموظف نفسه بلا صلاحية قراءة شاملة. */
  getInternalBookings: (status?: string) =>
    request<Booking[]>(`/bookings/internal${status ? `?status=${encodeURIComponent(status)}` : ''}`),
  /** فاتورة شغل داخل الشركة — الحجز إجباري، والخادم يرفض أي حجز
   *  نوعه مو «داخل الشركة». */
  createInternalInvoice: (data: {
    bookingId: string
    work: string
    price: number
    note?: string
  }) => request<LeaderInvoice>('/leader-invoices/internal', { method: 'POST', body: JSON.stringify(data) }),
  /** إكسل حجوزات داخل الشركة لشهر — المالك والمدير والمحاسب والمراقب. */
  exportInternalBookings: (month: string) =>
    downloadFile(`/internal-bookings/export?month=${encodeURIComponent(month)}`,
      `حجوزات-داخل-الشركة-${month}.xlsx`),
  getDesignAssets: (archived?: boolean) =>
    request<DesignAsset[]>(`/design-assets${archived ? '?archived=1' : ''}`),
  /** التصنيفات من الخادم — تسمية وحدة بالمكانين. */
  getDesignCategories: () => request<Record<string, string>>('/design-assets/categories'),
  createDesignAsset: (data: {
    title: string
    category: string
    fileKey: string
    fileType?: string
    notes?: string
  }) => request<DesignAsset>('/design-assets', { method: 'POST', body: JSON.stringify(data) }),
  setDesignAssetArchived: (id: string, archived: boolean) =>
    request<DesignAsset>(`/design-assets/${id}/archive`, {
      method: 'PUT', body: JSON.stringify({ archived }),
    }),
  getDepartments: (all?: boolean) =>
    request<Department[]>(`/departments${all ? '?all=1' : ''}`),
  createDepartment: (name: string) =>
    request<Department>('/departments', { method: 'POST', body: JSON.stringify({ name }) }),
  updateDepartment: (id: string, body: { name?: string; active?: boolean }) =>
    request<Department>(`/departments/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  createDepartmentHead: (body: { departmentId: string; name: string; phone?: string }) =>
    request<DepartmentHead>('/department-heads', { method: 'POST', body: JSON.stringify(body) }),
  updateDepartmentHead: (id: string, body: { name?: string; phone?: string; active?: boolean }) =>
    request<DepartmentHead>(`/department-heads/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  getLeaderSkills: (employeeId: string) =>
    request<LeaderSkillRating[]>(`/employees/${employeeId}/leader-skills`),
  setLeaderSkills: (employeeId: string, scores: Record<string, number>) =>
    request<LeaderSkillRating[]>(`/employees/${employeeId}/leader-skills`, {
      method: 'PUT', body: JSON.stringify({ scores }),
    }),
  getRoleDefaults: () =>
    request<Record<string, string[]>>('/permissions/role-defaults'),

  // Attendance
  checkIn: () => request<AttendanceRecord>('/attendance/checkin', { method: 'POST' }),
  checkOut: () => request<AttendanceRecord>('/attendance/checkout', { method: 'POST' }),
  getAttendanceGate: () => request<AttendanceGateState>('/attendance/gate'),
  techSuppliersCan: () => request<{ can: boolean }>('/tech-suppliers/can'),
  getMyTechSuppliers: () => request<TechSupplier[]>('/tech-suppliers/mine'),
  // (ع) 10-15: علاقات الموردين
  supplierRelations: () => request<{ rows: unknown[]; contactDays: number; kinds: Record<string, string> }>('/supplier-relations'),
  supplierTimeline: (kind: string, id: string) => request<{ at: string; type: string; text: string | null; by: string | null }[]>(`/supplier-relations/${kind}/${id}/timeline`),
  addSupplierContact: (kind: string, id: string, type: string, note: string) =>
    request<{ ok: boolean }>(`/supplier-relations/${kind}/${id}/contact`, { method: 'POST', body: JSON.stringify({ type, note }) }),
  setSupplierContactDays: (contactDays: number) => request<{ contactDays: number }>('/supplier-relations/policy', { method: 'PUT', body: JSON.stringify({ contactDays }) }),
  addTechSupplier: (body: TechSupplierIn) => request<{ id: string; message: string }>('/tech-suppliers', { method: 'POST', body: JSON.stringify(body) }),
  getAllTechSuppliers: () => request<{ suppliers: TechSupplier[]; people: { id: string; name: string; role: string; kind: string }[] }>('/tech-suppliers'),
  deleteTechSupplier: (id: string) => request<{ ok: boolean }>(`/tech-suppliers/${id}`, { method: 'DELETE' }),
  setTechSupplierAccess: (employeeId: string, supplierIds: string[]) => request<{ ok: boolean }>(`/tech-suppliers/access/${employeeId}`, { method: 'PUT', body: JSON.stringify({ supplierIds }) }),
  answerAutoCheckout: (id: string, kind: 'ACK' | 'WORKED' | 'BACK', until?: string, note?: string) =>
    request<{ message: string }>(`/attendance/auto/${id}/answer`, { method: 'POST', body: JSON.stringify({ kind, until, note }) }),
  getAttendanceClaims: (status = 'PENDING') => request<AttendanceClaim[]>(`/attendance/claims?status=${status}`),
  decideAttendanceClaim: (id: string, approve: boolean) => request<{ ok: boolean }>(`/attendance/claims/${id}/decide`, { method: 'POST', body: JSON.stringify({ approve }) }),
  getProjectPayments: (projectId: string) => request<ProjectPaymentsState>(`/projects/${projectId}/payments`),
  // ═══ مركز التدريب (ع 10-13) ═══
  tcInsights: () => request<TcInsight[]>('/training-center/insights'),
  tcTracks: () => request<{ id: string; name: string; division: string; materials: number; skills: number; masters: number; assigned: number; programs: number }[]>('/training-center/tracks'),
  tcCurriculum: () => request<TcCurriculum[]>('/training-center/curriculum'),
  tcAddCurriculum: (b: { role: string; title: string; details?: string }) => request<TcCurriculum[]>('/training-center/curriculum', { method: 'POST', body: JSON.stringify(b) }),
  tcDeleteCurriculum: (id: string) => request<TcCurriculum[]>(`/training-center/curriculum/${id}`, { method: 'DELETE' }),
  tcQuestions: (role = '') => request<TcQuestion[]>(`/training-center/questions?role=${encodeURIComponent(role)}`),
  tcAddQuestion: (b: { role: string; text: string; options: string[]; correct: number }) => request<{ ok: boolean }>('/training-center/questions', { method: 'POST', body: JSON.stringify(b) }),
  tcGenerateQuestions: (role: string, roleLabel: string, count: number) => request<{ added: number }>('/training-center/questions/generate', { method: 'POST', body: JSON.stringify({ role, roleLabel, count }) }),
  tcApproveQuestion: (id: string) => request<{ ok: boolean }>(`/training-center/questions/${id}/approve`, { method: 'POST' }),
  tcDeleteQuestion: (id: string) => request<{ ok: boolean }>(`/training-center/questions/${id}`, { method: 'DELETE' }),
  tcCreateExam: (b: { role: string; title?: string; count: number; minutes: number }) => request<{ questions: number }>('/training-center/exams', { method: 'POST', body: JSON.stringify(b) }),
  tcExamReport: (month = '') => request<TcExamReportRow[]>(`/training-center/exams/report?month=${month}`),
  tcPlans: (month = '') => request<TcPlan[]>(`/training-center/plans?month=${month}`),
  tcAddPlan: (b: { title: string; role?: string; employeeIds: string[]; plannedDate?: string; trainerId?: string; month?: string }) => request<{ ok: boolean }>('/training-center/plans', { method: 'POST', body: JSON.stringify(b) }),
  tcInspectPlan: (id: string, result: 'DONE' | 'PARTIAL' | 'NOT_DONE', notes: string) => request<{ ok: boolean }>(`/training-center/plans/${id}/inspect`, { method: 'POST', body: JSON.stringify({ result, notes }) }),
  tcDeletePlan: (id: string) => request<{ ok: boolean }>(`/training-center/plans/${id}`, { method: 'DELETE' }),
  tcMentors: () => request<{ employeeId: string; mentorId: string | null; mentorName: string | null; notes: string | null }[]>('/training-center/mentors'),
  tcSetMentor: (employeeId: string, mentorId: string, notes: string) => request<{ ok: boolean }>(`/training-center/mentors/${employeeId}`, { method: 'PUT', body: JSON.stringify({ mentorId, notes }) }),
  tcMyExam: () => request<TcMyExam>('/training-center/my-exam'),
  tcStartExam: () => request<TcMyExam>('/training-center/my-exam/start', { method: 'POST' }),
  tcSubmitExam: (answers: Record<string, number>) => request<{ score: number; total: number }>('/training-center/my-exam/submit', { method: 'POST', body: JSON.stringify({ answers }) }),
  // فترة الحجز وموظفها (ع 10-13)
  getPendingShifts: () => request<{ bookingId: string; shift: 'MORNING' | 'EVENING'; owners: { id: string; name: string; strikes: number }[] }[]>('/coordination-alerts/shifts'),
  // تبرير الموظف (ع 10-13)
  getMyJustifications: () => request<Justification[]>('/justifications/mine'),
  answerJustification: (id: string, answer: string) => request<{ ok: boolean }>(`/justifications/${id}/answer`, { method: 'POST', body: JSON.stringify({ answer }) }),
  getJustifications: () => request<Justification[]>('/justifications'),
  decideJustification: (id: string, verdict: 'ACCEPTED' | 'REJECTED', note: string) =>
    request<{ ok: boolean }>(`/justifications/${id}/decide`, { method: 'POST', body: JSON.stringify({ verdict, note }) }),
  // سجل الدفعات + تصحيح المبلغ — للمالك (ع 10-13)
  getPaymentAudit: (by = '', since = '') => request<PaymentAuditRow[]>(`/project-payments/audit?by=${encodeURIComponent(by)}&since=${encodeURIComponent(since)}`),
  setPaymentAmount: (id: string, amount: number, reason: string) =>
    request<{ ok: boolean }>(`/project-payments/${id}/amount`, { method: 'PUT', body: JSON.stringify({ amount, reason }) }),
  addProjectPayment: (projectId: string, body: { amount: number; paidAt: string; method: string; receiptNo?: string; note?: string }) =>
    request<ProjectPaymentsState>(`/projects/${projectId}/payments`, { method: 'POST', body: JSON.stringify(body) }),
  verifyProjectPayment: (id: string) => request<ProjectPaymentsState>(`/project-payments/${id}/verify`, { method: 'POST' }),
  cancelProjectPayment: (id: string, reason: string) => request<ProjectPaymentsState>(`/project-payments/${id}/cancel`, { method: 'POST', body: JSON.stringify({ reason }) }),
  fillProjectInfo: (projectId: string, body: { price?: string; responsibleId?: string }) =>
    request<{ ok: boolean }>(`/projects/${projectId}/fill-info`, { method: 'PUT', body: JSON.stringify(body) }),
  setProjectContractValue: (projectId: string, amount: number) => request<ProjectPaymentsState>(`/projects/${projectId}/contract-value`, { method: 'PUT', body: JSON.stringify({ amount }) }),
  getProjectMoneyOverview: () => request<ProjectMoney[]>('/project-payments/overview'),
  getMyOpenSession: () => request<OpenSessionResponse>('/attendance/open'),
  getTodaySummary: () => request<EmployeeDailyAttendanceSummary[]>('/attendance/today-summary'),
  // لوحة الدوام (ع 10-13): كل الموظفين بيوم — بدا/انتهى/إجازة
  getAttendanceBoard: (date: string) => request<{ date: string; rows: AttendanceBoardRow[] }>(`/attendance/board?date=${date}`),
  getMonthlyAttendance: (employeeId: string, month: string) =>
    request<MonthlyAttendanceReport>(`/attendance/employee/${employeeId}?month=${month}`),
  exportEmployeeAttendance: (employeeId: string, month: string) =>
    downloadFile(`/attendance/export/employee/${employeeId}?month=${month}`, `attendance-${employeeId}-${month}.xlsx`),
  exportTodayAttendance: (date?: string) =>
    downloadFile(`/attendance/export/today${date ? `?date=${date}` : ''}`, `attendance-today-${date || 'now'}.xlsx`),

  // KPI
  getKpiEvaluations: () => request<KpiEvaluation[]>('/kpi'),
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
  createVehicle: (data: { name: string; plateNumber: string; color?: string; type?: string; [k: string]: string | number | undefined }) =>
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

  /**
   * كم مرة انغسلت كل سيارة بشهر معيّن ومنو غسلها.
   *
   * ⚠️ «انغسلت» = اكو شخص مسمّى غسلها (صف VehicleWashRating) — مو
   * `wash > 0`. عمود wash درجة جودة ٠–٤ وممكن يكون فارغاً، وصف
   * التقييم ينكتب حتى بيوم ماكو غسل.
   */
  getVehicleWashMonthly: (month: string, vehicleId?: string) =>
    request<VehicleWashMonthly[]>(
      `/vehicles/ratings/wash-monthly?month=${encodeURIComponent(month)}${vehicleId ? `&vehicleId=${vehicleId}` : ''}`,
    ),
  /** ينزّل نفس إحصاء الغسل ملف إكسل — نفس البيانة ونفس الحارس */
  exportVehicleWashMonthly: (month: string, vehicleId?: string) =>
    downloadFile(
      `/vehicles/ratings/wash-monthly/export?month=${encodeURIComponent(month)}${vehicleId ? `&vehicleId=${vehicleId}` : ''}`,
      `wash-stats-${month}.xlsx`,
    ),

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
    /** (ع 10-16) نموذج «تقرير تنفيذ المشروع» كامل */
    sheet?: unknown
  }) => request<WorkReport>('/work-reports', { method: 'POST', body: JSON.stringify(data) }),
  getMyDuties: () => request<DutiesReport>('/duties/me'),
  getEmployeeDuties: (employeeId: string) => request<DutiesReport>(`/duties/${encodeURIComponent(employeeId)}`),
  getEmployeeLedger: (employeeId: string, period: 'day' | 'month', day?: string) =>
    request<EmployeeLedger>(`/ai/employee-ledger?employeeId=${encodeURIComponent(employeeId)}&period=${period}${day ? `&day=${day}` : ''}`),
  reviewWorkReport: (id: string, note?: string) =>
    request<{ ok: boolean }>(`/work-reports/${id}/review`, { method: 'PUT', body: JSON.stringify({ note }) }),
  reviewProjectReports: (bookingId: string, note?: string) =>
    request<{ reviewed: number }>(`/bookings/${bookingId}/progress/review`, { method: 'PUT', body: JSON.stringify({ note }) }),
  getWorkReports: (employeeId?: string, review?: 'PENDING' | 'DONE' | 'ALL') => {
    const q = new URLSearchParams()
    if (employeeId) q.set('employeeId', employeeId)
    if (review) q.set('review', review)
    const qs = q.toString()
    return request<WorkReport[]>(`/work-reports${qs ? `?${qs}` : ''}`)
  },
  // (دازاي 10-28) العدد الحقيقي — العرض لحد ٢٠٠ بس
  getWorkReportCounts: () => request<{ pending: number; done: number }>('/work-reports/counts'),
  /** (دازاي 10-31) هل انرسل تقرير عمل لهذا الحجز؟ — بالكود (B964 أو 964) بلا حد الـ٢٠٠ */
  getWorkReportsByBooking: (code: string) => request<{
    booking: { id: string; code: string; status: string; completedAt: string | null; leaderName: string | null } | null
    reports: WorkReport[]
  }>(`/work-reports/by-booking?code=${encodeURIComponent(code)}`),

  // Smart KPI
  getTechnicianKpi: (employeeId: string, month?: string) =>
    request<TechnicianKpi>(`/smart-kpi/technician/${employeeId}${month ? `?month=${month}` : ''}`),
  getKpiLeaderboard: (month?: string) =>
    request<TechnicianKpi[]>(`/smart-kpi/leaderboard${month ? `?month=${month}` : ''}`),
  getRoleKpiLeaderboard: (role: string, scope?: { leader?: string; division?: string }) => {
    const q = new URLSearchParams()
    if (scope?.leader) q.set('leader', scope.leader)
    if (scope?.division) q.set('division', scope.division)
    const qs = q.toString()
    return request<RoleKpiLeaderboard>(`/kpi/leaderboard/${role}${qs ? `?${qs}` : ''}`)
  },
  /** «تقييم بين الإداريين» — ترتيب ADMIN/OWNER/MONITOR/HR_COORDINATOR حسب
   *  عدد الحجوزات الي راجعوها. `points`/`evaluationCount` هنا عدد
   *  المراجعات لا نقاط KPI. */
  getEvaluatorLeaderboard: () => request<RoleKpiLeaderboard>('/performance-reviews/evaluator-leaderboard'),
  /** الترتيب حسب الشغل — أصحاب نفس الصلاحية ينقارنون ببعض. */
  getPermissionKpiLeaderboard: (permission: string) =>
    request<RoleKpiLeaderboard>(`/kpi/leaderboard-by-permission/${permission}`),

  // Cart
  getCartItems: (bookingId: string) => request<CartItem[]>(`/cart/booking/${bookingId}`),
  addCartItem: (bookingId: string, data: { productName: string; quantity: number; unitPrice: number; notes?: string }) =>
    request<CartItem>(`/cart/booking/${bookingId}`, { method: 'POST', body: JSON.stringify(data) }),
  deleteCartItem: (id: string) => request<void>(`/cart/${id}`, { method: 'DELETE' }),

  // Notifications
  getNotifications: () => request<{ notifications: Notification[]; unreadCount: number }>('/notifications'),
  markNotificationRead: (id: string) => request<{ success: boolean }>(`/notifications/${id}/read`, { method: 'POST' }),
  markAllNotificationsRead: () => request<{ success: boolean }>('/notifications/read-all', { method: 'POST' }),

  // Inventory
  createInventoryCheck: (data: { complete: boolean; missingItems?: string; bookingId?: string }) =>
    request<InventoryCheck>('/inventory/checks', { method: 'POST', body: JSON.stringify(data) }),
  /** حالة جرد كل واحد بكادر حجز معيّن — «جرد أدوات فريقي». */
  getBookingCrewInventory: (bookingId: string) =>
    request<BookingCrewInventoryState[]>(`/inventory/checks/booking/${bookingId}`),
  getTodaysInventoryChecks: () => request<InventoryCheck[]>('/inventory/checks/today'),
  // آخر جرد للفني نفسه — يرجّع null لو ما جرد أبداً
  getMyLastInventoryCheck: () => request<InventoryCheck | null>('/inventory/checks/mine'),
  resolveInventoryCheck: (id: string) => request<InventoryCheck>(`/inventory/checks/${id}/resolve`, { method: 'POST' }),
  // (ع) 11-04: كل النواقص المفتوحة بكل أعمارها
  // (ع) 11-07
  getRevisits: (days: number) => request<RevisitReport>(`/revisits?days=${days}`),
  getCustomerSignature: (bookingId: string) => request<CustomerSignature | null>(`/bookings/${bookingId}/customer-signature`),
  saveCustomerSignature: (bookingId: string, signerName: string, image: string) =>
    request<{ ok: boolean }>(`/bookings/${bookingId}/customer-signature`, { method: 'POST', body: JSON.stringify({ signerName, image }) }),
  getOpenShortages: () => request<InventoryCheck[]>('/inventory/checks/open'),
  setShortagePriority: (id: string, priority: string) => request<{ ok: boolean }>(`/inventory/checks/${id}/priority`, { method: 'PUT', body: JSON.stringify({ priority }) }),
  // (ع) 11-04: رفع طلب للمدير
  createManagerRequest: (body: { title: string; details: string; priority: ManagerRequest['priority']; attachInventory: boolean }) =>
    request<{ id: string }>('/manager-requests', { method: 'POST', body: JSON.stringify(body) }),
  getMyManagerRequests: () => request<ManagerRequest[]>('/manager-requests/mine'),
  getManagerRequestInventoryContext: () => request<{ context: string }>('/manager-requests/inventory-context'),
  getManagerRequests: () => request<ManagerRequest[]>('/manager-requests'),
  handleManagerRequest: (id: string, status: 'DONE' | 'REJECTED', reply: string) =>
    request<{ ok: boolean }>(`/manager-requests/${id}/handle`, { method: 'POST', body: JSON.stringify({ status, reply }) }),
  getPersonalTools: (employeeId?: string) =>
    request<PersonalTool[]>(`/inventory/personal${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createPersonalTool: (data: { employeeId: string; name: string; barcode: string }) =>
    request<PersonalTool>('/inventory/personal', { method: 'POST', body: JSON.stringify(data) }),
  // سجل حركة الأدوات: toolId لأداة وحدة، employeeId لكل عدة موظف، بلا شي = الكل
  // سياسة الخصوصية
  getPrivacyPolicy: (all?: boolean) =>
    request<PrivacyPolicyPoint[]>(`/privacy-policy${all ? '?all=true' : ''}`),
  getPrivacyPolicyStatus: () => request<PrivacyPolicyStatus>('/privacy-policy/status'),
  acceptPrivacyPolicy: (hash?: string) =>
    request<{ accepted: boolean }>('/privacy-policy/accept', { method: 'POST', body: JSON.stringify({ hash: hash ?? '' }) }),
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
  /** (ع) 10-16: أدوات المخزن التخصصية — أبو الكميات ينطي موظف من المخزن ويرجّع */
  issueStoreTool: (toolId: string, employeeId: string, note?: string) =>
    request<ToolRequestItem>(`/store-tools/${toolId}/issue`, { method: 'POST', body: JSON.stringify({ employeeId, note: note || null }) }),
  returnStoreTool: (requestId: string) =>
    request<ToolRequestItem>(`/store-tools/issued/${requestId}/return`, { method: 'POST' }),
  createOnDemandTool: (data: { name: string; barcode: string; totalQuantity: number }) =>
    request<OnDemandTool>('/inventory/ondemand', { method: 'POST', body: JSON.stringify(data) }),
  updateOnDemandTool: (id: string, data: Partial<OnDemandTool>) =>
    request<OnDemandTool>(`/inventory/ondemand/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  // يفتح رابط الخريطة بالسيرفر ويرجع النقطة — ضروري للروابط المختصرة
  // (maps.app.goo.gl) الي ما تحتوي إحداثيات إلا بعد ما تنفتح.
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

  /**
   * ═══ استثناءات العدة القياسية لموظف بعينه ═══
   *
   * حذف أداة من عدة موظف **يستثنيها تلقائياً** بالخادم — بدون هذا،
   * اسمها لسه بالقالب القياسي فترجع فوراً بتقرير النواقص كـ«خالصة
   * من المخزن»، وهاي بالضبط شكوى أبو الكميات.
   *
   * ⚠️ الاستثناء لهذا الموظف وحده — القالب مشترك، وحذفها منه يعني
   * ماكو ولا فني يتحاسب عليها.
   */
  getToolExemptions: () => request<PersonalToolExemption[]>('/inventory/tool-exemptions'),
  /**
   * يحذف أداة من نواقص موظف بعينه (استثناء) — القالب يبقى مشترك، فالأداة
   * تبقى محسوبة على باقي الفنيين. الحذف يظهر بقائمة «مستثناة من نواقصه»
   * ويُرجَّع بضغطة، حتى ما يصير استثناء مخفي يخلي التقرير يكذب بصمت.
   */
  exemptTool: (employeeId: string, toolName: string, note?: string) =>
    request<{ ok: boolean }>('/inventory/tool-exemptions', {
      method: 'POST',
      body: JSON.stringify({ employeeId, toolName, note: note ?? null }),
    }),
  /** يرجّع الأداة لنواقص الموظف — تراجع عن الاستثناء */
  unexemptTool: (employeeId: string, toolName: string) =>
    request<{ ok: boolean }>(
      `/inventory/tool-exemptions?employeeId=${encodeURIComponent(employeeId)}&toolName=${encodeURIComponent(toolName)}`,
      { method: 'DELETE' },
    ),
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
  getComplaints: (opts?: { open?: boolean }) => request<Complaint[]>(`/complaints${opts?.open ? '?open=1' : ''}`),
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

  // Procurement
  getProcurementRequests: () => request<ProcurementRequest[]>('/procurement'),
  getProcurementStats: () => request<ProcurementStats>('/procurement/stats'),
  createProcurementRequest: (data: { requestedById: string; bookingId?: string; requestType: ProcurementRequestType; notes?: string; items: { productName: string; quantity: number; unitPrice?: number | null; model?: string | null; spec?: string | null; itemNotes?: string | null }[] }) =>
    request<ProcurementRequest>('/procurement', { method: 'POST', body: JSON.stringify(data) }),
  updateProcurementStatus: (id: string, status: string) =>
    request<ProcurementRequest>(`/procurement/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
  fulfillProcurementRequest: (id: string, data: { fulfilledById: string; supplierId: string; totalCost?: number; fulfillmentNotes?: string; items?: { id: string; unitPrice?: number; totalPrice?: number; fulfilled?: boolean }[]; quotes?: { supplierId: string; totalPrice: number }[] }) =>
    request<ProcurementRequest>(`/procurement/${id}/fulfill`, { method: 'PUT', body: JSON.stringify(data) }),

  // صيانة الأجهزة العامة (شيت "صيانة الاجهزة") — ليدر فقط
  getDeviceMaintenanceTickets: () => request<DeviceMaintenanceTicket[]>('/device-maintenance'),
  createDeviceMaintenanceTicket: (data: { appointmentDate?: string | null; customerCode: number; deviceTypeName: string; problem: string; deviceSerial?: string | null }) =>
    request<DeviceMaintenanceTicket>('/device-maintenance', { method: 'POST', body: JSON.stringify(data) }),
  updateDeviceMaintenanceTicket: (id: string, data: { appointmentDate?: string | null; deviceTypeName?: string; problem?: string; deviceSerial?: string | null; markReceived?: boolean; markDelivered?: boolean }) =>
    request<DeviceMaintenanceTicket>(`/device-maintenance/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // جرد الفريق ("جرد العدد") — ليدر فقط
  getTeamInventoryChecks: () => request<TeamInventoryCheck[]>('/team-inventory/checks'),

  // تقدير مدة العمل المتعلَّم تلقائياً (بدون رقم مفروض يدوياً) — يرجع expectedMinutes:
  // null لو البيانات التاريخية غير كافية بعد (sampleCount < minSamples).
  getJobDurationEstimate: (params: { systemName: string; jobType: 'INSTALL' | 'MAINTENANCE'; itemCount: number; crewSize: number }) =>
    request<JobDurationEstimate>(
      `/job-duration-estimate?systemName=${encodeURIComponent(params.systemName)}&jobType=${params.jobType}&itemCount=${params.itemCount}&crewSize=${params.crewSize}`,
    ),

  // ماتركس: توقع التأخير (أدوار/صلاحية تنسيق الحجوزات) وفرص البيع (sales_booking)
  getDelayRisks: () => request<DelayRisk[]>('/ai/delay-risks'),
  getSalesOpportunities: () => request<SalesOpportunity[]>('/ai/opportunities'),
  // ماتركس: الكادر الأنسب (حارس التكليف)، منحنى الجدد والتقرير الأسبوعي (ADMIN/OWNER)،
  // المواد (inventory)، الاستبدال (ADMIN/OWNER أو it_assets/vehicle_management)
  getCrewRecommendation: (bookingId: string) =>
    request<CrewRecommendation>(`/ai/crew-recommendation?bookingId=${encodeURIComponent(bookingId)}`),
  getNewEmployeeCurves: () => request<NewEmployeeCurve[]>('/ai/new-employee-curves'),
  getStockForecast: () => request<StockForecast>('/ai/stock-forecast'),
  setMaterialStock: (id: string, quantity: number) =>
    request<{ ok: boolean }>(`/materials/${encodeURIComponent(id)}/stock`, { method: 'PUT', body: JSON.stringify({ quantity }) }),
  getReplacementSuggestions: () => request<ReplacementSuggestions>('/ai/replacement-suggestions'),
  getMyWatch: () => request<MatrixWatch>('/ai/my-watch'),
  // رسول ماتركس (ع 10-13)
  getCourierState: () => request<CourierState>('/matrix-courier/state'),
  getCourierInbox: () => request<CourierMessage[]>('/matrix-courier/inbox'),
  courierDelivered: (id: string) => request<void>(`/matrix-courier/${id}/delivered`, { method: 'POST' }),
  courierSeen: (link: string) => request<void>(`/matrix-courier/seen?link=${encodeURIComponent(link)}`, { method: 'POST' }),
  getMatrixGuide: (path: string, label: string) =>
    request<{ text: string; source: 'RULES' | 'MODEL' | 'NONE' }>(`/ai/guide?path=${encodeURIComponent(path)}&label=${encodeURIComponent(label)}`),
  getEmployeeReport: (employeeId: string, day?: string) =>
    request<EmployeeReport>(`/ai/employee-report?employeeId=${encodeURIComponent(employeeId)}${day ? `&day=${day}` : ''}`),
  getMatrixBusiness: () => request<MatrixBusiness>('/ai/business'),
  getMatrixProposals: (status = 'PENDING') => request<MatrixProposal[]>(`/ai/proposals?status=${encodeURIComponent(status)}`),
  approveMatrixProposal: (id: string, payload?: unknown) =>
    request<{ ok: boolean }>(`/ai/proposals/${encodeURIComponent(id)}/approve`, { method: 'POST', body: JSON.stringify({ payload: payload ?? null }) }),
  rejectMatrixProposal: (id: string, note: string) =>
    request<{ ok: boolean }>(`/ai/proposals/${encodeURIComponent(id)}/reject`, { method: 'POST', body: JSON.stringify({ note }) }),
  getGuideRules: () => request<{ rules: MatrixGuideRule[]; modelEnabled: boolean }>('/ai/guide-rules'),
  createGuideRule: (r: Partial<MatrixGuideRule>) => request<MatrixGuideRule>('/ai/guide-rules', { method: 'POST', body: JSON.stringify(r) }),
  updateGuideRule: (id: string, r: Partial<MatrixGuideRule>) => request<MatrixGuideRule>(`/ai/guide-rules/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(r) }),
  deleteGuideRule: (id: string) => request<{ ok: boolean }>(`/ai/guide-rules/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  getMatrixFeed: () => request<MatrixFeedItem[]>('/ai/feed'),
  getMatrixTrend: () => request<MatrixTrendDay[]>('/ai/daily-trend'),
  getLateFocus: () => request<LateFocus>('/ai/late-focus'),
  askMatrix: (question: string) => request<{ answer: string; source: 'AGENT' | 'MODEL' | 'RULES'; steps?: string[] }>('/ai/ask', { method: 'POST', body: JSON.stringify({ question }) }),
  getGroupPerformance: (group: string) => request<GroupPerformance>(`/ai/group-performance?group=${encodeURIComponent(group)}`),
  getRoleWatch: () => request<{ group: string; employees: MatrixWatch[]; red: number; alert: number }[]>('/ai/role-watch'),
  getMatrixDecisions: (day?: string) =>
    request<MatrixDecisions>(`/ai/decisions${day ? `?day=${encodeURIComponent(day)}` : ''}`),
  resumeMatrixKind: (kind: string) =>
    request<{ ok: boolean }>(`/ai/actions/kinds/${encodeURIComponent(kind)}/resume`, { method: 'POST' }),
  undoMatrixAction: (id: string) => request<{ ok: boolean }>(`/ai/actions/${encodeURIComponent(id)}/undo`, { method: 'POST' }),
  getWeeklyReport: (week?: string) =>
    request<WeeklyReport>(`/ai/weekly-report${week ? `?week=${encodeURIComponent(week)}` : ''}`),

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
  getProjectMonths: () => request<{ month: string; new: number; completed: number; inProgress: number }[]>('/stats-management/projects/monthly'),
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
export const purchaseApi = {
  control: () => request<PurchaseControlRow[]>('/procurement/control'),
  review: (id: string, r: SupplierReviewInput) => request<{ ok: boolean }>(`/procurement/${id}/review`, { method: 'POST', body: JSON.stringify(r) }),
  verify: (id: string) => request<{ ok: boolean }>(`/procurement/${id}/verify`, { method: 'PUT', body: JSON.stringify({}) }),
  prices: (q: string) => request<PurchasePriceRow[]>(`/procurement/prices?q=${encodeURIComponent(q)}`),
  alerts: () => request<PurchaseAlert[]>('/procurement/alerts'),
  reviewAlert: (id: string, note: string) => request<{ ok: boolean }>(`/procurement-alerts/${id}`, { method: 'PUT', body: JSON.stringify({ note }) }),
  supplierScores: () => request<SupplierScore[]>('/suppliers/scores'),
  policy: () => request<{ quoteThreshold: number }>('/procurement-policy'),
  setPolicy: (quoteThreshold: number) => request<{ quoteThreshold: number }>('/procurement-policy', { method: 'PUT', body: JSON.stringify({ quoteThreshold }) }),
  quotes: (id: string) => request<{ supplier: string; totalPrice: number; chosen: boolean; note: string | null }[]>(`/procurement-quotes/${id}`),
  bias: () => request<PurchaseBias[]>('/procurement-bias'),
  standards: () => request<{ standards: StandardPrice[]; suggestions: StandardSuggestion[]; ceilingPct: number }>('/procurement-standards'),
  setStandard: (product: string, price: number) => request<{ ok: boolean }>('/procurement-standards', { method: 'PUT', body: JSON.stringify({ product, price }) }),
  acceptStandard: (key: string, accept: boolean) => request<{ ok: boolean }>('/procurement-standards/accept', { method: 'POST', body: JSON.stringify({ key, accept }) }),
  setCeiling: (ceilingPct: number) => request<{ ceilingPct: number }>('/procurement-policy/ceiling', { method: 'PUT', body: JSON.stringify({ ceilingPct }) }),
  monthly: (month: string) => request<PurchaseMonth>(`/procurement-monthly?month=${month}`),
  monthlyExcel: async (month: string) => {
    const token = currentToken()
    const res = await fetch(`${API_URL}/procurement-monthly?month=${month}&format=xlsx`, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (!res.ok) throw new Error('تعذر تنزيل الإكسل')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(await res.blob())
    a.download = `المشتريات-${month}.xlsx`
    a.click()
  },
}
