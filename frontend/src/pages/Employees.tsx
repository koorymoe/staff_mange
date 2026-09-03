import { useEffect, useState } from 'react'
import { api, LEADER_SKILLS, type Employee, type Service, type Stats } from '../api'
import EmployeeHRPanel from '../components/EmployeeHRPanel'
import EmployeeAvatar from '../components/EmployeeAvatar'
import { useSession } from '../session'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'
import StatTile from '../components/StatTile'
import AddEmployeeWizard from '../components/AddEmployeeWizard'
import { openManagerChat } from '../components/openManagerChat'
import { matches } from '../utils/search'
import { toIntlPhone } from '../utils/phone'

// زرّا واتساب وتلغرام جنب رقم الهاتف — دائماً الاثنان معاً (ماكو
// حقل بالنظام يحدد أي تطبيق يستخدمه الموظف)، ويظهران بس لمن الهاتف موجود.
function PhoneContactLinks({ phone }: { phone: string | null | undefined }) {
  if (!phone) return null
  const intl = toIntlPhone(phone)
  if (!intl) return null
  return (
    <span className="inline-flex items-center gap-1.5">
      <a href={`https://wa.me/${intl}`} target="_blank" rel="noreferrer" title="فتح واتساب"
        onClick={(e) => e.stopPropagation()}
        className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600 hover:bg-emerald-500/25">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor"><path d="M12.04 2c-5.5 0-9.96 4.46-9.96 9.96 0 1.76.46 3.45 1.32 4.95L2 22l5.24-1.37a9.9 9.9 0 0 0 4.8 1.23h.01c5.5 0 9.96-4.46 9.96-9.96S17.54 2 12.04 2zm0 18.2c-1.5 0-2.97-.4-4.25-1.16l-.3-.18-3.12.82.83-3.04-.2-.31a8.2 8.2 0 0 1-1.26-4.37c0-4.55 3.7-8.24 8.25-8.24 4.54 0 8.24 3.7 8.24 8.24s-3.7 8.24-8.24 8.24zm4.52-6.17c-.25-.12-1.47-.72-1.7-.81-.23-.08-.4-.12-.56.13-.17.25-.65.81-.79.97-.15.17-.29.19-.54.06-.25-.12-1.04-.38-1.98-1.22-.73-.65-1.23-1.46-1.37-1.71-.14-.25-.02-.38.11-.51.11-.11.25-.29.37-.43.12-.15.16-.25.24-.42.08-.17.04-.31-.02-.43-.06-.12-.56-1.36-.77-1.86-.2-.49-.41-.42-.56-.43h-.48c-.17 0-.43.06-.66.31-.23.25-.86.85-.86 2.07 0 1.22.89 2.4 1.01 2.57.12.17 1.75 2.67 4.24 3.74.59.26 1.05.41 1.41.52.59.19 1.13.16 1.55.1.47-.07 1.47-.6 1.68-1.18.21-.58.21-1.08.15-1.18-.06-.1-.23-.16-.48-.28z"/></svg>
      </a>
      <a href={`https://t.me/+${intl}`} target="_blank" rel="noreferrer" title="فتح تلغرام"
        onClick={(e) => e.stopPropagation()}
        className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-sky-500/15 text-sky-600 hover:bg-sky-500/25">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor"><path d="M21.9 4.3 18.6 20c-.25 1.1-.9 1.37-1.83.86l-5.06-3.73-2.44 2.35c-.27.27-.5.5-1.02.5l.36-5.16 9.4-8.5c.41-.36-.09-.56-.63-.2L6.3 12.7 1.3 11.1c-1.09-.34-1.1-1.09.23-1.6l19.4-7.48c.9-.34 1.7.2 1.4 1.28z"/></svg>
      </a>
    </span>
  )
}

const levels = [
  { level: 1, label: 'متدرب', min: 0 },
  { level: 2, label: 'فني مبتدئ', min: 3 },
  { level: 3, label: 'فني', min: 6 },
  { level: 4, label: 'فني متمرس', min: 10 },
  { level: 5, label: 'فني خبير', min: 15 },
]

const BOOKINGS_PER_RANK = 10

const roleLabels: Record<string, string> = {
  ADMIN: 'مدير النظام',
  SALES: 'موظف مبيعات',
  HR_COORDINATOR: 'إداري الكوادر',
  TECHNICIAN: 'فني',
  PROJECT_MANAGER: 'مدير مشاريع',
  MONITOR: 'مراقب',
  FINANCE: 'محاسب',
  GPS_ADMIN: 'مسؤول GPS',
  QUALITY_ENGINEER: 'مهندس جودة',
  ENGINEER: 'مهندس',
  PROCUREMENT_ADMIN: 'إداري الكميات',
  DESIGNER: 'مصمم',
  SERVICE_MANAGER: 'مسؤول خدمة',
  // التقني يتولى أكثر من خدمة — منفصل عن «فني» الي يشتغل بخدمة وحدة
  TECHNICAL: 'تقني',
}

const roleColors: Record<string, { bg: string; text: string; dot: string }> = {
  ADMIN: { bg: 'bg-amber-50', text: 'text-amber-700', dot: 'bg-amber-500' },
  SALES: { bg: 'bg-emerald-50', text: 'text-emerald-700', dot: 'bg-emerald-500' },
  HR_COORDINATOR: { bg: 'bg-violet-50', text: 'text-violet-700', dot: 'bg-violet-500' },
  TECHNICIAN: { bg: 'bg-sky-50', text: 'text-sky-700', dot: 'bg-sky-500' },
  PROJECT_MANAGER: { bg: 'bg-rose-50', text: 'text-rose-700', dot: 'bg-rose-500' },
  MONITOR: { bg: 'bg-cyan-50', text: 'text-cyan-700', dot: 'bg-cyan-500' },
  FINANCE: { bg: 'bg-lime-50', text: 'text-lime-700', dot: 'bg-lime-500' },
  GPS_ADMIN: { bg: 'bg-indigo-50', text: 'text-indigo-700', dot: 'bg-indigo-500' },
  QUALITY_ENGINEER: { bg: 'bg-fuchsia-50', text: 'text-fuchsia-700', dot: 'bg-fuchsia-500' },
  ENGINEER: { bg: 'bg-teal-50', text: 'text-teal-700', dot: 'bg-teal-500' },
  PROCUREMENT_ADMIN: { bg: 'bg-orange-50', text: 'text-orange-700', dot: 'bg-orange-500' },
  DESIGNER: { bg: 'bg-pink-50', text: 'text-pink-700', dot: 'bg-pink-500' },
  SERVICE_MANAGER: { bg: 'bg-teal-50', text: 'text-teal-700', dot: 'bg-teal-500' },
}

// ⚠️ تدرّجات الأفاتار انشالت: الصورة صارت من `EmployeeAvatar`
// (صورة الموظف، وإلا حرفه بلون ثابت من اسمه).

