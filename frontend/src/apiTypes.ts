// ═══ أنواع واجهة السيرفر (٧ح، 11-04) ═══
// انفصلت من api.ts بلا أي تغيير بالمحتوى. api.ts يعيد تصديرها كلها، فكل الاستيرادات القديمة تشتغل.

export interface AttendanceRecord {
  id: string
  employeeId: string
  checkIn: string
  checkOut: string | null
  date: string
  employee: { id: string; name: string } | null
  /** الساعات المحسوبة (حد حسب ساعة الحضور) وسبب الانصراف المفترض إذا انقطع شي. */
  countedMinutes?: number
  assumedNote?: string
}

export interface DailyAttendance {
  date: string
  sessions: AttendanceRecord[]
  firstCheckIn: string
  lastCheckOut: string | null
  stillOpen: boolean
  totalMinutes: number
  assumed?: boolean
  /** الخادم يحسبه من بداية الدوام الفعلية (بتوقيت بغداد). */
  late?: boolean
}

export interface AttendanceBoardRow {
  id: string; name: string; role: string; shift: string | null; shiftStart: string | null; shiftEnd: string | null
  firstIn: string | null; lastOut: string | null; open: boolean; minutes: number; leaveKind: string | null
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
  /** نوع خدمة الفاتورة اليدوية: جي بي اس أو داش كام. null = ما انتأشّرت */
  serviceKind?: 'GPS' | 'DASHCAM' | null
  category: string | null
  division: Division
  skills: Skill[]
  /** الخدمة تطلب عدد أجهزة ونوع مركبة وقت الحجز (جي بي اس) */
  requiresDeviceInfo?: boolean
  /** ═══ الورق (تقرير + فاتورة) على مسؤول الخدمة مو على الفني ═══
   *  للخدمات الي يكفيها فني واحد — جي بي اس وداش كام.
   *  ⚠️ منفصل عن `requiresDeviceInfo` عمداً: معنيان مختلفان. */
  managerHandlesPaperwork?: boolean
}

export interface ServiceManager {
  id: string
  createdAt: string
  employee: { id: string; name: string; position: string | null } | null
  service: Service | null
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
  | 'IT_SUPPORT'
  | 'MEDIA'
  | 'PUBLIC_RELATIONS'
  | 'CAR_WASHER'
  | 'SERVICE_DEVELOPER'
  | 'OWNER'

export interface Employee {
  /** الأدوار الفعلية من منصبه بالهيكل (دازاي 11-01) — من /auth/me */
  orgRoles?: string[]
  /** (11-04) جدول «شنو يشوف» من محرك الصلاحيات بالسيرفر (/auth/me) */
  can?: Record<string, boolean>
  id: string
  name: string
  certificate: string | null
  /** مسار صورة الموظف — فاضي يعني نعرض أول حرف اسمه */
  photoUrl?: string | null
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
  contractEndDate?: string | null
  shift: 'MORNING' | 'EVENING' | null
  shiftStart: string | null
  shiftEnd: string | null
  monthlyLeaves: number
  jobTitle: string | null
  division: Division
  /**
   * أدوار إضافية — **حزمة صلاحيات مو رتبة**.
   *
   * 🔴 الواحد عندنا يسوي أكثر من شغلة: فني بالميدان ويتولى حاسبات
   * المكتب. والدور الثاني يصبّ صلاحيات دوره وبس — ما يبدّل `role`
   * ولا يمرّ بأي حارس يقرا الدور.
   *
   * ⚠️ و**ما يكون OWNER ولا ADMIN أبداً** — الخادم يشيلهم من القائمة
   * مهما انبعثت، فالقيد مو على الواجهة وحدها.
   */
  secondaryRoles?: EmployeeRole[]
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
  /**
   * الرصيد مفصَّلاً لكل دوار.
   *
   * ⚠️ المجموع وحده ما يكفي: موظف أخذ من دوارين «عليه ٣٠ ألف»، بس
   * كل مبلغ لازم يرجع للدوار الي طلع منه. بلا هذا التفصيل الواجهة
   * تضطر تخمّن الدوار — وهاي چانت العلّة الي رجّعت الفلوس لدوار غلط.
   */
  funds: EmployeeFundLine[]
}

/** الي بيد الموظف من دوار واحد بعينه */
export interface EmployeeFundLine {
  fundId: string
  fundName: string
  totalTaken: number
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

/** حالتا الإجازة — الفرق بينهن مهلة التقديم مو تصنيف إداري. */
export type LeaveKind = 'REGULAR' | 'URGENT'

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
  kind: LeaveKind
  kindLabel: string
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

export interface ReturnedFromProjects {
  id: string; code: string; customer: string | null; phone: string | null; status: string
  returnedAt: string; note: string | null; by: string | null; scheduledAt: string | null
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

/** حجز منجز ينتظر تقييم كادره. */
export interface BookingAwaitingReview {
  bookingId: string
  code: string
  customerName: string
  serviceName: string | null
  customerPhone: string | null
  customerAddress: string | null
  completedAt: string | null
  crew: CrewReviewState[]
}

/** ═══ أنواع التقييم ═══
 *
 * «يحتاج تدريب» غير «مخالفة سلوك»: الأول نقص مهارة علاجه دورة،
 * والثاني إجراء إداري. خلطهن يظلم الاثنين.
 *
 * ⚠️ الليدر **يبلّغ** بالمخالفة ما يغرّم — الغرامة قرار الإدارة. */
export type ReviewRating = 'POSITIVE' | 'NEEDS_TRAINING' | 'MISCONDUCT' | 'COMMITMENT'

export interface CrewReviewState {
  employeeId: string
  name: string
  position: string | null
  /** فاضي إذا لسه ما انقيّم بهذا الحجز. */
  rating: ReviewRating | null
  reason: string | null
  /** درجات ١-٥ اختيارية — فاضية يعني ما انطّى نجوم، مو صفر. */
  commitmentScore: number | null
  speedScore: number | null
  qualityScore: number | null
}

export interface PerformanceReview {
  id: string
  employeeId: string
  evaluatorId: string
  rating: ReviewRating
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
  // حجز ديكور مربوط (ع 10-16)
  parentBookingId?: string | null
  decorTiming?: 'BEFORE' | 'WITH' | null
  /** (ع 10-16) ليدر الحجز بالتعريف الموحّد بالخادم (booking_leader_id) — استعمله بدل الحساب بالمتصفح */
  leaderId?: string | null
  leaderName?: string | null
  /** قرار (ع) 10-09: قيمة الشغل الداخلي من فواتيره الداخلية */
  internalInvoiceTotal?: number
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
  /** انأجّل بلا موعد — منزاح من جدول اليوم لحد ما ينحدد له موعد */
  awaitingReschedule: boolean
  /** تسوية إدارية: انقفل «منجز بدون تفاصيل» من المالك */
  settledLegacyAt?: string | null
  settledLegacyNote?: string | null
  // ═══ تتبّع المراحل ═══
  createdByName?: string
  // (ع) 10-14: matrix_app = سجّله ماتركس من تطبيق الزبائن
  source?: string | null
  crewNotes?: string
  crewNotesByName?: string
  crewNotesAt?: string
  projectNotes?: string
  projectNotesByName?: string
  projectNotesAt?: string
  cancelledAt?: string
  cancelledByName?: string
  cancelReason?: string
  /** محسوبة بالسيرفر — أي سلّة ينتمي لها الحجز (فاضية = ولا وحدة) */
  stageBucket?: StageBucket | ''
  lastPostponedAt: string | null
  postponeReason: string | null

