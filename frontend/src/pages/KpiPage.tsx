import { useEffect, useState } from 'react'
import { api, type KpiCriterion, type KpiEvaluation, type Employee, type TechnicianKpi } from '../api'
import { useSession } from '../session'

// نقاط الكي بي اي صارت تتحمّل من الباك إند (قابلة للإضافة والحذف من الواجهة
// بدل ما تكون مثبتة هنا بالكود) — راجع KpiCriterion بـ api.ts.
const POINTS_PER_WEEK = 8
const IQD_PER_POINT = 10_000

// منو يقدر يخصم نقاط. الكي بي اي مو خاص بالإداريين — المدير والمالك
// يخصمون من أي موظف بالشركة: فني، مصمم، مبيعات، محاسب... الكل.
// كانوا ناقصين من القائمة فما كان يطلعلهم فورم الخصم أصلاً.
const EVALUATOR_ROLES = ['ADMIN', 'OWNER', 'MONITOR', 'HR_COORDINATOR']

function getWeekStart(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  d.setDate(d.getDate() - day)
  d.setHours(0, 0, 0, 0)
  return d
}

function isThisWeek(dateStr: string): boolean {
  const weekStart = getWeekStart(new Date())
  const d = new Date(dateStr)
  return d >= weekStart
}

function getCurrentMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

const BREAKDOWN_LABELS: Record<string, string> = {
  completedBookings: 'الحجوزات المنجزة',
  completionSpeed: 'سرعة الإنجاز',
  workReports: 'تقارير العمل',
  attendance: 'الحضور',
  complaints: 'الشكاوى',
  manualDeductions: 'خصومات يدوية',
}

// ─── Technician Tab ───────────────────────────────────────────────────────────

