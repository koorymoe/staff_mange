import { useEffect, useState } from 'react'
import { api, type KpiEvaluation, type Employee } from '../api'
import { useSession } from '../session'

const KPI_CRITERIA = [
  { id: 'colleagues', label: 'العلاقة مع الزملاء', icon: '🤝' },
  { id: 'tasks', label: 'تنفيذ المهام الموكلة إليه', icon: '📋' },
  { id: 'skills', label: 'استطيع ولا استطيع', icon: '🎯' },
  { id: 'compliance', label: 'الالتزام بالآليات وتوجيهات المسؤول', icon: '📐' },
  { id: 'vehicle', label: 'تنظيف السيارة', icon: '🚗' },
  { id: 'response', label: 'سرعة الاستجابة بالمهام', icon: '⚡' },
  { id: 'tools', label: 'ترتيب العدد', icon: '🔧' },
  { id: 'complaints', label: 'شكوى الزبائن', icon: '📞' },
]

const POINTS_PER_WEEK = 8
const IQD_PER_POINT = 10_000

const EVALUATOR_ROLES = ['ADMIN', 'MONITOR', 'HR_COORDINATOR']

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

export default function KpiPage() {
  const { employee: currentUser } = useSession()
  const [evaluations, setEvaluations] = useState<KpiEvaluation[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Evaluator form state
  const [selectedEmployeeId, setSelectedEmployeeId] = useState('')
  const [deductCriteria, setDeductCriteria] = useState<string | null>(null)
  const [deductPoints, setDeductPoints] = useState(1)
  const [deductNotes, setDeductNotes] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const isEvaluator = currentUser && EVALUATOR_ROLES.includes(currentUser.role)

  const load = () => {
    setLoading(true)
    Promise.all([api.getKpiEvaluations(), api.getEmployees()])
      .then(([evals, emps]) => {
        setEvaluations(evals)
        setEmployees(emps)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const weeklyEvals = evaluations.filter((ev) => isThisWeek(ev.createdAt))

  // Per-employee weekly deductions
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

  // Current user's weekly data
  const myWeekly = currentUser ? weeklyByEmployee[currentUser.id] : null
  const myDeductedPoints = myWeekly?.totalPoints || 0
  const myRemainingPoints = POINTS_PER_WEEK - myDeductedPoints

  const handleDeduct = async () => {
    if (!currentUser || !selectedEmployeeId || !deductCriteria) return
    const criteriaLabel = KPI_CRITERIA.find((c) => c.id === deductCriteria)?.label || ''
    const reason = deductNotes ? `${criteriaLabel}: ${deductNotes}` : criteriaLabel
    setSubmitting(true)
    try {
      await api.createKpiEvaluation({
        employeeId: selectedEmployeeId,
        evaluatorId: currentUser.id,
        points: deductPoints,
        reason,
      })
      setDeductCriteria(null)
      setDeductPoints(1)
      setDeductNotes('')
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

  const getCardBorder = (deducted: number) => {
    if (deducted === 0) return 'border-green-200'
    if (deducted <= 3) return 'border-amber-200'
    return 'border-red-200'
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
      <div>
        <h2 className="text-2xl font-bold text-brand-900">تقييم الأداء (KPI)</h2>
        <p className="mt-1 text-slate-500">
          نظام النقاط الأسبوعي - {POINTS_PER_WEEK} نقاط بقيمة {IQD_PER_POINT.toLocaleString()} د.ع لكل نقطة
        </p>
      </div>

      {/* My weekly points (all employees see this) */}
      {currentUser && (
        <div className="mt-6 rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
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
                {((myWeekly?.totalIQD || 0)).toLocaleString()} د.ع
              </p>
            </div>
          </div>
          <div className="mt-4 h-4 w-full overflow-hidden rounded-full bg-slate-100">
            <div
              className={`h-full rounded-full transition-all ${getBarColor(myDeductedPoints)}`}
              style={{ width: `${(myRemainingPoints / POINTS_PER_WEEK) * 100}%` }}
            />
          </div>
          {/* My deduction history */}
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

      {/* Evaluator section */}
      {isEvaluator && (
        <>
          {/* Employee selector */}
          <div className="mt-6 rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="mb-4 text-lg font-bold text-brand-800">خصم نقاط موظف</h3>
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
                {/* Selected employee weekly summary */}
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

                {/* KPI criteria cards */}
                <p className="mb-3 text-sm font-medium text-slate-600">اختر معيار الخصم</p>
                <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                  {KPI_CRITERIA.map((c) => (
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
                      <span className="block text-2xl">{c.icon}</span>
                      <span className="mt-1 block text-xs font-medium text-slate-700">
                        {c.label}
                      </span>
                    </button>
                  ))}
                </div>

                {/* Deduction form */}
                {deductCriteria && (
                  <div className="mt-4 rounded-xl border border-red-100 bg-red-50/50 p-4">
                    <p className="mb-3 font-medium text-red-800">
                      خصم: {KPI_CRITERIA.find((c) => c.id === deductCriteria)?.label}
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

          {/* Weekly overview table */}
          <div className="mt-6 rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] overflow-hidden">
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

          {/* Full deduction history */}
          <div className="mt-6 rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] overflow-hidden">
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
                    <tr key={ev.id} className="transition-colors hover:bg-slate-50">
                      <td className="px-4 py-3 font-medium">{ev.employee.name}</td>
                      <td className="px-4 py-3 text-slate-500">{ev.evaluator.name}</td>
                      <td className="px-4 py-3">
                        <span className="rounded-full bg-brand-50 px-3 py-1 text-sm font-bold text-brand-700">
                          {ev.points}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-slate-600">{ev.reason || '-'}</td>
                      <td className="px-4 py-3 font-bold text-red-600">
                        {ev.deductionAmount.toLocaleString()} د.ع
                      </td>
                      <td className="px-4 py-3 text-slate-500">
                        {new Date(ev.createdAt).toLocaleDateString('ar-IQ')}
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() => handleDelete(ev.id)}
                          className="rounded-lg bg-red-100 px-3 py-1 text-sm font-medium text-red-700 transition-colors hover:bg-red-200"
                        >
                          حذف
                        </button>
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
