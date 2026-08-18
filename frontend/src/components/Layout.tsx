import { useState, useEffect, useRef } from 'react'
import { useAutoRefresh } from '../useAutoRefresh'
import PrivacyPolicyGate from './PrivacyPolicyGate'
import EmployeeAvatar from './EmployeeAvatar'
import LiveAlerts from './LiveAlerts'
import ThemeToggle from './ThemeToggle'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { api, type Employee } from '../api'
import { SessionContext, roleLabels } from '../session'
import { ensureFileToken } from '../api'
import Login from '../pages/Login'
import CommandApp from '../command/CommandApp'
import TrainingPage from '../pages/TrainingPage'
import AssistantWidget from './AssistantWidget'
import ManagerAssistantChat from './ManagerAssistantChat'
import SettingsPanel from './SettingsPanel'

// وجهة كل نوع إشعار: ضغطة على الإشعار توديك للشاشة الي تخصه بدل ما
// تدوّر عليها. نوع مو موجود بالخريطة = الإشعار يتأشر مقروء وبس.
const notifTargets: Record<string, string> = {
  leave_request: '/leaves',
  leave_preliminary: '/leaves',
  leave_decision: '/leaves',
  booking_delete_request: '/booking-delete-requests',
  booking_confirmed: '/bookings',
  booking_returned_to_crew: '/bookings',
  audit_issue: '/audit-issues',
  kpi_deduction: '/kpi',
  kpi_leaderboard: '/kpi',
  authz_violation: '/owner-security',
  job_duration_overrun: '/missions',
  // ═══ قرارات الطلبات ═══
  // الإشعار بلا وجهة يخلي الموظف يقرا «انوافق على طلبك» ويضغط عليه
  // وما يصير شي — فيضطر يدوّر على الشاشة بنفسه.
  tool_request_decision: '/my-inventory',
  procurement_decision: '/procurement',
  staff_request_decision: '/staff-requests',
  booking_delete_decision: '/bookings',
  employee_letter_decision: '/letters',
  extra_task: '/my-extra-tasks',
  extra_task_done: '/extra-tasks',
  extra_task_cancelled: '/my-extra-tasks',
}
import AnnouncementTicker from './AnnouncementTicker'
import { navItems, isNavVisible, type NavItem } from './navTree'


const loadStoredEmployee = (): Employee | null => {
  const raw = localStorage.getItem('currentEmployee')
  if (!raw) return null
  try { return JSON.parse(raw) as Employee } catch { return null }
}

function hasActiveChild(item: NavItem, pathname: string): boolean {
  if (!item.children) return pathname === item.to || pathname.startsWith(item.to + '/')
  return item.children.some(c => hasActiveChild(c, pathname))
}

const roleColors: Record<string, string> = {
  ADMIN: 'from-amber-500 to-orange-600',
  SALES: 'from-emerald-500 to-teal-600',
  HR_COORDINATOR: 'from-violet-500 to-purple-600',
  TECHNICIAN: 'from-sky-500 to-blue-600',
  PROJECT_MANAGER: 'from-rose-500 to-pink-600',
  MONITOR: 'from-cyan-500 to-teal-600',
  FINANCE: 'from-lime-500 to-green-600',
  GPS_ADMIN: 'from-indigo-500 to-blue-600',
  QUALITY_ENGINEER: 'from-fuchsia-500 to-purple-600',
  ENGINEER: 'from-teal-500 to-cyan-700',
}

