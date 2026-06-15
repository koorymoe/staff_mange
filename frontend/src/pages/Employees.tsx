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
  const [submitting, setSubmitting] = useState(false)

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
      })
      setName('')
      setCertificate('')
      setPosition('')
      setPhone('')
      setUsername('')
      setPassword('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const visibleEmployees = isHR ? employees.filter((emp) => emp.role === 'TECHNICIAN') : employees

  const selectedEmployee = employees.find((emp) => emp.id === selectedId) || null

  useEffect(() => {
    setCredUsername(selectedEmployee?.username || '')
    setCredPassword('')
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

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة الكوادر</h2>
      <p className="mt-1 text-slate-500">
        إدارة بيانات الموظفين وتحديد المهارات (الخدمات) التي يستطيع كل موظف تنفيذها.
      </p>

      {isAdmin && (
      <form
        onSubmit={handleAddEmployee}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4"
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
          <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-1">
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

          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-2">
            {!selectedEmployee && (
              <p className="text-slate-400">اختر موظفاً من القائمة لعرض/تعديل مهاراته.</p>
            )}
            {selectedEmployee && isHR && (
              <div>
                <h3 className="text-lg font-bold text-brand-800">
                  بيانات: {selectedEmployee.name}
                </h3>
                <p className="text-sm text-slate-500">{selectedEmployee.position || 'فني'}</p>

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
                      <div className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 p-4 text-white shadow-lg shadow-brand-900/20">
                        <p className="text-xs text-brand-100">المستوى</p>
                        <p className="mt-1 text-xl font-extrabold">
                          {currentLevel.level} - {currentLevel.label}
                        </p>
                        <p className="mt-1 text-xs text-brand-100">
                          {skillCount} مهارة معتمدة
                          {nextLevel ? ` - يحتاج ${nextLevel.min - skillCount} للترقي` : ' - أعلى مستوى'}
                        </p>
                      </div>
                      <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
                        <p className="text-xs text-slate-500">الرانك</p>
                        <p className="mt-1 text-xl font-extrabold text-emerald-700">{rank}</p>
                        <p className="mt-1 text-xs text-slate-500">{completedCount} حجز منجز</p>
                      </div>
                      <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
                        <p className="text-xs text-slate-500">الترتيب بين الفنيين</p>
                        <p className="mt-1 text-xl font-extrabold text-brand-700">
                          {position >= 0 ? `#${position + 1}` : '-'}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {sortedTechs.length > 0 ? `من ${sortedTechs.length}` : '-'}
                        </p>
                      </div>
                      <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
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

                <h4 className="mt-5 font-bold text-brand-800">المهارات المتقنة</h4>
                <div className="mt-2 flex flex-wrap gap-2">
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
              </div>
            )}
            {selectedEmployee && !isHR && (
              <div>
                <h3 className="text-lg font-bold text-brand-800">
                  بيانات: {selectedEmployee.name}
                </h3>

                <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
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
                <h4 className="mt-5 font-bold text-brand-800">بيانات تسجيل الدخول</h4>
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

                <h4 className="mt-5 font-bold text-brand-800">المهارات</h4>
                <p className="mt-1 text-sm text-slate-500">
                  حدد المهارات الدقيقة التي يستطيع هذا الموظف تنفيذها ضمن كل خدمة. يستخدم
                  النظام هذه القائمة لاقتراح الموظف تلقائياً عند إنشاء حجز جديد.
                </p>
                <div className="mt-4 flex flex-col gap-4">
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
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
