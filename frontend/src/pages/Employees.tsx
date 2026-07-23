import { useEffect, useState } from 'react'
import { api, type Employee, type Service, type Stats } from '../api'
import { useSession } from '../session'
import AddEmployeeWizard from '../components/AddEmployeeWizard'
import { openManagerChat } from '../components/ManagerAssistantChat'

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
}

const avatarGradients: string[] = [
  'from-blue-500 to-indigo-600',
  'from-emerald-500 to-teal-600',
  'from-violet-500 to-purple-600',
  'from-rose-500 to-pink-600',
  'from-amber-500 to-orange-600',
  'from-cyan-500 to-sky-600',
  'from-fuchsia-500 to-purple-600',
  'from-lime-500 to-green-600',
]

export default function Employees() {
  const { employee: currentUser, permissions: userPermissions } = useSession()
  const isAdmin = currentUser?.role === 'ADMIN'
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

  const [skillTab, setSkillTab] = useState<'technical' | 'leader'>('technical')
  const [leaderRatings, setLeaderRatings] = useState<Record<string, number>>({})
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
    setLoading(true)
    Promise.all([api.getEmployees(), api.getServices()])
      .then(([emps, svcs]) => { setEmployees(emps); setServices(svcs) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  useEffect(() => {
    if (isHR) api.getStats().then(setStats).catch(() => setStats(null))
  }, [isHR])
  useEffect(() => {
    if (isAdmin && showArchived) api.getArchivedEmployees().then(setArchivedEmployees).catch(() => setArchivedEmployees([]))
  }, [isAdmin, showArchived])

  const baseEmployees = showArchived ? archivedEmployees : (isHR ? employees.filter((emp) => emp.role === 'TECHNICIAN') : employees)
  const visibleEmployees = baseEmployees.filter(emp => {
    if (searchQuery && !emp.name.includes(searchQuery) && !(emp.position || '').includes(searchQuery)) return false
    if (filterRole && emp.role !== filterRole) return false
    return true
  })

  const selectedEmployee = [...employees, ...archivedEmployees].find((emp) => emp.id === selectedId) || null

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

  useEffect(() => {
    setCredUsername(selectedEmployee?.username || '')
    setCredPassword('')
    setEditSalary('')
    setEditShiftStart('')
    setEditShiftEnd('')
    setEditMonthlyLeaves('')
    setEditJobTitle('')
    setEditIsLeader((selectedEmployee as any)?.isLeader || false)
    setEditName('')
    setEditPhone('')
    setEditPosition('')
    setEditCertificate('')
    setShowCompare(false)
    setCompareId(null)
    setSkillTab('technical')
  }, [selectedId])

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

  const handleFieldBlur = async (field: string, value: any) => {
    if (!selectedEmployee) return
    try {
      const updated = await api.updateEmployee(selectedEmployee.id, { [field]: value } as any)
      setEmployees((prev) => prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)))
    } catch {}
  }

  const toggleSkill = async (employee: Employee, skillId: string) => {
    const current = new Map(employee.skills.map((s) => [s.skillId, s.canPerform]))
    current.set(skillId, !current.get(skillId))
    const skills = services.flatMap((svc) => svc.skills.map((sk) => ({ skillId: sk.id, canPerform: current.get(sk.id) ?? false })))
    const updated = await api.updateEmployeeSkills(employee.id, skills)
    setEmployees((prev) => prev.map((emp) => (emp.id === employee.id ? updated : emp)))
  }

  const getEmployeeStats = (empId: string) => {
    const techStat = stats?.technicianStats.find((s) => s.employeeId === empId)
    return { completed: techStat?.completed || 0, inProgress: techStat?.totalAssigned || 0, overtime: 0 }
  }

  const getAvatarGradient = (idx: number) => avatarGradients[idx % avatarGradients.length]

  const uniqueRoles = [...new Set(baseEmployees.map(e => e.role))]

  return (
    <div className="max-w-[1400px] mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-extrabold text-[#0f2040]">إدارة الكوادر</h2>
          <p className="mt-1 text-sm text-slate-400">إدارة بيانات الموظفين وتحديد المهارات</p>
        </div>
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
          {isAdmin && !showArchived && (
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
      {isAdmin && showAddForm && (
        <AddEmployeeWizard onClose={() => setShowAddForm(false)} onCreated={load} />
      )}

      {loading && <div className="flex items-center justify-center py-20"><div className="h-8 w-8 animate-spin rounded-full border-4 border-[#2c5aad] border-t-transparent"/></div>}
      {error && <p className="mt-6 rounded-2xl bg-red-50 p-5 text-red-600 text-sm">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="grid grid-cols-1 gap-6 xl:grid-cols-[340px_1fr]">
          {/* Employee Cards Grid */}
          <div className="space-y-3 max-h-[calc(100vh-200px)] overflow-y-auto pl-2 scrollbar-thin">
            <p className="text-xs font-medium text-slate-400 mb-2">{visibleEmployees.length} موظف</p>
            {visibleEmployees.map((emp, idx) => {
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
                  <div className={`relative flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${getAvatarGradient(idx)} text-base font-bold text-white shadow-md`}>
                    {emp.name.charAt(0)}
                    <span className={`absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-white ${emp.onDuty ? 'bg-emerald-500' : 'bg-slate-300'}`}/>
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
              <div className="flex flex-col items-center py-12 text-slate-300">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                <p className="mt-3 text-sm">لا يوجد موظفين</p>
              </div>
            )}
          </div>

          {/* Detail Panel */}
          <div className="rounded-2xl bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.06)] max-h-[calc(100vh-200px)] overflow-y-auto">
            {!selectedEmployee && (
              <div className="flex flex-col items-center justify-center py-20 text-slate-300">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
                <p className="mt-4 text-sm">اختر موظفاً لعرض التفاصيل</p>
              </div>
            )}

            {selectedEmployee && (() => {
              const skillCount = selectedEmployee.skills.filter(s => s.canPerform).length
              const currentLevel = [...levels].reverse().find(l => skillCount >= l.min) || levels[0]
              const nextLevel = levels.find(l => l.min > skillCount)
              const empIdx = employees.findIndex(e => e.id === selectedEmployee.id)

              return (
                <div>
                  {/* Hero Card */}
                  <div className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-[#2c5aad] to-[#0f2040] p-6 text-white">
                    <div className="absolute top-0 left-0 w-40 h-40 rounded-full bg-white/5 -translate-x-1/2 -translate-y-1/2"/>
                    <div className="absolute bottom-0 right-0 w-32 h-32 rounded-full bg-white/5 translate-x-1/3 translate-y-1/3"/>
                    <div className="relative flex items-start gap-5">
                      <div className={`flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br ${getAvatarGradient(empIdx)} text-2xl font-bold text-white shadow-xl ring-4 ring-white/20`}>
                        {selectedEmployee.name.charAt(0)}
                      </div>
                      <div className="flex-1">
                        <h3 className="text-xl font-extrabold">{selectedEmployee.name}</h3>
                        <div className="mt-1.5 flex flex-wrap items-center gap-2">
                          <span className="rounded-full bg-white/15 backdrop-blur-sm px-3 py-1 text-xs font-semibold">
                            {roleLabels[selectedEmployee.role] || selectedEmployee.role}
                          </span>
                          <span className="text-sm text-blue-200/80">{selectedEmployee.position || '-'}</span>
                          {(selectedEmployee as any)?.isLeader && (
                            <span className="rounded-full bg-amber-500/20 px-2.5 py-0.5 text-[10px] font-bold text-amber-300">ليدر</span>
                          )}
                          {selectedEmployee.status === 'SUSPENDED' && (
                            <span className="rounded-full bg-red-500/30 px-2.5 py-0.5 text-[10px] font-bold text-red-200">⚠ موقوف تلقائياً — محاولات وصول غير مخوّلة</span>
                          )}
                        </div>
                        {selectedEmployee.phone && (
                          <p className="mt-2 text-sm text-blue-200/60">{selectedEmployee.phone}</p>
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
                          <p className="mt-1 text-lg font-extrabold text-[#0f2040]">{(selectedEmployee as any).monthlyLeaves ?? 0}</p>
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
                          <label className="mb-1 block text-xs font-medium text-slate-400">الهاتف</label>
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
                          { label: 'العنوان الوظيفي', value: (selectedEmployee as any).jobTitle || '-' },
                        ].map(f => (
                          <div key={f.label}>
                            <label className="mb-1 block text-xs font-medium text-slate-400">{f.label}</label>
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
                          {Object.entries(roleLabels).filter(([k]) => k !== 'OWNER').map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                        </select>
                      </div>
                    )}
                    <div>
                      <label className="mb-1 block text-xs font-medium text-slate-400">حالة الدوام</label>
                      <label className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm cursor-pointer hover:bg-slate-50 transition-colors">
                        <input type="checkbox" checked={selectedEmployee.onDuty}
                          onChange={async (e) => {
                            const updated = await api.updateEmployee(selectedEmployee.id, { onDuty: e.target.checked })
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
                                const updated = await api.updateEmployee(selectedEmployee.id, { hasDrivingLicense: e.target.checked })
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
                                const updated = await api.updateEmployee(selectedEmployee.id, { hasSafetyCertificate: e.target.checked })
                                setEmployees(prev => prev.map(emp => emp.id === updated.id ? { ...emp, ...updated } : emp))
                              }}
                              className="h-4 w-4 accent-[#2c5aad] rounded" />
                            يملك شهادة السلامة
                          </label>
                        </div>
                      </>
                    )}
                  </div>

                  {/* Login Credentials — Admin only */}
                  {isAdmin && (
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
                        <div className="grid grid-cols-3 gap-3">
                          <div className="rounded-xl bg-emerald-50/80 p-3 text-center ring-1 ring-emerald-100/50">
                            <p className="text-[10px] text-emerald-500 font-medium">المهام المنجزة</p>
                            <p className="mt-1 text-xl font-extrabold text-emerald-700">{empStats.completed}</p>
                          </div>
                          <div className="rounded-xl bg-blue-50/80 p-3 text-center ring-1 ring-blue-100/50">
                            <p className="text-[10px] text-blue-500 font-medium">المهام الجارية</p>
                            <p className="mt-1 text-xl font-extrabold text-blue-700">{empStats.inProgress}</p>
                          </div>
                          <div className="rounded-xl bg-amber-50/80 p-3 text-center ring-1 ring-amber-100/50">
                            <p className="text-[10px] text-amber-500 font-medium">ساعات إضافية</p>
                            <p className="mt-1 text-xl font-extrabold text-amber-700">{empStats.overtime}</p>
                          </div>
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
                                    ['ساعات إضافية', statsA.overtime, statsB.overtime],
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
                              {services.map(service => (
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
                                    {service.skills.length === 0 && <p className="text-xs text-slate-300">لا توجد مهارات.</p>}
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
                        {['القيادة', 'إدارة الفريق', 'حل المشكلات', 'التواصل', 'اتخاذ القرار'].map(skill => (
                          <div key={skill} className="flex items-center gap-3">
                            <span className="w-24 text-xs font-medium text-slate-500">{skill}</span>
                            <input type="range" min={1} max={10} value={leaderRatings[skill] || 5}
                              onChange={(e) => setLeaderRatings(prev => ({ ...prev, [skill]: Number(e.target.value) }))}
                              className="flex-1 accent-[#2c5aad]" />
                            <span className="w-8 text-center text-sm font-extrabold text-[#2c5aad]">{leaderRatings[skill] || 5}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )
            })()}
          </div>
        </div>
      )}
    </div>
  )
}