function TechnicianTab() {
  const [leaderboard, setLeaderboard] = useState<TechnicianKpi[]>([])
  const [month, setMonth] = useState(getCurrentMonth)
  const [loading, setLoading] = useState(true)
  const [selectedTech, setSelectedTech] = useState<TechnicianKpi | null>(null)

  useEffect(() => {
    // month can change after mount; re-arm loading via a microtask so the setState
    // isn't synchronous within the effect body (react-hooks/set-state-in-effect).
    queueMicrotask(() => setLoading(true))
    api
      .getKpiLeaderboard(month)
      .then(setLeaderboard)
      .catch(() => setLeaderboard([]))
      .finally(() => setLoading(false))
  }, [month])

  const rankBadge = (index: number) => {
    if (index === 0)
      return (
        <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-yellow-300 to-yellow-500 text-sm font-extrabold text-yellow-900 shadow">
          1
        </span>
      )
    if (index === 1)
      return (
        <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-gray-200 to-gray-400 text-sm font-extrabold text-gray-800 shadow">
          2
        </span>
      )
    if (index === 2)
      return (
        <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-amber-500 to-amber-700 text-sm font-extrabold text-white shadow">
          3
        </span>
      )
    return (
      <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-slate-100 text-sm font-bold text-slate-500">
        {index + 1}
      </span>
    )
  }

  return (
    <div>
      {/* Month selector */}
      <div className="mb-6 flex items-center gap-3">
        <label className="text-sm font-medium text-slate-600">الشهر:</label>
        <input
          type="month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
          className="rounded-xl border border-gray-300 px-4 py-2 text-right outline-none focus:border-brand-500"
        />
      </div>

      {loading ? (
        <p className="mt-6 text-slate-400">جاري التحميل...</p>
      ) : leaderboard.length === 0 ? (
        <p className="mt-6 text-center text-slate-400">لا يوجد فنيون لعرض النتائج</p>
      ) : (
        <>
          {/* Leaderboard table */}
          <div className="overflow-hidden rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-right">
                <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                  <tr>
                    <th className="px-4 py-3 text-sm font-semibold">الترتيب</th>
                    <th className="px-4 py-3 text-sm font-semibold">الفني</th>
                    <th className="px-4 py-3 text-sm font-semibold">الحجوزات</th>
                    <th className="px-4 py-3 text-sm font-semibold">التقارير</th>
                    <th className="px-4 py-3 text-sm font-semibold">الحضور</th>
                    <th className="px-4 py-3 text-sm font-semibold">إجمالي النقاط</th>
                    <th className="px-4 py-3 text-sm font-semibold">تفاصيل</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {leaderboard.map((tech, i) => (
                    <tr
                      key={tech.employeeId}
                      className={`transition-colors hover:bg-slate-50 ${i < 3 ? 'bg-slate-50/50' : ''}`}
                    >
                      <td className="px-4 py-3">{rankBadge(i)}</td>
                      <td className="px-4 py-3 font-medium">{tech.employeeName}</td>
                      <td className="px-4 py-3 text-sm">
                        {tech.breakdown.completedBookings.count}
                      </td>
                      <td className="px-4 py-3 text-sm">
                        {tech.breakdown.workReports.count}
                      </td>
                      <td className="px-4 py-3 text-sm">
                        {tech.breakdown.attendance.daysPresent}/{tech.breakdown.attendance.totalDays}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`rounded-full px-3 py-1 text-sm font-bold ${
                            tech.totalPoints >= 100
                              ? 'bg-green-100 text-green-700'
                              : tech.totalPoints >= 50
                                ? 'bg-amber-100 text-amber-700'
                                : 'bg-red-100 text-red-700'
                          }`}
                        >
                          {tech.totalPoints}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() =>
                            setSelectedTech(
                              selectedTech?.employeeId === tech.employeeId ? null : tech,
                            )
                          }
                          className="rounded-lg bg-brand-50 px-3 py-1 text-sm font-medium text-brand-700 transition-colors hover:bg-brand-100"
                        >
                          {selectedTech?.employeeId === tech.employeeId ? 'إخفاء' : 'عرض'}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Detail card */}
          {selectedTech && (
            <div className="mt-6 rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <h3 className="mb-4 text-lg font-bold text-brand-800">
                تفاصيل نقاط: {selectedTech.employeeName}
              </h3>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                {(
                  Object.keys(selectedTech.breakdown) as Array<
                    keyof TechnicianKpi['breakdown']
                  >
                ).map((key) => {
                  const item = selectedTech.breakdown[key]
                  return (
                    <div
                      key={key}
                      className="rounded-xl border border-slate-100 bg-slate-50 p-4 text-center"
                    >
                      <p className="text-sm font-medium text-slate-500">
                        {BREAKDOWN_LABELS[key]}
                      </p>
                      <p
                        className={`mt-2 text-2xl font-extrabold ${
                          item.points >= 0 ? 'text-brand-700' : 'text-red-600'
                        }`}
                      >
                        {item.points > 0 ? `+${item.points}` : item.points}
                      </p>
                      {'count' in item && (
                        <p className="mt-1 text-xs text-slate-400">
                          العدد: {item.count}
                        </p>
                      )}
                      {'avgMinutes' in item && item.avgMinutes > 0 && (
                        <p className="mt-1 text-xs text-slate-400">
                          متوسط: {item.avgMinutes} دقيقة
                        </p>
                      )}
                      {'daysPresent' in item && (
                        <p className="mt-1 text-xs text-slate-400">
                          {item.daysPresent}/{item.totalDays} يوم
                        </p>
                      )}
                    </div>
                  )
                })}
              </div>
              <div className="mt-4 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 p-4 text-center">
                <p className="text-sm text-white/80">إجمالي النقاط</p>
                <p className="text-3xl font-extrabold text-white">
                  {selectedTech.totalPoints}
                </p>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

// ─── Administrative Tab ───────────────────────────────────────────────────────

function AdministrativeTab() {
  const { employee: currentUser, permissions } = useSession()
  const [evaluations, setEvaluations] = useState<KpiEvaluation[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [criteria, setCriteria] = useState<KpiCriterion[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [selectedEmployeeId, setSelectedEmployeeId] = useState('')
  const [deductCriteria, setDeductCriteria] = useState<string | null>(null)
  const [deductPoints, setDeductPoints] = useState(1)
  const [deductNotes, setDeductNotes] = useState('')
  // نشر المخالفة بلوحة الإعلانات — اختياري، ولمدة 3 أيام بس
  const [announceDeduction, setAnnounceDeduction] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [newCriterionLabel, setNewCriterionLabel] = useState('')

  const isEvaluator = currentUser && (EVALUATOR_ROLES.includes(currentUser.role) || permissions.includes('auditing') || permissions.includes('kpi_management'))
  const canManageCriteria = currentUser?.role === 'ADMIN' || permissions.includes('kpi_criteria_management')

  const load = () => {
    Promise.all([api.getKpiEvaluations(), api.getEmployees(), api.getKpiCriteria()])
      .then(([evals, emps, crit]) => {
        setEvaluations(evals)
        setEmployees(emps.filter((e) => e.role !== 'TECHNICIAN' && e.role !== 'ADMIN'))
        setCriteria(crit)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  // التقييمات الملغاة (المسترجعة) ما تدخل بحساب النقاط المتبقية ولا الخصم المالي
  const weeklyEvals = evaluations.filter((ev) => isThisWeek(ev.createdAt) && !ev.cancelled)

  const weeklyByEmployee = weeklyEvals.reduce<
    Record<string, { name: string; totalPoints: number; totalIQD: number; deductions: KpiEvaluation[] }>
  >((acc, ev) => {
    if (!acc[ev.employeeId]) {
      acc[ev.employeeId] = { name: ev.employee.name, totalPoints: 0, totalIQD: 0, deductions: [] }
    }
    acc[ev.employeeId].totalPoints += ev.points
    acc[ev.employeeId].totalIQD += ev.deductionAmount
    acc[ev.employeeId].deductions.push(ev)
    return acc
  }, {})

  const myWeekly = currentUser ? weeklyByEmployee[currentUser.id] : null
  const myDeductedPoints = myWeekly?.totalPoints || 0
  const myRemainingPoints = POINTS_PER_WEEK - myDeductedPoints

  const getPointColor = (deducted: number) => {
    if (deducted === 0) return 'text-green-600'
    if (deducted <= 3) return 'text-amber-600'
    return 'text-red-600'
  }

  const getBarColor = (deducted: number) => {
    if (deducted === 0) return 'bg-green-500'
    if (deducted <= 3) return 'bg-amber-500'
    return 'bg-red-500'
  }

  const handleDeduct = async () => {
    if (!currentUser || !selectedEmployeeId || !deductCriteria) return
    const criteriaLabel = criteria.find((c) => c.id === deductCriteria)?.label || ''
    const reason = deductNotes ? `${criteriaLabel}: ${deductNotes}` : criteriaLabel
    setSubmitting(true)
    try {
      await api.createKpiEvaluation({
        employeeId: selectedEmployeeId,
        evaluatorId: currentUser.id,
        points: deductPoints,
        reason,
        announce: announceDeduction,
      })
      setDeductCriteria(null)
      setDeductPoints(1)
      setDeductNotes('')
      setAnnounceDeduction(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف هذا التقييم؟')) return
    try {
      await api.deleteKpiEvaluation(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  // إرجاع نقطة: ما تنحذف — تضل بالسجل معلّمة "ملغاة" حتى المراقب يشوف تاريخها
  const handleCancel = async (id: string) => {
    if (!confirm('استرجاع هذي النقطة؟ راح توقف تأثيرها المالي بس تضل بالسجل.')) return
    try {
      await api.cancelKpiEvaluation(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleAddCriterion = async () => {
    if (!newCriterionLabel.trim()) return
    try {
      await api.createKpiCriterion(newCriterionLabel.trim())
      setNewCriterionLabel('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleDeleteCriterion = async (id: string) => {
    if (!confirm('حذف نقطة الكي بي اي هذي نهائياً؟')) return
    try {
      await api.deleteKpiCriterion(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  if (loading) return <p className="mt-6 text-slate-400">جاري التحميل...</p>
  if (error)
    return (
      <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
        تعذر الاتصال بالخادم: {error}
      </p>
    )

  return (
    <div>
      {/* My weekly points */}
      {currentUser && currentUser.role !== 'TECHNICIAN' && (
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-slate-500">نقاطك هذا الأسبوع</p>
              <p className={`mt-1 text-3xl font-extrabold ${getPointColor(myDeductedPoints)}`}>
                {myRemainingPoints} / {POINTS_PER_WEEK}
              </p>
            </div>
            <div className="text-left">
              <p className="text-sm text-slate-500">إجمالي الخصم</p>
              <p className="mt-1 text-2xl font-bold text-red-600">
                {(myWeekly?.totalIQD || 0).toLocaleString()} د.ع
              </p>
            </div>
          </div>
          <div className="mt-4 h-4 w-full overflow-hidden rounded-full bg-slate-100">
            <div
              className={`h-full rounded-full transition-all ${getBarColor(myDeductedPoints)}`}
              style={{ width: `${(myRemainingPoints / POINTS_PER_WEEK) * 100}%` }}
            />
          </div>
          {myWeekly && myWeekly.deductions.length > 0 && (
            <div className="mt-4 space-y-2">
              <p className="text-sm font-medium text-slate-600">سجل الخصومات</p>
              {myWeekly.deductions.map((ev) => (
                <div
                  key={ev.id}
                  className="flex items-center justify-between rounded-xl bg-red-50 px-4 py-2"
                >
                  <div>
                    <span className="font-medium text-red-700">-{ev.points} نقطة</span>
                    <span className="mr-3 text-sm text-slate-600">{ev.reason}</span>
                  </div>
                  <span className="text-sm text-slate-400">
                    {new Date(ev.createdAt).toLocaleDateString('ar-IQ')}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* إدارة نقاط الكي بي اي — إضافة/حذف (صلاحية kpi_criteria_management) */}
      {canManageCriteria && (
        <div className="mt-6 rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-4 text-lg font-bold text-brand-800">إدارة نقاط الكي بي اي</h3>
          <div className="flex gap-2">
            <input
              value={newCriterionLabel}
              onChange={(e) => setNewCriterionLabel(e.target.value)}
              placeholder="عنوان نقطة جديدة..."
              className="flex-1 rounded-xl border border-gray-300 px-4 py-2.5 text-right outline-none focus:border-brand-500"
            />
            <button
              onClick={handleAddCriterion}
              className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2.5 font-medium text-white"
            >
              إضافة
            </button>
          </div>
          <div className="mt-4 flex flex-wrap gap-2">
            {criteria.map((c) => (
              <span
                key={c.id}
                className="flex items-center gap-2 rounded-full bg-slate-100 px-3 py-1.5 text-sm text-slate-700"
              >
                {c.label}
                <button
                  onClick={() => handleDeleteCriterion(c.id)}
                  className="text-red-500 hover:text-red-700"
                  title="حذف"
                >
                  ✕
                </button>
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Evaluator section */}
      {isEvaluator && (
        <>
          <div className="mt-6 rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="text-lg font-bold text-brand-800">خصم نقاط موظف</h3>
            <p className="mb-4 mt-1 text-sm text-slate-500">
              يشمل كل موظفي الشركة — فنيين، مصممين، مبيعات، محاسبة، إداريين.
            </p>
            <div className="mb-4">
              <label className="mb-1 block text-sm font-medium text-slate-600">اختر الموظف</label>
              <select
                value={selectedEmployeeId}
                onChange={(e) => setSelectedEmployeeId(e.target.value)}
                className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              >
                <option value="">اختر الموظف</option>
                {employees
                  .filter((e) => e.status === 'ACTIVE')
                  .map((emp) => {
                    const empWeekly = weeklyByEmployee[emp.id]
                    const remaining = POINTS_PER_WEEK - (empWeekly?.totalPoints || 0)
                    return (
                      <option key={emp.id} value={emp.id}>
                        {emp.name} ({remaining}/{POINTS_PER_WEEK} نقطة)
                      </option>
                    )
                  })}
              </select>
            </div>

            {selectedEmployeeId && (
              <>
                {(() => {
                  const empW = weeklyByEmployee[selectedEmployeeId]
                  const deducted = empW?.totalPoints || 0
                  const remaining = POINTS_PER_WEEK - deducted
                  return (
                    <div className="mb-4 flex items-center gap-4">
                      <div className="flex-1">
                        <div className="h-3 w-full overflow-hidden rounded-full bg-slate-100">
                          <div
                            className={`h-full rounded-full ${getBarColor(deducted)}`}
                            style={{ width: `${(remaining / POINTS_PER_WEEK) * 100}%` }}
                          />
                        </div>
                      </div>
                      <span className={`font-bold ${getPointColor(deducted)}`}>
                        {remaining}/{POINTS_PER_WEEK}
                      </span>
                    </div>
                  )
                })()}

                <p className="mb-3 text-sm font-medium text-slate-600">اختر معيار الخصم</p>
                <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                  {criteria.map((c) => (
                    <button
                      key={c.id}
                      type="button"
                      onClick={() => setDeductCriteria(deductCriteria === c.id ? null : c.id)}
                      className={`rounded-2xl border-2 p-4 text-center transition-all ${
                        deductCriteria === c.id
                          ? 'border-red-400 bg-red-50 shadow-md'
                          : 'border-slate-200 bg-white hover:border-red-200 hover:bg-red-50/50'
                      }`}
                    >
                      <span className="mt-1 block text-xs font-medium text-slate-700">
                        {c.label}
                      </span>
                    </button>
                  ))}
                  {criteria.length === 0 && (
                    <p className="col-span-full text-center text-sm text-slate-400">لا توجد نقاط كي بي اي بعد</p>
                  )}
                </div>

                {deductCriteria && (
                  <div className="mt-4 rounded-xl border border-red-100 bg-red-50/50 p-4">
                    <p className="mb-3 font-medium text-red-800">
                      خصم: {criteria.find((c) => c.id === deductCriteria)?.label}
                    </p>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          عدد النقاط
                        </label>
                        <select
                          value={deductPoints}
                          onChange={(e) => setDeductPoints(Number(e.target.value))}
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        >
                          {Array.from({ length: 8 }, (_, i) => i + 1).map((n) => (
                            <option key={n} value={n}>
                              {n} نقطة ({(n * IQD_PER_POINT).toLocaleString()} د.ع)
                            </option>
                          ))}
                        </select>
                      </div>
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          ملاحظات
                        </label>
                        <input
                          value={deductNotes}
                          onChange={(e) => setDeductNotes(e.target.value)}
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                          placeholder="تفاصيل إضافية..."
                        />
                      </div>
                    </div>
                    {/* نشر المخالفة بالشريط — قرار المدير، ولثلاثة أيام
                        بعدها تنطفي لحالها */}
                    <label className="mt-3 flex cursor-pointer items-start gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3">
                      <input
                        type="checkbox"
                        checked={announceDeduction}
                        onChange={(e) => setAnnounceDeduction(e.target.checked)}
                        className="mt-0.5"
                      />
                      <span className="text-sm text-amber-900">
                        <span className="font-bold">📢 انشرها بلوحة الإعلانات</span>
                        <span className="block text-xs text-amber-700">
                          تطلع بالشريط المتحرك لكل الموظفين لمدة ٣ أيام، وبعدها تنطفي لحالها.
                        </span>
                      </span>
                    </label>
                    <button
                      onClick={handleDeduct}
                      disabled={submitting}
                      className="mt-3 w-full rounded-xl bg-red-600 px-6 py-3 font-bold text-white shadow transition-all hover:bg-red-700 disabled:opacity-50"
                    >
                      {submitting ? 'جاري الحفظ...' : `خصم ${deductPoints} نقطة`}
                    </button>
                  </div>
                )}
              </>
            )}
          </div>

          {/* Weekly overview */}
          <div className="mt-6 overflow-hidden rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="p-5">
              <h3 className="text-lg font-bold text-brand-800">ملخص الأسبوع</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-right">
                <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                  <tr>
                    <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                    <th className="px-4 py-3 text-sm font-semibold">النقاط المتبقية</th>
                    <th className="px-4 py-3 text-sm font-semibold">الخصومات</th>
                    <th className="px-4 py-3 text-sm font-semibold">المبلغ (د.ع)</th>
                    <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {employees
                    .filter((e) => e.status === 'ACTIVE')
                    .map((emp) => {
                      const data = weeklyByEmployee[emp.id]
                      const deducted = data?.totalPoints || 0
                      const remaining = POINTS_PER_WEEK - deducted
                      return { emp, deducted, remaining, iqd: data?.totalIQD || 0 }
                    })
                    .sort((a, b) => b.deducted - a.deducted)
                    .map(({ emp, deducted, remaining, iqd }) => (
                      <tr key={emp.id} className="transition-colors hover:bg-slate-50">
                        <td className="px-4 py-3 font-medium">{emp.name}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <div className="h-2.5 w-20 overflow-hidden rounded-full bg-slate-100">
                              <div
                                className={`h-full rounded-full ${getBarColor(deducted)}`}
                                style={{ width: `${(remaining / POINTS_PER_WEEK) * 100}%` }}
                              />
                            </div>
                            <span className={`text-sm font-bold ${getPointColor(deducted)}`}>
                              {remaining}/{POINTS_PER_WEEK}
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          {deducted > 0 ? (
                            <span className="rounded-full bg-red-100 px-3 py-1 text-sm font-bold text-red-700">
                              -{deducted}
                            </span>
                          ) : (
                            <span className="rounded-full bg-green-100 px-3 py-1 text-sm font-bold text-green-700">
                              0
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3 font-bold text-red-600">
                          {iqd > 0 ? `${iqd.toLocaleString()} د.ع` : '-'}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-block h-3 w-3 rounded-full ${
                              deducted === 0
                                ? 'bg-green-500'
                                : deducted <= 3
                                  ? 'bg-amber-500'
                                  : 'bg-red-500'
                            }`}
                          />
                        </td>
                      </tr>
                    ))}
                  {employees.filter((e) => e.status === 'ACTIVE').length === 0 && (
                    <tr>
                      <td colSpan={5} className="px-4 py-6 text-center text-slate-400">
                        لا يوجد موظفون
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Full history */}
          <div className="mt-6 overflow-hidden rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="p-5">
              <h3 className="text-lg font-bold text-brand-800">سجل التقييمات</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-right">
                <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                  <tr>
                    <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                    <th className="px-4 py-3 text-sm font-semibold">المقيّم</th>
                    <th className="px-4 py-3 text-sm font-semibold">النقاط</th>
                    <th className="px-4 py-3 text-sm font-semibold">السبب</th>
                    <th className="px-4 py-3 text-sm font-semibold">مبلغ الخصم</th>
                    <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                    <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {evaluations.map((ev) => (
                    <tr key={ev.id} className={`transition-colors hover:bg-slate-50 ${ev.cancelled ? 'opacity-50' : ''}`}>
                      <td className="px-4 py-3 font-medium">{ev.employee.name}</td>
                      <td className="px-4 py-3 text-slate-500">{ev.evaluator.name}</td>
                      <td className="px-4 py-3">
                        <span className={`rounded-full px-3 py-1 text-sm font-bold ${ev.cancelled ? 'bg-slate-100 text-slate-400 line-through' : 'bg-brand-50 text-brand-700'}`}>
                          {ev.points}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-slate-600">
                        {ev.reason || '-'}
                        {ev.cancelled && (
                          <div className="mt-1 text-xs font-bold text-amber-600">
                            ⚠ ملغاة {ev.cancelledByEmployee ? `(بواسطة ${ev.cancelledByEmployee.name})` : ''}
                            {ev.cancelledAt && ` — ${new Date(ev.cancelledAt).toLocaleDateString('ar-IQ')}`}
                          </div>
                        )}
                      </td>
                      <td className={`px-4 py-3 font-bold ${ev.cancelled ? 'text-slate-300 line-through' : 'text-red-600'}`}>
                        {ev.deductionAmount.toLocaleString()} د.ع
                      </td>
                      <td className="px-4 py-3 text-slate-500">
                        {new Date(ev.createdAt).toLocaleDateString('ar-IQ')}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex gap-2">
                          {!ev.cancelled && (
                            <button
                              onClick={() => handleCancel(ev.id)}
                              className="rounded-lg bg-amber-100 px-3 py-1 text-sm font-medium text-amber-700 transition-colors hover:bg-amber-200"
                            >
                              استرجاع
                            </button>
                          )}
                          <button
                            onClick={() => handleDelete(ev.id)}
                            className="rounded-lg bg-red-100 px-3 py-1 text-sm font-medium text-red-700 transition-colors hover:bg-red-200"
                          >
                            حذف
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {evaluations.length === 0 && (
                    <tr>
                      <td colSpan={7} className="px-4 py-6 text-center text-slate-400">
                        لا توجد تقييمات بعد
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

// ─── Main Page ────────────────────────────────────────────────────────────────

export default function KpiPage() {
  const [activeTab, setActiveTab] = useState<'technician' | 'admin'>('technician')

  return (
    <div>
      <div>
        <h2 className="text-2xl font-bold text-brand-900">تقييم الأداء (KPI)</h2>
        <p className="mt-1 text-slate-500">نظام النقاط الذكي للفنيين والنقاط الإدارية</p>
      </div>

      {/* Tabs */}
      <div className="mt-6 flex gap-2">
        <button
          onClick={() => setActiveTab('technician')}
          className={`rounded-xl px-6 py-3 text-sm font-bold transition-all ${
            activeTab === 'technician'
              ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow'
              : 'border bg-white text-slate-600 hover:bg-slate-50'
          }`}
        >
          نقاط الفنيين
        </button>
        <button
          onClick={() => setActiveTab('admin')}
          className={`rounded-xl px-6 py-3 text-sm font-bold transition-all ${
            activeTab === 'admin'
              ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow'
              : 'border bg-white text-slate-600 hover:bg-slate-50'
          }`}
        >
          نقاط إدارية
        </button>
      </div>

      {/* Tab content */}
      <div className="mt-6">
        {activeTab === 'technician' ? <TechnicianTab /> : <AdministrativeTab />}
      </div>
    </div>
  )
}
