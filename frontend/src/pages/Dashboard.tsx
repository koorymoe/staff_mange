import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import MyFundBalance from '../components/MyFundBalance'
import { api } from '../api'
import type { Booking, Expense, AttendanceRecord, StaffRequest, LeaveRequest, InventoryCheck, FinanceSummary, DailyAuditReport, TodayPulse } from '../api'
import { useSession, hasGpsSkill } from '../session'
import { timeGreeting, GREETING_HOLD_MS } from '../greeting'
import { MapViewer } from '../components/MapLazy'

/* ───── Attendance helpers ───── */

function elapsedSince(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime()
  if (diffMs < 0) return '٠ دقائق'
  const diffMin = Math.floor(diffMs / 60000)
  const hours = Math.floor(diffMin / 60)
  const mins = diffMin % 60
  if (hours === 0) return `${mins} دقيقة`
  return `${hours} ساعة و ${mins} دقيقة`
}

function fmtTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
}

function formatTime(): string {
  return new Date().toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })
}

function formatDate(): string {
  return new Date().toLocaleDateString('ar-IQ', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

/* ───── Types ───── */

interface GpsStats {
  totalDevices: number
  totalCustomers: number
  devicesByStatus: { status: string; count: number }[]
  totalSims: number
  availableSims: number
  inUseSims: number
}

/* ───── Animated Counter ───── */
function AnimatedNumber({ value, loading }: { value: number; loading: boolean }) {
  const [display, setDisplay] = useState(0)
  useEffect(() => {
    if (loading) return
    let start = 0
    const duration = 800
    const step = Math.ceil(value / (duration / 16))
    const timer = setInterval(() => {
      start += step
      if (start >= value) { setDisplay(value); clearInterval(timer) }
      else setDisplay(start)
    }, 16)
    return () => clearInterval(timer)
  }, [value, loading])
  if (loading) return <span className="inline-block h-8 w-16 animate-pulse rounded-lg bg-slate-200" />
  return <>{display.toLocaleString('ar-SA')}</>
}

/* ───── Progress Ring ───── */
function ProgressRing({ percent, size = 56, stroke = 5, color }: { percent: number; size?: number; stroke?: number; color: string }) {
  const r = (size - stroke) / 2
  const circ = 2 * Math.PI * r
  const offset = circ - (percent / 100) * circ
  return (
    <svg width={size} height={size} className="-rotate-90">
      <circle cx={size/2} cy={size/2} r={r} fill="none" stroke="currentColor" strokeWidth={stroke} className="text-slate-100" />
      <circle cx={size/2} cy={size/2} r={r} fill="none" stroke={color} strokeWidth={stroke}
        strokeDasharray={circ} strokeDashoffset={offset} strokeLinecap="round"
        className="transition-all duration-1000 ease-out" />
    </svg>
  )
}

export default function Dashboard() {
  const { employee, permissions, gpsServiceId } = useSession()
  const navigate = useNavigate()
  const [gpsStats, setGpsStats] = useState<GpsStats | null>(null)
  const [bookingCount, setBookingCount] = useState(0)
  const [employeeCount, setEmployeeCount] = useState(0)
  const [customerCount, setCustomerCount] = useState(0)
  const [bookings, setBookings] = useState<Booking[]>([])
  const [myTasks, setMyTasks] = useState<Booking[]>([])
  const [lastCheck, setLastCheck] = useState<InventoryCheck | null>(null)
  // الفني/التقني بالميدان (مو ليدر) — واجهته الرئيسية شغله هو بس
  const fieldOnly = !!employee
    && (employee.role === 'TECHNICIAN' || employee.role === 'TECHNICAL')
    && !employee.isLeader

  // ── متى يصير الجرد مستحق؟ ──
  // قاعدتان مثل ما الشغل ماشي بالواقع:
  //  ١. جرد أسبوعي — مرّت ٧ أيام أو أكثر على آخر جرد (أو ما جرد أبداً).
  //  ٢. قبل كل حجز — عنده مهمة مو مكتملة وما جرد اليوم، لأن العدة تنتفحص
  //     قبل ما يطلع للموقع مو بعد ما يوصل ويكتشف إنه ناقصه شي.
  // «الآن» تنقرأ مرة وحدة عند فتح الصفحة — قراءة الوقت أثناء الرسم
  // تخلي النتيجة تفرق بين رسمة وأخرى.
  const [openedAt] = useState(() => Date.now())
  const lastCheckAt = lastCheck ? new Date(lastCheck.checkedAt) : null
  const daysSinceCheck = lastCheckAt
    ? Math.floor((openedAt - lastCheckAt.getTime()) / 86400000)
    : null
  const checkedToday = !!lastCheckAt && lastCheckAt.toDateString() === new Date(openedAt).toDateString()
  const dueWeekly = daysSinceCheck === null || daysSinceCheck >= 7
  const dueBeforeTask = myTasks.length > 0 && !checkedToday
  const inventoryDue = fieldOnly && (dueWeekly || dueBeforeTask)
  const inventoryHint = daysSinceCheck === null
    ? 'ما سويت جرد بعد — راجع عدتك وأشّر الموجود'
    : dueBeforeTask && !dueWeekly
      ? `عندك ${myTasks.length} مهمة — جرد عدتك قبل ما تطلع`
      : dueWeekly
        ? (daysSinceCheck === 0 ? 'حان وقت الجرد' : `آخر جرد قبل ${daysSinceCheck} يوم`)
        : 'جردت اليوم ✓ — تكدر تراجع عدتك'
  const [mapTask, setMapTask] = useState<Booking | null>(null)
  const [taskAmounts, setTaskAmounts] = useState<Record<string, string>>({})
  const [taskAdvances, setTaskAdvances] = useState<Record<string, string>>({})
  const [taskNotes, setTaskNotes] = useState<Record<string, string>>({})
  const [completedBookings, setCompletedBookings] = useState<Booking[]>([])
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [finance, setFinance] = useState<FinanceSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [currentTime, setCurrentTime] = useState(formatTime())

  // ═══ التحية ═══
  // تطلع زمنية («صباح الخير») وبعد ثواني تستقر على «مرحباً».
  //
  // ⚠️ التحية تنحسب مرة وحدة عند فتح الشاشة وتنثبت: لو حسبناها بكل
  // رندر، الموظف الي فاتح النظام الساعة ١١:٥٩ ظهراً يشوف التحية
  // تنقلب قدامه من «صباح الخير» لـ«نهارك سعيد» بلا سبب واضح.
  // نبض اليوم — أرقام الإداري. تفشل بهدوء: الشاشة الرئيسية ما تصير
  // تنكسر لأن ويدجت وحدة ما وصلها رد.
  const [pulse, setPulse] = useState<TodayPulse | null>(null)
  useEffect(() => {
    api.getTodayPulse().then(setPulse).catch(() => setPulse(null))
  }, [])

  const [greet] = useState(() => timeGreeting())
  const [greetingPhase, setGreetingPhase] = useState<'time' | 'rest'>('time')
  useEffect(() => {
    const t = setTimeout(() => setGreetingPhase('rest'), GREETING_HOLD_MS)
    return () => clearTimeout(t)
  }, [])
  const [projectStats, setProjectStats] = useState<Record<string, number> | null>(null)
  const [pendingStaffReqs, setPendingStaffReqs] = useState<StaffRequest[]>([])
  // طلبات الإجازة الي تنتظر قرار هذا المدير — بضمنها الي انطاها موافقة
  // أولية. تظهر بالشاشة الرئيسية حتى ما تنتسى: الموافقة الأولية بلا
  // قرار نهائي معناها موظف ناطر بلا جواب.
  const [openLeaves, setOpenLeaves] = useState<LeaveRequest[]>([])
  // التدقيق اليومي على الشاشة الرئيسية — طلب صاحب العمل: المحاسب يفتح
  // النظام ويشوف شغل اليوم كَبل ما يدور عليه بالقائمة.
  const [dailyAudit, setDailyAudit] = useState<DailyAuditReport | null>(null)

  useEffect(() => {
    if (employee?.role !== 'FINANCE' && !permissions.includes('finance')) return
    api.getDailyAudit().then(setDailyAudit).catch(() => setDailyAudit(null))
  }, [employee?.role, permissions])

  useEffect(() => {
    const t = setInterval(() => setCurrentTime(formatTime()), 60000)
    return () => clearInterval(t)
  }, [])

  useEffect(() => {
    if (!employee) return
    // مدير المشاريع مدير مو فني — ما ينستلم مهام مثل الفنيين
    const isTech = employee.role === 'TECHNICIAN' || employee.role === 'TECHNICAL' || employee.isLeader
    const needsFinance = employee.role === 'FINANCE' || permissions.includes('monitoring')
    // الفني/التقني بالميدان (مو ليدر): صفحته الرئيسية = مهامه هو، مو أكثر.
    // كان ينزّل كل حجوزات الشركة ويفلترها بالمتصفح حتى يلكه مهامه — يعني
    // كل ما تكبر الشركة تصير صفحته أبطأ، وهو أصلاً ما إله علاقة بالباقي.
    // هسه الفلترة بقاعدة البيانات، وشلنا عنه طلبين ما يشوف نتيجتهم:
    // إحصاءات GPS وملخّص الأرقام (موظفين/زبائن) — مالتهم للإداريين.
    // الأرقام تجي من مسار ملخّص واحد (بضع عشرات البايتات) بدل ما ننزّل
    // كل الموظفين وكل العملاء على جهاز المستخدم عشان نعدّهم. الحجوزات
    // تُطلب بس للي يحتاج قائمتها فعلاً (فني عنده مهام، أو مالية).
    const needsBookingList = isTech || needsFinance
    Promise.all([
      fieldOnly ? Promise.resolve(null) : api.getGpsStats().catch(() => null),
      needsBookingList
        ? api
            .getBookings(
              fieldOnly
                ? { assignedTo: 'me' }
                // المراقب والمحاسب يعرضون الشغل الحيّ بس — كل الفلاتر
                // بلوحاتهم على هذي الحالات الثلاث. الباقي (منجز/ملغى)
                // جان ينزل بلا فايدة ويثقل الصفحة كل ما كبر الأرشيف.
                : { status: ['PENDING', 'CONFIRMED', 'IN_PROGRESS'] },
            )
            .catch(() => [] as Booking[])
        : Promise.resolve([] as Booking[]),
      fieldOnly ? Promise.resolve(null) : api.getDashboardSummary().catch(() => null),
      // الأرقام (المجاميع والعدادات) تجي محسوبة من السيرفر على الأرشيف
      // الكامل، والقائمة تجي مقصوصة لأنها للعرض بس — «آخر المنجز» و«منجز
      // اليوم». يعني الأرقام دقيقة ١٠٠٪ بدون ما ننزّل أرشيف الشركة.
      needsFinance ? api.getBookings({ status: 'COMPLETED', limit: 200 }).catch(() => [] as Booking[]) : Promise.resolve([] as Booking[]),
      needsFinance ? api.getExpenses().catch(() => [] as Expense[]) : Promise.resolve([] as Expense[]),
      needsFinance ? api.getFinanceSummary().catch(() => null) : Promise.resolve(null),
    ]).then(([gps, bk, summary, cb, exp, fin]) => {
      setFinance(fin as FinanceSummary | null)
      const allBookings = bk
      setGpsStats(gps as GpsStats | null)
      setBookings(bk as Booking[])
      const sum = summary as { employeeCount: number; customerCount: number; bookingCount: number } | null
      setBookingCount(sum?.bookingCount ?? (bk as Booking[]).length)
      setEmployeeCount(sum?.employeeCount ?? 0)
      setCustomerCount(sum?.customerCount ?? 0)
      setCompletedBookings(cb as Booking[])
      setExpenses(exp as Expense[])
      const taskList = (allBookings as Booking[]).filter(b =>
        b.status !== 'COMPLETED' && b.status !== 'CANCELLED' &&
        b.assignments.some(a => a.employee.id === employee.id)
      )
      setMyTasks(taskList)
    }).finally(() => setLoading(false))

    // آخر جرد للفني — حتى نعرف هل حان وقت جرده الأسبوعي
    if (fieldOnly) api.getMyLastInventoryCheck().then(setLastCheck).catch(() => setLastCheck(null))

    // طلبات الكادر المعلقة — تنبيه لإداري الكوادر والأدمن
    if (employee.role === 'ADMIN' || employee.role === 'HR_COORDINATOR') {
      api.getStaffRequests()
        .then(reqs => setPendingStaffReqs(reqs.filter(r => r.status === 'PENDING' || r.status === 'APPROVED')))
        .catch(() => setPendingStaffReqs([]))
    }

    // طلبات الإجازة الي تنتظر قراره. صندوق الموافقات مو مفتوح للكل —
    // الراوت يرد 403 على غير المخوّل، فكل موظف عادي (وأغلب الموظفين
    // كذلك) جان يطلع بكونسوله خطأ ٤٠٣ مع كل فتحة للرئيسية بلا فايدة.
    // نسأل أول بمسار العدّ الرخيص: مخوّل؟ وقتها بس ننزّل الصندوق.
    api.getLeavePendingCount()
      .then(({ canApprove }) => (canApprove ? api.getLeaveInbox('OPEN') : []))
      .then((rows) => setOpenLeaves(rows ?? []))
      .catch(() => setOpenLeaves([]))

    // إحصائيات المشاريع للوحة مدير المشاريع (أو الأدمن)
    if (employee.role === 'PROJECT_MANAGER' || employee.role === 'ADMIN' || permissions.includes('project_management')) {
      const token = localStorage.getItem('authToken')
      fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:4000/api'}/projects`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
        .then(r => r.ok ? r.json() : null)
        .then(d => setProjectStats(d?.stats || null))
        .catch(() => setProjectStats(null))
    }
  }, [employee, permissions])

  /* ── Attendance widget state ── */
  const [activeRecord, setActiveRecord] = useState<AttendanceRecord | null>(null)
  const [todayTotalMinutes, setTodayTotalMinutes] = useState(0)
  const [elapsed, setElapsed] = useState('')
  const [showCheckoutConfirm, setShowCheckoutConfirm] = useState(false)

  const refreshOpenSession = useCallback(() => {
    if (!employee) return
    api.getMyOpenSession()
      .then(res => {
        setActiveRecord(res.open)
        setTodayTotalMinutes(res.totalMinutes)
      })
      .catch(() => setActiveRecord(null))
  }, [employee])

  useEffect(() => {
    refreshOpenSession()
  }, [refreshOpenSession])

  useEffect(() => {
    if (!activeRecord?.checkIn || activeRecord.checkOut) return
    // Same intentional immediate-init as AttendancePage's elapsed timer.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setElapsed(elapsedSince(activeRecord.checkIn))
    const interval = setInterval(() => {
      setElapsed(elapsedSince(activeRecord.checkIn))
      setTodayTotalMinutes(m => m + 1)
    }, 60000)
    return () => clearInterval(interval)
  }, [activeRecord])

  const handleAttCheckIn = useCallback(async () => {
    if (!employee) return
    const rec = await api.checkIn()
    setActiveRecord(rec)
    refreshOpenSession()
  }, [employee, refreshOpenSession])

  const handleAttCheckOut = useCallback(async () => {
    if (!employee || !activeRecord) return
    await api.checkOut()
    setActiveRecord(null)
    setShowCheckoutConfirm(false)
    refreshOpenSession()
  }, [employee, activeRecord, refreshOpenSession])


  const handleTaskStart = async (b: Booking) => {
    const updated = await api.startBooking(b.id)
    setMyTasks(prev => prev.map(t => t.id === updated.id ? updated : t))
  }

  const handleTaskComplete = async (b: Booking) => {
    await api.completeBooking(b.id, {
      completionNotes: taskNotes[b.id] || undefined,
      amountCollected: taskAmounts[b.id] ? Number(taskAmounts[b.id]) : undefined,
      advancePaid: taskAdvances[b.id] ? Number(taskAdvances[b.id]) : undefined,
    })
    setMyTasks(prev => prev.filter(t => t.id !== b.id))
  }

  if (!employee) return null

  const isAdmin = employee.role === 'ADMIN'
  const pendingMaintenance = gpsStats?.devicesByStatus?.find((d) => d.status === 'MAINTENANCE')?.count || 0

  const quickCards = [
    {
      title: 'حجز جديد',
      desc: 'إنشاء حجز خدمة جديد للعميل',
      gradient: 'from-blue-500 via-blue-600 to-blue-700',
      iconPath: 'M12 4v16m8-8H4',
      path: '/sales',
      // نفس بوابة القائمة الجانبية: بالدور أو بصلاحية ممنوحة يدوياً. قبلها
      // كانت الكروت بالدور بس، فموظف عنده صلاحية الحجز يشوف الرابط بالقائمة
      // ولوحته فارغة.
      visible: ['ADMIN', 'SALES', 'HR_COORDINATOR', 'QUALITY_ENGINEER'].includes(employee.role) || permissions.includes('sales_booking'),
    },
    {
      title: 'حجز شكوى',
      desc: 'تسجيل شكوى عميل جديدة',
      gradient: 'from-red-500 via-red-600 to-rose-700',
      iconPath: 'M12 9v4m0 4h.01M10.29 3.86 1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z',
      path: '/complaints',
      visible: employee.role === 'SALES' || permissions.includes('complaints'),
    },
    {
      title: 'طلب GPS جديد',
      desc: 'تسجيل طلب اشتراك GPS جديد للعميل',
      gradient: 'from-indigo-500 via-indigo-600 to-blue-700',
      iconPath: 'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z',
      path: '/gps/purchase',
      visible: employee.role === 'SALES' || permissions.includes('gps_system'),
    },
    {
      title: 'تسجيل مشكلة جودة',
      desc: 'مشكلة تنفيذية ميدانية أو مشكلة رقابية جديدة',
      gradient: 'from-purple-500 via-purple-600 to-fuchsia-700',
      iconPath: 'M12 15a3 3 0 100-6 3 3 0 000 6z M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 11-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9',
      path: '/quality',
      visible: employee.role === 'QUALITY_ENGINEER' || permissions.includes('quality_control'),
    },
    {
      title: 'طلبات GPS المعلقة',
      desc: 'طلبات الأجهزة والتجديد والصيانة بانتظار المراجعة',
      gradient: 'from-amber-500 via-amber-600 to-orange-700',
      iconPath: 'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z',
      path: '/gps/requests',
      visible: employee.role === 'GPS_ADMIN',
    },
    {
      title: 'عرض سعر جديد',
      desc: 'إنشاء عرض سعر احترافي للعميل',
      gradient: 'from-emerald-500 via-emerald-600 to-emerald-700',
      iconPath: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
      path: '/quotations/new',
      visible: isAdmin || ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system'].some((p) => permissions.includes(p)),
    },
    {
      title: 'إضافة عميل',
      desc: 'تسجيل عميل جديد في النظام',
      gradient: 'from-violet-500 via-violet-600 to-violet-700',
      iconPath: 'M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z',
      path: '/customers',
      visible: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'].includes(employee.role) || permissions.includes('manage_customers'),
    },
    {
      title: 'تتبع المهام',
      desc: 'متابعة المهام الميدانية والفرق',
      gradient: 'from-amber-500 via-amber-600 to-amber-700',
      iconPath: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
      path: '/missions',
      visible: isAdmin || employee.role === 'HR_COORDINATOR' || permissions.includes('monitoring'),
    },
  ].filter(c => c.visible)



  const kpiCards = [
    { title: 'الموظفين', value: employeeCount, color: '#3b82f6', bg: 'from-blue-500/10 to-blue-500/5', ring: 75, path: '/employees',
      icon: <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 7a4 4 0 100-8 4 4 0 000 8zM23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" /> },
    { title: 'العملاء', value: customerCount, color: '#10b981', bg: 'from-emerald-500/10 to-emerald-500/5', ring: 60, path: '/customers',
      icon: <><path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></> },
    { title: 'الحجوزات', value: bookingCount, color: '#8b5cf6', bg: 'from-violet-500/10 to-violet-500/5', ring: 45, path: '/bookings',
      icon: <><rect x="3" y="4" width="18" height="18" rx="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" /></> },
    { title: 'أجهزة GPS', value: gpsStats?.totalDevices || 0, color: '#f59e0b', bg: 'from-amber-500/10 to-amber-500/5', ring: 85, path: '/gps',
      icon: <><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z" /><circle cx="12" cy="10" r="3" /></> },
  ]

  return (
    <div dir="rtl" className="space-y-6">
      {/* ═══ Hero Section ═══ */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-l from-[#0a1628] via-[#1a3a6e] to-[#2c5aad] p-8 text-white">
        {/* Decorative elements */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -left-20 -top-20 h-64 w-64 rounded-full bg-white/5 blur-3xl" />
          <div className="absolute -bottom-10 left-1/3 h-48 w-48 rounded-full bg-blue-400/10 blur-2xl" />
          <div className="absolute right-10 top-10 h-32 w-32 rounded-full bg-blue-300/5 blur-xl" />
          {/* Grid pattern */}
          <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'radial-gradient(circle, white 1px, transparent 1px)', backgroundSize: '24px 24px' }} />
        </div>

        <div className="relative z-10 flex flex-wrap items-center justify-between gap-6">
          <div className="flex-1">
            <div className="mb-1 flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 rounded-full bg-white/10 px-3 py-1 text-xs font-medium backdrop-blur-sm">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
                متصل
              </span>
            </div>
            {/* التحية الزمنية أول، وبعدها تستقر على «مرحباً».
                key يخلي React يعيد تشغيل الحركة عند التبديل. */}
            <h1 key={greetingPhase} className="greet-swap text-3xl font-bold tracking-tight">
              {greetingPhase === 'time' ? (
                <>
                  <span className="greet-icon ml-1 inline-block">{greet.icon}</span>
                  {greet.text}، {employee.name}
                </>
              ) : (
                <>مرحباً، {employee.name}</>
              )}
            </h1>
            <p className="mt-2 text-sm text-blue-200/80">{formatDate()}</p>
            <p className="mt-1 text-xs text-blue-300/60">نظام إدارة شامل — شركة الأماني</p>
          </div>

          {/* Time display */}
          <div className="flex flex-col items-center rounded-2xl bg-white/10 px-6 py-4 backdrop-blur-md">
            <span className="text-4xl font-bold tracking-wider tabular-nums">{currentTime}</span>
            <span className="mt-1 text-xs text-blue-200/70">التوقيت المحلي</span>
          </div>
        </div>
      </div>

      {/* رصيد الدوار — يظهر بس للموظف الي أخذ منه */}
      <MyFundBalance />

      {/* ═══ KPI Cards - ADMIN ═══ */}
      {isAdmin && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          {kpiCards.map((kpi) => (
            <button key={kpi.title} onClick={() => navigate(kpi.path)}
              className="group relative overflow-hidden rounded-2xl bg-white p-5 text-right shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)] transition-all duration-300 hover:shadow-[0_8px_32px_rgba(44,90,173,0.12)] hover:-translate-y-0.5">
              {/* Gradient background */}
              <div className={`absolute inset-0 bg-gradient-to-bl ${kpi.bg} opacity-0 transition-opacity group-hover:opacity-100`} />
              <div className="relative flex items-start justify-between">
                <div className="relative">
                  <ProgressRing percent={kpi.ring} color={kpi.color} />
                  <div className="absolute inset-0 flex items-center justify-center">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={kpi.color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      {kpi.icon}
                    </svg>
                  </div>
                </div>
                <div>
                  <p className="text-xs font-medium text-slate-400">{kpi.title}</p>
                  <p className="mt-1 text-3xl font-black text-slate-800">
                    <AnimatedNumber value={kpi.value} loading={loading} />
                  </p>
                </div>
              </div>
              {/* Hover arrow */}
              <div className="absolute bottom-3 left-3 opacity-0 transition-all group-hover:opacity-100 group-hover:translate-x-0 translate-x-2">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={kpi.color} strokeWidth="2.5" strokeLinecap="round"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
              </div>
            </button>
          ))}
        </div>
      )}

      {/* ═══ إدارة الإحصائيات — يومية/أسبوعية/شهرية/مشاريع، حصراً لمدير النظام ═══ */}
      {isAdmin && (
        <button
          onClick={() => navigate('/stats-management')}
          className="flex w-full items-center justify-between rounded-2xl bg-gradient-to-l from-[#1a237e] to-[#283593] p-5 text-right text-white shadow-lg transition hover:shadow-xl"
        >
          <div>
            <h3 className="text-base font-extrabold">📊 إدارة الإحصائيات</h3>
            {/* ⚠️ bookingCount هو COUNT(*) على كل جدول الحجوزات — مجموع
                تاريخي مو رقم اليوم. اللافتة كانت تسميه «حجوزات اليوم»
                فتنطي رقم أكبر بمرات من الحقيقة. */}
            <p className="mt-1 text-sm text-blue-100">مجموع الحجوزات: {bookingCount} — إحصائيات يومية وأسبوعية وشهرية لكل موظف، وإحصائية المشاريع</p>
          </div>
          <span className="text-lg">←</span>
        </button>
      )}

      {/* ═══ طلبات كادر معلقة — تنبيه لإداري الكوادر حتى ما يفوته طلب من إدارة المشاريع ═══ */}
      {['ADMIN', 'HR_COORDINATOR'].includes(employee.role) && pendingStaffReqs.length > 0 && (
        <div className="rounded-2xl border-2 border-violet-200 bg-gradient-to-l from-violet-50 to-purple-50 p-5 shadow-lg shadow-violet-100/50">
          <div className="mb-3 flex items-center justify-between">
            <button onClick={() => navigate('/staff-requests')} className="rounded-lg bg-violet-600 px-4 py-1.5 text-xs font-bold text-white transition hover:bg-violet-700">
              عرض الطلبات ←
            </button>
            <div className="flex items-center gap-3">
              <span className="relative flex h-3 w-3">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-violet-400 opacity-75" />
                <span className="relative inline-flex h-3 w-3 rounded-full bg-violet-500" />
              </span>
              <h3 className="text-base font-extrabold text-violet-800">
                طلبات كادر بانتظار التلبية
                <span className="mr-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-violet-600 text-sm font-black text-white">{pendingStaffReqs.length}</span>
              </h3>
            </div>
          </div>
          <div className="space-y-2">
            {pendingStaffReqs.slice(0, 3).map(r => (
              <div key={r.id} onClick={() => navigate('/staff-requests')} className="flex cursor-pointer items-center justify-between rounded-xl bg-white/80 px-4 py-3 transition hover:bg-white hover:shadow-md">
                <span className="text-xs text-slate-400">{new Date(r.neededAt).toLocaleString('ar-IQ')}</span>
                <div className="text-right text-sm">
                  <span className="font-bold text-violet-700">{r.requester?.name || '—'}</span>
                  <span className="mr-2 text-slate-600">يطلب {r.employees.length} موظف · {r.durationHours} ساعة</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ═══ طلبات إجازة مفتوحة — بضمنها الي انطاها موافقة أولية، حتى ما تنتسى ═══ */}
      {openLeaves.length > 0 && (
        <div className="rounded-2xl border-2 border-sky-200 bg-gradient-to-l from-sky-50 to-cyan-50 p-5 shadow-lg shadow-sky-100/50">
          <div className="mb-3 flex items-center justify-between">
            <button onClick={() => navigate('/leaves')} className="rounded-lg bg-sky-600 px-4 py-1.5 text-xs font-bold text-white transition hover:bg-sky-700">
              عرض الإجازات ←
            </button>
            <div className="flex items-center gap-3">
              <span className="relative flex h-3 w-3">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-sky-400 opacity-75" />
                <span className="relative inline-flex h-3 w-3 rounded-full bg-sky-500" />
              </span>
              <h3 className="text-base font-extrabold text-sky-800">
                طلبات إجازة تنتظر قرارك
                <span className="mr-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-sky-600 text-sm font-black text-white">{openLeaves.length}</span>
              </h3>
            </div>
          </div>
          <div className="space-y-2">
            {openLeaves.slice(0, 4).map(l => (
              <div
                key={l.id}
                onClick={() => navigate('/leaves')}
                className={`flex cursor-pointer items-center justify-between rounded-xl px-4 py-3 transition hover:shadow-md ${
                  l.status === 'PRELIMINARY' ? 'bg-amber-50 ring-1 ring-amber-300 hover:bg-amber-100' : 'bg-white/80 hover:bg-white'
                }`}
              >
                <span className="shrink-0 text-[11px] font-bold">
                  {l.status === 'PRELIMINARY' ? (
                    <span className="rounded-full bg-amber-500 px-2 py-0.5 text-white">موافقة أولية — ناقصها تأكيد</span>
                  ) : (
                    <span className="rounded-full bg-slate-200 px-2 py-0.5 text-slate-700">جديد</span>
                  )}
                </span>
                <div className="text-right text-sm">
                  <span className="font-bold text-sky-700">{l.employeeName}</span>
                  <span className="mr-2 text-slate-600">
                    {l.days} يوم · {new Date(l.startDate).toLocaleDateString('ar-IQ')} → {new Date(l.endDate).toLocaleDateString('ar-IQ')}
                  </span>
                </div>
              </div>
            ))}
            {openLeaves.length > 4 && (
              <div className="pt-1 text-center text-xs font-semibold text-sky-700">و{openLeaves.length - 4} طلب غيرها…</div>
            )}
          </div>
        </div>
      )}

      {/* ═══ Bookings Notification Panel ═══ */}
      {(['ADMIN', 'HR_COORDINATOR'].includes(employee.role) || permissions.includes('monitoring')) && (() => {
        const pending = bookings.filter(b => b.status === 'PENDING')
        const confirmed = bookings.filter(b => b.status === 'CONFIRMED')
        const inProgress = bookings.filter(b => b.status === 'IN_PROGRESS')
        if (pending.length === 0 && confirmed.length === 0 && inProgress.length === 0) return null
        return (
          <div className="space-y-4">
            {/* Pending - urgent notification */}
            {pending.length > 0 && (
              <div className="rounded-2xl border-2 border-amber-200 bg-gradient-to-l from-amber-50 to-orange-50 p-5 shadow-lg shadow-amber-100/50">
                <div className="mb-4 flex items-center justify-between">
                  <button onClick={() => navigate('/coordinator')} className="rounded-lg bg-amber-500 px-4 py-1.5 text-xs font-bold text-white transition hover:bg-amber-600">
                    تنسيق الحجوزات ←
                  </button>
                  <div className="flex items-center gap-3">
                    <span className="relative flex h-3 w-3">
                      <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-75" />
                      <span className="relative inline-flex h-3 w-3 rounded-full bg-amber-500" />
                    </span>
                    <h3 className="text-base font-extrabold text-amber-800">
                      حجوزات بانتظار التثبيت
                      <span className="mr-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-amber-500 text-sm font-black text-white">{pending.length}</span>
                    </h3>
                  </div>
                </div>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {pending.map(b => (
                    <div key={b.id} onClick={() => navigate('/coordinator')} className="flex cursor-pointer items-center justify-between rounded-xl bg-white/80 px-4 py-3 transition hover:bg-white hover:shadow-md">
                      <div className="flex items-center gap-2">
                        {b.priority === 'URGENT' && <span className="rounded-full bg-red-100 px-2 py-0.5 text-[10px] font-bold text-red-600">عاجل</span>}
                        <span className="text-xs text-slate-400">{new Date(b.createdAt).toLocaleDateString('ar-IQ')}</span>
                      </div>
                      <div className="flex items-center gap-3 text-right">
                        <div>
                          <span className="text-sm font-bold text-amber-700">{b.code}</span>
                          <span className="mr-2 text-sm font-medium text-slate-700">{b.customer?.name}</span>
                        </div>
                        {b.service && <span className="rounded-lg bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-700">{b.service.name}</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Confirmed + In Progress - summary row */}
            {(confirmed.length > 0 || inProgress.length > 0) && (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                {confirmed.length > 0 && (
                  <button onClick={() => navigate('/coordinator')} className="group flex items-center justify-between rounded-2xl border-2 border-blue-400 bg-gradient-to-l from-blue-600 to-blue-500 p-5 text-right shadow-lg shadow-blue-200 transition hover:shadow-xl hover:shadow-blue-300">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" className="opacity-60 group-hover:opacity-100"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <div className="flex items-center gap-4">
                      <div>
                        <span className="text-sm font-extrabold text-white">حجوزات مثبتة بحاجة تنسيق</span>
                        <p className="text-xs text-blue-100 mt-0.5">تعيين كوادر وموعد ومركبة</p>
                      </div>
                      <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-xl font-black text-blue-600 shadow-lg">{confirmed.length}</div>
                    </div>
                  </button>
                )}
                {inProgress.length > 0 && (
                  <button onClick={() => navigate('/bookings')} className="group flex items-center justify-between rounded-2xl border-2 border-violet-400 bg-gradient-to-l from-violet-600 to-violet-500 p-5 text-right shadow-lg shadow-violet-200 transition hover:shadow-xl hover:shadow-violet-300">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" className="opacity-60 group-hover:opacity-100"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <div className="flex items-center gap-4">
                      <div>
                        <span className="text-sm font-extrabold text-white">حجوزات جاري تنفيذها</span>
                        <p className="text-xs text-violet-100 mt-0.5">متابعة الكوادر الميدانية</p>
                      </div>
                      <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-xl font-black text-violet-600 shadow-lg">{inProgress.length}</div>
                    </div>
                  </button>
                )}
              </div>
            )}
          </div>
        )
      })()}

      {/* ═══ Monitor Overview Panel ═══ */}
      {permissions.includes('monitoring') && !isAdmin && (() => {
        // الأرقام من ملخّص السيرفر (محسوبة على الأرشيف الكامل)، والقوائم
        // من الشغل الحيّ ومن آخر ٢٠٠ منجز — للعرض بس.
        const inProgress = bookings.filter(b => b.status === 'IN_PROGRESS')
        const completed = completedBookings
        const todayCompleted = completed.filter(b => b.completedAt && new Date(b.completedAt).toDateString() === new Date().toDateString())
        const nPending = finance?.pendingCount ?? bookings.filter(b => b.status === 'PENDING').length
        const nConfirmed = finance?.confirmedCount ?? bookings.filter(b => b.status === 'CONFIRMED').length
        const nInProgress = finance?.inProgressCount ?? inProgress.length
        const nUnverified = finance?.unverifiedCount ?? 0
        const nVerified = finance?.verifiedCount ?? 0
        const nTodayCompleted = finance?.todayCompleted ?? todayCompleted.length
        const nActiveTechs = finance?.activeCrewCount ?? new Set(inProgress.flatMap(b => b.assignments.map(a => a.employee.id))).size
        const totalCollected = finance?.totalCollected ?? 0
        const nPendingExpenses = finance?.pendingExpenses ?? expenses.filter(e => e.status === 'PENDING').length

        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <button onClick={() => navigate('/monitor')} className="text-xs font-medium text-brand-500 hover:underline">لوحة المراقبة الكاملة ←</button>
              <h3 className="flex items-center gap-2 text-base font-extrabold text-brand-900">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                نظرة عامة للمراقب
              </h3>
            </div>

            {/* Stats grid */}
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <div className="rounded-2xl border-2 border-violet-200 bg-gradient-to-b from-violet-50 to-white p-4 text-center">
                <p className="text-3xl font-black text-violet-600">{nInProgress}</p>
                <p className="mt-1 text-xs font-medium text-violet-500">مهام قيد التنفيذ</p>
              </div>
              <div className="rounded-2xl border-2 border-blue-200 bg-gradient-to-b from-blue-50 to-white p-4 text-center">
                <p className="text-3xl font-black text-blue-600">{nActiveTechs}</p>
                <p className="mt-1 text-xs font-medium text-blue-500">فنيين في الميدان</p>
              </div>
              <div className="rounded-2xl border-2 border-amber-200 bg-gradient-to-b from-amber-50 to-white p-4 text-center">
                <p className="text-3xl font-black text-amber-600">{nUnverified}</p>
                <p className="mt-1 text-xs font-medium text-amber-500">بانتظار تدقيق مالي</p>
              </div>
              <div className="rounded-2xl border-2 border-emerald-200 bg-gradient-to-b from-emerald-50 to-white p-4 text-center">
                <p className="text-3xl font-black text-emerald-600">{nTodayCompleted}</p>
                <p className="mt-1 text-xs font-medium text-emerald-500">أنجزت اليوم</p>
              </div>
            </div>

            {/* Secondary stats */}
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <div className="rounded-xl bg-white p-3 text-center shadow-sm">
                <p className="text-xl font-bold text-brand-700">{nPending}</p>
                <p className="text-[10px] text-slate-400">بانتظار التثبيت</p>
              </div>
              <div className="rounded-xl bg-white p-3 text-center shadow-sm">
                <p className="text-xl font-bold text-brand-700">{nConfirmed}</p>
                <p className="text-[10px] text-slate-400">مثبتة (بحاجة تنسيق)</p>
              </div>
              <div className="rounded-xl bg-white p-3 text-center shadow-sm">
                <p className="text-xl font-bold text-emerald-600">{nVerified}</p>
                <p className="text-[10px] text-slate-400">تم تدقيقها</p>
              </div>
              <div className="rounded-xl bg-white p-3 text-center shadow-sm">
                <p className="text-xl font-bold text-brand-700">{totalCollected.toLocaleString()}</p>
                <p className="text-[10px] text-slate-400">إجمالي المحصّل (د.ع)</p>
              </div>
            </div>

            {/* Active crews list */}
            {inProgress.length > 0 && (
              <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
                <h4 className="mb-3 flex items-center gap-2 text-sm font-bold text-brand-900">
                  <span className="relative flex h-2.5 w-2.5"><span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-violet-400 opacity-75" /><span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-violet-500" /></span>
                  الكوادر النشطة الآن
                </h4>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {inProgress.map(b => (
                    <div key={b.id} className="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-2.5">
                      <div className="flex flex-wrap items-center gap-1">
                        {b.assignments.map(a => (
                          <span key={a.id} className="rounded-full bg-violet-100 px-2 py-0.5 text-[10px] font-bold text-violet-700">{a.employee.name}</span>
                        ))}
                      </div>
                      <div className="flex items-center gap-3 text-right">
                        <span className="text-sm font-bold text-brand-700">{b.code}</span>
                        <span className="text-xs text-slate-500">{b.customer?.name}</span>
                        {b.service && <span className="rounded-md bg-brand-50 px-2 py-0.5 text-[10px] text-brand-600">{b.service.name}</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Recent completed */}
            {todayCompleted.length > 0 && (
              <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
                <h4 className="mb-3 text-sm font-bold text-brand-900">المهام المنجزة اليوم</h4>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {todayCompleted.map(b => {
                    const cartTotal = (b.cartItems ?? []).reduce((s, c) => s + c.totalPrice, 0)
                    return (
                      <div key={b.id} className="flex items-center justify-between rounded-xl bg-emerald-50 px-4 py-2.5">
                        <div className="flex items-center gap-3">
                          <span className="rounded-full bg-emerald-200 px-2 py-0.5 text-[10px] font-bold text-emerald-700">
                            {b.amountVerified ? 'مدقق' : 'غير مدقق'}
                          </span>
                          {(b.amountCollected || 0) > 0 && (
                            <span className="text-xs font-bold text-emerald-700">{((b.amountCollected || 0) + (b.advancePaid || 0)).toLocaleString()} د.ع</span>
                          )}
                          {cartTotal > 0 && <span className="text-[10px] text-slate-400">مواد: {cartTotal.toLocaleString()}</span>}
                        </div>
                        <div className="flex items-center gap-2 text-right">
                          <span className="text-sm font-bold text-emerald-700">{b.code}</span>
                          <span className="text-xs text-slate-600">{b.customer?.name}</span>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Alerts row */}
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {nUnverified > 0 && (
                <button onClick={() => navigate('/finance')} className="rounded-2xl border-2 border-amber-300 bg-gradient-to-l from-amber-50 to-orange-50 p-4 text-right transition hover:shadow-lg">
                  <div className="flex items-center justify-between">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#d97706" strokeWidth="2.5" strokeLinecap="round"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <div className="flex items-center gap-3">
                      <div>
                        <span className="text-sm font-extrabold text-amber-800">{nUnverified} حجز بحاجة تدقيق</span>
                        <p className="text-xs text-amber-600 mt-0.5">مبالغ لم يتم التحقق منها</p>
                      </div>
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-200">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#d97706" strokeWidth="2" strokeLinecap="round"><path d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                      </div>
                    </div>
                  </div>
                </button>
              )}
              {nPendingExpenses > 0 && (
                <button onClick={() => navigate('/expenses')} className="rounded-2xl border-2 border-red-200 bg-gradient-to-l from-red-50 to-rose-50 p-4 text-right transition hover:shadow-lg">
                  <div className="flex items-center justify-between">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#dc2626" strokeWidth="2.5" strokeLinecap="round"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <div className="flex items-center gap-3">
                      <div>
                        <span className="text-sm font-extrabold text-red-800">{nPendingExpenses} مصروف بانتظار الموافقة</span>
                        <p className="text-xs text-red-600 mt-0.5">طلبات استرجاع من الفنيين</p>
                      </div>
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-red-200">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#dc2626" strokeWidth="2" strokeLinecap="round"><path d="M12 1v22M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6" /></svg>
                      </div>
                    </div>
                  </div>
                </button>
              )}
            </div>
          </div>
        )
      })()}

      {/* ═══ Accountant Dashboard Panel ═══ */}
      {employee.role === 'FINANCE' && (() => {
        // المجاميع تجي من السيرفر محسوبة على كل الأرشيف المنجز — دقيقة
        // بالكامل. القائمة المعروضة تحت (آخر ٨) تكفيها آخر دفعة نزلت.
        const completed = completedBookings
        const nUnverified = finance?.unverifiedCount ?? 0
        const nVerified = finance?.verifiedCount ?? 0
        const nCompleted = finance?.completedCount ?? completed.length
        const totalCollected = finance?.totalCollected ?? 0
        const totalQuoted = finance?.totalQuoted ?? 0
        const totalCartValue = finance?.totalCartValue ?? 0
        const nPendingExpenses = finance?.pendingExpenses ?? expenses.filter(e => e.status === 'PENDING').length
        const totalExpenses = finance?.totalExpenseValue ?? 0
        const recentCompleted = [...completed].sort((a, b) => new Date(b.completedAt || 0).getTime() - new Date(a.completedAt || 0).getTime()).slice(0, 8)

        return (
          <div className="space-y-4">
            <h3 className="flex items-center gap-2 text-base font-extrabold text-brand-900">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M12 1v22M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6" /></svg>
              لوحة المحاسب
            </h3>

            {/* ═══ التدقيق اليومي — أول شي يشوفه المحاسب ═══
                الأرقام كلها من حجوزات **منجزة** بس (السيرفر يفلترها)،
                فالي يظهر هنا شغل يكدر يشتغله فعلاً مو طابور وهمي. */}
            {dailyAudit && (
              <button
                type="button"
                onClick={() => navigate('/daily-audit')}
                className="w-full rounded-2xl border-2 border-brand-200 bg-gradient-to-l from-brand-50 to-white p-4 text-right transition hover:shadow-md"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <p className="font-extrabold text-brand-900">📅 التدقيق اليومي — {dailyAudit.date}</p>
                  <span className="text-xs font-bold text-brand-600">افتح ←</span>
                </div>
                <div className="mt-3 flex flex-wrap gap-2 text-xs">
                  <span className="rounded-full bg-amber-100 px-3 py-1 font-bold text-amber-800">
                    بانتظار التدقيق: {Math.max(0, dailyAudit.completedCount - dailyAudit.rows.filter((r) => r.amountVerified).length)}
                  </span>
                  <span className="rounded-full bg-emerald-100 px-3 py-1 font-bold text-emerald-800">
                    مدقق: {dailyAudit.rows.filter((r) => r.amountVerified).length}
                  </span>
                  <span className="rounded-full bg-slate-100 px-3 py-1 font-bold text-slate-600">
                    المحصّل: {dailyAudit.collectedTotal.toLocaleString()}
                  </span>
                  {dailyAudit.issuesCount > 0 && (
                    <span className="rounded-full bg-red-100 px-3 py-1 font-bold text-red-700">
                      محوّلة للرقابة: {dailyAudit.issuesCount}
                    </span>
                  )}
                </div>
                {dailyAudit.rows.length === 0 && (
                  <p className="mt-3 text-xs text-slate-500">ماكو حجز منجز اليوم — ماكو شي ينتدقق.</p>
                )}
              </button>
            )}

            {/* Financial KPIs */}
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <div className="rounded-2xl border-2 border-emerald-200 bg-gradient-to-b from-emerald-50 to-white p-4 text-center">
                <p className="text-2xl font-black text-emerald-600">{totalCollected.toLocaleString()}</p>
                <p className="mt-1 text-xs font-medium text-emerald-500">إجمالي المحصّل</p>
              </div>
              <div className="rounded-2xl border-2 border-blue-200 bg-gradient-to-b from-blue-50 to-white p-4 text-center">
                <p className="text-2xl font-black text-blue-600">{totalQuoted.toLocaleString()}</p>
                <p className="mt-1 text-xs font-medium text-blue-500">إجمالي التكاليف المقدرة</p>
              </div>
              <div className="rounded-2xl border-2 border-violet-200 bg-gradient-to-b from-violet-50 to-white p-4 text-center">
                <p className="text-2xl font-black text-violet-600">{totalCartValue.toLocaleString()}</p>
                <p className="mt-1 text-xs font-medium text-violet-500">قيمة المواد المستخدمة</p>
              </div>
              <div className="rounded-2xl border-2 border-rose-200 bg-gradient-to-b from-rose-50 to-white p-4 text-center">
                <p className="text-2xl font-black text-rose-600">{totalExpenses.toLocaleString()}</p>
                <p className="mt-1 text-xs font-medium text-rose-500">مصاريف الليدر المعتمدة</p>
              </div>
            </div>

            {/* Audit status */}
            <div className="grid grid-cols-3 gap-3">
              <button onClick={() => navigate('/finance')} className="rounded-xl bg-white p-3 text-center shadow-sm transition hover:shadow-md">
                <p className="text-xl font-bold text-brand-700">{nCompleted}</p>
                <p className="text-[10px] text-slate-400">إجمالي المنجزة</p>
              </button>
              <button onClick={() => navigate('/finance')} className="rounded-xl bg-amber-50 p-3 text-center shadow-sm transition hover:shadow-md">
                <p className="text-xl font-bold text-amber-600">{nUnverified}</p>
                <p className="text-[10px] text-amber-500">بانتظار التدقيق</p>
              </button>
              <button onClick={() => navigate('/finance')} className="rounded-xl bg-emerald-50 p-3 text-center shadow-sm transition hover:shadow-md">
                <p className="text-xl font-bold text-emerald-600">{nVerified}</p>
                <p className="text-[10px] text-emerald-500">تم التدقيق</p>
              </button>
            </div>

            {/* Pending expenses alert */}
            {nPendingExpenses > 0 && (
              <button onClick={() => navigate('/expenses')} className="w-full rounded-2xl border-2 border-red-200 bg-gradient-to-l from-red-50 to-rose-50 p-4 text-right transition hover:shadow-lg">
                <div className="flex items-center justify-between">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#dc2626" strokeWidth="2.5" strokeLinecap="round"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                  <div className="flex items-center gap-3">
                    <div>
                      <span className="text-sm font-extrabold text-red-800">{nPendingExpenses} طلب استرجاع بانتظار الموافقة</span>
                      <p className="text-xs text-red-600 mt-0.5">مصاريف من الفنيين تحتاج مراجعة</p>
                    </div>
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-red-200">
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#dc2626" strokeWidth="2" strokeLinecap="round"><path d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    </div>
                  </div>
                </div>
              </button>
            )}

            {/* Recent completed bookings with amounts */}
            {recentCompleted.length > 0 && (
              <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
                <div className="mb-3 flex items-center justify-between">
                  <button onClick={() => navigate('/finance')} className="text-xs font-medium text-brand-500 hover:underline">عرض الكل ←</button>
                  <h4 className="text-sm font-bold text-brand-900">آخر الحجوزات المنجزة</h4>
                </div>
                <div className="overflow-hidden rounded-lg border border-slate-200">
                  <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 text-slate-600">
                      <tr>
                        <th className="p-2 text-center font-medium">الحالة</th>
                        <th className="p-2 text-center font-medium">المواد</th>
                        <th className="p-2 text-center font-medium">المحصّل</th>
                        <th className="p-2 text-center font-medium">التكلفة</th>
                        <th className="p-2 text-start font-medium">الزبون</th>
                        <th className="p-2 text-start font-medium">الكود</th>
                      </tr>
                    </thead>
                    <tbody>
                      {recentCompleted.map(b => {
                        const cartTotal = (b.cartItems ?? []).reduce((s, c) => s + c.totalPrice, 0)
                        const collected = (b.amountCollected || 0) + (b.advancePaid || 0)
                        return (
                          <tr key={b.id} className="border-t border-slate-100 hover:bg-slate-50">
                            <td className="p-2 text-center">
                              <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${b.amountVerified ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'}`}>
                                {b.amountVerified ? 'مدقق' : 'بانتظار'}
                              </span>
                            </td>
                            <td className="p-2 text-center text-slate-600">{cartTotal > 0 ? cartTotal.toLocaleString() : '—'}</td>
                            <td className="p-2 text-center font-bold text-emerald-700">{collected > 0 ? collected.toLocaleString() : '—'}</td>
                            <td className="p-2 text-center text-slate-600">{b.quotedPrice ? b.quotedPrice.toLocaleString() : '—'}</td>
                            <td className="p-2 text-slate-700">{b.customer?.name}</td>
                            <td className="p-2 font-mono font-bold text-brand-600">{b.code}</td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </table>
                  </div>
                </div>
              </div>
            )}
          </div>
        )
      })()}

      {/* ═══ Sales Level Card ═══ */}
      {employee.role === 'SALES' && (
        <div className="rounded-2xl bg-white p-5 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]">
          <div className="flex items-center gap-4">
            <div className="relative">
              <ProgressRing percent={Math.min(100, (bookingCount % 10) * 10)} color="#10b981" />
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-lg font-extrabold text-emerald-600">{Math.floor(bookingCount / 10) + 1}</span>
              </div>
            </div>
            <div className="flex-1">
              <h3 className="text-sm font-bold text-[#0f2040]">مستوى المبيعات</h3>
              <p className="text-xs text-slate-400 mt-0.5">كل 10 حجوزات = مستوى جديد</p>
              <div className="mt-2 flex items-center gap-4">
                <div className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-emerald-500"/>
                  <span className="text-xs text-slate-500"><span className="font-bold text-[#0f2040]">{bookingCount}</span> حجز</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-blue-500"/>
                  <span className="text-xs text-slate-500"><span className="font-bold text-[#0f2040]">{10 - (bookingCount % 10)}</span> للمستوى التالي</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ═══ Compact Attendance Bar ═══ */}
      <div className="flex items-center justify-between rounded-xl bg-white px-4 py-2.5 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
        <div className="flex items-center gap-2">
          {activeRecord ? (
            <>
              <div className="relative">
                <button onClick={() => setShowCheckoutConfirm(!showCheckoutConfirm)}
                  className="rounded-lg bg-red-500 px-3 py-1 text-xs font-bold text-white hover:bg-red-600">
                  انصراف
                </button>
                {showCheckoutConfirm && (
                  <div className="absolute right-4 top-full z-[60] mt-2 w-44 rounded-xl border border-slate-100 bg-white p-3 shadow-2xl">
                    <p className="mb-2 text-xs font-semibold text-gray-800">تسجيل انصراف؟</p>
                    <div className="flex gap-2">
                      <button onClick={handleAttCheckOut} className="flex-1 rounded-lg bg-red-500 px-2 py-1 text-xs font-bold text-white">نعم</button>
                      <button onClick={() => setShowCheckoutConfirm(false)} className="flex-1 rounded-lg border border-slate-200 px-2 py-1 text-xs text-slate-600">إلغاء</button>
                    </div>
                  </div>
                )}
              </div>
            </>
          ) : (
            <button onClick={handleAttCheckIn}
              className="rounded-lg bg-emerald-500 px-3 py-1 text-xs font-bold text-white hover:bg-emerald-600">
              تسجيل حضور
            </button>
          )}
        </div>
        <div className="flex items-center gap-3 text-right">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>
          {activeRecord ? (
            <span className="flex items-center gap-2 text-xs">
              <span className="relative flex h-2 w-2"><span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" /><span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-400" /></span>
              <span className="font-bold text-emerald-700">متواجد</span>
              <span className="text-slate-400">منذ {fmtTime(activeRecord.checkIn)} • {elapsed}</span>
            </span>
          ) : (
            <span className="text-xs text-slate-400">{todayTotalMinutes > 0 ? 'خارج الدوام حالياً' : 'لم تسجل حضورك بعد'}</span>
          )}
          {todayTotalMinutes > 0 && (
            <span className="text-xs font-semibold text-[#0f2040]">
              دوامك اليوم: {Math.floor(todayTotalMinutes / 60)} ساعة و{todayTotalMinutes % 60} دقيقة
            </span>
          )}
        </div>
      </div>

      {/* ═══ شريط الفني: جرد أدواته + طلب إجازة ═══
          الفني ما إله شاشات إدارية يتنقل بيها، فالشغلتين الي يحتاجهن
          فعلاً — جرده وطلب إجازته — تطلعن له بالرئيسية مباشرة بدل ما
          يدور عليهن بالقائمة. */}
      {fieldOnly && (
        <div className="grid gap-3 sm:grid-cols-2">
          <button
            onClick={() => navigate('/my-inventory')}
            className={`flex items-center justify-between gap-3 rounded-2xl p-4 text-right shadow-[0_1px_3px_rgba(0,0,0,0.04)] transition hover:shadow-md ${
              inventoryDue ? 'bg-amber-50 ring-1 ring-amber-300' : 'bg-white'
            }`}
          >
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke={inventoryDue ? '#d97706' : '#64748b'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
              <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
            </svg>
            <div className="flex-1">
              <p className={`text-sm font-extrabold ${inventoryDue ? 'text-amber-900' : 'text-brand-900'}`}>
                {inventoryDue ? 'حان وقت جرد أدواتك' : 'جرد أدواتي'}
              </p>
              <p className={`text-xs ${inventoryDue ? 'text-amber-700' : 'text-slate-400'}`}>{inventoryHint}</p>
            </div>
          </button>

          <button
            onClick={() => navigate('/leaves')}
            className="flex items-center justify-between gap-3 rounded-2xl bg-white p-4 text-right shadow-[0_1px_3px_rgba(0,0,0,0.04)] transition hover:shadow-md"
          >
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#64748b" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
              <rect x="3" y="4" width="18" height="18" rx="2" /><path d="M16 2v4M8 2v4M3 10h18" /><path d="m9 16 2 2 4-4" />
            </svg>
            <div className="flex-1">
              <p className="text-sm font-extrabold text-brand-900">تقديم طلب إجازة</p>
              <p className="text-xs text-slate-400">محتاج إجازة؟ قدّم طلبك من هنا</p>
            </div>
          </button>
        </div>
      )}

      {/* ═══ My Tasks Panel (Technician/Leader) ═══ */}
      {(employee.role === 'TECHNICIAN' || employee.role === 'TECHNICAL' || employee.isLeader) && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <button onClick={() => navigate('/my-tasks')} className="text-xs font-medium text-brand-500 hover:underline">عرض الكل ←</button>
            <h3 className="flex items-center gap-2 text-base font-extrabold text-brand-900">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M9 11l3 3L22 4" /><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" /></svg>
              مهامي
              {myTasks.length > 0 && <span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-brand-600 text-xs font-black text-white">{myTasks.length}</span>}
            </h3>
          </div>

          {myTasks.length === 0 ? (
            <div className="rounded-2xl bg-white p-8 text-center shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" strokeWidth="1.5" strokeLinecap="round" className="mx-auto mb-3">
                <path d="M9 11l3 3L22 4" /><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" />
              </svg>
              <p className="text-sm text-slate-400">لا توجد مهام حالياً — أحسنت!</p>
            </div>
          ) : (
            <div className="space-y-3">
              {myTasks.map(b => {
                const myRole = b.assignments.find(a => a.employee.id === employee?.id)?.role
                return (
                  <div key={b.id} className="overflow-hidden rounded-2xl bg-white shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
                    {/* Task header */}
                    <div className={`flex items-center justify-between px-5 py-3 ${b.status === 'IN_PROGRESS' ? 'bg-gradient-to-l from-violet-600 to-violet-700' : 'bg-gradient-to-l from-amber-500 to-amber-600'}`}>
                      <div className="flex items-center gap-2">
                        {b.priority === 'URGENT' && <span className="rounded-full bg-red-500 px-2 py-0.5 text-[10px] font-bold text-white">عاجل</span>}
                        <span className="rounded-full bg-white/20 px-2 py-0.5 text-[11px] font-bold text-white">
                          {myRole === 'TECH_1' ? 'الفني الأول' : myRole === 'TECH_2' ? 'الفني الثاني' : 'الفني الثالث'}
                        </span>
                      </div>
                      <div className="flex items-center gap-3 text-white">
                        <span className="text-sm font-bold">{b.code}</span>
                        <span className="text-xs">{b.status === 'IN_PROGRESS' ? '🔄 جاري التنفيذ' : '📋 بانتظار الاستلام'}</span>
                      </div>
                    </div>

                    <div className="p-5">
                      {/* Customer & service info */}
                      <div className="mb-3 grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                        <div className="text-right">
                          <span className="text-xs text-slate-400">الزبون</span>
                          <p className="font-bold text-brand-900">{b.customer?.name}</p>
                        </div>
                        <div className="text-right">
                          <span className="text-xs text-slate-400">الخدمة</span>
                          <p className="font-medium text-brand-800">{b.service?.name || '—'}</p>
                        </div>
                        <div className="text-right">
                          <span className="text-xs text-slate-400">الهاتف</span>
                          <p className="font-medium text-slate-700" dir="ltr">{b.customer?.phone}</p>
                        </div>
                        <div className="text-right">
                          <span className="text-xs text-slate-400">العنوان</span>
                          <p className="font-medium text-slate-700">{b.address || b.customer?.location || '—'}</p>
                        </div>
                        {b.assignedVehicle && (
                          <div className="text-right">
                            <span className="text-xs text-slate-400">السيارة</span>
                            <p className="font-medium text-slate-700">{b.assignedVehicle}</p>
                          </div>
                        )}
                        {b.quotedPrice != null && (
                          <div className="text-right">
                            <span className="text-xs text-slate-400">التكلفة</span>
                            <p className="font-bold text-emerald-700">{b.quotedPrice.toLocaleString()} د.ع</p>
                          </div>
                        )}
                      </div>

                      {b.notes && <p className="mb-3 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-600">{b.notes}</p>}

                      {/* الخريطة تنفتح داخل النظام نفسه (نافذة عرض)، مو برابط خارجي بتاب جديد */}
                      {b.mapLatitude && b.mapLongitude && (
                        <button onClick={() => setMapTask(b)}
                          className="mb-3 flex w-full items-center justify-center gap-2 rounded-lg bg-blue-50 px-3 py-2 text-sm font-bold text-blue-700 hover:bg-blue-100">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z" /><circle cx="12" cy="10" r="3" /></svg>
                          عرض الموقع على الخريطة
                        </button>
                      )}

                      {/* Actions */}
                      {b.status === 'CONFIRMED' ? (
                        <button onClick={() => handleTaskStart(b)}
                          className="w-full rounded-xl bg-gradient-to-l from-amber-500 to-amber-600 px-4 py-3 text-sm font-bold text-white shadow-lg shadow-amber-500/20 hover:shadow-xl">
                          ✅ تم الاستلام — بدأت بالعمل
                        </button>
                      ) : (
                        <div className="space-y-2">
                          <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                            <input type="number" placeholder="المبلغ المستلم" value={taskAmounts[b.id] || ''}
                              onChange={e => setTaskAmounts(p => ({ ...p, [b.id]: e.target.value }))}
                              className="rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none focus:border-brand-500" />
                            <input type="number" placeholder="دفعة مقدمة" value={taskAdvances[b.id] || ''}
                              onChange={e => setTaskAdvances(p => ({ ...p, [b.id]: e.target.value }))}
                              className="rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none focus:border-brand-500" />
                            <input placeholder="ملاحظات" value={taskNotes[b.id] || ''}
                              onChange={e => setTaskNotes(p => ({ ...p, [b.id]: e.target.value }))}
                              className="rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none focus:border-brand-500" />
                            <button onClick={() => handleTaskComplete(b)}
                              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-bold text-white shadow-md hover:shadow-lg">
                              تم الإنجاز
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* ═══ نبض اليوم ═══
          الإداري يفتح النظام الصبح ولازم يعرف وضع يومه بنظرة وحدة:
          كم حجز اليوم، منو بالميدان، شنو ينتظر شغله، وشنو متأخر.
          قبل، كل رقم منهن يحتاج يفتح شاشة منفصلة يدور بيها.

          كل بطاقة تودّي لشاشتها بضغطة — الرقم بلا طريق يوصلك له
          يخلّي الإداري يشوف المشكلة وما يقدر يتصرّف. */}
      {pulse && (isAdmin || employee.role === 'HR_COORDINATOR' || permissions.includes('coordinator')) && (
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <PulseCard
            label="حجوزات اليوم" value={pulse.todayBookings} icon="📅" tone="blue"
            compare={pulse.yesterdayBookings} compareLabel="أمس"
            onClick={() => navigate('/bookings')}
          />
          <PulseCard
            label="تحتاج تنسيق" value={pulse.needsCoordination} icon="⏳" tone="amber"
            hint="مثبّتة بلا موعد أو كادر"
            onClick={() => navigate('/coordinator')}
          />
          <PulseCard
            label="بالميدان الآن" value={pulse.crewInField} icon="🚗" tone="emerald"
            hint={`${pulse.openMissions} مهمة مفتوحة`}
            onClick={() => navigate('/missions')}
          />
          <PulseCard
            label="شكاوى جديدة" value={pulse.newComplaints} icon="⚠️" tone="red"
            hint={pulse.overdueMissions > 0 ? `+${pulse.overdueMissions} مهمة متأخرة` : undefined}
            onClick={() => navigate('/complaints')}
          />
        </div>
      )}

      {/* ═══ Quick Access ═══ */}
      {quickCards.length > 0 && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {quickCards.map((card) => (
            <button key={card.path} onClick={() => navigate(card.path)}
              className={`group relative overflow-hidden rounded-2xl bg-gradient-to-bl ${card.gradient} p-5 text-right text-white shadow-lg transition-all duration-300 hover:shadow-2xl hover:-translate-y-1`}>
              <div className="absolute inset-0 bg-gradient-to-l from-white/0 via-white/10 to-white/0 translate-x-full transition-transform duration-700 group-hover:-translate-x-full" />
              <div className="relative flex items-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur-sm">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d={card.iconPath} /></svg>
                </div>
                <div className="flex-1">
                  <p className="text-base font-bold">{card.title}</p>
                  <p className="mt-0.5 text-xs text-white/70">{card.desc}</p>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}

      {/* ═══ Projects Panel — مدير المشاريع (نظرة سريعة على مراحل المشاريع من دون فتح الصفحة) ═══ */}
      {!isAdmin && employee.role === 'PROJECT_MANAGER' && projectStats && (
        <div className="grid grid-cols-1 gap-4">
          <SystemPanel title="المشاريع" color="#8b5cf6" dotColor="bg-violet-400" actionLabel="عرض الكل" onAction={() => navigate('/projects')}>
            <div className="grid grid-cols-3 gap-3 sm:grid-cols-7">
              <MiniKPI label="اتصال" value={projectStats['اتصال'] || 0} color="#3b82f6" />
              <MiniKPI label="كشف" value={projectStats['كشف'] || 0} color="#10b981" />
              <MiniKPI label="سعر" value={projectStats['سعر'] || 0} color="#f59e0b" />
              <MiniKPI label="تنفيذ" value={projectStats['تنفيذ'] || 0} color="#ef4444" />
              <MiniKPI label="مكتمل" value={projectStats['مكتمل'] || 0} color="#2563eb" />
              <MiniKPI label="مرفوض" value={projectStats['مرفوض'] || 0} color="#6b7280" />
              <MiniKPI label="حجوزات محولة" value={bookings.filter(b => b.transferToProjects && b.status !== 'COMPLETED' && b.status !== 'CANCELLED').length} color="#8b5cf6" />
            </div>
          </SystemPanel>
        </div>
      )}

      {/* ═══ GPS Panel ═══
          مسؤول الجي بي اس يشوفها دائماً. أما الفني فلازم يجتمع عنده
          شرطين: مهارة الجي بي اس *و* صلاحية gps_system الي ينطيها
          المدير بيده. قبل، مجرد تأشير المهارة كان يفتح اللوحة —
          فالفني العادي يلكه لوحة جي بي اس بلوحته وهي ما تخصه. */}
      {!isAdmin && (employee.role === 'GPS_ADMIN'
        || (employee.role === 'TECHNICIAN' && hasGpsSkill(employee, gpsServiceId) && permissions.includes('gps_system'))) && (
        <div className="grid grid-cols-1 gap-4">
          <SystemPanel title="نظام GPS" color="#f59e0b" dotColor="bg-amber-400" actionLabel="عرض الكل" onAction={() => navigate('/gps')}>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <MiniKPI label="الأجهزة" value={gpsStats?.totalDevices || 0} color="#10b981" />
              <MiniKPI label="المشتركين" value={gpsStats?.totalCustomers || 0} color="#3b82f6" />
              <MiniKPI label="شرائح SIM" value={gpsStats?.totalSims || 0} color="#8b5cf6" />
              <MiniKPI label="صيانة معلقة" value={pendingMaintenance} color={pendingMaintenance > 0 ? '#ef4444' : '#f59e0b'} />
            </div>
          </SystemPanel>
        </div>
      )}

      {/* ═══ System Panels - ADMIN ═══ */}
      {isAdmin && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {/* GPS Panel */}
          <SystemPanel title="نظام GPS" color="#f59e0b" dotColor="bg-amber-400" actionLabel="عرض الكل" onAction={() => navigate('/gps')}>
            <div className="grid grid-cols-2 gap-3">
              <MiniKPI label="الأجهزة" value={gpsStats?.totalDevices || 0} color="#10b981" />
              <MiniKPI label="المشتركين" value={gpsStats?.totalCustomers || 0} color="#3b82f6" />
              <MiniKPI label="شرائح SIM" value={gpsStats?.totalSims || 0} color="#8b5cf6" />
              <MiniKPI label="صيانة معلقة" value={pendingMaintenance} color={pendingMaintenance > 0 ? '#ef4444' : '#f59e0b'} />
            </div>
          </SystemPanel>

          {/* Sales Panel */}
          <SystemPanel title="المبيعات والمالية" color="#3b82f6" dotColor="bg-blue-400" actionLabel="عرض الكل" onAction={() => navigate('/finance')}>
            <div className="grid grid-cols-2 gap-3">
              <MiniKPI label="الحجوزات" value={bookingCount} color="#3b82f6" />
              <MiniKPI label="العملاء" value={customerCount} color="#10b981" />
              <MiniKPI label="عروض الأسعار" value={0} color="#8b5cf6" />
              <MiniKPI label="المعاملات" value={0} color="#f59e0b" />
            </div>
          </SystemPanel>

          {/* HR Panel */}
          <SystemPanel title="الموارد البشرية" color="#10b981" dotColor="bg-emerald-400" actionLabel="عرض الكل" onAction={() => navigate('/employees')}>
            <div className="grid grid-cols-2 gap-3">
              <MiniKPI label="الموظفين" value={employeeCount} color="#3b82f6" />
              <MiniKPI label="الشكاوى" value={0} color="#ef4444" />
            </div>
          </SystemPanel>

          {/* Activity Panel */}
          <div className="rounded-2xl bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]">
            <h3 className="mb-4 flex items-center gap-2 text-right text-sm font-bold text-slate-700">
              <span className="h-2 w-2 rounded-full bg-red-400" />
              آخر النشاطات
            </h3>
            <div className="flex h-28 items-center justify-center rounded-xl border-2 border-dashed border-slate-100">
              <div className="text-center">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" strokeWidth="1.5" strokeLinecap="round" className="mx-auto mb-2">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                <p className="text-xs text-slate-300">لا توجد نشاطات حديثة</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {mapTask && mapTask.mapLatitude != null && mapTask.mapLongitude != null && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => setMapTask(null)}>
          <div className="w-full max-w-2xl rounded-2xl bg-white p-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-3 flex items-center justify-between">
              <h3 className="font-bold text-brand-900">موقع المهمة {mapTask.code}</h3>
              <button onClick={() => setMapTask(null)} className="rounded-lg px-3 py-1 text-sm text-slate-500 hover:bg-slate-100">✕ إغلاق</button>
            </div>
            <MapViewer lat={mapTask.mapLatitude} lng={mapTask.mapLongitude} height={380} />
          </div>
        </div>
      )}
    </div>
  )
}

/* ═══ Sub-components ═══ */

function SystemPanel({ title, dotColor, actionLabel, onAction, children }: {
  title: string; color: string; dotColor: string; actionLabel: string; onAction: () => void; children: React.ReactNode
}) {
  return (
    <div className="rounded-2xl bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]">
      <div className="mb-4 flex items-center justify-between">
        <button onClick={onAction}
          className="rounded-lg bg-slate-50 px-3 py-1.5 text-xs font-bold text-slate-500 transition hover:bg-brand-50 hover:text-brand-600">
          {actionLabel} ←
        </button>
        <h3 className="flex items-center gap-2 text-sm font-bold text-slate-700">
          <span className={`h-2 w-2 rounded-full ${dotColor}`} />
          {title}
        </h3>
      </div>
      {children}
    </div>
  )
}

// ═══ بطاقة نبض ═══
//
// الرقم كبير ويُقرا من بعيد، واللون يگول الحالة بلا ما تقرا.
//
// ⚠️ الصفر ما ينعرض بلون تحذير: «٠ شكاوى» خبر زين، وتلوينه أحمر
// يخلي الإداري يتوتر على شي مو موجود. اللون ينشتغل بس لو الرقم > 0.
function PulseCard({
  label, value, icon, tone, hint, compare, compareLabel, onClick,
}: {
  label: string
  value: number
  icon: string
  tone: 'blue' | 'amber' | 'emerald' | 'red'
  hint?: string
  compare?: number
  compareLabel?: string
  onClick: () => void
}) {
  const active = value > 0
  const tones: Record<string, { ring: string; text: string; bg: string }> = {
    blue:    { ring: 'border-sky-200',     text: 'text-sky-700',     bg: 'bg-sky-50' },
    amber:   { ring: 'border-amber-300',   text: 'text-amber-700',   bg: 'bg-amber-50' },
    emerald: { ring: 'border-emerald-200', text: 'text-emerald-700', bg: 'bg-emerald-50' },
    red:     { ring: 'border-red-300',     text: 'text-red-700',     bg: 'bg-red-50' },
  }
  const t = active ? tones[tone] : { ring: 'border-slate-200', text: 'text-slate-500', bg: 'bg-white' }

  // فرق اليوم عن أمس — الاتجاه يفيد أكثر من الرقم لحاله
  let delta: { text: string; up: boolean } | null = null
  if (compare !== undefined && compare > 0) {
    const diff = Math.round(((value - compare) / compare) * 100)
    if (diff !== 0) delta = { text: `${Math.abs(diff)}٪ عن ${compareLabel}`, up: diff > 0 }
  }

  return (
    <button
      onClick={onClick}
      className={`rounded-2xl border-2 ${t.ring} ${t.bg} p-4 text-right transition-all hover:-translate-y-0.5 hover:shadow-md`}
    >
      <div className="flex items-start justify-between gap-2">
        <span className="text-xl">{icon}</span>
        <span className={`text-3xl font-black tabular-nums ${t.text}`}>{value}</span>
      </div>
      <p className="mt-1.5 text-xs font-bold text-slate-700">{label}</p>
      {hint && <p className="mt-0.5 text-[11px] text-slate-500">{hint}</p>}
      {delta && (
        <p className={`mt-0.5 text-[11px] font-bold ${delta.up ? 'text-emerald-600' : 'text-slate-500'}`}>
          {delta.up ? '▲' : '▼'} {delta.text}
        </p>
      )}
    </button>
  )
}

function MiniKPI({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="group flex items-center justify-between rounded-xl border border-slate-100 p-3.5 transition-colors hover:border-slate-200 hover:bg-slate-50/50">
      <div className="flex h-10 w-10 items-center justify-center rounded-lg" style={{ backgroundColor: color + '12' }}>
        <span className="text-lg font-black" style={{ color }}>{value}</span>
      </div>
      <p className="text-xs font-medium text-slate-500">{label}</p>
    </div>
  )
}
