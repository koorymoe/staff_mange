import { useEffect, useState } from 'react'
import { api, type KpiEvaluation, type Employee } from '../api'
import { useSession } from '../session'

export default function KpiPage() {
  const { employee: currentUser } = useSession()
  const [evaluations, setEvaluations] = useState<KpiEvaluation[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)

  // Form state
  const [employeeId, setEmployeeId] = useState('')
  const [points, setPoints] = useState('')
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

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

  // Summary: total deductions per employee
  const deductionSummary = evaluations.reduce<Record<string, { name: string; total: number }>>((acc, ev) => {
    if (!acc[ev.employeeId]) {
      acc[ev.employeeId] = { name: ev.employee.name, total: 0 }
    }
    acc[ev.employeeId].total += ev.deductionAmount
    return acc
  }, {})

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!currentUser) return
    setSubmitting(true)
    try {
      await api.createKpiEvaluation({
        employeeId,
        evaluatorId: currentUser.id,
        points: Number(points),
        reason,
      })
      setEmployeeId('')
      setPoints('')
      setReason('')
      setShowForm(false)
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

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">تقييم الأداء (KPI)</h2>
          <p className="mt-1 text-slate-500">إدارة تقييمات أداء الموظفين والخصومات المترتبة.</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
        >
          إضافة تقييم جديد
        </button>
      </div>

      {/* Deduction Summary */}
      {Object.keys(deductionSummary).length > 0 && (
        <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Object.entries(deductionSummary).map(([empId, { name, total }]) => (
            <div key={empId} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-sm text-slate-500">إجمالي الخصومات</p>
              <p className="mt-1 text-lg font-bold text-brand-800">{name}</p>
              <p className="mt-1 text-xl font-extrabold text-red-600">
                {total.toLocaleString()} د.ع
              </p>
            </div>
          ))}
        </div>
      )}

      {/* Add Form */}
      {showForm && (
        <form
          onSubmit={handleSubmit}
          className="mt-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
        >
          <h3 className="mb-4 text-lg font-bold text-brand-800">تقييم جديد</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الموظف</label>
              <select
                required
                value={employeeId}
                onChange={(e) => setEmployeeId(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              >
                <option value="">اختر الموظف</option>
                {employees.map((emp) => (
                  <option key={emp.id} value={emp.id}>{emp.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">النقاط</label>
              <input
                required
                type="number"
                value={points}
                onChange={(e) => setPoints(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">المقيّم</label>
              <input
                disabled
                value={currentUser?.name || ''}
                className="w-full rounded-lg border border-gray-300 bg-slate-50 px-4 py-3 text-right"
              />
            </div>
          </div>
          <div className="mt-4">
            <label className="mb-1 block text-sm font-medium text-slate-600">السبب</label>
            <textarea
              required
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
            />
          </div>
          <div className="mt-4">
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
            >
              {submitting ? 'جاري الحفظ...' : 'حفظ التقييم'}
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
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
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
      )}
    </div>
  )
}