// ═══ منو يشوف شنو — مصدر وحيد ═══
//
// نقية بلا حالة، حتى نقدر نسألها «شنو راح يشوفه فلان؟» بدون ما نسجّل
// دخوله. شاشة «شوف بعين الموظف» تناديها بنفسها، فما تنحرف عن الواقع
// أبداً — نفس الدالة الي ترسم القائمة الحقيقية.
export default function Layout() {
  const [employee, setEmployeeState] = useState<Employee | null>(loadStoredEmployee)
  const [employeePermissions, setEmployeePermissions] = useState<string[]>([])
  const [gpsServiceId, setGpsServiceId] = useState<string | null>(null)
  // هل عند الموظف المتدرب مواد فعلاً؟ null = لسّه ما فحصنا.
  // بدونها ينحبس بشاشة تدريب فارغة ما بيها شي يدرسه.
  const [traineeHasMaterials, setTraineeHasMaterials] = useState<boolean | null>(null)
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({})
  const [collapsed, setCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [notifications, setNotifications] = useState<import('../api').Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [notifOpen, setNotifOpen] = useState(false)
  const notifPanelRef = useRef<HTMLDivElement>(null)
  const notifBtnRef = useRef<HTMLButtonElement>(null)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()

  // Closing the mobile nav on route change is a one-line UI reset tied to router
  // navigation, not data fetching; safe to keep as a synchronous effect.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { setMobileOpen(false) }, [location.pathname])

  // وسم عرض الملفات: ينجاب مرة عند الدخول ويتجدد كل 10 دقائق. بدونه
  // الصور المخزّنة برّا القاعدة ترجع 401 لأن <img> ما يرسل ترويسة
  // Authorization.
  useEffect(() => {
    if (!employee) return
    void ensureFileToken()
    const timer = setInterval(() => { void ensureFileToken() }, 10 * 60 * 1000)
    return () => clearInterval(timer)
  }, [employee])

  // نفحص إذا اكو مواد تدريبية منشورة لهذا المتدرب. ما اكو مواد = ما
  // اكو تدريب، فيدخل النظام عادي بدل ما ينحبس بشاشة فارغة.
  useEffect(() => {
    // Resetting the flag when the trainee state changes is derived-state sync, not a fetch.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee?.isTrainee) { setTraineeHasMaterials(null); return }
    let alive = true
    api.getMyTraining(employee.id)
      .then((res) => { if (alive) setTraineeHasMaterials((res?.materials?.length ?? 0) > 0) })
      // فشل الطلب ما يصير يحبس الموظف — نخليه يدخل عادي
      .catch(() => { if (alive) setTraineeHasMaterials(false) })
    return () => { alive = false }
  }, [employee])

  // تحديث تلقائي كل نص ساعة + معالجة رجوع الموظف للتبويب بعد غياب طويل
  useAutoRefresh(!!employee)

  const toggleSidebar = () => {
    if (typeof window !== 'undefined' && window.innerWidth < 1024) setMobileOpen((o) => !o)
    else setCollapsed((c) => !c)
  }

  const setEmployee = (emp: Employee | null) => {
    // المالك (OWNER) لازم يشوف ويسوي كل شي مدير النظام (ADMIN) يسويه، وأكثر —
    // نطبّع role إلى 'ADMIN' حتى كل شرط role === 'ADMIN' بالواجهة يشتغل له
    // تلقائياً بدون تعديل كل مكان، ونحفظ الدور الحقيقي بـ actualRole للعرض
    // وللتحقق الحصري بصفحة المراقبة الخلفية.
    const normalized = emp && emp.role === 'OWNER' ? { ...emp, actualRole: 'OWNER' as const, role: 'ADMIN' as const } : emp
    setEmployeeState(normalized)
    if (normalized) {
      localStorage.setItem('currentEmployee', JSON.stringify(normalized))
    } else {
      localStorage.removeItem('currentEmployee')
      localStorage.removeItem('authToken')
    }
  }


  // نتحقق من هوية الموظف الحقيقية من السيرفر مرة وحدة عند فتح النظام —
  // هذا يصحح تلقائياً أي بيانات جلسة قديمة/معدَّلة (مثلاً بأدوات المطورين
  // بالمتصفح) بقيت محفوظة بذاكرة المتصفح المحلية من قبل، ويسجل خروج
  // الحساب تلقائياً إذا صار موقوف (SUSPENDED) بينما الجلسة القديمة لسه مفتوحة
  useEffect(() => {
    if (!employee) return
    api.getMe()
      .then((fresh) => setEmployee(fresh))
      .catch(() => setEmployee(null))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    // Clearing notifications when the user logs out is a guard-clause reset of local
    // state tied to the `employee` dependency, not a fetch; safe to keep as-is.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setNotifications([]); setUnreadCount(0); return }
    const load = () => api.getNotifications().then((r) => { setNotifications(r.notifications); setUnreadCount(r.unreadCount) }).catch(() => {})
    load()
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [employee])

  // ═══ إغلاق قائمة الإشعارات ═══
  //
  // ⚠️ بمستمع على المستند مو بحجاب شفاف فوگ الشاشة. الحجاب أسهل
  // بالكتابة بس يبلع كل تفاعل: التمرير ينضرب بيه فالصفحة تتجمّد،
  // وضغطة الجرس نفسه ما توصله.
  //
  // ⚠️ و`pointerdown` مو `click`: لو الموظف ضغط على زر ثاني بالصفحة،
  // الإغلاق يصير قبل ما ينفّذ الزر — فما تنسد الطريق على ضغطة وحدة.
  //
  // ⚠️ والتمرير يغلقها بعد: القائمة معلّقة بمكان الجرس، فلو الصفحة
  // تحركت وهي مفتوحة تبقى طايفة بالفراغ فوگ محتوى ما يخصّها.
  useEffect(() => {
    if (!notifOpen) return
    const close = () => setNotifOpen(false)
    const onDown = (e: PointerEvent) => {
      const t = e.target as Node
      if (notifPanelRef.current?.contains(t)) return
      if (notifBtnRef.current?.contains(t)) return   // الجرس يبدّل بنفسه
      close()
    }
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') close() }
    document.addEventListener('pointerdown', onDown)
    window.addEventListener('keydown', onKey)
    // `true` = مرحلة الالتقاط: التمرير داخل `main` ما يوصل `window` بالفقاعة
    document.addEventListener('scroll', close, true)
    return () => {
      document.removeEventListener('pointerdown', onDown)
      window.removeEventListener('keydown', onKey)
      document.removeEventListener('scroll', close, true)
    }
  }, [notifOpen])

  const handleNotifClick = async (n: import('../api').Notification) => {
    if (!n.read) {
      try {
        await api.markNotificationRead(n.id)
        setNotifications((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: true } : x)))
        setUnreadCount((c) => Math.max(0, c - 1))
      } catch { /* ignore */ }
    }
    // الإشعار بلا وجهة يخلي الموظف يدوّر بيده على الشاشة الي تخص الخبر —
    // وأحياناً ما يلكها أصلاً. كل نوع يوديه لمحله مباشرة.
    const target = notifTargets[n.type]
    if (target) {
      setNotifOpen(false)
      navigate(target)
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await api.markAllNotificationsRead()
      setNotifications((prev) => prev.map((x) => ({ ...x, read: true })))
      setUnreadCount(0)
    } catch { /* ignore */ }
  }

  useEffect(() => {
    // Guard-clause reset when logged out; not part of the fetch itself.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setEmployeePermissions([]); return }
    api.getEmployeePermissions(employee.id)
      .then((perms) => setEmployeePermissions(perms.map((p) => p.name)))
      .catch(() => setEmployeePermissions([]))
  }, [employee])

  useEffect(() => {
    // Guard-clause reset when logged out; not part of the fetch itself.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setGpsServiceId(null); return }
    api.getServices()
      .then((services) => setGpsServiceId(services.find((s) => s.name === 'GPS')?.id || null))
      .catch(() => setGpsServiceId(null))
  }, [employee])

  useEffect(() => {
    const autoExpand = (items: NavItem[]) => {
      items.forEach((item) => {
        if (item.children && hasActiveChild(item, location.pathname)) {
          setExpandedGroups((prev) => ({ ...prev, [item.label]: true }))
          autoExpand(item.children)
        }
      })
    }
    autoExpand(navItems)
  }, [location.pathname])

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  // ═══ مركز القيادة ═══
  // نفس اليوزر دخل بباسورد ثاني → نظام ثاني بالكامل (فكرة PPSK).
  // ⚠️ الفصل الحقيقي بالسيرفر (توكن القيادة ينرفض على مسارات
  // الموظفين) — هذا الشرط للعرض بس، مو حماية.
  if (api.currentRealm() === 'command') {
    return (
      <CommandApp
        onExit={() => {
          localStorage.removeItem('authToken')
          localStorage.removeItem('authRealm')
          window.location.reload()
        }}
      />
    )
  }

  // موظف قيد التدريب: يشوف صفحة التدريب فقط، بدون أي وصول لباقي النظام.
  //
  // بس بشرط: لازم يكون المدير ناشر مواد تدريبية فعلاً. قبل، مجرد ما
  // ينتأشر الموظف «متدرب» ينحبس بشاشة مكتوب بيها «لا توجد مواد تدريبية
  // لهذه الخدمة بعد» — ما يشوف مهامه ولا يسجّل دوام ولا يسوي ولا شي،
  // وما عنده حتى شنو يدرس. حبس بلا فايدة.
  //
  // هسه: ما اكو مواد = ما اكو تدريب، والموظف يدخل النظام عادي. الحبس
  // يشتغل بس لمن المدير يحدد المواد — وهو صاحب القرار.
  if (employee.isTrainee && traineeHasMaterials === true) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
        <div dir="rtl" className="min-h-screen bg-[#f0f4f9]">
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-8">
            <span className="text-lg font-extrabold text-[#0f2040] tracking-tight">نظام شركة الأماني — التدريب</span>
            <button onClick={() => setEmployee(null)}
              className="rounded-lg px-3 py-1.5 text-sm font-bold text-red-500 hover:bg-red-50">
              تسجيل الخروج
            </button>
          </header>
          <main className="p-3 sm:p-5 lg:p-8">
            <TrainingPage />
          </main>
        </div>
      </SessionContext.Provider>
    )
  }

  const role = employee?.role
  const isVisible = (item: NavItem, unitGranted = false): boolean =>
    isNavVisible(item, { employee, permissions: employeePermissions, gpsServiceId }, unitGranted)
  // الفني العادي (مو ليدر): قائمته مسطّحة — تنستعمل بترتيب العرض تحت
  const isPlainTechnician = (role === 'TECHNICIAN' || role === 'TECHNICAL') && !employee?.isLeader

  // ═══ تنظيف الشجرة قبل العرض ═══
  //
  // القائمة كانت تتبنى مثل ما هي مكتوبة بالكود، فطلعت:
  //  • ٣٣ شاشة مكررة — نفس الصفحة تحت «الإدارة» ومرة ثانية تحت «الوحدات»
  //    (تدقيق الحسابات، الدوار، المشاريع، الحجوزات... كلها مرتين)
  //  • مجموعات بولد واحد: «إدارة المركبات ← إدارة المركبات»
  //  • تفرعات لخمس مستويات للوصول لشاشة وحدة
  //
  // بدل ما نحذف بالإيد عنصر عنصر (ويرجع الخلل أول ما ينضاف عنصر جديد)،
  // ننضّف بقاعدتين عامتين تشتغلن لحالهن على أي عنصر ينضاف بالمستقبل:
  //   ١. الشاشة تظهر بأول محل توصلها بيه بس — التكرار بعدها ينشال.
  //   ٢. المجموعة الي ما بقى بيها إلا ولد واحد تنفك، والولد يطلع بمحلها.
  type PrunedItem = Omit<NavItem, 'children'> & { granted: boolean; children?: PrunedItem[] }

  // insideUnit: هل احنا جوّا وحدة إدارية؟ الوحدات مستثناة من قاعدة
  // «الشاشة تظهر بأول محل بس» — لأن محتواها كله موجود أصلاً تحت
  // «الإدارة»، فالقاعدة جانت تفرّغ الوحدات وتشيلها كاملة عن مدير
  // النظام. الوحدة باب مستقل يلمّ شغل قسم معيّن، مو تكرار.
  const prune = (items: NavItem[], unitGranted: boolean, seen: Set<string>, depth = 0, insideUnit = false): PrunedItem[] => {
    const out: PrunedItem[] = []
    for (const item of items) {
      if (!isVisible(item, unitGranted)) continue
      const granted =
        unitGranted ||
        (!!item.unitPermission && (role === 'ADMIN' || employeePermissions.includes(item.unitPermission)))

      if (item.divider) { out.push({ ...item, children: undefined, granted }); continue }

      if (item.children) {
        const kids = prune(item.children, granted, seen, depth + 1, insideUnit || !!item.unitPermission)
        if (!kids.length) continue
        // مجموعة بولد واحد = تفرع بلا فايدة: نطلّع الولد محلها.
        // بس مو بالمستوى الأول — هناك الولد يطلع يتيم بلا عنوان يدل
        // على وين هو (مثلاً «تقييم الأداء» طايح جنب «سياسة الخصوصية»).
        // قائمة الفني العادي مسطّحة بالكامل: مجموعات المستوى الأول تنفك
        // وأولادها يطلعون مباشرة. عنده ست شاشات بس، فأي تفرع فوقهن
        // ضغطة زايدة بلا فايدة.
        if (isPlainTechnician && depth === 0) { out.push(...kids); continue }
        if (depth > 0 && kids.length === 1 && !kids[0].divider) { out.push(kids[0]); continue }
        // بالمستوى الأول ما ننفك عموماً (الولد يطلع يتيم بلا عنوان يدل
        // على وين هو). الاستثناء: المجموعة الي ما بقى بيها إلا نفس شاشة
        // عنوانها — مثل «تصنيفي ← تصنيفي» لمن الموظف مو ليدر. هذي
        // تفرع على نفسها، تنفك.
        if (depth === 0 && kids.length === 1 && kids[0].to === item.to) { out.push(kids[0]); continue }
        out.push({ ...item, granted, children: kids })
        continue
      }

      // نفس الشاشة ما تتكرر — أول محل يوصلها هو محلها.
      // المفتاح يشمل الـ query لأن نفس الصفحة تشتغل شغلتين مختلفتين حسبه:
      // ‎/leader-invoices/new?mode=estimate‎ حساب استفساري ما ينحفظ، و
      // ‎/leader-invoices/new‎ فاتورة مربوطة بحجز.
      const key = item.to || ''
      if (!insideUnit) {
        if (key && seen.has(key)) continue
        if (key) seen.add(key)
      }
      out.push({ ...item, children: undefined, granted })
    }
    // فاصل ما يتبعه ولا عنصر (مثلاً "── الوحدات ──" وكل الوحدات انشالت
    // لأن محتواها ظاهر فوق) ما إله معنى
    return out.filter((it, idx, arr) => {
      if (!it.divider) return true
      const next = arr[idx + 1]
      return !!next && !next.divider
    })
  }

  const visibleItems = prune(navItems, false, new Set<string>())

  const toggle = (label: string) => setExpandedGroups((p) => ({ ...p, [label]: !p[label] }))

  const gradientClass = roleColors[employee.role] || 'from-blue-500 to-indigo-600'

  // unitGranted ينتقل للأولاد: لما الموظف عنده صلاحية الوحدة، كل صفحاتها
  // تنعرض له بدون فحص صلاحياتها التفصيلية.
  // العنصر وصلنا منضّف من prune — ما نفلتر ولا نحسب صلاحيات هنا من جديد.
  const renderNavItem = (item: PrunedItem, depth: number = 0): React.ReactNode => {

    if (item.divider) {
      // عنوان قسم بسيط (بلا خطوط) مقابل الفاصل الي يفصل مجموعتين
      if (item.plain) {
        if (collapsed) return <div key={item.to || item.label} className="my-2 h-px bg-white/10" />
        return (
          <p key={item.to || item.label} className="mb-1 mt-2.5 px-2 text-[9.5px] font-bold tracking-wide text-white/55">
            {item.label}
          </p>
        )
      }
      return (
        <div key={item.to || item.label} className="my-2 flex items-center gap-2 px-1">
          <span className="h-px flex-1 bg-white/20" />
          {!collapsed && <span className="text-[11px] font-bold text-white/50">{item.label}</span>}
          <span className="h-px flex-1 bg-white/20" />
        </div>
      )
    }

    if (item.children) {
      const kids = item.children
      const open = expandedGroups[item.label]
      const active = hasActiveChild(item, location.pathname)

      if (depth === 0) {
        return (
          <div key={item.to || item.label}>
            <button
              onClick={() => toggle(item.label)}
              className={`group relative flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-[13px] font-bold transition-all duration-300 ${
                active
                  ? 'bg-gradient-to-l from-[#2563eb] to-[#1e40af] text-white shadow-[0_0_18px_rgba(59,130,246,0.45)] ring-1 ring-sky-400/40'
                  : 'bg-white/[0.04] text-blue-100/70 hover:bg-white/[0.08] hover:text-white hover:shadow-[0_0_12px_rgba(59,130,246,0.18)]'
              }`}
            >
              {active && <span className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-sky-300 shadow-[0_0_10px_2px_rgba(125,211,252,0.9)]" />}
              {!collapsed && (
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"
                  className={`ml-0.5 transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              )}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              {/* الأيقونة بمربّع — نفس التصميم: صندوق صغير بحافة دائرية
                  يميّز العنصر النشط ويخلي الأيقونات على خط واحد. */}
              <span className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-all duration-300 ${
                active ? 'bg-white/25 text-white' : 'bg-white/[0.07] text-blue-100/70 group-hover:bg-white/[0.12]'
              }`}>{item.icon}</span>
            </button>
            {open && !collapsed && (
              <div className="mt-1 flex flex-col gap-1 rounded-lg bg-black/25 p-1 animate-in">
                {kids.map(child => renderNavItem(child, 1))}
              </div>
            )}
          </div>
        )
      }

      return (
        <div key={item.to || item.label}>
          <button
            onClick={() => toggle(item.label)}
            className={`group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              active ? 'bg-white/[0.08] text-white' : 'text-slate-400 hover:bg-white/[0.04] hover:text-slate-200'
            }`}
          >
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"
              className={`transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span className="flex-1 text-right">{item.label}</span>
            {active && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
          </button>
          {open && (
            <div className="mr-4 flex flex-col gap-0.5 border-r border-white/[0.06] pr-2">
              {kids.map(child => renderNavItem(child, 2))}
            </div>
          )}
        </div>
      )
    }

    // Leaf
    if (depth === 0) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group relative flex items-center gap-2 rounded-lg px-2 py-1.5 text-[13px] font-bold transition-all duration-300 ${
              isActive
                ? 'bg-gradient-to-l from-[#2563eb] to-[#1e40af] text-white shadow-[0_0_18px_rgba(59,130,246,0.45)] ring-1 ring-sky-400/40'
                : 'bg-white/[0.04] text-blue-100/70 hover:bg-white/[0.08] hover:text-white hover:shadow-[0_0_12px_rgba(59,130,246,0.18)]'
            }`
          }>
          {({ isActive }) => (
            <>
              {/* شريط أزرق على الحافة اليسرى للعنصر النشط */}
              {/* شعاع الإنارة على الحافة: يمشي مع العنصر النشط */}
              {isActive && <span className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-sky-300 shadow-[0_0_10px_2px_rgba(125,211,252,0.9)]" />}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              <span className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-all duration-300 ${
                isActive ? 'bg-white/25 text-white' : 'bg-white/[0.07] text-blue-100/70 group-hover:bg-white/[0.12]'
              }`}>{item.icon}</span>
            </>
          )}
        </NavLink>
      )
    }

    // Leaf مباشرة تحت مجموعة رئيسية (زي "الإدارة") — نفس شكل عناوين المجموعات
    // الشقيقة (بولد وأبيض) حتى ما توهم إنها ابن تابع لمجموعة ثانية فوقها.
    if (depth === 1) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              isActive ? 'bg-white/[0.08] text-white' : 'text-slate-300 hover:bg-white/[0.04] hover:text-white'
            }`
          }>
          {({ isActive }) => (
            <>
              <span className="flex-1 text-right">{item.label}</span>
              {isActive && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
            </>
          )}
        </NavLink>
      )
    }

    // ── بند داخل مجموعة ──
    // نقطة زرقاء صغيرة بدل الخط الجانبي: تربط البند بمجموعته بلمحة
    // وتخلي العنصر النشط يبيّن بلا ما ياخذ لون كامل يزاحم عنوان
    // المجموعة فوقه.
    return (
      <NavLink key={item.to} to={item.to} end={item.end}
        className={({ isActive }) =>
          `group flex items-center gap-2 rounded-md px-2.5 py-1.5 text-right text-[12px] transition-all duration-300 ${
            isActive
              ? 'bg-sky-500/25 font-bold text-white shadow-[0_0_12px_rgba(56,189,248,0.35)] ring-1 ring-sky-400/30'
              : 'bg-white/[0.03] font-medium text-blue-100/55 hover:bg-white/[0.08] hover:text-blue-100/85'
          }`
        }>
        {({ isActive }) => (
          <>
            <span className={`h-1.5 w-1.5 shrink-0 rounded-full transition-all duration-300 ${
              isActive ? 'bg-sky-300 shadow-[0_0_8px_2px_rgba(125,211,252,0.8)]' : 'bg-white/20 group-hover:bg-white/40'
            }`} />
            <span className="flex-1 text-right">{item.label}</span>
          </>
        )}
      </NavLink>
    )
  }

  return (
    <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
      {/* سياسة الخصوصية: تنعرض أول دخول ولما تنضاف نقاط جديدة */}
      <PrivacyPolicyGate />
      <div dir="ltr" className="app-shell flex bg-[#f0f4f9]">

        {/* ===== Main Area ===== */}
        <div dir="rtl" className="flex min-w-0 flex-1 flex-col">
          {/* Top Header — Glass effect */}
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-3 sm:px-5 lg:px-8">
            <div className="flex items-center gap-2 sm:gap-3">
              <button
                onClick={toggleSidebar}
                className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="15" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
                </svg>
              </button>
              <div className="hidden h-8 w-px bg-slate-200 sm:block"/>
              <span className="truncate text-sm font-extrabold text-[#0f2040] tracking-tight sm:text-lg">نظام شركة الأماني</span>
            </div>
            <div className="flex items-center gap-2 sm:gap-3">
              <div className="relative">
                <button
                  onClick={() => setSettingsOpen((o) => !o)}
                  className="flex h-10 w-10 items-center justify-center rounded-full text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
                  title="الإعدادات"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="3"/>
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09A1.65 1.65 0 0 0 15 4.6a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                  </svg>
                </button>
                {settingsOpen && <SettingsPanel onClose={() => setSettingsOpen(false)} />}
              </div>
              {/* ═══ التنبيهات الحيّة ═══
                  «أريد توصل إشعارات لأجهزتهم وإشعارات للحاسبات».
                  الزر يطلع لمن ما ينطلب الإذن بعد — وبعدها يختفي. */}
              <LiveAlerts />
              {/* الوضع الليلي — اختياره ينحفظ بالجهاز */}
              <ThemeToggle />
              <div className="relative">
                <button
                  ref={notifBtnRef}
                  onClick={() => setNotifOpen((o) => !o)}
                  aria-expanded={notifOpen}
                  aria-label={`الإشعارات${unreadCount > 0 ? ` — ${unreadCount} غير مقروء` : ''}`}
                  className="relative flex h-10 w-10 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0"/>
                  </svg>
                  {unreadCount > 0 && (
                    <span className="absolute -top-0.5 -left-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-600 px-1 text-[10px] font-bold text-white">
                      {unreadCount > 9 ? '9+' : unreadCount}
                    </span>
                  )}
                </button>
                {notifOpen && (
                  <>
                    {/* ⚠️ ماكو حجاب شفاف يغطّي الشاشة — الإغلاق يصير
                        بمستمع على المستند (تحت). الحجاب چان يبلع كل شي:
                        عجلة الماوس تنضرب بيه فالصفحة تتجمّد ورا القائمة
                        («تضل فوگ» والموظف يظن النظام علّق)، وحتى ضغطة
                        الجرس نفسه ما توصله فما يكدر يغلقها من نفس الزر.
                        ⚠️ ظل قوي وحلقة: بلاهن اللوحة بيضة على خلفية
                        فاتحة بلا فصل، فتبين ملزوقة بالصفحة مو طايفة
                        فوگها.
                        ⚠️ وبالموبايل تنشدّ للحافتين: الجرس بحافة الشاشة،
                        و`left-0` چانت تطلّعها برّا الحافة. */}
                    <div
                      ref={notifPanelRef}
                      className="absolute top-12 z-50 w-80 overflow-hidden rounded-2xl bg-white shadow-2xl ring-1 ring-slate-900/10 max-sm:fixed max-sm:inset-x-3 max-sm:w-auto sm:left-0"
                    >
                      <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
                        <span className="text-sm font-bold text-slate-800">الإشعارات</span>
                        {unreadCount > 0 && (
                          <button onClick={handleMarkAllRead} className="text-xs font-medium text-brand-600 hover:underline">تحديد الكل كمقروء</button>
                        )}
                      </div>
                      <div className="max-h-96 overflow-y-auto">
                        {notifications.length === 0 && (
                          <p className="px-4 py-6 text-center text-sm text-slate-400">ماكو إشعارات</p>
                        )}
                        {notifications.map((n) => (
                          <button
                            key={n.id}
                            onClick={() => handleNotifClick(n)}
                            className={`block w-full border-b border-slate-50 px-4 py-3 text-right text-sm transition-colors hover:bg-slate-50 ${n.read ? 'text-slate-500' : 'bg-brand-50/50 font-medium text-slate-800'}`}
                          >
                            <p>{n.message}</p>
                            <div className="mt-1 flex items-center justify-between">
                              <span className="text-[11px] text-slate-400">{new Date(n.createdAt).toLocaleString('ar-IQ')}</span>
                              {/* سهم يبيّن إن الإشعار ينفتح على شاشة */}
                              {notifTargets[n.type] && (
                                <span className="text-[11px] font-bold text-brand-600">افتحها ←</span>
                              )}
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  </>
                )}
              </div>
              <div className="flex items-center gap-2 rounded-xl bg-slate-50/80 px-2 py-1.5 transition-colors hover:bg-slate-100 sm:gap-3 sm:px-3">
                <div className="hidden text-left sm:block">
                  <p className="text-sm font-bold text-slate-800">{employee.name}</p>
                  <p className="text-[11px] text-slate-400">{roleLabels[employee.actualRole || employee.role]}</p>
                </div>
                <div className={`relative flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-md sm:h-10 sm:w-10`}>
                  {employee.attendanceIcon || employee.name.charAt(0)}
                  <span className="absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-white bg-emerald-500"/>
                </div>
              </div>
            </div>
          </header>

          {/* شريط الإعلانات — يشوفه كل موظف تحت الهيدر مباشرة */}
          <AnnouncementTicker />

          {/* Content */}
          <main className="min-w-0 flex-1 overflow-x-hidden overflow-y-auto p-3 sm:p-5 lg:p-8">
            <Outlet />
          </main>
        </div>

        {/* Mobile backdrop */}
        {mobileOpen && (
          <div onClick={() => setMobileOpen(false)} className="fixed inset-0 z-40 bg-black/40 lg:hidden" />
        )}

        {/* ===== Right Sidebar — Premium ===== */}
        <aside
          dir="rtl"
          // ⚠️ isolate + overflow-hidden: هالة الضوء تحت مطلقة الموضع،
          // وبدونهن تطلع برّا القائمة وتغطي محتوى الصفحة.
          className={`app-sidebar glossy-dark isolate fixed inset-y-0 right-0 z-50 flex flex-col overflow-hidden bg-[#0f2040] transition-transform duration-300 ease-in-out lg:sticky lg:top-0 lg:z-auto lg:h-auto lg:translate-x-0 ${
            mobileOpen ? 'translate-x-0' : 'translate-x-full'
          }`}
          style={{ width: collapsed ? 72 : 270, minWidth: collapsed ? 72 : 270 }}
        >
          {/* ═══ هالة الإنارة ═══
              ضوء أزرق خفيف ينزل من أعلى القائمة، يعطيها العمق الي
              بالتصميم بدل الكحلي المسطّح.
              ⚠️ pointer-events-none: طبقة زينة، وبدونها تبلع ضغطات
              أول بندين بالقائمة. */}
          <div aria-hidden className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-72 bg-[radial-gradient(120%_70%_at_50%_0%,rgba(56,120,255,0.28),transparent_70%)]" />

          {/* Logo */}
          <div className={`flex items-center ${collapsed ? 'justify-center' : 'justify-center gap-3'} px-4 py-5`}>
            {/* شعار الشركة الحقيقي بدل المكعّب العام.
                ⚠️ على خلفية بيضا: الشعار أزرق غامق، وعلى خلفية القائمة
                الكحلية ما يبيّن — نفس اللونين تقريباً. */}
            <div className="relative flex h-11 w-11 items-center justify-center rounded-xl bg-white shadow-lg shadow-blue-900/40">
              <img src={`${import.meta.env.BASE_URL}favicon.png?v=3`} alt="شعار شركة الأماني" className="h-8 w-8 object-contain" />
              <span className="absolute -top-1 -left-1 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
            </div>
            {!collapsed && (
              <div>
                <p className="text-sm font-extrabold text-white tracking-tight">الأماني</p>
                <p className="text-[10px] text-blue-300/75 font-medium">Management System</p>
              </div>
            )}
          </div>

          {/* User card */}
          {!collapsed ? (
            <div className="mx-3 mb-4 overflow-hidden rounded-2xl bg-gradient-to-br from-white/[0.08] to-white/[0.03] p-4 shadow-[inset_0_0_0_1px_rgba(255,255,255,0.06)]">
              <div className="flex items-center gap-3">
                <div className="flex-1 text-right">
                  <p className="text-sm font-bold text-white">{employee.name}</p>
                  <p className="text-[11px] text-blue-300/80">{roleLabels[employee.actualRole || employee.role]}</p>
                </div>
                {/* ═══ صورة الموظف — عرض بس ═══
                    «اقفلها بيد الإدارة بس».
                    ⚠️ والسيرفر كان يفرضها من الأصل: تعديل بيانات
                    الموظف (`PUT /employees/{id}`) محصور بمدير النظام،
                    فزر الرفع هنا چان يفشل بـ٤٠٣ لأي موظف عادي —
                    يعني زر يوعد بشي ما يكدر يسويه. الصورة تنضاف من
                    شاشة الموظفين، وهنا تنشاف وتنفتح بس. */}
                <div className="relative">
                  <EmployeeAvatar
                    name={employee.name}
                    photoUrl={employee.photoUrl}
                    size="md"
                    rounded="xl"
                  />
                  <span className="pointer-events-none absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
                </div>
              </div>
              <div className="mt-3 flex items-center gap-2">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"/>
                <span className="text-[10px] text-emerald-400/80 font-medium">متصل الآن</span>
                <span className={`mr-auto rounded-full bg-gradient-to-l ${gradientClass} px-2.5 py-0.5 text-[10px] font-bold text-white shadow-sm`}>
                  {roleLabels[employee.actualRole || employee.role]}
                </span>
              </div>
            </div>
          ) : (
            <div className="mx-auto mb-3">
              <div className="relative">
                <EmployeeAvatar name={employee.name} photoUrl={employee.photoUrl} size="md" rounded="xl" />
                <span className="pointer-events-none absolute -bottom-0.5 -left-0.5 h-2.5 w-2.5 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
              </div>
            </div>
          )}

          {/* Divider */}
          <div className="mx-4 mb-2 h-px bg-gradient-to-l from-transparent via-white/10 to-transparent"/>

          {/* Nav */}
          <nav className="flex flex-1 flex-col gap-1 overflow-y-auto px-2.5 pb-3 scrollbar-thin">
            {visibleItems.map(item => renderNavItem(item, 0))}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/owner-security"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                👁️ لوحة المراقبة الخلفية
              </NavLink>
            )}
            {/* النسخ الاحتياطية: للمالك وحده — actualRole مو role، لأن
                role يتطبّع لـ'ADMIN' فوق وهذا يكشفها لكل مدير */}
            {/* ⚠️ «الطلبات» انشالت من هنا — صارت بطاقة بالواجهة
                الرئيسية جنب «جردي». الموظف يفتح النظام ويشوفها
                قدامه، بدل ما ينزّل بالقائمة الجانبية يدوّر عليها.
                (المسار ‎/letters‎ باقي شغّال لمن عنده رابط مباشر.) */}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/command-code"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                🔐 رمز مركز القيادة
              </NavLink>
            )}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/owner-backups"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                🔐 النسخ الاحتياطية
              </NavLink>
            )}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/assistant-conversations"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                💬 محادثات المساعد الذكي
              </NavLink>
            )}
          </nav>

          {/* سياسة الخصوصية — قراءة متاحة لكل موظف بدون أي صلاحية، ثابتة فوق
              زر تسجيل الخروج. إضافة/تعديل النقاط شي ثاني تماماً: صلاحية
              "privacy_policy_manage" وتظهر داخل وحدة الرقابة. */}
          <div className="mx-3 mt-1">
            <NavLink
              to="/privacy-policy"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                }`
              }
            >
              {collapsed ? '🔒' : '🔒 سياسة الخصوصية'}
            </NavLink>
          </div>

          {/* لوحة الإعلانات — ثابتة بالأسفل تحت سياسة الخصوصية، مو داخل
              أي وحدة، لأنها تخص المالك ومدير النظام وحدهم. */}
          {(role === 'ADMIN' || role === 'OWNER') && (
            <div className="mx-3 mt-1">
              <NavLink
                to="/announcements"
                className={({ isActive }) =>
                  `flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                {collapsed ? '📢' : '📢 لوحة الإعلانات'}
              </NavLink>
            </div>
          )}

          {/* Logout */}
          <div className="mx-3 mb-3 mt-1">
            <div className="h-px bg-gradient-to-l from-transparent via-white/10 to-transparent mb-3"/>
            <button onClick={() => setEmployee(null)}
              className="group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-red-300/80 transition-all duration-200 hover:bg-red-500/10 hover:text-red-300">
              {!collapsed && <span className="flex-1 text-right">تسجيل الخروج</span>}
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="opacity-60 group-hover:opacity-100 transition-opacity">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
              </svg>
            </button>
          </div>
        </aside>
      </div>
      {(employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || employee?.actualRole === 'OWNER')
        ? <ManagerAssistantChat />
        : <AssistantWidget />}
    </SessionContext.Provider>
  )
}