  workStoppedAt: string | null
  workStopReason: string | null
  hasInvoice: boolean
  hasReport: boolean
  completionState: CompletionState
  /** منو راجع الفلوس وطابقها بشاشة التدقيق اليومي (آخر فاتورة للحجز) —
   *  غير «أكّده» بالهوية (ذاك إنجاز الشغل الفني). فاضي لو الحجز صُحّح
   *  بسكربت تسوية جماعية ولا مرّ بشاشة التدقيق أصلاً. */
  financeAuditedByName?: string | null
  financeAuditedAt?: string | null
  /** ربط بحجز آخر كإنجاز جزئي لنفس الشغلة — شغلة حقيقية استوردت
   *  تاريخياً كصفوف منفصلة ليوم ١ ويوم ٢... غير حالة PARTIAL الحية. */
  partialJobBooking?: { id: string; code: string } | null
  sequenceNumber: number | null
  scheduledAt: string | null
  // نهاية المدى المتفق عليه مع الزبون (ساعة بعد البداية) — الموعد
  // ينحكى للزبون «من ... إلى ...» مو وقت واحد.
  scheduledEndAt: string | null
  scheduleLogs: { id: string; changedById: string; changedBy: { id: string; name: string; role: string }; oldTime: string | null; newTime: string; kind: string; reason: string | null; createdAt: string }[]
  materialsReadyAt: string | null
  materialsReadyBy: { id: string; name: string } | null
  acceptedAt?: string | null
  departedAt?: string | null
  departOrigin?: DepartOrigin | null
  autoDepartedAt?: string | null
  autoArrivedAt?: string | null
  autoCompletedAt?: string | null
  responseMinutes: number | null
  arrivedAt: string | null
  startedAt: string | null
  confirmationContactedAt: string | null
  // ترحيل للتقني (0326)
  handoverToId?: string | null
  handoverAt?: string | null
  handoverReason?: string | null
  handoverTo?: { id: string; name: string } | null
  /** (ع) 11-04: وين واصل الحجز المرحّل وشكد صارله — بمحطتي المشاريع والتقنيين بس */
  tracking?: { kind: 'PROJECT' | 'TECH'; stage: string; since: string; days: number; limit: number; late: boolean } | null
  handoverBy?: { id: string; name: string } | null
  techContactedAt?: string | null
  techDiagnosis?: string | null
  techDiagnosedAt?: string | null
  techCrewRequestedAt?: string | null
  techDecision?: 'PHONE' | 'VISIT' | null
  techDecidedAt?: string | null
  techVisitAt?: string | null
  techVisitedAt?: string | null
  techVisitMoves?: number
  // وقت تحويل الحجز لتنسيق الحجوزات (التثبيت)
  confirmedAt: string | null
  // الموقع: عنوان كلامي + نقطة على الخريطة + رابط (الرابط يغني عن التحديد)
  locationUrl: string | null
  confirmationContactedBy: { id: string; name: string } | null
  systemType: string | null
  systemCount: number | null
  deviceCount: number | null
  bookingType: string
  /** فترة الشغل: صباحي أو مسائي. `null` = غير محدَّدة (ما نخمّنها). */
  shift: 'MORNING' | 'EVENING' | null
  /** حجز داخل الشركة — الطالب وقسمه وملاحظة إداري الكوادر. */
  internalEmployeeName?: string | null
  internalEmployeePhone?: string | null
  internalDepartment?: string | null
  internalDepartmentId?: string | null
  internalHeadId?: string | null
  internalHrNote?: string | null
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
  /** وصل مرحلة «البدء بالتنفيذ» — الي يفتح الحجز للتنسيق */
  projectExecutionAt?: string | null
  /** عند إدارة المشاريع وما وصل التنفيذ: يشوفه المنسّق بس ما يلمسه */
  projectLocked: boolean
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

// ═══ الطلعات ═══
//
// «كل مرة طلعناله تنحسب حجز للموظف، وكل مرة ينكتب بيها تاريخ وكادر
// طلع — لأن يجوز الكادر يتغيّر». الحجز الي ياخذ أربع أيام أربع
// طلعات، وكل طلعة إلها كادرها الي طلع بيها هي.
export interface BookingVisitCrewMember {
  employeeId: string
  name: string
  role: string
  isLeader: boolean
}

export interface BookingVisit {
  id: string
  bookingId: string
  visitNumber: number
  /** PARTIAL = يوم شغل انقفل والحجز ما خلص · DONE = الطلعة الي خلّصته */
  outcome: 'PARTIAL' | 'DONE'
  percentDone: number | null
  progressReportId: string | null
  scheduledAt: string | null
  occurredAt: string
  createdAt: string
  crew: BookingVisitCrewMember[]
}

/** ═══ لوحة اليوم ═══ «شنو صاير اليوم» و«شغلي اليوم» ولوحة الشرف. */
export interface TodayBoardData {
  bookingsToday: number
  inField: number
  completedToday: number
  newToday: number
  needsContact: number
  needsCrew: number
  needsPaper: number
  needsFinish: number
  /** نافذة ٧ أيام متدحرجة تنتهي اليوم — لحساب «نسبة الإنجاز الأسبوعية» */
  weekTotal: number
  weekCompleted: number
  topCrew: { employeeId: string; name: string; photoUrl: string | null; visits: number; done: number }[]
  last14: { day: string; count: number }[]
}

export interface JobDurationEstimate {
  expectedMinutes: number | null
  sampleCount: number
  minSamples: number
}

/** ماتركس — حجز متوقع ياخذ أكثر من وقته المحجوز (GET /ai/delay-risks) */
export interface DelayRisk {
  bookingId: string
  bookingCode: string
  scheduledAt: string
  expectedMinutes: number
  availableMinutes: number
  samples: number
  basis: 'LEADER' | 'SERVICE'
  limitedBy: 'SHIFT_END' | 'NEXT_BOOKING'
  nextBookingCode?: string
}

/** ماتركس — فرصة بيع (GET /ai/opportunities) */
export interface SalesOpportunity {
  reason: 'CROSS_SELL' | 'MAINTENANCE_DUE'
  reasonLabel: string
  customerId: string
  customerCode: string
  customerName: string
  customerPhone: string
  basedOnService: string
  suggestedService: string
  suggestedServiceId?: string
  lastBookingCompleted: string
}

/** ماتركس — مرشح كادر (GET /ai/crew-recommendation) */
export interface CrewSuggestion {
  employeeId: string
  name: string
  isLeader: boolean
  score: number
  hasSkill: boolean
  doneCount: number
  problemRatePct?: number
  dayLoad: number
  reasons: string[]
}
export interface CrewRecommendation {
  bookingId: string
  serviceName: string
  day: string
  insufficient: boolean
  note?: string
  leaders: CrewSuggestion[]
  technicians: CrewSuggestion[]
}

/** ماتركس — منحنى الموظف الجديد (GET /ai/new-employee-curves) */
export interface CurveWeek {
  week: number
  complete: boolean
  completed: number
  stops: number
  complaints: number
  paperDue: number
  paperOk: number
  paperPct?: number
  score: number
}
export interface NewEmployeeCurve {
  employeeId: string
  name: string
  isLeader: boolean
  startDate: string
  daysIn: number
  weeks: CurveWeek[]
  status: 'OK' | 'NEEDS_FOLLOWUP' | 'INSUFFICIENT'
  statusNote: string
}

/** ماتركس — استهلاك المواد (GET /ai/stock-forecast) */
export interface StockForecast {
  stockTracked: boolean
  note: string
  windowDays: number
  minUsages: number
  horizonDays: number
  items: { key: string; name: string; usages: number; totalQuantity: number; avgDaily: number; need14Days: number; remaining?: number; daysLeft?: number; countedAt?: string }[]
}

/** ماتركس — اقتراح استبدال (GET /ai/replacement-suggestions) */
export interface ReplacementSuggestions {
  itIncluded: boolean
  vehiclesIncluded: boolean
  it: { id: string; name: string; kind: string; status: string; repairCount: number; repairCost: number; reason: string }[]
  vehicles: { id: string; name: string; plateNumber: string; cost12m: number; incidents180d: number; reasons: string[] }[]
  fleetMedian?: number
  fleetNote?: string
}

/** ماتركس — تقرير المالك الأسبوعي (GET /ai/weekly-report) */
export interface WeekTotals {
  completed: number
  revenue: number
  complaints: number
  workStops: number
  latePaperwork: number
}
export interface WeeklyMover { employeeId: string; name: string; thisScore: number; lastScore: number; delta: number; note: string }
export interface WeeklyReport {
  week: string
  from: string
  to: string
  partial: boolean
  this: WeekTotals
  last: WeekTotals
  improving: WeeklyMover[]
  declining: WeeklyMover[]
  openDecisions: { pendingLeaves: number; deleteRequests: number; unreviewedVerdicts: number }
  weeks: string[]
  matrix?: MatrixDecisions['accuracy']
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
  /** ═══ الحجز الي انصرف عليه ═══
   *  ⚠️ **قابل للفراغ**: المصاريف المسجّلة قبل الربط ما إلها حجز،
   *  وتنعرض بجدول منفصل «ما تنحسب» بدل ما تنحسب غلط أو تختفي بصمت. */
  bookingId?: string | null
  bookingCode?: string
}

export interface Customer {
  id: string
  createdAt: string
  customerCode: number
  code: string
  name: string
  phone: string
  location: string | null
  mapLatitude: number | null
  mapLongitude: number | null
  services: string[]
  previousBookingsCount?: number
  /** حجوزات الزبون المفتوحة — تنبيه قبل حجز ثاني (أكثر سبب للتكرار). */
  openBookings?: { id: string; code: string; status: string; scheduledAt: string | null; address: string | null; service: string | null; createdBy: string | null; createdAt: string }[]
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

export type AuditVerdict = 'NEEDS_FOLLOWUP' | 'RECHECK_RATING' | 'APPROVED'

export interface ComplaintEvent {
  id: string
  complaintId: string
  kind: 'CREATED' | 'CONTACTED' | 'RATED' | 'NOTED' | 'ASSIGNED' | 'AUDITED' | 'RESOLVED'
  detail: string | null
  byName: string | null
  createdAt: string
}

export interface LeaderSkillRating {
  id: string
  employeeId: string
  skill: string
  score: number
  ratedByName: string | null
  ratedAt: string
}

/** إجراء إغلاق بلاغ التدقيق — نفس ثوابت الخادم. */
export type AuditCloseAction = 'PENALIZE' | 'NO_FAULT'

export interface MissingRoleDefault {
  employeeId: string
  employeeName: string
  role: string
  missing: string[]
  missingLabels: string[]
  /** صلاحيات مو معرّفة بقاعدة البيانات أصلاً — مو «الموظف ناقصه». */
  notSeeded: string[] | null
}

export interface ComplaintCustomerStat {
  customerId: string
  customerName: string
  customerPhone: string
  complaintCount: number
  openCount: number
  notContactedCount: number
  /** ⚠️ null لمن ماكو ولا تقييم — **مو صفر**. اعرضها «—». */
  avgRating: number | null
  lastContactAt: string | null
  contactedLast30: number
  needsAuditCount: number
  /** حالة ومهندس **آخر** شكوى للزبون — الصف مجمّع. */
  latestStatus: Complaint['status'] | null
  latestEngineer: string | null
  latestComplaintId: string | null
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
  status: 'PENDING' | 'CONTACTED_OK' | 'CONTACTED_ISSUE' | 'CONVERTED' | 'RECONTACT' | 'CLOSED'
  contactNotes: string | null
  contactedAt: string | null
  createdAt: string
  booking: Booking
  customer: Customer
  contactedByEmployee: { id: string; name: string } | null
  /** حجز الصيانة المربوط — تبقى المتابعة معلّقة لحد ما ينجز، وقتها تتحول RECONTACT تلقائياً */
  linkedBooking: { id: string; code: string; status: string } | null
  /** حكم الجودة */
  reportType: 'POSITIVE' | 'NEGATIVE' | 'STUBBORN' | null
  inspectionStatus: 'NONE' | 'PENDING' | 'DONE'
  inspectionResult: 'CUSTOMER_RIGHT' | 'CUSTOMER_WRONG' | null
  inspectionNotes: string | null
  inspectedAt: string | null
  inspectedBy: { id: string; name: string } | null
  penalizedEmployee: { id: string; name: string } | null
  monitorVerdict: 'OK' | 'NOT_OK' | 'SHORTCOMING' | null
  monitorNote: string | null
  monitorAt: string | null
  monitorBy: { id: string; name: string } | null
  monitorTarget: { id: string; name: string } | null
  execution: BookingExecutionDetail | null
  // تفاصيل المشروع والمبالغ — مهندس الجودة يحتاجها وهو يتصل بالزبون
  financials?: QualityFollowUpFinancials | null
}

// ═══ 🗂️ السكرتير — نفس شكل العقد (docs/secretary.md) ═══
export interface SecretaryCard {
  kind: 'booking' | 'project' | 'customer' | 'employee' | 'attendance' | 'task' | 'summary' | 'info'
  title: string; status: string; owner: string | null; updatedAt: string | null; lines: string[]; link: string | null
}
export interface SecretaryProposal {
  actionId: string; kind: 'task'; summary: string
  fields: { employeeId: string; employeeName: string; title: string; description: string; priority: 'NORMAL' | 'URGENT'; dueAt: string }
  candidates: { id: string; name: string }[]
}
export interface SecretaryConversation { id: string; title: string; messages: number; createdAt: string; updatedAt: string }
export interface SecretaryReply { messageId: string; conversationId: string; transcript: string | null; reply: string; cards: SecretaryCard[]; proposal: SecretaryProposal | null }
export interface SecretaryChat {
  id: string; at: string; role: 'user' | 'assistant'; text: string; transcript: string | null; cards: SecretaryCard[]
  // (دازاي 10-29) البصمة تضل ظاهرة: hasAudio من السجل، audioUrl للبصمة الي انرسلت هسه
  hasAudio?: boolean; audioUrl?: string
  action: { id: string; status: 'PROPOSED' | 'SENT' | 'CANCELLED' | 'FAILED'; result: string | null } | null
}
export interface SecretaryPending { id: string; kind: string; question: string; why: string; impact: string; affects: string; link: string; createdAt: string }
export interface SecretaryToday { greeting: string; summary: { lines: string[]; cards: SecretaryCard[] }; pending: SecretaryPending[] }

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
  /** مجموع التقييمات اليدوية بس (خصومات/مكافآت المدير) */
  kpiPoints: number
  kpiPointsValue: number
  /** النقاط الكاملة المحسوبة من الستة عناصر — هذا الرقم الي يفصّله المخطط */
  smartKpiPoints: number
  // نسبة زمن الموظف للمتوسط العام بنفس المنظومة ونوع الشغل:
  // فوق ١ أسرع، تحت ١ أبطأ. null يعني ماكو عيّنات كافية.
  workSpeedScore: number | null
  workSpeedSamples: number
  vehicleCleanlinessScore: number | null
  vehicleRatingsCount: number
  complaintsCount: number
  salesCount: number
  completedBookingsCount: number
  /** المبالغ الي دخّلها فعلاً — صافي فواتير الليدر الي رفعها هو بالشهر */
  revenueBrought: number
  /** راتبه الشهري — ⚠️ `null` يعني ماكو راتب مكتوب، **مو راتب صفر** */
  salary: number | null
  /**
   * نسبة تغطية الراتب = الإيراد ÷ الراتب × ١٠٠.
   * ⚠️ `null` لمّا الراتب فاضي أو صفر — تُعرض «—» مو ٠٪.
   * وقرار (ع): **١٠٠٪ تكفي** (يعبر راتبه)، مو أضعافاً.
   */
  salaryCoverage: number | null
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


export interface ExecutionCostItem {
  systemName: string
  itemName: string
  count: number
  /** سعر الوحدة لبنود «الإضافات» بفاتورة الخدمة */
  unitPrice?: number
  heightMeters: number
  wiringItemName?: string
  // ارتفاع التسليك — قاعدته ثنائية بالاكسل (٥م فما فوق = ضعف السعر)، غير
  // متدرجة مثل ارتفاع التركيب
  wiringHeightMeters?: number
  cableLengthMeters?: number
  programmingItem?: string
  /** ملاحظة قصيرة للبند — ما تدخل بالحساب، بس تفرق بالتنفيذ والتدقيق */
  notes?: string
}

/** ═══ وين هذا الحجز؟ ═══
 *  يرجّع **مكاناً** مو بيانات — الوصول للتفاصيل يبقى بصلاحية الشاشة. */
export interface LocateResult {
  found: boolean
  code?: string
  where?: string
  route?: string
  hint?: string
  invoiceStage?: string
  invoiceRoute?: string
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
  /** فاتورة الخدمة (الداش كام): وقت بداية ونهاية الشغل من الحجز — دازاي 10-31 */
  serviceStartAt?: string | null
  serviceEndAt?: string | null
  isFree?: boolean
  freeReasonId?: string | null
  freeReasonNote?: string | null
  freeReasonLabel?: string | null
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
  /** CATALOG = السعر من جدول الكلفة · MANUAL = كتبه صاحب الصلاحية بإيده. */
  pricingMode?: 'CATALOG' | 'MANUAL'
  /** شنو انعمل للزبون — يجي بالفاتورة اليدوية بس، وهو مرجع تدقيقها. */
  manualWork?: string
  manualPriceNote?: string
  status: string
  createdAt: string
  approvedByEmployeeId: string | null
  approvedAt: string | null
  systems: string[]
  items: ExecutionCostItem[]
  materials: LeaderInvoiceMaterialItem[]
  // ═══ التدقيق ═══
  /** MATCHED مطابق · MISMATCH غير مطابق · PRICE_ERROR خطأ بالسعر · FREE صيانة مجانية */
  auditVerdict?: 'MATCHED' | 'MISMATCH' | 'PRICE_ERROR' | 'FREE' | 'CREDIT'
  auditNote?: string
  auditedByName?: string
  auditedAt?: string
  /** المبلغ الي دخل فعلاً — أساس المطابقة */
  auditedAmount?: number
  // ═══ مراجعة المراقب ═══
  //
  // ⚠️ الحالة تنشتق من الطوابع: `monitorRequestedAt` موجود و
  // `monitorDecidedAt` فارغ = الفاتورة **عند المراقب الآن**.
  monitorRequestedAt?: string
  monitorRequestedByName?: string
  monitorRequestNote?: string
  monitorDecidedAt?: string
  monitorDecidedByName?: string
  monitorVerdict?: 'OK' | 'FLAGGED'
  monitorNote?: string
  // ═══ إرجاع المالك للمحاسب ═══
  returnedAt?: string
  returnedByName?: string
  returnReason?: string
  returnedCount?: number
  // ═══ سحب الاعتماد ═══
  revokedAt?: string
  revokedByName?: string
  revokeReason?: string
  revokedCount?: number
}

export interface CreditSale {
  id: string; kind: 'BOOKING' | 'PROJECT' | 'OTHER'; bookingId: string | null; bookingCode: string | null; projectId: string | null; projectCode: string | null
  title: string; customerName: string | null; phone: string | null; amount: number; paid: number; note: string | null; dueDate: string | null
  createdByName: string | null; createdAt: string; settledAt: string | null; lastPaidAt: string | null
}

export type BookingDeleteChannel = 'WEBSITE' | 'MOBILE_APP' | 'CALL_CENTER'
export type BookingDeleteRequestType = 'RECURRING_DUPLICATE' | 'CUSTOMER_CANCEL' | 'DATA_CORRECTION' | 'CUSTOMER_INQUIRY'

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
  /** رقم الزبون — الي يوافق على الحذف يتصل يتأكد إنه فعلاً ألغى */
  customerPhone: string
  /** الخدمة الي طلبها الزبون (والإضافية معها إذا الحجز متعدد) */
  serviceNames: string
  /** موعد الحجز — يفرّق بين حجز باچر وحجز فات */
  bookingScheduledAt: string | null
  bookingStatus: string
  requestedByName: string
  decidedByName: string | null
  channel: BookingDeleteChannel | null
  requestType: BookingDeleteRequestType | null
  channelLabel: string
  requestTypeLabel: string
  needsInfo: boolean
  needsInfoNote: string | null
  needsInfoAt: string | null
  needsInfoByName: string | null
}

export interface BookingDeleteRequestCounts {
  approved: number
  needsInfo: number
  awaitingReview: number
  rejected: number
  total: number
}

/** تنسيق الحجوزات: تنبيه تقصير الإداري بتثبيت الحجز. */
export interface CoordinationAlert {
  id: string
  bookingId: string
  coordinatorName: string | null
  reason: string | null
  byName: string | null
  createdAt: string
  resolvedAt: string | null
  resolvedByName: string | null
  resolveNote: string | null
}

export interface CoordinationAlertSummary {
  bookingId: string
  openCount: number
  totalCount: number
  lastAlertAt: string | null
}

/** التدقيق اليومي: حجوزات يوم واحد بمجاميعه الأربعة */
export type ItAssetKind = 'COMPUTER' | 'LAPTOP' | 'SERVER' | 'NETWORK' | 'PRINTER' | 'CAMERA' | 'PHONE' | 'OTHER'
export type ItAssetStatus = 'ACTIVE' | 'REPAIR' | 'SPARE' | 'RETIRED'
export interface ItAsset {
  id: string
  name: string
  kind: ItAssetKind
  status: ItAssetStatus
  brand: string | null
  model: string | null
  serialNumber: string | null
  ipAddress: string | null
  location: string | null
  assignedEmployeeId: string | null
  assignedEmployeeName: string | null
  purchaseDate: string | null
  warrantyUntil: string | null
  notes: string | null
  createdAt: string
  updatedAt: string
  lastLogAt: string | null
}
export type ItAssetInput = Omit<ItAsset, 'id' | 'assignedEmployeeName' | 'createdAt' | 'updatedAt' | 'lastLogAt'>
export interface ItAssetLog {
  id: string
  assetId: string
  kind: string
  note: string
  cost: number | null
  employeeName: string | null
  assetName?: string | null
  createdAt: string
}
export interface ItStats {
  total: number
  byKind: Record<string, number>
  byStatus: Record<string, number>
  warrantySoon: ItAsset[]
  inRepair: ItAsset[]
  repairCost90d: number
  repairCount90d: number
  unassigned: number
  recentLogs: ItAssetLog[]
}

export type SurveyPhotoOwner = { bookingId: string; projectId?: undefined } | { projectId: string; bookingId?: undefined }

export interface SurveyPhoto {
  id: string
  bookingId: string | null
  projectId: string | null
  contentType: string
  fileName: string
  uploadedById: string
  uploadedByName: string
  createdAt: string
}

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
  /** رقم الفاتورة بالنظام المحاسبي الثاني — يدخله المحاسب وقت الاعتماد */
  invoiceExternalNumber: string | null
  /** فاتورة الحجز مؤشَّرة مجانية (ضمان/إعادة عمل/…) وسببها */
  invoiceIsFree: boolean
  invoiceFreeReason: string | null
  expectedAmount: number
  openIssues: number
  /** القصة: منو الليدر المسؤول ومنو الإداري الي أكّد الحجز */
  leaderName: string | null
  leaderPhone: string | null
  confirmedByName: string | null
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
  // الليدر صاحب فاتورة الحجز — للمراقب حتى يتأكد منه ليش عنده أخطاء
  leaderName?: string | null
  // منو أغلق البلاغ وليش وشنو سوّى
  resolvedByName?: string | null
  resolveReason?: string | null
  actionKind?: AuditCloseAction | null
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

/** ═══ صندوق المراقب ═══
 *  كل حدث لازم عين المراقب عليه ينوصل هنا كصف بانتظار قرار. */
export type MonitorStage =
  | 'INVOICE_BEFORE_AUDIT' | 'INVOICE_AFTER_AUDIT' | 'INVOICE_ADJUSTED'
  | 'BOOKING_BEFORE_CONFIRM' | 'BOOKING_AFTER_CONFIRM' | 'BOOKING_AFTER_COMPLETE'
  | 'PROCUREMENT_FULFILLED' | 'QUALITY_VERDICT' | 'GPS_DEVICE_DONE'
  | 'SOLAR_QUOTED' | 'AI_VERDICT'

/** ═══ سلال مراحل الحجز ═══
 *  «هاي تصير بيها حالتين — قبل التثبيت وبعد التثبيت»: زبون ألغى قبل ما
 *  نثبتله موعد شي، وزبون ألغى بعد ما وعدناه وحضّرنا كادر شي ثاني. */
export type StageBucket =
  | 'POSTPONED_AFTER_CONFIRM'
  | 'NO_ANSWER_BEFORE_CONFIRM'
  | 'NO_ANSWER_AFTER_CONFIRM'
  | 'CANCELLED_BEFORE_CONFIRM'
  | 'CANCELLED_AFTER_CONFIRM'
  // «الزبون يستفسر وبعدين يرجع خبر» — ننتظر قراره، مو إنه ما رد
  | 'AWAITING_CUSTOMER_DECISION'

/** ═══ نواة الذكاء الاصطناعي ═══
 *  الأدلة حقائق نحسبها من الجداول، والحكم تفسير. الفصل مقصود:
 *  حكم بلا أدلة = تخمين، والتخمين هنا يتحول لظلم موظف. */
export interface AiVerdict {
  id: string
  signalId: string
  source: 'RULES' | 'MODEL'
  modelName?: string
  headline: string
  reasoning?: string
  confidence: number
  severity: 'INFO' | 'WATCH' | 'WARN' | 'CRITICAL'
  blameEmployeeId?: string
  blameEmployeeName?: string
  suggestion?: string
  createdAt: string
}

export interface AiEvidence {
  id: string
  signalId: string
  facts: Record<string, unknown>
  /** الي ما قدرنا نجمعه — ينعرض مو ينخبّى */
  gaps: string[]
  collectedAt: string
}

export interface AiSignal {
  id: string
  kind: string
  entityType: string
  entityId: string
  employeeId?: string
  employeeName?: string
  status: 'PENDING' | 'COLLECTED' | 'ANALYZED' | 'SKIPPED'
  occurredAt: string
  createdAt: string
  evidence?: AiEvidence
  verdict?: AiVerdict
}

export interface AiCatalogItem {
  key: string
  label: string
  ready: boolean
  detail: string
}

export interface AiCatalog {
  signals: AiCatalogItem[]
  metrics: AiCatalogItem[]
  judge: string
  platformLinked: boolean
}

export interface AiMetric {
  id: string
  metricKey: string
  scope: string
  scopeId?: string
  periodStart: string
  periodEnd: string
  value: number
  /** عدد العيّنات — «٥٠٪» من عيّنتين مو مثل «٥٠٪» من مئتين */
  sampleCount: number
  computedAt: string
  /** خلايا الاستكشاف الأسبوعي بس (scope غير COMPANY) — axisLabel/valueLabel/overallAvg/narration */
  details?: { axisLabel?: string; valueLabel?: string; overallAvg?: number; narration?: string }
}


/** ═══ المهام الإضافية ═══
 *  شغل موجّه من المدير لموظف، مو مربوط بحجز: «خرّج فواتير الشهر»،
 *  «رتّب المخزن». تطلع عند الموظف الي انتوجّهت له. */
export interface ExtraTask {
  clause?: number | null
  id: string
  title: string
  description?: string
  assignedToId: string
  assignedToName?: string
  assignedById?: string
  assignedByName?: string
  priority: 'NORMAL' | 'URGENT'
  dueAt?: string
  status: 'NEW' | 'IN_PROGRESS' | 'DONE' | 'CANCELLED'
  seenAt?: string
  startedAt?: string
  doneAt?: string
  doneNote?: string
  cancelledAt?: string
  cancelReason?: string
  createdAt: string
  /** محسوبة بالسيرفر: فات موعدها وما انخلصت */
  overdue: boolean
}

/** ═══ الخط الزمني للحجز ═══
 *  قصة الحجز من سبعة جداول + التأخيرات الستة. ماكو بيانات جديدة —
 *  كلها موجودة بس متفرقة. */
export interface TimelineEvent {
  at: string
  kind: string
  title: string
  detail?: string
  actor?: string
}

export interface DelayMetric {
  key: string
  label: string
  /** null = ما ينطبق (مو صفر) — الواجهة تخفيه كلياً */
  minutes: number | null
  owner: string
  thresholdMinutes: number
  breached: boolean
}

// ═══ ماتركس ٢٠٥٠ — سلسلة الحجز ═══
export type ChainStatus = 'OK' | 'LATE' | 'MISSED' | 'WAITING' | 'ISSUE' | 'NA'
export interface ChainStation {
  no: number; key: string; title: string; role: string; roleTitle: string
  owners: { id: string; name: string; status?: ChainStatus }[]
  startAt: string | null; endAt: string | null; minutes: number | null; threshold: number | null
  status: ChainStatus; verdict: string; issues: string[]; facts: string[]
}
export interface CrewRatingRow { bookingId: string; bookingCode: string; leaderId: string; leaderName: string; technicianId: string; score: number; note: string | null; createdAt: string }
export interface BookingChain {
  bookingId: string; code: string; service: string; solo: boolean; legacy: boolean; project?: boolean; trail?: string[]
  stations: ChainStation[]; summary: string
  score: { ok: number; late: number; missed: number; issue: number; waiting: number }
  ratings: CrewRatingRow[] | null
}
export interface ChainStationStat { key: string; title: string; count: number; ok: number; late: number; missed: number; issue: number; medianMin: number | null; avgMin: number | null; teamMedianMin: number | null }
export interface ChainEmployeeReport {
  id: string; name: string; total: number; onTime: number; late: number; missed: number; issues: number
  onTimePct: number; prevPct: number | null; stations: ChainStationStat[]; worst: string
  bad: { id: string; code: string; station: string; status: ChainStatus; note: string; trail?: string[]
    who?: { createdBy: string; createdAt: string; contactedBy: string; contactedAt: string; tries: number; confirmedBy: string; confirmedAt: string; scheduledAt: string; leader: string; status: string; explain: string } }[] | null
  lateWeekday: string; lateHour: number | null; ratingAvg: number | null; ratingCount: number; ratingNotes: string[]; insights: string[]
}
export interface RoleChainReport {
  role: string; title: string; days: number; bookings: number; teamPct: number; prevPct: number | null
  stations: ChainStationStat[]; employees: ChainEmployeeReport[]; insights: string[]
}
export interface ChainLearning {
  stations: { key: string; title: string; samples: number; medianMin: number; limitMin: number; source: string }[]
  servicesOwnLimits: number; learnedAt: string
}
export interface ProjectVerdict {
  id: string; code: string; name: string; stage: string; workType: string | null; deliveryDate: string | null
  ownerId: string; ownerName: string; daysInStage: number; stageLimit: number; idleDays: number
  status: ChainStatus; problems: string[]; verdict: string; touchers: number; lastActivityBy: string | null; bookingCode: string | null
}
export interface ProjectsReport {
  projects: ProjectVerdict[]
  people: { id: string; name: string; owned: number; open: number; stuck: number; done: number; avgIdle: number; insights: string[] }[]
  stageMedianDays: Record<string, number>; insights: string[]
}
export interface ProjectChain {
  verdict: ProjectVerdict
  touchers: { employeeId: string; name: string; role: string; firstAt: string; lastAt: string; actions: number }[]
  delegations: { employeeId: string; name: string | null; action: string; by: string | null; at: string }[]
  order: string[]
}
export interface MatrixSuggestedPerson { id: string; name: string; why: string[]; score: number; isLeader: boolean }
export interface MatrixBookingSuggestion {
  bookingId: string; code: string
  schedule: { at: string; durationMin: number; freeLeaders: number } | null; scheduleWhy: string[]
  crew: { leader: MatrixSuggestedPerson | null; techs: MatrixSuggestedPerson[]; altLeaders: MatrixSuggestedPerson[]; altTechs: MatrixSuggestedPerson[]; for: string; crewSize: number; solo: boolean } | null
  crewWhy: string[]; note: string
}
export interface MatrixSuggestAccuracy {
  days: number; pending: number; insights: string[]
  kinds: { kind: string; title: string; decided: number; accepted: number; partial: number; changed: number; acceptPct: number; onTimeAccepted: number | null; onTimeChanged: number | null; sampleAccepted: number; sampleChanged: number }[]
  recent: { bookingId: string; code: string; kind: string; outcome: string; at: string }[]
}
export interface MatrixAutonomyStatus {
  autopilotOn: boolean
  autoCrew: { on: boolean; decided: number; accepted: number; pct: number; eligible: boolean; needDecided: number; needPct: number }
  monitorEscalate: { on: boolean; overdue: number }
  recent: { id: string; kind: string; summary: string; status: string; createdAt: string; details: unknown }[]
}
export interface AIUsageReport {
  today: number; month: number; last30: number; forecast: number
  features: { feature: string; label: string; calls: number; tokens: number; costToday: number; cost30: number }[]
  switches: Record<string, string>
  days: { day: string; cost: number }[]
}
export interface MatrixEyeItem { eye: string; kind: string; severity: 'HIGH' | 'MEDIUM' | 'LOW'; title: string; detail: string; advice: string; ownerId?: string; ownerName?: string; bookingId?: string; at?: string }
export interface MatrixEyesReport {
  eyes: { key: string; title: string; items: MatrixEyeItem[]; summary: string; high: number }[]
  customers: { employeeId: string; name: string; followUps: number; happy: number; unhappy: number; complaints: number; complaints30: number; avgRating: number | null; happyPct: number | null; notes: string[] }[]
  teamHappyPct: number | null; insights: string[]
}
export interface MediaBrief {
  id: string; projectId: string; projectCode: string; projectName: string; workType: string | null; stage: string
  location: string | null; locationUrl: string | null; lat: number | null; lng: number | null; deliveryDate: string | null; projectTask: string | null
  engineerId: string | null; engineerName: string | null; engineerPhone: string | null
  startAt: string | null; expectedEndAt: string | null; duration: string | null; notes: string | null
  status: 'NEW' | 'SCHEDULED' | 'SHOT' | 'PUBLISHED' | 'CANCELLED'; shootAt: string | null
  customerRep: string | null; customerPhone: string | null
  mediaEmployeeName: string | null; publishedUrl: string | null; mediaNotes: string | null; createdByName: string | null; createdAt: string
}
export interface DisciplineEntry {
  source: 'KPI' | 'POINTS' | 'LEDGER'; id: string; employeeId: string; employeeName: string; amount: number; points: number
  reason: string; kind: string | null; byName: string | null; bookingId: string | null; bookingCode: string | null
  at: string; returned: boolean; returnedAt: string | null; returnNote: string | null
}
export interface DisciplineRecord {
  employees: { employeeId: string; employeeName: string; count: number; amount: number; returned: number; net: number; pointsLost: number; pointsBack: number; thisMonth: number; lastMonth: number; topReason: string; last: string | null; insights: string[] }[]
  entries: DisciplineEntry[]; totals: Record<string, number>; insights: string[]
}
export interface PeerNote {
  id: string; authorId: string; authorName: string; targetId: string; targetName: string; score: number; kind: 'GOOD' | 'NOTE' | 'PROBLEM'
  text: string | null; needsFollowup: boolean; fWhat: string | null; fWhy: string | null; fWish: string | null; suggestion: string | null
  accepted: boolean | null; acceptNote: string | null; severity: 'NORMAL' | 'ATTENTION' | 'URGENT'; reportedAt: string | null; createdAt: string
}
export interface WorkplaceIssue {
  id: string; title: string; partyAId: string; partyAName: string; partyBId: string | null; partyBName: string | null; openedByName: string | null
  description: string; severity: 'NORMAL' | 'ATTENTION' | 'URGENT'; status: 'OPEN' | 'IN_PROGRESS' | 'RESOLVED'; aiSummary: string | null
  decision: string | null; decidedByName: string | null; decidedAt: string | null; followUpAt: string | null; createdAt: string
}
export interface WorkplaceIssueDetail extends WorkplaceIssue {
  training?: { id: string; employeeId: string; verdict: 'NEEDS' | 'NO_NEED' | 'UNKNOWN'; areas: string[]; evidence: string; baselinePct: number | null; followUpAt: string; status: string; currentPct: number | null; resultNote: string | null }[]
  entries: { id: string; authorName: string | null; kind: string; text: string; createdAt: string }[]
  context: { sharedBookings: number; aComplaints: number; bComplaints: number; aPeerAvg: number | null; bPeerAvg: number | null }
}
export interface MyPeerState {
  open: boolean; week: string
  checkin: { id: string; mood: number | null; worry: string | null; suggestion: string | null; nothing: boolean } | null
  notes: PeerNote[]; colleagues: { id: string; name: string; worked: boolean }[]; followups: WorkplaceIssue[]
}
export interface PeerVoiceReport {
  full: boolean; insights: string[]
  mood: { week: string; avg: number | null; count: number }[]
  scores: { employeeId: string; name: string; avg: number; count: number; problems: number; prevAvg: number | null }[]
  tensions: { aId: string; aName: string; bId: string; bName: string; aToB: number; bToA: number }[]
  notes: PeerNote[]
  checkins: { id: string; employeeId: string; name: string; mood: number | null; worry: string | null; suggestion: string | null; urgent: boolean; createdAt: string }[]
}
export interface PendingCrewRating { bookingId: string; bookingCode: string; completedAt: string; techs: { id: string; name: string }[] }

export interface BookingTimeline {
  bookingId: string
  code: string
  events: TimelineEvent[]
  delays: DelayMetric[]
}

/** عدّاد حي لطوابير مكتب المراقب الخمسة. */
export interface MonitorDeskCounts {
  inbox: number
  issues: number
  invoices: number
  quality: number
  crew: number
}

/** ═══ معرض التصاميم ═══
 *  معرض **مستقل** ما ينربط بحجز — المصممة ما تشوف الحجوزات. */
export interface DesignAsset {
  id: string
  title: string
  category: string
  notes: string | null
  /** مفتاح التخزين (مسار /api/files/…) — الصورة مو بقاعدة البيانات. */
  fileKey: string
  fileType: string | null
  uploadedById: string | null
  uploadedByName: string
  createdAt: string
  archivedAt: string | null
}

/** ═══ سجل الأقسام ومسؤوليها — للحجز داخل الشركة ═══ */
export interface DepartmentHead {
  id: string
  departmentId: string
  name: string
  phone: string | null
  employeeId: string | null
  active: boolean
  createdAt: string
}
export interface Department {
  id: string
  name: string
  active: boolean
  createdAt: string
  /** أكثر من مسؤول للقسم الواحد — كلهم يكدرون يطلبون حجزاً له. */
  heads?: DepartmentHead[]
}

export interface MonitorReview {
  id: string
  stage: MonitorStage
  entityType: 'BOOKING' | 'LEADER_INVOICE' | 'INVOICE_ADJUSTMENT' | 'PROCUREMENT' | 'QUALITY_FOLLOW_UP' | 'GPS_DEVICE' | 'AI_VERDICT'
  entityId: string
  title: string
  summary: string | null
  ownerRole: string | null
  ownerEmployeeId: string | null
  status: 'PENDING' | 'OK' | 'FLAGGED'
  note: string | null
  reviewedAt: string | null
  createdAt: string
  /** صف عاجل — «غير مطابق» بالفواتير. يطلع بأول الصندوق بشارة حمرا. */
  urgent: boolean
  ownerEmployee?: { id: string; name: string }
  reviewedBy?: { id: string; name: string }
  /** هوية الحجز وراء الصف — تنبني بالسيرفر. غايبة لو الصف ما إله حجز. */
  identity?: {
    bookingId: string
    bookingCode: string
    customerCode: number
    customerName: string
    customerPhone: string | null
    address: string | null
    leaderName: string | null
    /** رقم الفاتورة المحاسبية — لصفوف الفواتير وتعديلاتها بس */
    externalInvoiceNumber?: string
  }
  /** (ع 10-16) الفاتورة والحجز بكل تفاصيلهم — يتعبّى بالخادم */
  details?: {
    invoice?: {
      invoiceId: string; leaderName: string | null; executionCost: number; materialsTotal: number; discountValue: number
      netTotal: number; isFree: boolean; status: string; collected: number | null; externalNo: string | null; createdAt: string
      adjustments: { oldNetTotal: number; newNetTotal: number; oldExecutionCost: number; newExecutionCost: number; oldMaterialsTotal: number; newMaterialsTotal: number; reason: string; byName: string | null; createdAt: string }[]
    }
    booking?: {
      status: string; station: string; service: string | null; createdByName: string | null; createdAt: string
      contactedAt: string | null; confirmedAt: string | null; confirmedByName: string | null; scheduledAt: string | null
      shift: string | null; arrivedAt: string | null; startedAt: string | null; completedAt: string | null
      leaderName: string | null; crew: string | null; collected: number | null; quality: string | null
    }
  }
  // تفاصيل «تأخر بالخروج للزبون» (أحكام ماتركس)
  late?: { employeeId: string | null; employeeName: string | null; bookingId: string | null; bookingCode: string | null; scheduledAt: string | null; departedAt: string | null; minutesLate: number | null; threshold: number | null; count30d: number | null; reason: string | null
    /** «صارت قبل؟» — مرات تأخر نفس الموظف السابقة (٩٠ يوم، الأحدث أول). */
    history?: { occurredAt: string; bookingId: string | null; bookingCode: string | null; minutesLate: number | null; reason: string | null; review: 'PENDING' | 'OK' | 'FLAGGED' | null; reviewNote: string | null }[]
  }
}

/** ═══ تسعيرة الشبكات ═══
 *  الأسعار بقاعدة البيانات مو بالكود — المالك يعدّلها من الشاشة. */
export interface NetworkBracket {
  upTo: number      // 0 = «وأكثر»
  unitPrice: number
}

export interface NetworkPriceItem {
  id: string
  label: string
  unit: string
  pricingMode: 'FLAT' | 'TIERED' | 'BRACKET'
  basePrice: number
  includedQty: number
  extraPerUnit: number
  brackets: NetworkBracket[]
  note: string | null
  sortOrder: number
  active: boolean
}

export interface NetworkCostLineResult {
  itemId: string
  label: string
  unit: string
  quantity: number
  basePart: number
  extraQty: number
  extraPart: number
  unitPrice: number
  total: number
}

export interface NetworkCostResponse {
  lines: NetworkCostLineResult[]
  subtotal: number
  discount: number
  finalAmount: number
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
  ipCameraChangeCount: number
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

export interface DecorRate { id: string; serviceId: string; serviceName: string; name: string; unit: string; price: number; active: boolean }

export interface CreateLeaderInvoiceRequest {
  bookingId?: string
  customerName?: string
  customerPhone?: string
  customerAddress?: string
  systems: string[]
  items: ExecutionCostItem[]
  materials: CreateMaterialLineRequest[]
  discountValue: number
  // فاتورة الديكور (ع 10-16)
  decorLines?: { rateId: string; qty: number; notes?: string }[]
  /** الشغل المجاني: الصافي صفر بسبب من القائمة، والكلفة الحقيقية تبقى مسجّلة */
  isFree?: boolean
  freeReasonId?: string
  freeReasonNote?: string
}

/** خلاصة عيّنات سعر خدمة — الرقم مع عدد عيّناته دائماً */
export interface ServicePriceStats {
  serviceId: string
  serviceName: string
  sampleCount: number
  avgAmount: number
  minAmount: number
  maxAmount: number
  /** كم عيّنة باقية حتى يطلع الاقتراح — صفر يعني جاهز */
  remaining: number
}

/** اقتراح سعر من النظام — ما يصير سعراً رسمياً إلا باعتماد المالك */
export interface ServicePriceSuggestion {
  id: string
  serviceId: string
  serviceName: string | null
  avgAmount: number
  sampleCount: number
  minAmount: number
  maxAmount: number
  source: string
  status: 'PROPOSED' | 'REVIEWED' | 'APPROVED' | 'REJECTED'
  reviewedName: string | null
  reviewedAt: string | null
  reviewNote: string | null
  decidedName: string | null
  decidedAt: string | null
  createdAt: string
}

export interface QuotationItem {
  id?: string
  /** مرجع صورة المنتج وقت إصدار العرض — مرّرها عبر fileUrl() قبل الـsrc */
  imageBase64?: string | null
  productName: string
  unit: string
  quantity: number
  unitPrice: number
  totalPrice: number
  /** ترتيب البند بالعرض — محله بالقائمة المحفوظة */
  sortIndex?: number
}

export interface Quotation {
  id: string
  quotationNumber: string
  projectId?: string | null
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
  terms?: string[] | null
  status: 'NEW' | 'SENT' | 'ACCEPTED' | 'REJECTED'
  createdAt: string
}

/** ═══ نسخة مؤرشفة من عرض سعر ═══
 *  `snapshot` لقطة كاملة للعرض قبل التعديل — تجي **مرمّزة base64**
 *  لأن الخادم يخزّنها JSONB خاماً (`[]byte` بـGo). استعمل
 *  `parseQuotationSnapshot` لفكّها، ما تقراها مباشرة. */
export interface QuotationVersion {
  id: string
  quotationId: string
  version: number
  snapshot: string
  archivedByName: string | null
  archivedAt: string
}

/** سطر بسجل تعديلات المحاسب على فاتورة — قبل وبعد */
export interface LeaderInvoiceAdjustment {
  id: string
  invoiceId: string
  oldExecutionCost: number; newExecutionCost: number
  oldMaterialsTotal: number; newMaterialsTotal: number
  oldDiscountValue: number; newDiscountValue: number
  oldNetTotal: number; newNetTotal: number
  reason: string
  adjustedByName: string | null
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
  /** ⚠️ الخادم ما يرسلها لمن ماكو ركّاب — لازم ?? [] */
  passengers?: VehicleMissionPassenger[]
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

/** إحصاء غسل سيارة واحدة بشهر واحد */
export interface VehicleWashMonthly {
  vehicleId: string
  vehicleName: string
  plateNumber: string
  /** YYYY-MM */
  month: string
  /** عدد أيام الغسل (فيها غاسل مسمّى) */
  washCount: number
  /** متوسط درجة الغسل ٠–٤ — null لو ماكو ولا درجة. ⚠️ null مو صفر. */
  avgQuality: number | null
  washers: VehicleWasherCount[]
}

export interface VehicleWasherCount {
  employeeId: string
  employeeName: string
  times: number
  totalPoints: number
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
  /** كل الي انكلّف بيه (مو الملغى) — أساس معدل الإنجاز. */
  assignedBookings: number
  /** أيام حضوره بالفترة — أساس الالتزام بالدوام. */
  attendedDays: number
  /** المبيعات: الحجوزات الي سجّلها بالفترة. */
  createdBookings?: number
  /** المهام الإضافية الي خلّصها بالفترة. */
  extraTasksDone?: number
  /** المراقب: التدقيقات (صندوق المراقب + تقييمات التدقيق + الشكاوى). */
  monitorAudits?: number
  /** المراقب: تدقيق الفواتير (إضافي). */
  invoiceAudits?: number
  /** فرق النقاط عن الفترة السابقة. */
  pointsDelta: number
  /** فرق الترتيب — موجب = تقدّم (٣ → ١ يرجّع +2). */
  rankDelta: number
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
  cancelNote?: string | null // MONTHLY_RESET = رجع بالتصفير الشهري
  createdAt: string
}

export interface KpiCriterion {
  clause?: number | null
  id: string
  label: string
  createdAt: string
}

export interface EmployeeLedger {
  period: 'day' | 'month'; from: string; to: string; total: number
  sections: { key: string; title: string; count: number; items: { at: string; text: string; link?: string | null }[] }[]
}

export interface WorkReport {
  id: string
  // تدقيق المراقب (ع 10-16)
  reviewedAt?: string | null
  reviewNote?: string | null
  reviewedByName?: string | null
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
  /** (ع 10-16) نموذج «تقرير تنفيذ المشروع» — null للتقارير القديمة */
  sheet?: unknown
  employee: { id: string; name: string } | null
  booking: { id: string; code: string; customerName: string; customerPhone?: string | null; customerCode?: number; address?: string | null; leaderName?: string | null } | null
}

export interface TechnicianKpi {
  employeeId: string
  employeeName: string
  period: string
  breakdown: {
    completedBookings: { count: number; points: number }
    completionSpeed: { avgMinutes: number; points: number; entries?: number; basis?: 'MISSION' | 'BOOKING' }
    workReports: {
      count: number; fullReports: number; points: number
      // تفسير الفرق بين التقارير والحجوزات (ع 10-23)
      bookings?: number; extraVisits?: number; multiBookings?: number; outsideMonth?: number; noReport?: number
    }
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

/** ═══ حالة جرد موظف واحد بحجز معيّن ═══
 *  فاضي `checkedAt` = ما جرد لهذا الحجز بعد. */
export interface BookingCrewInventoryState {
  employeeId: string
  name: string
  position: string | null
  isLeader: boolean
  checkedAt: string | null
  complete: boolean | null
  missingItems: string | null
}

export interface InventoryCheck {
  id: string
  employeeId: string
  complete: boolean
  missingItems: string | null
  checkedAt: string
  /** الحجز الي انجرد قبله — فاضي بالجرود العامة والقديمة */
  bookingId: string | null
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

/** استثناء أداة من نواقص موظف بعينه — الأداة تبقى بالقالب لغيره */
export interface PersonalToolExemption {
  id: string
  employeeId: string
  toolName: string
  note: string | null
  byEmployeeId: string | null
  byName: string
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
  // رقم المبدأ (١-٢٠) لنقاط «المبادئ العشرين» — null للنقاط العادية
  principleNo?: number | null
}

export interface PrivacyPolicyStatus {
  needsAcceptance: boolean
  points: PrivacyPolicyPoint[]
  acceptedAt: string | null
  // وافق گبل بس النص تغيّر بعدها
  changed?: boolean
  // بصمة النص المعروض — ترجع ويا الموافقة حتى ما يوافق على نسخة ما قراها
  hash?: string
}

export type PersonalToolStatus = 'AVAILABLE' | 'LOST' | 'DAMAGED' | 'REPAIRING' | 'RETIRED' | 'CHECKED_OUT'

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

export type ComplaintType = 'DELAY' | 'DISORGANIZED' | 'TECHNICAL' | 'EXECUTION_ERROR' | 'INCOMPLETE' | 'OTHER'

export interface Complaint {
  id: string
  // (ع) 10-15: شكوى من شخصية مهمة — 👑 ومهلة ٢٤ ساعة
  isVip?: boolean
  vipNote?: string | null
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
  /** تقييم الزبون ١..٥ — null يعني «ما انسأل» مو «صفر» */
  customerRating: number | null
  auditVerdict: AuditVerdict | null
  auditNote: string | null
  auditedAt: string | null
  auditedBy: { id: string; name: string } | null
  contactedByName: string | null
  /** «الزبون ما رد» (ع 10-19): ٣ محاولات والشكوى تسقط */
  noAnswerCount?: number
  noAnswerAt?: string | null
  nextCallAt?: string | null
  droppedNoAnswerAt?: string | null
  notes: string | null
  createdAt: string
  resolvedAt: string | null
}

export type ProcurementRequestType = 'PERSONAL_SUPPLY' | 'CUSTOMER_PRODUCT' | 'MANUAL_SUPPLY'

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
/** نبض يوم الإداري — الأرقام الي تطلع بالشاشة الرئيسية. */
export interface TodayPulse {
  /** المجدولة اليوم فعلاً — مو مجموع الجدول كله. */
  todayBookings: number
  /** أمس للمقارنة: رقم بلا مرجع ما يگول شي. */
  yesterdayBookings: number
  /** مهام ميدانية لسه شغّالة (مو منجزة ولا متوقفة). */
  openMissions: number
  newComplaints: number
  /** مثبّتة بس بلا موعد أو بلا كادر — شغل ينتظر الإداري. */
  needsCoordination: number
  /** ليدرات بالميدان هسه (طلعوا وما خلصوا). */
  crewInField: number
  /** مهام مفتوحة من أكثر من ٢٤ ساعة. */
  overdueMissions: number
}

export interface FinanceSummary {
  completedCount: number
  unverifiedCount: number
  verifiedCount: number
  todayCompleted: number
  /** انفتحن اليوم — محسوبة بالخادم بتوقيت بغداد */
  todayCreated: number
  pendingCount: number
  confirmedCount: number
  inProgressCount: number
  activeCrewCount: number
  totalCollected: number
  /** قرار (ع) 10-09: المحصّل من الحجوزات بس */
  bookingsCollected?: number
  /** دفعات المشاريع */
  projectsCollected?: number
  /** غير مدققة — بنفس نطاق totalCollected (COMPLETED فعلاً، شامل المقدم) */
  unverifiedAmount: number
  /** مدققة — بنفس نطاق totalCollected */
  verifiedAmount: number
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

export interface CustomerPage {
  rows: GpsCustomerListItem[]
  total: number
  all: number
  newThisMonth: number
  vip: number
  gpsCount: number
  locations: string[]
}

export interface LeaderInvoicePage {
  rows: LeaderInvoice[]
  total: number
  sum: number
  counts: Partial<Record<'INTERNAL' | 'AUDIT' | 'MONITOR' | 'REVIEWED' | 'PENDING' | 'NO_NUMBER' | 'APPROVED' | 'ALL', number>>
  leaders: string[]
  customers: string[]
  systems: string[]
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
  verified?: boolean
  verifyNote?: string | null
  verifyRows?: number
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
  /** (ع 10-16) نموذج «تقرير تنفيذ المشروع» */
  sheet?: unknown
}

export interface ProgressRecentRow {
  bookingId: string
  bookingCode: string
  customerName: string | null
  reports: number
  lastDay: number
  lastPercent: number
  lastAt: string
  leaderName: string | null
  unreviewed: number
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
  /** (ع 10-16) نموذج «تقرير تنفيذ المشروع» — null للتقارير القديمة */
  sheet?: unknown
}

export interface SurveyReportBody {
  customerWants: string
  siteAreaSqm?: number | null
  siteDetails?: string | null
  otherNotes?: string | null
}

export interface BookingSurveyReport {
  id: string
  bookingId: string
  employeeId: string
  customerWants: string
  siteAreaSqm: number | null
  siteDetails: string | null
  otherNotes: string | null
  createdAt: string
  employee?: { id: string; name: string } | null
}

export interface DuplicateBookingBrief {
  id: string
  code: string
  customerName: string
  customerPhone: string
  customerCode: number
  address: string | null
  serviceName: string | null
  scheduledAt: string | null
  status?: string
  createdAt?: string
  createdByName?: string | null
  started?: boolean
  hasInvoice?: boolean
}

export interface DuplicateCustomerBrief {
  id: string
  name: string
  phone: string
  customerCode: number
  createdAt?: string
  bookings?: number
}

export interface DuplicateCandidate {
  id: string
  kind: 'BOOKING' | 'CUSTOMER'
  matchReason: string
  status: 'PENDING' | 'DISMISSED' | 'RESOLVED'
  resolution?: 'MERGED' | 'DELETE_REQUESTED'
  resolutionNote?: string
  /** تحليل ماتركس: ليش تكرار، وشنو الأرجح صار، والمقترح. */
  analysis?: string
  /** حجز: معرّف المرشّح للحذف. زبون: «الأصلي|المكرر». */
  suggested?: string
  reviewedAt: string | null
  reviewedByName?: string | null
  detectedAt: string
  bookingA?: DuplicateBookingBrief | null
  bookingB?: DuplicateBookingBrief | null
  customerA?: DuplicateCustomerBrief | null
  customerB?: DuplicateCustomerBrief | null
}

export interface Achievement {
  id: string
  employeeId: string
  employeeName?: string | null
  role: string
  bookingId: string | null
  bookingCode?: string | null
  reportText: string
  reviewStatus: 'PENDING' | 'GOOD' | 'NEEDS_REVIEW'
  reviewNote: string | null
  reviewedById: string | null
  reviewedByName?: string | null
  reviewedAt: string | null
  createdAt: string
  suggestedStatus?: 'GOOD' | 'NEEDS_REVIEW'
  suggestedReply?: string
  voiceSeconds?: number | null
}

export interface VoiceAnalysis {
  achievementId: string; status: 'PENDING' | 'DONE' | 'FAILED'
  transcript: string | null; summary: string | null; tasks: { what: string; count: number }[]
  style: string | null; error: string | null; analyzedAt: string | null
}

export interface SuggestedCrewMember {
  employeeId: string
  name: string
  role: string
  daysWorked: number
  available: boolean
  note: string
}

export interface FreeWorkReason {
  id: string
  label: string
  sortOrder: number
  active: boolean
  /** يطلب توضيح مكتوب إضافي (مثل «سبب آخر») */
  needsNote: boolean
}

export interface EmployeeLetter {
  id: string
  employeeId: string
  addressedTo: string
  subject: string
  body: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  decisionNote: string | null
  decidedAt: string | null
  createdAt: string
  employee: { id: string; name: string } | null
  decidedBy: { id: string; name: string } | null
  employeeJobTitle: string | null
}

/** نتائج الجي بي اس كما يشوفها المراقب — عرض بلا تنفيذ */
export interface GpsMonitorSnapshot {
  expiring: GpsMonitorSubscription[]
  expired: GpsMonitorSubscription[]
  problems: GpsMonitorProblem[]
  /** نافذة «قربت تنتهي» بالأيام — تنكتب بالشاشة حتى يعرف شنو يعني «قربت» */
  expiringWindowDays: number
}

export interface GpsMonitorSubscription {
  deviceRequestId: string
  customerName: string
  customerPhone: string
  gpsNumber: string | null
  subscriptionEnd: string | null
  subscriptionType: string | null
  /** موجب = باقي أيام · سالب = فاتت أيام */
  daysLeft: number
}

export interface GpsMonitorProblem {
  id: string
  customerName: string
  customerPhone: string
  problemDescription: string
  status: string
  createdAt: string
  resolvedAt: string | null
  ageDays: number
}

// ═══ صندوق قرارات ماتركس (ADMIN/OWNER) ═══
export interface MatrixAction {
  id: string
  kind: string
  entityType: string
  entityId: string
  period: string
  targetEmployeeId: string | null
  targetLabel: string
  summary: string
  status: 'DONE' | 'UNDONE'
  createdAt: string
  undoneById: string | null
  undoneAt: string | null
  details: Record<string, unknown> | null
  resolvedAt: string | null
  escalatedAt: string | null
}
export interface MatrixPendingDecision {
  id: string
  entityType: string
  entityId: string
  title: string
  summary: string | null
  createdAt: string
  // منو وأي حجز (من الإشارة الأصلية)
  employeeId?: string | null
  employeeName?: string | null
  bookingId?: string | null
  bookingCode?: string | null
  occurredAt?: string | null
}
export interface MatrixDecisions {
  day: string
  autopilotEnabled: boolean
  done: MatrixAction[]
  pending: MatrixPendingDecision[]
  unstaffed: { id: string; code: string; scheduledAt: string }[]
  escalated: MatrixAction[]
  paused: { kind: string; reason: string; pausedAt: string }[]
  accuracy: {
    actions: number; resolved: number; escalated: number; rejected: number
    delayPredicted: number; delayChecked: number; delayCorrect: number
  }
  labels: Record<string, string>
  proposals: number
}

// عين ماتركس — حالة الموظف نفسه.
export interface MatrixWorkload { key: string; label: string; done: number; left: number; route: string; verb: string; codes?: string[]; links?: { label: string; to: string }[] }
export interface TcInsight {
  id: string; name: string; role: string; daysIn: number; jobs: number
  avgMinutes: number | null; teamMinutes: number | null; complaints: number; stops: number; partials: number
  deductions: number; skills: number; lastExamPct: number | null; score: number; reasons: string[]; strengths: string[]
}
export interface TcCurriculum { id: string; role: string; title: string; details: string | null; ord: number }
export interface TcQuestion { id: string; role: string; text: string; options: string[]; correct: number; approved: boolean; source: 'MANUAL' | 'AI' }
export interface TcExamReportRow { examId: string; role: string; month: string; title: string; passPct: number; employeeId: string; employeeName: string; score: number | null; total: number | null; finishedAt: string | null }
export interface TcPlan {
  id: string; month: string; title: string; role: string | null; employeeIds: string[]; employeeNames: string[]
  plannedDate: string | null; trainerId: string | null; trainerName: string | null; inspectorName: string | null
  inspectionResult: 'DONE' | 'PARTIAL' | 'NOT_DONE' | null; inspectionNotes: string | null; inspectedAt: string | null
}
export interface TcMyExam {
  exam: { id: string; title: string; minutes: number; passPct: number; count: number } | null
  questions?: { id: string; text: string; options: string[] }[]
  attempt?: { score: number | null; total: number | null; finishedAt: string | null; startedAt: string }
}
export interface Justification {
  id: string; employeeId: string; employeeName: string; monitorName: string | null
  question: string; answer: string | null; answeredAt: string | null
  verdict: 'ACCEPTED' | 'REJECTED' | null; verdictNote: string | null; decidedAt: string | null; createdAt: string
}
export interface PaymentAuditRow {
  id: string; projectId: string; projectCode: string; projectName: string
  contractValue: number | null; received: number; amount: number; paidAt: string
  note: string | null; createdAt: string; createdById: string | null; createdByName: string | null
  verifiedAt: string | null; cancelledAt: string | null; cancelReason: string | null; duplicate: boolean
}

export interface CourierMessage {
  id: string
  kind: 'extra_task' | 'leave' | 'project' | 'booking'
  message: string
  link: string
  createdAt: string
  deliveredAt: string | null
  recipientName?: string
}

export interface CourierState {
  now: string
  outgoing: CourierMessage[]
  incoming: CourierMessage[]
  badges: Record<string, number>
}

export interface MatrixWatch {
  level: 'CALM' | 'ALERT' | 'RED'
  mood: 'CALM' | 'PLEASED' | 'ALERT' | 'ANGRY'
  group: string
  /** عنده مشروع مفتوح: العين خضرة (مشرف مشروع) وتراقب دوره ومشروعه. */
  projectMode?: boolean
  workload: MatrixWorkload[]
  name?: string
  id?: string
  open: number
  items: { summary: string; kind: string; label: string; escalated: boolean; since: string }[]
}

// تعليمات ماتركس — «الخطوات التعليمية» يعدّلها المدير.
export interface MatrixGuideRule {
  id: string
  route: string
  match: string
  groups: string
  text: string
  onlyIfPending: boolean
  priority: number
  enabled: boolean
  createdAt: string
  hits?: number
  lastHitAt?: string | null
  source?: string
}

// اقتراح من ماتركس ينتظر قرار المدير.
export interface MatrixProposal {
  id: string
  kind: 'GUIDE_RULE' | 'PREDICTION' | 'RULE_TUNE'
  title: string
  rationale: string
  evidence: Record<string, unknown> | null
  payload: Record<string, unknown> | null
  source: 'RULES' | 'MODEL'
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  note: string | null
  createdAt: string
  decidedAt?: string | null
  employeeId?: string | null
  employeeName?: string | null
  employeeRole?: string | null
}

// تقرير ماتركس المفصّل عن موظف.
export interface ReportLine { text: string; tone: 'OK' | 'WARN' | 'BAD' | 'INFO'; link?: string; fixes?: string[] }
export interface EmployeeReport {
  employeeId: string
  name: string
  group: string
  day: string
  attendance: ReportLine[]
  jobs: { code: string; service: string; status: string; stages: { label: string; at: string; gap: number }[]; lines: ReportLine[] }[]
  leaves: ReportLine[]
  behavior: ReportLine[]
  reminders: ReportLine[]
  performance: ReportLine[]
  activity: { at: string; text: string; count: number; area: string }[]
  slowJobs: { bookingId: string; code: string; service: string; actual: number; expected: number }[]
  summary: string
  summaryBy: 'RULES' | 'MODEL'
  workload: MatrixWorkload[] | null
  /** شغله حجوزات وطلعات؟ غير الميدانيين ينقاسون بشغل دورهم (role). */
  field: boolean
  role?: { title: string; metrics: RoleMetric[]; done: number; pending: number; lines: ReportLine[] }
}
export interface RoleMetric { label: string; value: number }
export interface MatrixBusiness {
  months: { month: string; bookings: number; revenue: number; invoiced: number }[]
  mtd: { bookings: number; revenue: number; lastBookings: number; lastRevenue: number; projectRevenue?: number; lastProjectRevenue?: number }
  forecast: {
    expected: number; expectedJobs: number; basis: string; insufficient: boolean; low?: number; high?: number; accuracy?: string
    // (ع) 10-15: التعلّم — التصحيح من الأغلاط، والثقة الحقيقية (null = لسه يتعلّم)، وسجل الأحكام
    rawExpected?: number; rawExpectedJobs?: number; calibrationPct?: number; jobsCalibrationPct?: number
    confidence?: number | null; jobsConfidence?: number | null
    history?: { month: string; kind: 'JOBS' | 'REVENUE'; forecast: number; actual: number; errorPct: number; verdict: 'CORRECT' | 'WRONG'; snapshotDay: string }[]
  }
  customers: { actual: number; inquiry: number; repeat: number; openOnly: number }
  inquirers: { id: string; name: string; archived: number; lastAt: string }[]
  /** منجزة هالشهر بلا فاتورة (بدون OLD والكشف). */
  uninvoiced?: { id: string; code: string; customer: string | null; service: string | null; completedAt: string; leader: string | null }[]
}

// جدول الدوام — كل موظف فعّال وحضوره اليوم.
export interface WorkScheduleRow {
  id: string
  name: string
  role: Employee['role']
  isLeader: boolean
  shift: string
  shiftStart: string
  shiftEnd: string
  checkIn: string | null
  checkOut: string | null
}

export interface GroupPerformance {
  group: string
  /** حجوزات وسرعة للميدانيين بس؛ الباقين بشغل دورهم (columns/metrics). */
  field: boolean
  columns?: string[]
  members: { id: string; name: string; checkIn?: string; late: number; absent: boolean; shiftChanged?: boolean; onLeave?: boolean; jobs: number; completed: number; partial: number; open: number; unstarted?: number; speed: number | null; prevSpeed: number | null; score: number; noSchedule?: boolean; metrics?: RoleMetric[]; pending?: number }[]
  problems: { text: string; link?: string; fixes: string[] }[]
}

// مركز قيادة ماتركس
export interface MatrixFeedItem { kind: 'BOOKING' | 'INVOICE' | 'ACTION' | 'ESCALATED' | 'PROPOSAL'; title: string; sub: string; at: string; ref: string }
export interface MatrixTrendDay { day: string; completed: number; revenue: number; onTime: number; scheduled: number }
export interface LateWindow { total: number; late: number; unstaffed: number; partial: number; openNow: number }
export interface LateFocus {
  recent: LateWindow; previous: LateWindow; recentPct: number; prevPct: number; changePct: number | null
  cause: string; impact: string; suggestion: string; severity: 'HIGH' | 'MEDIUM' | 'LOW'
  items?: { id: string; code: string; customer: string | null; service: string | null; status: string; scheduledAt: string; completedAt: string | null; leader: string | null; unstaffed: boolean; partial: boolean; daysLate: number; reason: string | null }[]
}

export interface RemoteEntry { id: string; employeeId: string; employeeName: string; workDate: string; seconds: number; note: string | null; addedById: string | null; addedBy: string; createdAt: string; source: 'MANUAL' | 'TIMER'; closedAt: string | null }
export interface RemoteTimer { running: boolean; elapsed: number; startedAt: string | null }
export interface RemoteSummary { employeeId: string; name: string; days: number; seconds: number }
export interface RemoteHistory {
  months: { month: string; employees: number; days: number; seconds: number }[]
  month?: string
  rows?: { employeeId: string; name: string; days: number; seconds: number; targetHours: number | null; hourlyRateIQD: number | null }[]
  entries?: RemoteEntry[]
}

// ═══ تقييم الموظفين (قرار (ع) 10-05) ═══
export type StaffRatingStage = 'COORD_LEADER' | 'AUDIT' | 'QUALITY_CALL'
export interface PendingStaffRating { bookingId: string; code: string; rateeId: string; rateeName: string; role: 'LEADER' | 'COORDINATOR' }
export interface RuleLoss { rule: string; title: string; lost: number; count: number; advice: string }
export interface StaffScore {
  id: string; name: string; role: string
  earned: number; max: number; matrixPct: number | null
  humanAvg: number | null; humanCount: number; humanPct: number | null
  final: number | null; prevFinal: number | null; noHuman: boolean
  topLosses: RuleLoss[]; bySource: Record<string, [number, number]>
  group: string; groupLabel: string
  reliability: number | null; reliabilityParts: { key: string; label: string; pct: number; detail: string }[]
  skills?: number; hasLicense?: boolean; isLeader?: boolean
}
export interface StaffScoreBoard { month: string; on: boolean; staff: StaffScore[]; insights: string[] }
export interface MatrixScorePoint {
  id: string; employeeId: string; source: string; sourceId: string; sourceLabel: string | null; rule: string
  points: number; maxPoints: number; reason: string; at: string; cancelledAt: string | null; cancelNote: string | null
}
export interface HumanRatingRow { raterName: string | null; stage: string; code: string | null; score: number; note: string | null; at: string }
export interface StaffScoreDetail extends StaffScore { month: string; points: MatrixScorePoint[]; human: HumanRatingRow[] }

// ═══ الحضور الإجباري (قرار (ع) 10-06) ═══
export interface AttendanceGateState { required: boolean; open: boolean; afterShift: boolean; shiftStart: string; shiftEnd: string; endLabel: string; startLabel?: string; evening: boolean; offer?: boolean; autoClosed?: { id: string; at: string; label: string } }
export interface AttendanceClaim {
  id: string; attendanceId: string; employeeId: string; name: string; kind: 'ACK' | 'WORKED' | 'BACK'
  autoAt: string; claimedUntil: string | null; evidenceAt: string | null; evidence: string | null; note: string | null
  status: 'OK' | 'PENDING' | 'APPROVED' | 'REJECTED'; createdAt: string
}

// ═══ دفعات المشاريع (قرار (ع) 10-06) ═══
export interface ProjectMoney {
  projectId: string; code: string; name: string; stage: string; priceText: string | null
  contractValue: number | null; valueFromText: boolean; paid: number; unverified: number; payments: number
  lastPaidAt: string | null; ownerId: string | null; ownerName: string | null
}
export interface ProjectPayment {
  id: string; projectId: string; amount: number; paidAt: string; method: 'CASH' | 'TRANSFER' | 'CHEQUE'
  receiptNo: string | null; note: string | null; createdBy: string | null; createdAt: string
  verifiedBy: string | null; verifiedAt: string | null; cancelledAt: string | null; cancelReason: string | null
}
export interface ProjectPaymentsState { money: ProjectMoney; payments: ProjectPayment[]; canVerify: boolean; canValue: boolean }

// ═══ موردين التقنيين (قرار (ع) 10-06) ═══
export interface TechSupplierIn { companyName: string; ownerName?: string; phone: string; address?: string; locationUrl?: string; specialty?: string; notes?: string }
export interface TechSupplier {
  id: string; companyName: string; ownerName: string | null; phone: string; address: string | null; locationUrl: string | null
  specialty: string | null; notes: string | null; createdBy: string | null; createdAt: string; assignedTo: string[]
}

// ═══ متابعة السيارات (ملف الإكسل — قرار (ع) 10-09) ═══
export interface VehicleTrackRating {
  id: string; vehicleId: string; date: string; score: number | null
  wash: number | null; exteriorClean: number | null; exteriorCondition: number | null; tireCondition: number | null
  glassClean: number | null; lightsCondition: number | null; technicalFaults: number | null; faultDescription: string | null
  interiorClean: number | null; seatsCondition: number | null; interiorDirt: number | null; smell: number | null
}
export interface VehicleTrackStats {
  vehicleId: string; vehicleName: string; completed: number; rawSum: number; rawMax: number; rawPct: number | null
  washPct: number | null; items: Record<string, number | null>; weighted: number | null; grade: string; washDone: number; washNot: number
}
export interface VehicleTracking {
  from: string; to: string; vehicles: { id: string; name: string; temp: boolean }[]; ratings: VehicleTrackRating[]; stats: VehicleTrackStats[]
}

// 🧽 غسل السيارات (عامل الغسل — (ع) 10-09)
export interface CarWashVehicle {
  vehicleId: string; name: string; plateNumber: string
  washedAt: string | null; washedBy: string | null; washedById: string | null
}

// ═══ مقارنة الأشهر (قرار (ع) 10-10) ═══
export interface MonthStats {
  month: string; bookingsCreated: number; bookingsCompleted: number; bookingsCancelled: number; internalWorks: number
  projectsCreated: number; newCustomers: number; complaints: number; bookingsRevenue: number; projectsRevenue: number; expenses: number
}

// ═══ تأخير المشاريع (قرار (ع) 10-10) ═══
export interface ProjectDelay {
  id: string; code: string; name: string; stage: string; workType: string | null
  ownerId: string; ownerName: string; daysInStage: number; stageLimit: number; idleDays: number
  status: string; problems: string[]; verdict: string; overLimit: boolean; workedToday: boolean
  lastActivityAt: string | null; lastActivityBy: string | null
  delayReason: string | null; delayReasonBy: string | null; delayReasonAt: string | null
}
export interface StageLimitRow { stage: string; learned: number; median: number; manual: number | null }

// ═══ شغل المخازن (قرار (ع) 10-10) ═══
export interface ProcDecision { id: string; label: string; createdAt: string; decidedAt: string | null; decidedBy: string | null; byName: string | null }
export interface ProcurementWatch {
  since: string; materials: ProcDecision[]; tools: ProcDecision[]; shortages: ProcDecision[]
  vehicleDays: { day: string; total: number; rated: number }[]; unratedToday: string[]; fleetOverdue: string[]
}

// ═══ دردشة ماتركس (قرار (ع) 10-10) ═══
export interface MatrixChatRow { id: string; title: string; updatedAt: string }
export interface MatrixChatMessage { id: string; role: 'USER' | 'ASSISTANT'; text: string; steps: string | null; image?: string | null; createdAt: string }
export interface MatrixMission { id: string; employeeId: string; employeeName: string; note: string | null; startsOn: string; endsOn: string; active: boolean; summary: string | null; createdAt: string; summaryAt: string | null }
export interface StaffPointsRow { id: string; name: string; role: string; group: string; groupLabel: string; kpi: number; kpiLost: string[]; attendance: number; work: number; total: number; lines: { label: string; count: number; each: number; points: number }[] }
export interface MatrixLesson { id: string; text: string; context: string | null; by: string | null; createdAt: string }

// ═══ الميدان (ع) 10-14 ═══
export type DepartOrigin = 'COMPANY' | 'PREV_BOOKING' | 'OUTSIDE' | 'UNKNOWN'
export interface CompanyLocation { lat: number; lng: number; radiusM: number }
export interface BookingTrack {
  times: {
    acceptedAt: string | null; materialsReadyAt: string | null; departedAt: string | null; departOrigin: DepartOrigin | null
    arrivedAt: string | null; startedAt: string | null; completedAt: string | null
    autoDepartedAt: string | null; autoArrivedAt: string | null; autoCompletedAt: string | null
    lat: number | null; lng: number | null
  }
  stops: { kind: 'ROAD' | 'ARRIVAL'; lat: number; lng: number; startAt: string; endAt: string; minutes: number }[]
  proofs: { stage: string; pressedAt: string; offline: boolean; verdict: string; distanceM: number | null }[]
  path: [number, number][]
  company: CompanyLocation | null
}
export interface FieldMapBooking {
  id: string; code: string; customer: string | null; status: string; lat: number | null; lng: number | null
  departedAt: string | null; departOrigin: DepartOrigin | null; arrivedAt: string | null; completedAt: string | null
  autoArrivedAt: string | null; autoCompletedAt: string | null; team: string | null
}
export interface FieldMap {
  company: CompanyLocation | null
  bookings: FieldMapBooking[]
  employees: { employeeId: string; name: string; lat: number; lng: number; at: string; bookingId: string | null }[]
  heat: [number, number, number][]
}

// ═══ رقابة المشتريات (ع) 10-15 ═══
export interface PurchaseControlRow { id: string; verifiedAt: string | null; verifiedBy: string | null; myReview: boolean; alerts: number }
export interface PurchasePriceRow { key: string; product: string; supplier: string; lastPrice: number; minPrice: number; times: number; lastAt: string }
export interface PurchaseAlert {
  id: string; kind: 'PRICIER' | 'JUMP' | 'QUOTE' | 'OVER_STD'; code: string; product: string; supplier: string | null; unitPrice: number; quantity: number
  refSupplier: string | null; refPrice: number; loss: number; buyer: string | null; createdAt: string; reviewedAt: string | null; reviewNote: string | null
}
export interface SupplierScore { supplierId: string; n: number; price: number; quality: number; speed: number; punctuality: number; dealing: number; alerts: number }
export interface SupplierReviewInput { price: number; quality: number; speed: number; punctuality: number; dealing: number; note?: string }
export interface StandardPrice { key: string; product: string; price: number; source: 'AUTO' | 'MANUAL'; setBy: string | null; updatedAt: string; market: number | null }
export interface StandardSuggestion { key: string; product: string; current: number; suggested: number; basis: string }
export interface PurchaseGroup { name: string; total: number; count: number; saving: number }
export interface PurchaseMonth {
  month: string; total: number; count: number; requests: number; prevTotal: number; prevCount: number
  saving: number; alerts: number; unverified: number
  byCategory: PurchaseGroup[]; bySupplier: PurchaseGroup[]; byBuyer: PurchaseGroup[]
  lines: { at: string; code: string; product: string; qty: number; unitPrice: number; supplier: string; category: string; buyer: string; bestPrice: number | null; bestSupplier: string | null; saving: number }[]
}
export interface PurchaseBias { buyer: string; supplier: string; n: number; total: number; share: number; extra: number; cheaperAvailable: number }

export interface DrillRow {
  id: string; code: string; customer: string | null; phone: string | null; service: string | null
  status: string; scheduledAt: string | null; createdAt: string; crew: string | null
  responsible: string | null; ageDays: number | null; station: string; missing: string | null
}
export interface DrillResult {
  key: string; title: string; explain: string; count: number; rows: DrillRow[]
  extra?: { completed: number; total: number }
}

/** «🎯 واجباتك اليوم» — (ع 10-17) */
export interface Duty {
  key: string
  label: string
  hint: string
  done: number
  pending: number | null
  /** المتأخر أكثر من ٢٤ ساعة بالطابور — الحكم عليه (ع 10-22) */
  late?: number
  status: 'DONE' | 'PARTIAL' | 'MISSED' | 'INFO'
  link?: string
}
export interface DutiesReport {
  employeeId: string
  position: string
  duties: Duty[] | null
  doneCount: number
  total: number
  verdict: 'GOOD' | 'PARTIAL' | 'BAD' | 'OFF'
  verdictAr: string
  /** جمعة أو إجازة معتمدة — الواجبات تبين بس ما تنحسب */
  off?: 'FRIDAY' | 'LEAVE'
}

/** صورة «قبل» أو «بعد» لشغلة حجز (دازاي 10-25) */
export interface WorkPhoto { id: string; phase: 'BEFORE' | 'AFTER'; employeeId: string | null; createdAt: string }

/** طلب حذف فاتورة (دازاي 10-25) */
export interface InvoiceDeleteRequest {
  id: string; invoiceId: string; accountingCode: string | null; bookingCode: string | null; netTotal: number | null
  requestedByName: string | null; reason: string; status: 'PENDING' | 'APPROVED' | 'REJECTED'
  decidedByName: string | null; decidedAt: string | null; decisionNote: string | null; createdAt: string
}

/** حجز قديم ناقصه ورق (دازاي 10-26) */
export interface CarryoverItem { id: string; code: string; completedAt: string | null; hasInvoice: boolean; hasReport: boolean }

/** «الصلاحيات والوصول» (دازاي 10-26) */
export interface AccessPermDef { name: string; label: string; grantable: boolean; lockReason?: string }
export interface AccessGroup { key: string; title: string; bundle?: string; perms: AccessPermDef[] }
export interface EmployeeAccess {
  employee: { id: string; name: string; role: string; secondaryRoles: string[]; isLeader: boolean }
  perms: { name: string; source: 'TEMPLATE' | 'MANUAL' | 'EXCEPTION'; expiresAt?: string; reason?: string }[]
  missingFromTemplate: string[]
  version: number
}
export interface RoleDriftRow {
  employeeId: string; name: string; role: string; status: 'MATCH' | 'DIFF' | 'NO_TEMPLATE'
  extra: string[]; missing: string[]; exceptions: string[]; version: number
}
export interface PermissionAuditRow {
  id: string; employeeId: string; employeeName: string | null; actorName: string | null
  action: 'ADD' | 'REMOVE' | 'EXPIRE' | 'UNPIN'; permission: string; reason: string | null; createdAt: string
}

export interface RemotePlan { targetHours: number; hourlyRateIQD: number }

/** (ع) 11-04: طلب مرفوع للمدير */
export interface ManagerRequest {
  id: string
  requesterId: string
  requesterName: string
  title: string
  details: string
  priority: 'NORMAL' | 'HIGH' | 'URGENT'
  context: string | null
  source: 'SCREEN' | 'MATRIX'
  status: 'OPEN' | 'DONE' | 'REJECTED'
  reply: string | null
  handledByName: string | null
  handledAt: string | null
  createdAt: string
}