export default function Employees({ embedded }: { embedded?: boolean } = {}) {
  // كل حفظ بهاي الشاشة يمر من هنا — الفشل ينعرض بدل ما ينبلع
  const guard = useSaveGuard()
  const { employee: currentUser, permissions: userPermissions } = useSession()
  const isAdmin = currentUser?.role === 'ADMIN'

  // ═══ حفظ صورة موظف ═══
  // نحدّث القائمة والمختار بالجواب الراجع من السيرفر — مو بالقيمة الي
  // دزّيناها، حتى لو رفض أو عدّل تبقى الشاشة تعرض الحقيقة.
  const savePhoto = async (employeeId: string, url: string | null) => {
    try {
      const updated = await api.updateEmployee(employeeId, { photoUrl: url ?? '' })
      // ⚠️ `selectedEmployee` مشتق من القائمة مو حالة مستقلة، فتحديث
      // القائمة يكفي — ولو خزّناه بحالة ثانية نفتح باب اختلافهن.
      setEmployees((prev) => prev.map((e) => (e.id === updated.id ? { ...e, photoUrl: updated.photoUrl } : e)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حفظ الصورة')
    }
  }
  // فتح حساب جديد = المالك وحده. `role` ينزّل OWNER لـADMIN حتى تشتغل
  // بقية الشاشات، فالدور الحقيقي بـ`actualRole`.
  //
  // ⚠️ هذا إخفاء للزر بس — المنع الحقيقي بالسيرفر (RequireOwnerOnly على
  // POST /api/employees). إخفاء الزر لحاله ما يمنع أحد يدزّ الطلب بيده.
  const canCreateAccounts = currentUser?.actualRole === 'OWNER'
  const isHR = currentUser?.role === 'HR_COORDINATOR'
  const [employees, setEmployees] = useState<Employee[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [credUsername, setCredUsername] = useState('')
  const [credPassword, setCredPassword] = useState('')
  const [savingCreds, setSavingCreds] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterRole, setFilterRole] = useState<string>('')

  const [showAddForm, setShowAddForm] = useState(false)

  // تبويبا لوحة التفاصيل بأعلى: معلومات الموظف مقابل مهاراته —
  // نفس قالب تبديل المحتوى الموجود بـ`KpiPage.tsx`.
  const [panelTab, setPanelTab] = useState<'info' | 'skills'>('info')
  const [skillTab, setSkillTab] = useState<'technical' | 'leader'>('technical')
  // ⚠️⚠️ چانت حالة محلية ما تنحفظ ولا تنجلب — والتصفير عند تبديل
  // الموظف ما چان يشملها، فالمدير يقيّم واحداً ويشوف نفس درجاته
  // على الي بعده. صارت تنجلب من الخادم وتنصفّر بالتبديل.
  const [leaderRatings, setLeaderRatings] = useState<Record<string, number>>({})
  const [leaderSaved, setLeaderSaved] = useState<Record<string, number>>({})
  const [leaderBusy, setLeaderBusy] = useState(false)
  const [editSalary, setEditSalary] = useState('')
  const [editShiftStart, setEditShiftStart] = useState('')
  const [editShiftEnd, setEditShiftEnd] = useState('')
  const [editMonthlyLeaves, setEditMonthlyLeaves] = useState('')
  const [editJobTitle, setEditJobTitle] = useState('')
  const [editIsLeader, setEditIsLeader] = useState(false)
  const [editName, setEditName] = useState('')
  const [editPhone, setEditPhone] = useState('')
  const [editPosition, setEditPosition] = useState('')
  const [editCertificate, setEditCertificate] = useState('')

  const [showCompare, setShowCompare] = useState(false)
  const [compareId, setCompareId] = useState<string | null>(null)

  // أرشفة/حذف الموظفين — الأدمن بس يقدر يشوف قائمة المؤرشفين/المحذوفين وتاريخهم
  const [showArchived, setShowArchived] = useState(false)
  const [archivedEmployees, setArchivedEmployees] = useState<Employee[]>([])

  const load = () => {
    Promise.all([api.getEmployees(), api.getServices()])
      .then(([emps, svcs]) => { setEmployees(emps); setServices(svcs) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  useEffect(() => {
    // ⚠️ چانت `if (isHR)` — فالمالك والمدير ما يجلبونها أصلاً
    // وعدّادات «إحصائيات الموظف» تطلعلهم **صفراً دائماً**.
    api.getStats().then(setStats).catch(() => setStats(null))
  }, [isHR])
  useEffect(() => {
    if (isAdmin && showArchived) api.getArchivedEmployees().then(setArchivedEmployees).catch(() => setArchivedEmployees([]))
  }, [isAdmin, showArchived])

  const baseEmployees = showArchived ? archivedEmployees : (isHR ? employees.filter((emp) => emp.role === 'TECHNICIAN') : employees)
  const visibleEmployees = baseEmployees.filter(emp => {
    if (searchQuery && !matches([emp.name, emp.position, emp.phone, emp.jobTitle], searchQuery)) return false
    if (filterRole && emp.role !== filterRole) return false
    return true
  })

  const selectedEmployee = [...employees, ...archivedEmployees].find((emp) => emp.id === selectedId) || null

  // ⚠️ نفس قائمة الخادم بالضبط (main.go: ADMIN/OWNER/HR_COORDINATOR)،
  // والمالك يوصل بدوره المطبَّع ADMIN. والخادم يرفض تقييم النفس هم.
  const canRateLeader = !!currentUser
    && ['ADMIN', 'OWNER', 'HR_COORDINATOR'].includes(currentUser.role)
    && currentUser.id !== selectedEmployee?.id

  const handleArchive = async (status: 'ARCHIVED' | 'DELETED' | 'ACTIVE') => {
    if (!selectedEmployee) return
    const label = status === 'ARCHIVED' ? 'أرشفة' : status === 'DELETED' ? 'حذف' : 'استرجاع'
    if (!confirm(`متأكد تريد ${label} الموظف "${selectedEmployee.name}"؟`)) return
    try {
      await api.updateEmployee(selectedEmployee.id, { status })
      setSelectedId(null)
      load()
      if (showArchived) api.getArchivedEmployees().then(setArchivedEmployees).catch(() => {})
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تنفيذ العملية')
    }
  }
  const compareEmployee = employees.find((emp) => emp.id === compareId) || null

  const canGenerateAiReport = isAdmin || currentUser?.role === 'MONITOR'

  const handleGenerateAiReport = async () => {
    if (!selectedEmployee) return
    openManagerChat(`أعطني تقرير شامل عن الموظف ${selectedEmployee.name}`)
  }

  const [linkingHistorical, setLinkingHistorical] = useState(false)
  const handleLinkHistorical = async () => {
    if (!selectedEmployee) return
    setLinkingHistorical(true)
    try {
      const result = await api.linkHistoricalRecords(selectedEmployee.id)
      if (result.bookingsLinked === 0 && result.complaintsLinked === 0) {
        alert('ماكو سجلات تاريخية بنفس اسم هذا الموظف بالضبط')
      } else {
        alert(`تم ربط ${result.bookingsLinked} حجز و ${result.complaintsLinked} شكوى بحساب هذا الموظف`)
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر ربط السجلات التاريخية')
    } finally {
      setLinkingHistorical(false)
    }
  }

  // جلب تقييم القيادة للموظف المختار.
  // ⚠️ تأثير مستقل عن تصفير النموذج: التصفير مزامنة حالة، وهذا
  // جلب بيانات — خلطهما يخلي React يشتكي من setState متزامن.
  useEffect(() => {
    const id = selectedEmployee?.id
    if (!id) return
    let alive = true
    api.getLeaderSkills(id)
      .then((rows) => {
        if (!alive) return
        const map: Record<string, number> = {}
        for (const r of rows) map[r.skill] = r.score
        setLeaderRatings(map)
        setLeaderSaved(map)
      })
      .catch(() => { /* ما تنعرض درجات = ماكو تقييم بعد */ })
    return () => { alive = false }
  }, [selectedEmployee?.id])

  useEffect(() => {
    // Resetting the edit-form fields whenever the selected employee changes is a
    // derived-state sync from a prop, not data fetching.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setCredUsername(selectedEmployee?.username || '')
    setCredPassword('')
    setEditSalary('')
    setEditShiftStart('')
    setEditShiftEnd('')
    setEditMonthlyLeaves('')
    setEditJobTitle('')
    setEditIsLeader(selectedEmployee?.isLeader || false)
    setEditName('')
    setEditPhone('')
    setEditPosition('')
    setEditCertificate('')
    setShowCompare(false)
    setCompareId(null)
    // ⚠️ هاي السطر هو إصلاح التسريب: بدونه درجات الموظف السابق
    // تبقى معروضة على الي بعده.
    setLeaderRatings({})
    setLeaderSaved({})
    setSkillTab('technical')
    setPanelTab('info')
  }, [selectedId, selectedEmployee?.isLeader, selectedEmployee?.username])

  const handleSaveCredentials = async () => {
    if (!selectedEmployee) return
    setSavingCreds(true)
    try {
      const updated = await api.updateEmployee(selectedEmployee.id, {
        username: credUsername, ...(credPassword ? { password: credPassword } : {}),
      })
      setEmployees((prev) => prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)))
      setCredPassword('')
    } catch (e) { alert(e instanceof Error ? e.message : 'حدث خطأ') }
    finally { setSavingCreds(false) }
  }

  const handleFieldBlur = async <K extends keyof Employee>(field: K, value: Employee[K]) => {
    if (!selectedEmployee) return
    try {
      const updated = await api.updateEmployee(selectedEmployee.id, { [field]: value })
      setEmployees((prev) => prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)))
    } catch (e) {
      console.error(e)
    }
  }

  const toggleSkill = async (employee: Employee, skillId: string) => {
    const current = new Map(employee.skills.map((s) => [s.skillId, s.canPerform]))
    current.set(skillId, !current.get(skillId))
    // بس مهارات شعبة الموظف. قبل، الحفظ جان يمر على كل الخدمات — يعني
    // موظف ديكور تنكتب له صفوف لمهارات هندسية (بـfalse)، والعكس. الشاشة
    // ما تعرضهن فما ينلاحظ، بس السجل يمتلئ بمهارات مو من شعبته ولو
    // انفتحت من أي مكان ثاني تطلع غلط.
    const skills = services
      .filter((svc) => svc.division === employee.division)
      .flatMap((svc) => svc.skills.map((sk) => ({ skillId: sk.id, canPerform: current.get(sk.id) ?? false })))
    const updated = await guard.run('حفظ المهارات', () => api.updateEmployeeSkills(employee.id, skills))
    if (!updated) return
    setEmployees((prev) => prev.map((emp) => (emp.id === employee.id ? updated : emp)))
  }

  const getEmployeeStats = (empId: string) => {
    const techStat = stats?.technicianStats.find((s) => s.employeeId === empId)
    // ⚠️ «الجارية» چانت totalAssigned — يعني **كل** المسند مو
    // الجاري، فالعنوان يناقض رقمه. والفرق هو الجاري فعلاً.
    // ⚠️ و«ساعات إضافية» انشالت: چانت 0 مثبتة بالكود، وصفر دائم
    // يوهم إن ماكو ساعات إضافية بينما ماكو بيانات أصلاً.
    const done = techStat?.completed || 0
    const assigned = techStat?.totalAssigned || 0
    return { completed: done, inProgress: Math.max(0, assigned - done) }
  }


  const uniqueRoles = [...new Set(baseEmployees.map(e => e.role))]

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
    <div className="max-w-[1400px] mx-auto">
      {/* Header — ⚠️ العنوان يختفي لمّن الشاشة مضمَّنة (مكتب إدارة
          الموظفين)، وشريط البحث/الفلتر يبقى. */}
      <div className="flex items-center justify-between mb-6">
        {!embedded && (
        <div>
          <h2 className="text-2xl font-extrabold text-[#0f2040]">إدارة الكوادر</h2>
          <p className="mt-1 text-sm text-slate-400">إدارة بيانات الموظفين وتحديد المهارات</p>
        </div>
        )}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 rounded-xl bg-white px-4 py-2.5 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_6px_16px_rgba(0,0,0,0.04)]">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="بحث عن موظف..."
              className="w-48 bg-transparent text-sm outline-none placeholder:text-slate-300"
            />
          </div>
          {!isHR && (
            <select
              value={filterRole}
              onChange={(e) => setFilterRole(e.target.value)}
              className="rounded-xl bg-white px-3 py-2.5 text-sm text-slate-600 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_6px_16px_rgba(0,0,0,0.04)] outline-none"
            >
              <option value="">كل الأدوار</option>
              {uniqueRoles.map(r => <option key={r} value={r}>{roleLabels[r] || r}</option>)}
            </select>
          )}
          {isAdmin && (
            <button
              onClick={() => { setShowArchived(!showArchived); setSelectedId(null) }}
              className={`rounded-xl px-4 py-2.5 text-sm font-semibold shadow-[0_1px_3px_rgba(0,0,0,0.04),0_6px_16px_rgba(0,0,0,0.04)] transition-all ${
                showArchived ? 'bg-slate-700 text-white' : 'bg-white text-slate-600'
              }`}
            >
              {showArchived ? '↩ رجوع للنشطين' : '🗄️ المؤرشفون/المحذوفون/الموقوفين'}
            </button>
          )}
          {canCreateAccounts && !showArchived && (
            <button
              onClick={() => setShowAddForm(!showAddForm)}
              className="flex items-center gap-2 rounded-xl bg-gradient-to-l from-[#2c5aad] to-[#1e3f7a] px-5 py-2.5 text-sm font-semibold text-white shadow-lg shadow-blue-900/20 transition-all hover:shadow-xl hover:shadow-blue-900/30"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              إضافة موظف
            </button>
          )}
        </div>
      </div>

      {/* Add Employee Wizard */}
      {canCreateAccounts && showAddForm && (
        <AddEmployeeWizard onClose={() => setShowAddForm(false)} onCreated={load} />
      )}

      {loading && <div className="flex items-center justify-center py-20"><div className="h-8 w-8 animate-spin rounded-full border-4 border-[#2c5aad] border-t-transparent"/></div>}
      {error && <p className="mt-6 rounded-2xl bg-red-50 p-5 text-red-600 text-sm">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="grid grid-cols-1 gap-6 xl:grid-cols-[340px_1fr]">
          {/* Employee Cards Grid */}
          <div className="space-y-3 max-h-[calc(100vh-200px)] overflow-y-auto pl-2 scrollbar-thin">
            <p className="text-xs font-medium text-slate-400 mb-2">{visibleEmployees.length} موظف</p>
            {visibleEmployees.map((emp) => {
              const rc = roleColors[emp.role] || { bg: 'bg-slate-50', text: 'text-slate-700', dot: 'bg-slate-500' }
              const skillCount = emp.skills.filter(s => s.canPerform).length
              const isSelected = selectedId === emp.id
              return (
                <button
                  key={emp.id}
                  onClick={() => setSelectedId(emp.id)}
                  className={`group relative flex w-full items-center gap-4 rounded-2xl p-4 text-right transition-all duration-200 ${
                    isSelected
                      ? 'bg-white shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.08)] ring-2 ring-[#2c5aad]/20'
                      : 'bg-white/60 shadow-[0_1px_3px_rgba(0,0,0,0.03)] hover:bg-white hover:shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.06)]'
                  }`}
                >
                  {isSelected && <span className="absolute right-0 top-1/2 -translate-y-1/2 h-8 w-1 rounded-l-full bg-[#2c5aad]"/>}
                  {/* صورة الموظف بالقائمة — الحرف بديل لمن ماكو صورة */}
                  <div className="relative shrink-0">
                    <EmployeeAvatar name={emp.name} photoUrl={emp.photoUrl} size="lg" rounded="xl" />
                    <span className={`pointer-events-none absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-white ${emp.onDuty ? 'bg-emerald-500' : 'bg-slate-300'}`}/>
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-bold text-[#0f2040] truncate">{emp.name}</p>
                    <div className="mt-1 flex items-center gap-2">
                      <span className={`inline-flex items-center gap-1 rounded-full ${rc.bg} px-2 py-0.5 text-[10px] font-semibold ${rc.text}`}>
                        <span className={`h-1.5 w-1.5 rounded-full ${rc.dot}`}/>
                        {roleLabels[emp.role] || emp.role}
                      </span>
                      {skillCount > 0 && (
                        <span className="text-[10px] text-slate-400">{skillCount} مهارة</span>
                      )}
                    </div>
                  </div>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                    className="shrink-0 opacity-0 transition-opacity group-hover:opacity-100"><polyline points="15 18 9 12 15 6"/></svg>
                </button>
              )
            })}
            {visibleEmployees.length === 0 && (
              <div className="flex flex-col items-center py-12 text-slate-500">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                <p className="mt-3 text-sm">لا يوجد موظفين</p>
              </div>
            )}
          </div>

          {/* Detail Panel */}
          <div className="rounded-2xl bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.06)] max-h-[calc(100vh-200px)] overflow-y-auto">
            {!selectedEmployee && (
              <div className="flex flex-col items-center justify-center py-20 text-slate-500">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
                <p className="mt-4 text-sm">اختر موظفاً لعرض التفاصيل</p>
              </div>
            )}

            {selectedEmployee && (() => {
              const skillCount = selectedEmployee.skills.filter(s => s.canPerform).length
              const currentLevel = [...levels].reverse().find(l => skillCount >= l.min) || levels[0]
              const nextLevel = levels.find(l => l.min > skillCount)

              return (
                <div>
                  {/* Hero Card */}
                  <div className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-[#2c5aad] to-[#0f2040] p-6 text-white">
                    <div className="absolute top-0 left-0 w-40 h-40 rounded-full bg-white/5 -translate-x-1/2 -translate-y-1/2"/>
                    <div className="absolute bottom-0 right-0 w-32 h-32 rounded-full bg-white/5 translate-x-1/3 translate-y-1/3"/>
                    <div className="relative flex items-start gap-5">
                      {/* ═══ المكان الوحيد الي تنضاف منه الصورة ═══
                          «اقفلها بيد الإدارة بس» — والسيرفر يفرضها
                          كمان: `PUT /employees/{id}` محصور بمدير
                          النظام. فالشرط هنا يطابق حارس السيرفر، مو
                          يعتمد عليه بس.
                          ⚠️ وصورة الموظف تحديداً حصراً بالمالك —
                          `actualRole` لا `role` (`role` تنزّل OWNER
                          لـADMIN حتى تشتغل بقية الشاشة، نفس
                          `canCreateAccounts` فوگ)، ونفس القيد مفروض
                          بالسيرفر (`req.PhotoURL` بـ`employee_handler.go`). */}
                      <EmployeeAvatar
                        name={selectedEmployee.name}
                        photoUrl={selectedEmployee.photoUrl}
                        size="xl"
                        rounded="xl"
                        canEdit={currentUser?.actualRole === 'OWNER'}
                        onPhotoChange={(url) => savePhoto(selectedEmployee.id, url)}
                        className="shrink-0 ring-4 ring-white/20 rounded-2xl"
                      />
                      <div className="flex-1">
                        <h3 className="text-xl font-extrabold">{selectedEmployee.name}</h3>
                        <div className="mt-1.5 flex flex-wrap items-center gap-2">
                          <span className="rounded-full bg-white/15 backdrop-blur-sm px-3 py-1 text-xs font-semibold">
                            {roleLabels[selectedEmployee.role] || selectedEmployee.role}
                          </span>
                          <span className="text-sm text-blue-200/80">{selectedEmployee.position || '-'}</span>
                          {selectedEmployee?.isLeader && (
                            <span className="rounded-full bg-amber-500/20 px-2.5 py-0.5 text-[10px] font-bold text-amber-300">ليدر</span>
                          )}
                          {selectedEmployee.status === 'SUSPENDED' && (
                            <span className="rounded-full bg-red-500/30 px-2.5 py-0.5 text-[10px] font-bold text-red-200">⚠ موقوف تلقائياً — محاولات وصول غير مخوّلة</span>
                          )}
                        </div>
                        {selectedEmployee.phone && (
                          <p className="mt-2 flex items-center gap-2 text-sm text-blue-200/60">
                            {selectedEmployee.phone}
                            <PhoneContactLinks phone={selectedEmployee.phone} />
                          </p>
                        )}
                      </div>
                      <div className={`flex flex-col items-center rounded-xl px-4 py-2 ${selectedEmployee.onDuty ? 'bg-emerald-500/20' : 'bg-white/10'}`}>
                        <span className={`h-2.5 w-2.5 rounded-full ${selectedEmployee.onDuty ? 'bg-emerald-400 animate-pulse' : 'bg-slate-400'}`}/>
                        <span className="mt-1 text-[10px] font-medium">{selectedEmployee.onDuty ? 'بالدوام' : 'خارج'}</span>
                      </div>
                    </div>
                    {canGenerateAiReport && selectedEmployee.status !== 'ARCHIVED' && selectedEmployee.status !== 'DELETED' && selectedEmployee.status !== 'SUSPENDED' && (
                      <div className="relative mt-4 flex flex-wrap gap-2">
                        <button
                          onClick={handleGenerateAiReport}
                          className="rounded-lg bg-indigo-500/20 px-3 py-1.5 text-xs font-bold text-indigo-200 hover:bg-indigo-500/30"
                        >
                          🧑‍💼🤖 اسأل الذكاء الاصطناعي عن هذا الموظف
                        </button>
                        {isAdmin && (
                          <button
                            onClick={handleLinkHistorical}
                            disabled={linkingHistorical}
                            className="rounded-lg bg-amber-500/20 px-3 py-1.5 text-xs font-bold text-amber-200 hover:bg-amber-500/30 disabled:opacity-50"
                          >
                            {linkingHistorical ? 'جاري الربط...' : '🔗 ربط السجلات التاريخية'}
                          </button>
                        )}
                      </div>
                    )}
                    {isAdmin && (
                      <div className="relative mt-4 flex flex-wrap gap-2">
                        {selectedEmployee.status === 'ARCHIVED' || selectedEmployee.status === 'DELETED' || selectedEmployee.status === 'SUSPENDED' ? (
                          <button
                            onClick={() => handleArchive('ACTIVE')}
                            className="rounded-lg bg-emerald-500/20 px-3 py-1.5 text-xs font-bold text-emerald-200 hover:bg-emerald-500/30"
                          >
                            ↩ استرجاع للنشطين
                          </button>
                        ) : (
                          <>
                            <button
                              onClick={() => handleArchive('ARCHIVED')}
                              className="rounded-lg bg-amber-500/20 px-3 py-1.5 text-xs font-bold text-amber-200 hover:bg-amber-500/30"
                            >
                              🗄️ أرشفة
                            </button>
                            <button
                              onClick={() => handleArchive('DELETED')}
                              className="rounded-lg bg-red-500/20 px-3 py-1.5 text-xs font-bold text-red-200 hover:bg-red-500/30"
                            >
                              🗑️ حذف
                            </button>
                          </>
                        )}
                      </div>
                    )}
                  </div>

                  {/* تبويبا لوحة التفاصيل: معلومات الموظف مقابل مهاراته —
                      بدل عمود واحد طويل يخلط الاثنين. */}
                  <div className="mt-5 flex border-b border-slate-200">
                    {[
                      { key: 'info' as const, label: '👤 معلومات' },
                      { key: 'skills' as const, label: '🛠️ المهارات' },
                    ].map(tab => (
                      <button key={tab.key} type="button" onClick={() => setPanelTab(tab.key)}
                        className={`px-5 py-2.5 text-sm font-bold transition-all border-b-2 -mb-px ${
                          panelTab === tab.key ? 'border-[#2c5aad] text-[#2c5aad]' : 'border-transparent text-slate-400 hover:text-slate-600'
                        }`}>{tab.label}</button>
                    ))}
                  </div>

                  {panelTab === 'info' && (
                  <>
                  {/* Stats Row */}
                  <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
                    <div className="rounded-2xl bg-gradient-to-br from-blue-50 to-indigo-50 p-4 text-center ring-1 ring-blue-100/50">
                      <p className="text-[10px] font-medium text-blue-400">المستوى</p>
                      <p className="mt-1 text-lg font-extrabold text-[#0f2040]">{currentLevel.level}</p>
                      <p className="text-[10px] text-blue-500 font-semibold">{currentLevel.label}</p>
                    </div>
                    <div className="rounded-2xl bg-gradient-to-br from-emerald-50 to-teal-50 p-4 text-center ring-1 ring-emerald-100/50">
                      <p className="text-[10px] font-medium text-emerald-400">المهارات</p>
                      <p className="mt-1 text-lg font-extrabold text-[#0f2040]">{skillCount}</p>
                      <p className="text-[10px] text-emerald-500 font-semibold">{nextLevel ? `${nextLevel.min - skillCount} للترقي` : 'أعلى مستوى'}</p>
                    </div>
                    {isHR && (() => {
                      const techStat = stats?.technicianStats.find(s => s.employeeId === selectedEmployee.id)
                      const completedCount = techStat?.completed || 0
                      const rank = Math.floor(completedCount / BOOKINGS_PER_RANK) + 1
                      const sortedTechs = stats ? [...stats.technicianStats].sort((a, b) => b.completed - a.completed) : []
                      const pos = sortedTechs.findIndex(s => s.employeeId === selectedEmployee.id)
                      return (
                        <>
                          <div className="rounded-2xl bg-gradient-to-br from-amber-50 to-orange-50 p-4 text-center ring-1 ring-amber-100/50">
                            <p className="text-[10px] font-medium text-amber-400">الرانك</p>
                            <p className="mt-1 text-lg font-extrabold text-[#0f2040]">{rank}</p>
                            <p className="text-[10px] text-amber-500 font-semibold">{completedCount} منجز</p>
                          </div>
                          <div className="rounded-2xl bg-gradient-to-br from-violet-50 to-purple-50 p-4 text-center ring-1 ring-violet-100/50">
                            <p className="text-[10px] font-medium text-violet-400">الترتيب</p>
                            <p className="mt-1 text-lg font-extrabold text-[#0f2040]">{pos >= 0 ? `#${pos + 1}` : '-'}</p>
                            <p className="text-[10px] text-violet-500 font-semibold">{sortedTechs.length > 0 ? `من ${sortedTechs.length}` : '-'}</p>
                          </div>
                        </>
                      )
                    })()}
                    {!isHR && (
                      <>
                        <div className="rounded-2xl bg-gradient-to-br from-amber-50 to-orange-50 p-4 text-center ring-1 ring-amber-100/50">
                          <p className="text-[10px] font-medium text-amber-400">الدوام</p>
                          <p className="mt-1 text-sm font-extrabold text-[#0f2040]">
                            {selectedEmployee.shiftStart && selectedEmployee.shiftEnd
                              ? `${selectedEmployee.shiftStart} - ${selectedEmployee.shiftEnd}`
                              : (selectedEmployee.shift === 'EVENING' ? 'مسائي' : 'صباحي')}
                          </p>
                        </div>
                        <div className="rounded-2xl bg-gradient-to-br from-violet-50 to-purple-50 p-4 text-center ring-1 ring-violet-100/50">
                          <p className="text-[10px] font-medium text-violet-400">الإجازات</p>
                          <p className="mt-1 text-lg font-extrabold text-[#0f2040]">{selectedEmployee?.monthlyLeaves ?? 0}</p>
                          <p className="text-[10px] text-violet-500 font-semibold">شهرياً</p>
                        </div>
                      </>
                    )}
                  </div>

                  {/* Admin/HR Fields */}
                  {(isAdmin || userPermissions.includes('edit_employee_profile')) && (
                    <div className="mt-5">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="h-1.5 w-1.5 rounded-full bg-[#2c5aad]"/>
                        <h4 className="text-sm font-bold text-[#0f2040]">بيانات الموظف</h4>
                      </div>
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">الاسم</label>
                          <input value={editName || selectedEmployee.name} onChange={(e) => setEditName(e.target.value)}
                            onBlur={() => handleFieldBlur('name', editName || selectedEmployee.name)} placeholder="اسم الموظف"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div>
                          <label className="mb-1 flex items-center gap-2 text-xs font-medium text-slate-400">
                            الهاتف
                            <PhoneContactLinks phone={selectedEmployee.phone} />
                          </label>
                          <input value={editPhone || selectedEmployee.phone || ''} onChange={(e) => setEditPhone(e.target.value)}
                            onBlur={() => handleFieldBlur('phone', editPhone || selectedEmployee.phone || '')} placeholder="07XXXXXXXXX"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">المنصب / التخصص</label>
                          <input value={editPosition || selectedEmployee.position || ''} onChange={(e) => setEditPosition(e.target.value)}
                            onBlur={() => handleFieldBlur('position', editPosition || selectedEmployee.position || '')} placeholder="مثال: كاميرات مراقبة" list="position-options-edit"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                          <datalist id="position-options-edit">
                            <option value="إداري كوادر" />
                            <option value="تقني" />
                            <option value="مهندس" />
                            <option value="مصمم" />
                          </datalist>
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">الشهادة</label>
                          <input value={editCertificate || selectedEmployee.certificate || ''} onChange={(e) => setEditCertificate(e.target.value)}
                            onBlur={() => handleFieldBlur('certificate', editCertificate || selectedEmployee.certificate || '')} placeholder="مثال: بكالوريوس هندسة"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">الراتب</label>
                          <input type="number" value={editSalary} onChange={(e) => setEditSalary(e.target.value)}
                            onBlur={() => handleFieldBlur('salary', Number(editSalary) || 0)} placeholder="0"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div className="sm:col-span-2">
                          <label className="mb-1 block text-xs font-medium text-slate-400">الدوام (من - إلى)</label>
                          <div className="flex items-center gap-2">
                            <input type="time"
                              value={editShiftStart || selectedEmployee.shiftStart || '08:00'}
                              onChange={(e) => setEditShiftStart(e.target.value)}
                              onBlur={() => handleFieldBlur('shiftStart', editShiftStart || selectedEmployee.shiftStart || '08:00')}
                              className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                            <span className="text-xs text-slate-400">إلى</span>
                            <input type="time"
                              value={editShiftEnd || selectedEmployee.shiftEnd || '16:00'}
                              onChange={(e) => setEditShiftEnd(e.target.value)}
                              onBlur={() => handleFieldBlur('shiftEnd', editShiftEnd || selectedEmployee.shiftEnd || '16:00')}
                              className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                          </div>
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">الإجازات الشهرية</label>
                          <input type="number" value={editMonthlyLeaves} onChange={(e) => setEditMonthlyLeaves(e.target.value)}
                            onBlur={() => handleFieldBlur('monthlyLeaves', Number(editMonthlyLeaves) || 0)} placeholder="0"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">العنوان الوظيفي</label>
                          <input value={editJobTitle} onChange={(e) => setEditJobTitle(e.target.value)}
                            onBlur={() => handleFieldBlur('jobTitle', editJobTitle)} placeholder="العنوان الوظيفي"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div className="flex items-center gap-3">
                          <label className="text-xs font-medium text-slate-400">ليدر فريق</label>
                          <button onClick={() => { const v = !editIsLeader; setEditIsLeader(v); handleFieldBlur('isLeader', v) }}
                            className={`relative h-7 w-12 rounded-full transition-colors duration-200 ${editIsLeader ? 'bg-emerald-500' : 'bg-slate-200'}`}>
                            <span className={`absolute top-0.5 h-6 w-6 rounded-full bg-white shadow-md transition-transform duration-200 ${editIsLeader ? 'right-0.5' : 'right-[22px]'}`} />
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* ملف الموارد البشرية — القسم والخبرة والتقييم، والحالة
                      الوظيفية تنحسب بالسيرفر مو تنكتب بالإيد */}
                  {(isAdmin || userPermissions.includes('edit_employee_profile')) && selectedEmployee && (
                    <EmployeeHRPanel
                      employee={selectedEmployee}
                      // selectedEmployee مشتق من القائمة، فتحديثها يحدّثه معاها
                      onUpdated={(u) => setEmployees((prev) => prev.map((x) => (x.id === u.id ? u : x)))}
                    />
                  )}

                  {!(isAdmin || userPermissions.includes('edit_employee_profile')) && (
                    <div className="mt-5">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="h-1.5 w-1.5 rounded-full bg-[#2c5aad]"/>
                        <h4 className="text-sm font-bold text-[#0f2040]">بيانات الموظف</h4>
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        {[
                          { label: 'الاسم', value: selectedEmployee.name },
                          { label: 'الهاتف', value: selectedEmployee.phone || '-' },
                          { label: 'المنصب / التخصص', value: selectedEmployee.position || '-' },
                          { label: 'الشهادة', value: selectedEmployee.certificate || '-' },
                          { label: 'الراتب', value: selectedEmployee.salary ?? '-' },
                          { label: 'الدوام', value: selectedEmployee.shiftStart && selectedEmployee.shiftEnd ? `${selectedEmployee.shiftStart} - ${selectedEmployee.shiftEnd}` : '-' },
                          { label: 'الإجازات الشهرية', value: selectedEmployee.monthlyLeaves ?? '-' },
                          { label: 'العنوان الوظيفي', value: selectedEmployee?.jobTitle || '-' },
                        ].map(f => (
                          <div key={f.label}>
                            <label className="mb-1 flex items-center gap-2 text-xs font-medium text-slate-400">
                              {f.label}
                              {f.label === 'الهاتف' && <PhoneContactLinks phone={selectedEmployee.phone} />}
                            </label>
                            <p className="rounded-xl border border-slate-100 bg-slate-50/50 px-3 py-2.5 text-sm text-slate-700">{f.value}</p>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Role & Status */}
                  <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
                    {isAdmin && (
                      <div>
                        <label className="mb-1 block text-xs font-medium text-slate-400">الصلاحية / الدور</label>
                        <select value={selectedEmployee.role}
                          onChange={async (e) => {
                            const nextRole = e.target.value as Employee['role']
                            try {
                              const updated = await api.updateEmployee(selectedEmployee.id, { role: nextRole })
                              setEmployees(prev => prev.map(emp => emp.id === updated.id ? { ...emp, ...updated } : emp))
                            } catch (err) {
                              alert(err instanceof Error ? err.message : 'تعذر تغيير الدور')
                            }
                          }}
                          className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white">
                          {Object.entries(roleLabels).filter(([k]) => k !== 'OWNER' && (k !== 'GPS_ADMIN' || selectedEmployee.role === 'GPS_ADMIN')).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                        </select>
                      </div>
                    )}
                    <div>
                      <label className="mb-1 block text-xs font-medium text-slate-400">حالة الدوام</label>
                      <label className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm cursor-pointer hover:bg-slate-50 transition-colors">
                        <input type="checkbox" checked={selectedEmployee.onDuty}
                          onChange={async (e) => {
                            const updated = await guard.run('تغيير حالة الدوام', () => api.updateEmployee(selectedEmployee.id, { onDuty: e.target.checked }))
                            if (!updated) return
                            setEmployees(prev => prev.map(emp => emp.id === updated.id ? { ...emp, ...updated } : emp))
                          }}
                          className="h-4 w-4 accent-[#2c5aad] rounded" />
                        متاح للتكليف حالياً
                      </label>
                    </div>
                    {isAdmin && (
                      <>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">رخصة القيادة</label>
                          <label className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm cursor-pointer hover:bg-slate-50 transition-colors">
                            <input type="checkbox" checked={selectedEmployee.hasDrivingLicense}
                              onChange={async (e) => {
                                const updated = await guard.run('تحديث رخصة السوق', () => api.updateEmployee(selectedEmployee.id, { hasDrivingLicense: e.target.checked }))
                                if (!updated) return
                                setEmployees(prev => prev.map(emp => emp.id === updated.id ? { ...emp, ...updated } : emp))
                              }}
                              className="h-4 w-4 accent-[#2c5aad] rounded" />
                            يملك رخصة قيادة
                          </label>
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">السلامة المهنية</label>
                          <label className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm cursor-pointer hover:bg-slate-50 transition-colors">
                            <input type="checkbox" checked={selectedEmployee.hasSafetyCertificate}
                              onChange={async (e) => {
                                const updated = await guard.run('تحديث شهادة السلامة', () => api.updateEmployee(selectedEmployee.id, { hasSafetyCertificate: e.target.checked }))
                                if (!updated) return
                                setEmployees(prev => prev.map(emp => emp.id === updated.id ? { ...emp, ...updated } : emp))
                              }}
                              className="h-4 w-4 accent-[#2c5aad] rounded" />
                            يملك شهادة السلامة
                          </label>
                        </div>
                      </>
                    )}
                  </div>

                  {/* بيانات الدخول — المالك وحده.
                      حط اسم مستخدم وباسورد على حساب موجود = استيلاء
                      عليه. لهذا نفس قيد فتح الحسابات ينطبق هنا،
                      والسيرفر يرفضها هم مو الواجهة بس. */}
                  {canCreateAccounts && (
                    <div className="mt-5">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="h-1.5 w-1.5 rounded-full bg-[#2c5aad]"/>
                        <h4 className="text-sm font-bold text-[#0f2040]">بيانات تسجيل الدخول</h4>
                      </div>
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">اسم المستخدم</label>
                          <input value={credUsername} onChange={(e) => setCredUsername(e.target.value)}
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs font-medium text-slate-400">كلمة مرور جديدة</label>
                          <input type="password" value={credPassword} onChange={(e) => setCredPassword(e.target.value)}
                            placeholder="اتركه فارغاً للإبقاء"
                            className="w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none transition-colors focus:border-[#2c5aad] focus:bg-white" />
                        </div>
                        <div className="flex items-end">
                          <button onClick={handleSaveCredentials} disabled={savingCreds}
                            className="w-full rounded-xl bg-gradient-to-l from-[#2c5aad] to-[#1e3f7a] px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-blue-900/20 transition-all hover:shadow-xl disabled:opacity-50">
                            {savingCreds ? 'جاري الحفظ...' : 'حفظ'}
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Stats & Compare */}
                  <div className="mt-5">
                    <div className="flex items-center gap-2 mb-3">
                      <div className="h-1.5 w-1.5 rounded-full bg-[#2c5aad]"/>
                      <h4 className="text-sm font-bold text-[#0f2040]">إحصائيات الموظف</h4>
                      <button type="button" onClick={() => setShowCompare(!showCompare)}
                        className="mr-auto rounded-lg bg-slate-100 px-3 py-1 text-[11px] font-semibold text-slate-500 transition-colors hover:bg-slate-200">
                        {showCompare ? 'إغلاق' : 'مقارنة'}
                      </button>
                    </div>
                    {(() => {
                      const empStats = getEmployeeStats(selectedEmployee.id)
                      return (
                        /* ⚠️ عدّادان مو ثلاثة: «ساعات إضافية» انشالت
                           لأن ماكو وراها بيانات — چانت صفراً مثبتاً. */
                        <div className="grid grid-cols-2 gap-3">
                          <StatTile label="المهام المنجزة" icon="✅" tone="success"
                            value={empStats.completed} />
                          <StatTile label="المهام الجارية" icon="⏳" tone="info"
                            value={empStats.inProgress} hint="المسند ناقص المنجز" />
                        </div>
                      )
                    })()}

                    {showCompare && (
                      <div className="mt-3 rounded-xl bg-slate-50 p-4 ring-1 ring-slate-100">
                        <select value={compareId || ''} onChange={(e) => setCompareId(e.target.value || null)}
                          className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-[#2c5aad] mb-3">
                          <option value="">-- اختر موظف --</option>
                          {employees.filter(e => e.id !== selectedEmployee.id).map(emp => (
                            <option key={emp.id} value={emp.id}>{emp.name}</option>
                          ))}
                        </select>
                        {compareEmployee && (() => {
                          const statsA = getEmployeeStats(selectedEmployee.id)
                          const statsB = getEmployeeStats(compareEmployee.id)
                          const skillCountA = selectedEmployee.skills.filter(s => s.canPerform).length
                          const skillCountB = compareEmployee.skills.filter(s => s.canPerform).length
                          return (
                            <div className="overflow-hidden rounded-xl ring-1 ring-slate-200">
                              <div className="overflow-x-auto">
                              <table className="w-full text-right text-sm">
                                <thead className="bg-gradient-to-l from-[#2c5aad] to-[#1e3f7a] text-white">
                                  <tr>
                                    <th className="px-3 py-2 text-xs font-semibold">المعيار</th>
                                    <th className="px-3 py-2 text-xs font-semibold">{selectedEmployee.name}</th>
                                    <th className="px-3 py-2 text-xs font-semibold">{compareEmployee.name}</th>
                                  </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100 bg-white text-xs">
                                  {[
                                    ['المهام المنجزة', statsA.completed, statsB.completed],
                                    ['المهام الجارية', statsA.inProgress, statsB.inProgress],
                                    ['عدد المهارات', skillCountA, skillCountB],
                                  ].map(([label, a, b]) => (
                                    <tr key={label as string}>
                                      <td className="px-3 py-2 font-medium text-slate-600">{label}</td>
                                      <td className="px-3 py-2">{a}</td>
                                      <td className="px-3 py-2">{b}</td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                              </div>
                            </div>
                          )
                        })()}
                      </div>
                    )}
                  </div>
                  </>
                  )}

                  {panelTab === 'skills' && (
                  <>
                  {/* Skills */}
                  <div className="mt-5">
                    <div className="flex items-center gap-2 mb-3">
                      <div className="h-1.5 w-1.5 rounded-full bg-[#2c5aad]"/>
                      <h4 className="text-sm font-bold text-[#0f2040]">المهارات</h4>
                      <span className="mr-auto rounded-full bg-blue-50 px-2.5 py-0.5 text-[10px] font-bold text-[#2c5aad]">{skillCount} مهارة</span>
                    </div>
                    <div className="mb-3 rounded-xl bg-gradient-to-l from-blue-50 to-indigo-50 px-4 py-2.5 text-xs text-[#2c5aad] font-medium ring-1 ring-blue-100/50">
                      المستوى {currentLevel.level} — {currentLevel.label}
                      {nextLevel && ` | يحتاج ${nextLevel.min - skillCount} مهارة إضافية للترقي`}
                    </div>
                    <div className="flex border-b border-slate-100 mb-3">
                      {[
                        { key: 'technical' as const, label: 'مهارات فنية' },
                        { key: 'leader' as const, label: 'مهارات قيادية' },
                      ].map(tab => (
                        <button key={tab.key} type="button" onClick={() => setSkillTab(tab.key)}
                          className={`px-4 py-2 text-xs font-semibold transition-all border-b-2 -mb-px ${
                            skillTab === tab.key ? 'border-[#2c5aad] text-[#2c5aad]' : 'border-transparent text-slate-400 hover:text-slate-600'
                          }`}>{tab.label}</button>
                      ))}
                    </div>
                    {skillTab === 'technical' && (
                      <>
                        {isHR ? (
                          <div className="flex flex-wrap gap-2">
                            {selectedEmployee.skills.filter(s => s.canPerform).map(s => (
                              <span key={s.id} className="rounded-full bg-blue-50 px-3 py-1.5 text-xs font-semibold text-[#2c5aad] ring-1 ring-blue-100/50">{s.skill.name}</span>
                            ))}
                            {selectedEmployee.skills.filter(s => s.canPerform).length === 0 && (
                              <p className="text-xs text-slate-400">لم يتم تحديد مهارات بعد.</p>
                            )}
                          </div>
                        ) : (
                          <>
                            <p className="mb-3 text-xs text-slate-400">حدد المهارات التي يستطيع الموظف تنفيذها.</p>
                            <div className="flex flex-col gap-4">
                              {services.filter(service => service.division === selectedEmployee.division).map(service => (
                                <div key={service.id}>
                                  <h5 className="mb-2 text-xs font-bold text-[#0f2040]">{service.name}</h5>
                                  <div className="grid grid-cols-1 gap-1.5 sm:grid-cols-2 lg:grid-cols-3">
                                    {service.skills.map(sk => {
                                      const skill = selectedEmployee.skills.find(s => s.skillId === sk.id)
                                      const canPerform = skill?.canPerform ?? false
                                      return (
                                        <label key={sk.id}
                                          className={`flex cursor-pointer items-center gap-2 rounded-xl border px-3 py-2 text-xs font-medium transition-all duration-200 ${
                                            canPerform
                                              ? 'border-[#2c5aad]/30 bg-blue-50/80 text-[#2c5aad] shadow-sm'
                                              : 'border-slate-100 text-slate-400 hover:border-slate-200 hover:bg-slate-50'
                                          }`}>
                                          <input type="checkbox" checked={canPerform} onChange={() => toggleSkill(selectedEmployee, sk.id)}
                                            className="h-3.5 w-3.5 accent-[#2c5aad] rounded" />
                                          {sk.name}
                                        </label>
                                      )
                                    })}
                                    {service.skills.length === 0 && <p className="text-xs text-slate-500">لا توجد مهارات.</p>}
                                  </div>
                                </div>
                              ))}
                            </div>
                          </>
                        )}
                      </>
                    )}
                    {skillTab === 'leader' && (
                      <div className="flex flex-col gap-3">
                        {LEADER_SKILLS.map(skill => (
                          <div key={skill} className="flex items-center gap-3">
                            <span className="w-24 text-xs font-medium text-slate-500">{skill}</span>
                            <input type="range" min={1} max={10}
                              value={leaderRatings[skill] ?? 5}
                              disabled={!canRateLeader}
                              onChange={(e) => setLeaderRatings(prev => ({ ...prev, [skill]: Number(e.target.value) }))}
                              className="flex-1 accent-[#2c5aad] disabled:opacity-60" />
                            <span className="w-8 text-center text-sm font-extrabold text-[#2c5aad]">
                              {leaderRatings[skill] ?? 5}
                            </span>
                          </div>
                        ))}

                        {canRateLeader ? (
                          <div className="flex items-center justify-between gap-2">
                            {/* ⚠️ «ماكو تقييم» مو «كلها ٥»: الأشرطة تبدي
                                بـ٥ لأنها منتصف المدى، فلازم يبين إذا
                                هذا رأي أحد لو مجرد وضع ابتدائي. */}
                            <span className="text-[11px] text-slate-400">
                              {Object.keys(leaderSaved).length === 0
                                ? 'ما انسجّل تقييم بعد — الأشرطة على منتصف المدى'
                                : `محفوظ ${Object.keys(leaderSaved).length} من ${LEADER_SKILLS.length}`}
                            </span>
                            <button
                              type="button"
                              disabled={leaderBusy}
                              onClick={async () => {
                                if (!selectedEmployee) return
                                setLeaderBusy(true)
                                const scores: Record<string, number> = {}
                                for (const k of LEADER_SKILLS) scores[k] = leaderRatings[k] ?? 5
                                const rows = await guard.run('حفظ تقييم القيادة', () =>
                                  api.setLeaderSkills(selectedEmployee.id, scores))
                                if (rows) {
                                  const map: Record<string, number> = {}
                                  for (const r of rows) map[r.skill] = r.score
                                  setLeaderRatings(map)
                                  setLeaderSaved(map)
                                }
                                setLeaderBusy(false)
                              }}
                              className="rounded-lg bg-gradient-to-l from-[#2c5aad] to-[#0f2040] px-4 py-1.5 text-xs font-bold text-white disabled:opacity-50">
                              {leaderBusy ? 'جاري الحفظ...' : 'احفظ التقييم'}
                            </button>
                          </div>
                        ) : (
                          <p className="text-[11px] text-slate-400">
                            التقييم للمالك والمدير وإداري الكوادر — إنت تشوف بس.
                          </p>
                        )}
                      </div>
                    )}
                  </div>
                  </>
                  )}
                </div>
              )
            })()}
          </div>
        </div>
      )}
    </div>
    </>
  )
}
