import { useEffect, useState } from 'react'
import { api, type Employee, type Service, type Stats } from '../api'
import { useSession } from '../session'

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
  GPS_ENGINEER: 'مهندس GPS',
  QUALITY_ENGINEER: 'مهندس جودة',
}

export default function Employees() {
  const { employee: currentUser } = useSession()
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

  // New employee form
  const [name, setName] = useState('')
  const [certificate, setCertificate] = useState('')
  const [position, setPosition] = useState('')
  const [phone, setPhone] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [formSalary, setFormSalary] = useState('')
  const [formShift, setFormShift] = useState<'morning' | 'evening'>('morning')
  const [formJobTitle, setFormJobTitle] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Detail panel state
  const [skillTab, setSkillTab] = useState<'technical' | 'leader'>('technical')
  const [leaderRatings, setLeaderRatings] = useState<Record<string, number>>({})
  const [editSalary, setEditSalary] = useState('')
  const [editShift, setEditShift] = useState<'morning' | 'evening'>('morning')
  const [editMonthlyLeaves, setEditMonthlyLeaves] = useState('')
  const [editJobTitle, setEditJobTitle] = useState('')

  // Compare
  const [showCompare, setShowCompare] = useState(false)
  const [compareId, setCompareId] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    Promise.all([api.getEmployees(), api.getServices()])
      .then(([emps, svcs]) => {
        setEmployees(emps)
        setServices(svcs)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  useEffect(() => {
    if (isHR) api.getStats().then(setStats).catch(() => setStats(null))
  }, [isHR])

  const handleAddEmployee = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createEmployee({
        name,
        certificate: certificate || null,
        position: position || null,
        phone: phone || null,
        username: username || undefined,
        password: password || undefined,
        salary: formSalary ? Number(formSalary) : undefined,
        shift: formShift,
        jobTitle: formJobTitle || undefined,
      } as any)
      setName('')
      setCertificate('')
      setPosition('')
      setPhone('')
      setUsername('')
      setPassword('')
      setFormSalary('')
      setFormShift('morning')
      setFormJobTitle('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const visibleEmployees = isHR ? employees.filter((emp) => emp.role === 'TECHNICIAN') : employees

  const selectedEmployee = employees.find((emp) => emp.id === selectedId) || null
  const compareEmployee = employees.find((emp) => emp.id === compareId) || null

  useEffect(() => {
    setCredUsername(selectedEmployee?.username || '')
    setCredPassword('')
    setEditSalary('')
    setEditShift('morning')
    setEditMonthlyLeaves('')
    setEditJobTitle('')
    setShowCompare(false)
    setCompareId(null)
    setSkillTab('technical')
  }, [selectedId])

  const handleSaveCredentials = async () => {
    if (!selectedEmployee) return
    setSavingCreds(true)
    try {
      const updated = await api.updateEmployee(selectedEmployee.id, {
        username: credUsername,
        ...(credPassword ? { password: credPassword } : {}),
      })
      setEmployees((prev) => prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)))
      setCredPassword('')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSavingCreds(false)
    }
  }

  const handleFieldBlur = async (field: string, value: any) => {
    if (!selectedEmployee) return
    try {
      const updated = await api.updateEmployee(selectedEmployee.id, { [field]: value } as any)
      setEmployees((prev) => prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)))
    } catch {
      // silently ignore if backend doesn't support the field
    }
  }

  const toggleSkill = async (employee: Employee, skillId: string) => {
    const current = new Map(employee.skills.map((s) => [s.skillId, s.canPerform]))
    const newValue = !current.get(skillId)
    current.set(skillId, newValue)

    const skills = services.flatMap((svc) =>
      svc.skills.map((sk) => ({
        skillId: sk.id,
        canPerform: current.get(sk.id) ?? false,
      })),
    )

    const updated = await api.updateEmployeeSkills(employee.id, skills)
    setEmployees((prev) => prev.map((emp) => (emp.id === employee.id ? updated : emp)))
  }

  const getEmployeeStats = (empId: string) => {
    const techStat = stats?.technicianStats.find((s) => s.employeeId === empId)
    return {
      completed: techStat?.completed || 0,
      inProgress: techStat?.today || 0,
      overtime: 0,
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة الكوادر</h2>
      <p className="mt-1 text-slate-500">
        إدارة بيانات الموظفين وتحديد المهارات (الخدمات) التي يستطيع كل موظف تنفيذها.
      </p>

      {isAdmin && (
      <form
        onSubmit={handleAddEmployee}
        className="mt-6 grid grid-cols-1 gap-4 rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الاسم</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الشهادة</label>
          <input
            value={certificate}
            onChange={(e) => setCertificate(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">المنصب</label>
          <input
            value={position}
            onChange={(e) => setPosition(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الهاتف</label>
          <input
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">اسم المستخدم</label>
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">كلمة المرور</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الراتب</label>
          <input
            type="number"
            value={formSalary}
            onChange={(e) => setFormSalary(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الدوام</label>
          <select
            value={formShift}
            onChange={(e) => setFormShift(e.target.value as 'morning' | 'evening')}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          >
            <option value="morning">صباحي</option>
            <option value="evening">مسائي</option>
          </select>
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">العنوان الوظيفي</label>
          <input
            value={formJobTitle}
            onChange={(e) => setFormJobTitle(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-4">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'إضافة موظف'}
          </button>
        </div>
      </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div className="overflow-hidden rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-1">
            <table className="w-full text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الاسم</th>
                  <th className="px-4 py-3 text-sm font-semibold">المنصب</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {visibleEmployees.map((emp) => (
                  <tr
                    key={emp.id}
                    onClick={() => setSelectedId(emp.id)}
                    className={`cursor-pointer transition-colors ${
                      selectedId === emp.id ? 'bg-brand-50' : 'hover:bg-slate-50'
                    }`}
                  >
                    <td className="px-4 py-3 font-medium">{emp.name}</td>
                    <td className="px-4 py-3 text-slate-500">{emp.position || '-'}</td>
                  </tr>
                ))}
                {visibleEmployees.length === 0 && (
                  <tr>
                    <td colSpan={2} className="px-4 py-6 text-center text-slate-400">
                      لا يوجد موظفين بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-2">
            {!selectedEmployee && (
              <p className="text-slate-400">اختر موظفاً من القائمة لعرض/تعديل مهاراته.</p>
            )}
            {selectedEmployee && isHR && (
              <div>
                {/* Employee Info Card */}
                <div className="rounded-2xl bg-gradient-to-l from-brand-500 to-brand-800 p-5 text-white shadow-lg shadow-brand-900/20">
                  <h3 className="text-xl font-extrabold">{selectedEmployee.name}</h3>
                  <div className="mt-1 flex items-center gap-3">
                    <span className="rounded-full bg-white/20 px-3 py-0.5 text-xs font-medium">
                      {roleLabels[selectedEmployee.role] || selectedEmployee.role}
                    </span>
                    <span className="text-sm text-brand-100">{selectedEmployee.position || 'فني'}</span>
                  </div>
                  {selectedEmployee.phone && (
                    <p className="mt-2 text-sm text-brand-100">{selectedEmployee.phone}</p>
                  )}
                </div>

                {/* Editable Info Fields */}
                <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">الراتب</label>
                    <input
                      type="number"
                      value={editSalary}
                      onChange={(e) => setEditSalary(e.target.value)}
                      onBlur={() => handleFieldBlur('salary', Number(editSalary) || 0)}
                      placeholder="0"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">الدوام</label>
                    <div className="flex rounded-lg border border-slate-300 overflow-hidden">
                      <button
                        type="button"
                        onClick={() => { setEditShift('morning'); handleFieldBlur('shift', 'morning') }}
                        className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${editShift === 'morning' ? 'bg-brand-500 text-white' : 'bg-white text-slate-600'}`}
                      >
                        صباحي
                      </button>
                      <button
                        type="button"
                        onClick={() => { setEditShift('evening'); handleFieldBlur('shift', 'evening') }}
                        className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${editShift === 'evening' ? 'bg-brand-500 text-white' : 'bg-white text-slate-600'}`}
                      >
                        مسائي
                      </button>
                    </div>
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">الإجازات الشهرية</label>
                    <input
                      type="number"
                      value={editMonthlyLeaves}
                      onChange={(e) => setEditMonthlyLeaves(e.target.value)}
                      onBlur={() => handleFieldBlur('monthlyLeaves', Number(editMonthlyLeaves) || 0)}
                      placeholder="0"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">العنوان الوظيفي</label>
                    <input
                      value={editJobTitle}
                      onChange={(e) => setEditJobTitle(e.target.value)}
                      onBlur={() => handleFieldBlur('jobTitle', editJobTitle)}
                      placeholder="العنوان الوظيفي"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                </div>

                {(() => {
                  const skillCount = selectedEmployee.skills.filter((s) => s.canPerform).length
                  const currentLevel =
                    [...levels].reverse().find((l) => skillCount >= l.min) || levels[0]
                  const nextLevel = levels.find((l) => l.min > skillCount)
                  const techStat = stats?.technicianStats.find(
                    (s) => s.employeeId === selectedEmployee.id,
                  )
                  const completedCount = techStat?.completed || 0
                  const rank = Math.floor(completedCount / BOOKINGS_PER_RANK) + 1
                  const sortedTechs = stats
                    ? [...stats.technicianStats].sort((a, b) => b.completed - a.completed)
                    : []
                  const position = sortedTechs.findIndex((s) => s.employeeId === selectedEmployee.id)

                  return (
                    <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-4">
                      <div className="rounded-2xl bg-gradient-to-l from-brand-500 to-brand-800 p-4 text-white shadow-lg shadow-brand-900/20">
                        <p className="text-xs text-brand-100">المستوى</p>
                        <p className="mt-1 text-xl font-extrabold">
                          {currentLevel.level} - {currentLevel.label}
                        </p>
                        <p className="mt-1 text-xs text-brand-100">
                          {skillCount} مهارة معتمدة
                          {nextLevel ? ` - يحتاج ${nextLevel.min - skillCount} للترقي` : ' - أعلى مستوى'}
                        </p>
                      </div>
                      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                        <p className="text-xs text-slate-500">الرانك</p>
                        <p className="mt-1 text-xl font-extrabold text-emerald-700">{rank}</p>
                        <p className="mt-1 text-xs text-slate-500">{completedCount} حجز منجز</p>
                      </div>
                      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                        <p className="text-xs text-slate-500">الترتيب بين الفنيين</p>
                        <p className="mt-1 text-xl font-extrabold text-brand-700">
                          {position >= 0 ? `#${position + 1}` : '-'}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {sortedTechs.length > 0 ? `من ${sortedTechs.length}` : '-'}
                        </p>
                      </div>
                      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                        <p className="text-xs text-slate-500">حالة الدوام</p>
                        <p
                          className={`mt-1 text-xl font-extrabold ${
                            selectedEmployee.onDuty ? 'text-emerald-700' : 'text-slate-400'
                          }`}
                        >
                          {selectedEmployee.onDuty ? 'بالدوام' : 'خارج الدوام'}
                        </p>
                      </div>
                    </div>
                  )
                })()}

                {/* Employee Statistics Section */}
                <div className="mt-6">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="h-2 w-2 rounded-full bg-brand-500" />
                    <h4 className="font-bold text-brand-800">إحصائيات الموظف</h4>
                  </div>
                  {(() => {
                    const empStats = getEmployeeStats(selectedEmployee.id)
                    return (
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center">
                          <p className="text-xs text-slate-500">عدد المهام المنجزة</p>
                          <p className="mt-1 text-2xl font-extrabold text-emerald-700">{empStats.completed}</p>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center">
                          <p className="text-xs text-slate-500">عدد المهام الجارية</p>
                          <p className="mt-1 text-2xl font-extrabold text-brand-700">{empStats.inProgress}</p>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center">
                          <p className="text-xs text-slate-500">الساعات الإضافية</p>
                          <p className="mt-1 text-2xl font-extrabold text-amber-600">{empStats.overtime}</p>
                        </div>
                      </div>
                    )
                  })()}
                  <button
                    type="button"
                    onClick={() => setShowCompare(!showCompare)}
                    className="mt-3 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
                  >
                    {showCompare ? 'إغلاق المقارنة' : 'مقارنة'}
                  </button>

                  {showCompare && (
                    <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-4">
                      <label className="mb-2 block text-sm font-medium text-slate-600">اختر موظف للمقارنة</label>
                      <select
                        value={compareId || ''}
                        onChange={(e) => setCompareId(e.target.value || null)}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 mb-4"
                      >
                        <option value="">-- اختر موظف --</option>
                        {employees.filter((e) => e.id !== selectedEmployee.id).map((emp) => (
                          <option key={emp.id} value={emp.id}>{emp.name}</option>
                        ))}
                      </select>
                      {compareEmployee && (() => {
                        const statsA = getEmployeeStats(selectedEmployee.id)
                        const statsB = getEmployeeStats(compareEmployee.id)
                        const skillCountA = selectedEmployee.skills.filter((s) => s.canPerform).length
                        const skillCountB = compareEmployee.skills.filter((s) => s.canPerform).length
                        return (
                          <div className="overflow-hidden rounded-2xl">
                            <table className="w-full text-right text-sm">
                              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                                <tr>
                                  <th className="px-4 py-2">المعيار</th>
                                  <th className="px-4 py-2">{selectedEmployee.name}</th>
                                  <th className="px-4 py-2">{compareEmployee.name}</th>
                                </tr>
                              </thead>
                              <tbody className="divide-y divide-slate-100 bg-white">
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">المهام المنجزة</td>
                                  <td className="px-4 py-2">{statsA.completed}</td>
                                  <td className="px-4 py-2">{statsB.completed}</td>
                                </tr>
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">المهام الجارية</td>
                                  <td className="px-4 py-2">{statsA.inProgress}</td>
                                  <td className="px-4 py-2">{statsB.inProgress}</td>
                                </tr>
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">الساعات الإضافية</td>
                                  <td className="px-4 py-2">{statsA.overtime}</td>
                                  <td className="px-4 py-2">{statsB.overtime}</td>
                                </tr>
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">عدد المهارات</td>
                                  <td className="px-4 py-2">{skillCountA}</td>
                                  <td className="px-4 py-2">{skillCountB}</td>
                                </tr>
                              </tbody>
                            </table>
                          </div>
                        )
                      })()}
                    </div>
                  )}
                </div>

                {/* Skills Section with Tabs */}
                <div className="mt-6">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="h-2 w-2 rounded-full bg-brand-500" />
                    <h4 className="font-bold text-brand-800">المهارات المتقنة</h4>
                    <span className="mr-auto text-xs text-slate-400">
                      {selectedEmployee.skills.filter((s) => s.canPerform).length} مهارة
                    </span>
                  </div>
                  {(() => {
                    const skillCount = selectedEmployee.skills.filter((s) => s.canPerform).length
                    const currentLevel = [...levels].reverse().find((l) => skillCount >= l.min) || levels[0]
                    const nextLevel = levels.find((l) => l.min > skillCount)
                    return (
                      <div className="mb-3 rounded-lg bg-brand-50 px-3 py-2 text-xs text-brand-700">
                        المستوى {currentLevel.level} - {currentLevel.label}
                        {nextLevel && ` | يحتاج ${nextLevel.min - skillCount} مهارة إضافية للترقي`}
                      </div>
                    )
                  })()}
                  <div className="flex border-b border-slate-200 mb-3">
                    <button
                      type="button"
                      onClick={() => setSkillTab('technical')}
                      className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${skillTab === 'technical' ? 'border-brand-500 text-brand-700' : 'border-transparent text-slate-400 hover:text-slate-600'}`}
                    >
                      مهارات فنية
                    </button>
                    <button
                      type="button"
                      onClick={() => setSkillTab('leader')}
                      className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${skillTab === 'leader' ? 'border-brand-500 text-brand-700' : 'border-transparent text-slate-400 hover:text-slate-600'}`}
                    >
                      مهارات قيادية
                    </button>
                  </div>
                  {skillTab === 'technical' && (
                    <div className="flex flex-wrap gap-2">
                      {selectedEmployee.skills
                        .filter((s) => s.canPerform)
                        .map((s) => (
                          <span
                            key={s.id}
                            className="rounded-full bg-brand-50 px-3 py-1 text-xs font-medium text-brand-700"
                          >
                            {s.skill.name}
                          </span>
                        ))}
                      {selectedEmployee.skills.filter((s) => s.canPerform).length === 0 && (
                        <p className="text-sm text-slate-400">لم يتم تحديد مهارات بعد.</p>
                      )}
                    </div>
                  )}
                  {skillTab === 'leader' && (
                    <div className="flex flex-col gap-3">
                      {['القيادة', 'إدارة الفريق', 'حل المشكلات', 'التواصل', 'اتخاذ القرار'].map((skill) => (
                        <div key={skill} className="flex items-center gap-3">
                          <span className="w-28 text-sm text-slate-600">{skill}</span>
                          <input
                            type="range"
                            min={1}
                            max={10}
                            value={leaderRatings[skill] || 5}
                            onChange={(e) => setLeaderRatings((prev) => ({ ...prev, [skill]: Number(e.target.value) }))}
                            className="flex-1 accent-brand-500"
                          />
                          <span className="w-8 text-center text-sm font-bold text-brand-700">
                            {leaderRatings[skill] || 5}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            )}
            {selectedEmployee && !isHR && (
              <div>
                {/* Employee Info Card */}
                <div className="rounded-2xl bg-gradient-to-l from-brand-500 to-brand-800 p-5 text-white shadow-lg shadow-brand-900/20">
                  <h3 className="text-xl font-extrabold">{selectedEmployee.name}</h3>
                  <div className="mt-1 flex items-center gap-3">
                    <span className="rounded-full bg-white/20 px-3 py-0.5 text-xs font-medium">
                      {roleLabels[selectedEmployee.role] || selectedEmployee.role}
                    </span>
                    <span className="text-sm text-brand-100">{selectedEmployee.position || '-'}</span>
                  </div>
                  {selectedEmployee.phone && (
                    <p className="mt-2 text-sm text-brand-100">{selectedEmployee.phone}</p>
                  )}
                </div>

                {/* Editable Info Fields */}
                <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">الراتب</label>
                    <input
                      type="number"
                      value={editSalary}
                      onChange={(e) => setEditSalary(e.target.value)}
                      onBlur={() => handleFieldBlur('salary', Number(editSalary) || 0)}
                      placeholder="0"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">الدوام</label>
                    <div className="flex rounded-lg border border-slate-300 overflow-hidden">
                      <button
                        type="button"
                        onClick={() => { setEditShift('morning'); handleFieldBlur('shift', 'morning') }}
                        className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${editShift === 'morning' ? 'bg-brand-500 text-white' : 'bg-white text-slate-600'}`}
                      >
                        صباحي
                      </button>
                      <button
                        type="button"
                        onClick={() => { setEditShift('evening'); handleFieldBlur('shift', 'evening') }}
                        className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${editShift === 'evening' ? 'bg-brand-500 text-white' : 'bg-white text-slate-600'}`}
                      >
                        مسائي
                      </button>
                    </div>
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">الإجازات الشهرية</label>
                    <input
                      type="number"
                      value={editMonthlyLeaves}
                      onChange={(e) => setEditMonthlyLeaves(e.target.value)}
                      onBlur={() => handleFieldBlur('monthlyLeaves', Number(editMonthlyLeaves) || 0)}
                      placeholder="0"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">العنوان الوظيفي</label>
                    <input
                      value={editJobTitle}
                      onChange={(e) => setEditJobTitle(e.target.value)}
                      onBlur={() => handleFieldBlur('jobTitle', editJobTitle)}
                      placeholder="العنوان الوظيفي"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                </div>

                <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {isAdmin && (
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      الصلاحية / الدور
                    </label>
                    <select
                      value={selectedEmployee.role}
                      onChange={async (e) => {
                        const updated = await api.updateEmployee(selectedEmployee.id, {
                          role: e.target.value as Employee['role'],
                        })
                        setEmployees((prev) =>
                          prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)),
                        )
                      }}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    >
                      <option value="ADMIN">مدير النظام</option>
                      <option value="SALES">موظف مبيعات</option>
                      <option value="HR_COORDINATOR">إداري الكوادر</option>
                      <option value="TECHNICIAN">فني</option>
                      <option value="PROJECT_MANAGER">مدير مشاريع</option>
                      <option value="MONITOR">مراقب</option>
                      <option value="FINANCE">محاسب</option>
                      <option value="GPS_ADMIN">مسؤول GPS</option>
                      <option value="GPS_ENGINEER">مهندس GPS</option>
                      <option value="QUALITY_ENGINEER">مهندس جودة</option>
                    </select>
                  </div>
                  )}
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      حالة الدوام
                    </label>
                    <label className="flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm">
                      <input
                        type="checkbox"
                        checked={selectedEmployee.onDuty}
                        onChange={async (e) => {
                          const updated = await api.updateEmployee(selectedEmployee.id, {
                            onDuty: e.target.checked,
                          })
                          setEmployees((prev) =>
                            prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)),
                          )
                        }}
                        className="h-4 w-4 accent-brand-700"
                      />
                      متاح للتكليف حالياً (بالدوام)
                    </label>
                  </div>
                  {isAdmin && (
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      رخصة القيادة
                    </label>
                    <label className="flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm">
                      <input
                        type="checkbox"
                        checked={selectedEmployee.hasDrivingLicense}
                        onChange={async (e) => {
                          const updated = await api.updateEmployee(selectedEmployee.id, {
                            hasDrivingLicense: e.target.checked,
                          })
                          setEmployees((prev) =>
                            prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)),
                          )
                        }}
                        className="h-4 w-4 accent-brand-700"
                      />
                      يملك رخصة قيادة
                    </label>
                  </div>
                  )}
                  {isAdmin && (
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      السلامة المهنية
                    </label>
                    <label className="flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm">
                      <input
                        type="checkbox"
                        checked={selectedEmployee.hasSafetyCertificate}
                        onChange={async (e) => {
                          const updated = await api.updateEmployee(selectedEmployee.id, {
                            hasSafetyCertificate: e.target.checked,
                          })
                          setEmployees((prev) =>
                            prev.map((emp) => (emp.id === updated.id ? { ...emp, ...updated } : emp)),
                          )
                        }}
                        className="h-4 w-4 accent-brand-700"
                      />
                      يملك شهادة السلامة المهنية
                    </label>
                  </div>
                  )}
                </div>

                {isAdmin && (
                <>
                <div className="mt-5 flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-brand-500" />
                  <h4 className="font-bold text-brand-800">بيانات تسجيل الدخول</h4>
                </div>
                <p className="mt-1 text-sm text-slate-500">
                  حدد اسم مستخدم وكلمة مرور حتى يدخل هذا الموظف لحسابه الخاص بالنظام.
                </p>
                <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      اسم المستخدم
                    </label>
                    <input
                      value={credUsername}
                      onChange={(e) => setCredUsername(e.target.value)}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      كلمة مرور جديدة
                    </label>
                    <input
                      type="password"
                      value={credPassword}
                      onChange={(e) => setCredPassword(e.target.value)}
                      placeholder="اتركه فارغاً للإبقاء على القديمة"
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div className="flex items-end">
                    <button
                      onClick={handleSaveCredentials}
                      disabled={savingCreds}
                      className="w-full rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg disabled:opacity-50"
                    >
                      {savingCreds ? 'جاري الحفظ...' : 'حفظ بيانات الدخول'}
                    </button>
                  </div>
                </div>
                </>
                )}

                {/* Employee Statistics Section */}
                <div className="mt-6">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="h-2 w-2 rounded-full bg-brand-500" />
                    <h4 className="font-bold text-brand-800">إحصائيات الموظف</h4>
                  </div>
                  {(() => {
                    const empStats = getEmployeeStats(selectedEmployee.id)
                    return (
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center">
                          <p className="text-xs text-slate-500">عدد المهام المنجزة</p>
                          <p className="mt-1 text-2xl font-extrabold text-emerald-700">{empStats.completed}</p>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center">
                          <p className="text-xs text-slate-500">عدد المهام الجارية</p>
                          <p className="mt-1 text-2xl font-extrabold text-brand-700">{empStats.inProgress}</p>
                        </div>
                        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-center">
                          <p className="text-xs text-slate-500">الساعات الإضافية</p>
                          <p className="mt-1 text-2xl font-extrabold text-amber-600">{empStats.overtime}</p>
                        </div>
                      </div>
                    )
                  })()}
                  <button
                    type="button"
                    onClick={() => setShowCompare(!showCompare)}
                    className="mt-3 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
                  >
                    {showCompare ? 'إغلاق المقارنة' : 'مقارنة'}
                  </button>

                  {showCompare && (
                    <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-4">
                      <label className="mb-2 block text-sm font-medium text-slate-600">اختر موظف للمقارنة</label>
                      <select
                        value={compareId || ''}
                        onChange={(e) => setCompareId(e.target.value || null)}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 mb-4"
                      >
                        <option value="">-- اختر موظف --</option>
                        {employees.filter((e) => e.id !== selectedEmployee.id).map((emp) => (
                          <option key={emp.id} value={emp.id}>{emp.name}</option>
                        ))}
                      </select>
                      {compareEmployee && (() => {
                        const statsA = getEmployeeStats(selectedEmployee.id)
                        const statsB = getEmployeeStats(compareEmployee.id)
                        const skillCountA = selectedEmployee.skills.filter((s) => s.canPerform).length
                        const skillCountB = compareEmployee.skills.filter((s) => s.canPerform).length
                        return (
                          <div className="overflow-hidden rounded-2xl">
                            <table className="w-full text-right text-sm">
                              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                                <tr>
                                  <th className="px-4 py-2">المعيار</th>
                                  <th className="px-4 py-2">{selectedEmployee.name}</th>
                                  <th className="px-4 py-2">{compareEmployee.name}</th>
                                </tr>
                              </thead>
                              <tbody className="divide-y divide-slate-100 bg-white">
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">المهام المنجزة</td>
                                  <td className="px-4 py-2">{statsA.completed}</td>
                                  <td className="px-4 py-2">{statsB.completed}</td>
                                </tr>
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">المهام الجارية</td>
                                  <td className="px-4 py-2">{statsA.inProgress}</td>
                                  <td className="px-4 py-2">{statsB.inProgress}</td>
                                </tr>
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">الساعات الإضافية</td>
                                  <td className="px-4 py-2">{statsA.overtime}</td>
                                  <td className="px-4 py-2">{statsB.overtime}</td>
                                </tr>
                                <tr>
                                  <td className="px-4 py-2 font-medium text-slate-600">عدد المهارات</td>
                                  <td className="px-4 py-2">{skillCountA}</td>
                                  <td className="px-4 py-2">{skillCountB}</td>
                                </tr>
                              </tbody>
                            </table>
                          </div>
                        )
                      })()}
                    </div>
                  )}
                </div>

                {/* Skills Section with Tabs */}
                <div className="mt-6">
                  <div className="flex items-center gap-2 mb-3">
                    <span className="h-2 w-2 rounded-full bg-brand-500" />
                    <h4 className="font-bold text-brand-800">المهارات</h4>
                    <span className="mr-auto text-xs text-slate-400">
                      {selectedEmployee.skills.filter((s) => s.canPerform).length} مهارة
                    </span>
                  </div>
                  {(() => {
                    const skillCount = selectedEmployee.skills.filter((s) => s.canPerform).length
                    const currentLevel = [...levels].reverse().find((l) => skillCount >= l.min) || levels[0]
                    const nextLevel = levels.find((l) => l.min > skillCount)
                    return (
                      <div className="mb-3 rounded-lg bg-brand-50 px-3 py-2 text-xs text-brand-700">
                        المستوى {currentLevel.level} - {currentLevel.label}
                        {nextLevel && ` | يحتاج ${nextLevel.min - skillCount} مهارة إضافية للترقي`}
                      </div>
                    )
                  })()}
                  <div className="flex border-b border-slate-200 mb-3">
                    <button
                      type="button"
                      onClick={() => setSkillTab('technical')}
                      className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${skillTab === 'technical' ? 'border-brand-500 text-brand-700' : 'border-transparent text-slate-400 hover:text-slate-600'}`}
                    >
                      مهارات فنية
                    </button>
                    <button
                      type="button"
                      onClick={() => setSkillTab('leader')}
                      className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${skillTab === 'leader' ? 'border-brand-500 text-brand-700' : 'border-transparent text-slate-400 hover:text-slate-600'}`}
                    >
                      مهارات قيادية
                    </button>
                  </div>
                  {skillTab === 'technical' && (
                    <>
                      <p className="mt-1 mb-3 text-sm text-slate-500">
                        حدد المهارات الدقيقة التي يستطيع هذا الموظف تنفيذها ضمن كل خدمة. يستخدم
                        النظام هذه القائمة لاقتراح الموظف تلقائياً عند إنشاء حجز جديد.
                      </p>
                      <div className="flex flex-col gap-4">
                        {services.map((service) => (
                          <div key={service.id}>
                            <h5 className="mb-2 text-sm font-bold text-brand-700">{service.name}</h5>
                            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
                              {service.skills.map((sk) => {
                                const skill = selectedEmployee.skills.find((s) => s.skillId === sk.id)
                                const canPerform = skill?.canPerform ?? false
                                return (
                                  <label
                                    key={sk.id}
                                    className={`flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors ${
                                      canPerform
                                        ? 'border-brand-500 bg-brand-50 text-brand-800'
                                        : 'border-slate-200 text-slate-500'
                                    }`}
                                  >
                                    <input
                                      type="checkbox"
                                      checked={canPerform}
                                      onChange={() => toggleSkill(selectedEmployee, sk.id)}
                                      className="h-4 w-4 accent-brand-700"
                                    />
                                    {sk.name}
                                  </label>
                                )
                              })}
                              {service.skills.length === 0 && (
                                <p className="text-sm text-slate-400">لا توجد مهارات محددة لهذه الخدمة.</p>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </>
                  )}
                  {skillTab === 'leader' && (
                    <div className="flex flex-col gap-3">
                      {['القيادة', 'إدارة الفريق', 'حل المشكلات', 'التواصل', 'اتخاذ القرار'].map((skill) => (
                        <div key={skill} className="flex items-center gap-3">
                          <span className="w-28 text-sm text-slate-600">{skill}</span>
                          <input
                            type="range"
                            min={1}
                            max={10}
                            value={leaderRatings[skill] || 5}
                            onChange={(e) => setLeaderRatings((prev) => ({ ...prev, [skill]: Number(e.target.value) }))}
                            className="flex-1 accent-brand-500"
                          />
                          <span className="w-8 text-center text-sm font-bold text-brand-700">
                            {leaderRatings[skill] || 5}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
