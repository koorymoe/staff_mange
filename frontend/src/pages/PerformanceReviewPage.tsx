import { useEffect, useMemo, useState } from 'react'
import { api, type Employee, type PerformanceReview } from '../api'
import { useSession } from '../session'

// تقييم الأداء — منفصل تماماً عن KPI (الغرامات المالية، حصراً للأدمن والمراقب).
// هذا التقييم يحدد بس هل الموظف يستحق تدريب أو لا:
//   - التيم ليدر يقيّم فنيي فريقه العاديين
//   - إداري الكوادر (أو الأدمن) يقيّم التيم ليدرات نفسهم
export default function PerformanceReviewPage() {
  const { employee } = useSession()
  const [employees, setEmployees] = useState<Employee[]>([])
  const [reviews, setReviews] = useState<PerformanceReview[]>([])
  const [loading, setLoading] = useState(true)
  const [reason, setReason] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState<string | null>(null)

  const isHR = employee?.role === 'ADMIN' || employee?.role === 'HR_COORDINATOR'
  const isLeader = !!employee?.isLeader && employee.role === 'TECHNICIAN'

  const load = () => {
    Promise.all([api.getEmployees(), api.getPerformanceReviews()])
      .then(([e, r]) => { setEmployees(e); setReviews(r) })
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  // اللي يقدر يقيمهم المستخدم الحالي حسب السلسلة الهرمية
  const evaluatable = useMemo(() => {
    if (isHR) return employees.filter(e => e.isLeader)
    if (isLeader) return employees.filter(e => e.role === 'TECHNICIAN' && !e.isLeader)
    return []
  }, [employees, isHR, isLeader])

  const submit = async (employeeId: string, rating: 'POSITIVE' | 'NEGATIVE') => {
    const r = (reason[employeeId] || '').trim()
    if (!r) return alert('اكتب سبب التقييم أول')
    setSaving(employeeId)
    try {
      await api.createPerformanceReview({ employeeId, rating, reason: r })
      setReason(prev => ({ ...prev, [employeeId]: '' }))
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حفظ التقييم')
    } finally {
      setSaving(null)
    }
  }

  const latestReview = (employeeId: string) =>
    reviews.filter(r => r.employeeId === employeeId).sort((a, b) => b.createdAt.localeCompare(a.createdAt))[0]

  if (loading) return <p className="py-16 text-center text-slate-400">جاري التحميل...</p>

  if (!isHR && !isLeader) {
    return (
      <div className="rounded-2xl bg-white p-10 text-center shadow-sm">
        <div className="mb-3 text-5xl">🔒</div>
        <p className="text-slate-400">هذي الصفحة لتيم ليدرات الفرق وإدارة الكوادر بس.</p>
      </div>
    )
  }

  return (
    <div>
      <h2 className="mb-2 text-2xl font-bold text-brand-900">تقييم الأداء 📝</h2>
      <p className="mb-6 text-sm text-slate-500">
        هذا التقييم منفصل تماماً عن نظام KPI (الغرامات المالية) — يحدد بس هل الموظف يستحق تدريب أو لا.
        {isLeader && ' تقيّم هنا فنيي فريقك العاديين.'}
        {isHR && ' تقيّم هنا تيم ليدرات الفرق.'}
      </p>

      <div className="flex flex-col gap-4">
        {evaluatable.map(emp => {
          const last = latestReview(emp.id)
          return (
            <div key={emp.id} className="rounded-2xl bg-white p-5 shadow-sm">
              <div className="mb-3 flex items-center justify-between">
                <div>
                  <span className="font-bold text-slate-800">{emp.name}</span>
                  {emp.isTrainee && <span className="mr-2 rounded-full bg-amber-50 px-2 py-0.5 text-xs font-bold text-amber-700">🎓 بالتدريب حالياً</span>}
                </div>
                {last && (
                  <span className={`rounded-full px-3 py-1 text-xs font-bold ${last.rating === 'POSITIVE' ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-600'}`}>
                    آخر تقييم: {last.rating === 'POSITIVE' ? 'إيجابي' : 'سلبي'} — {new Date(last.createdAt).toLocaleDateString('ar-IQ')}
                  </span>
                )}
              </div>
              <textarea value={reason[emp.id] || ''} onChange={e => setReason(prev => ({ ...prev, [emp.id]: e.target.value }))}
                placeholder="سبب التقييم..." rows={2}
                className="mb-3 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <div className="flex gap-2">
                <button onClick={() => submit(emp.id, 'POSITIVE')} disabled={saving === emp.id}
                  className="rounded-lg bg-emerald-500 px-4 py-2 text-xs font-bold text-white hover:bg-emerald-600 disabled:opacity-50">
                  👍 تقييم إيجابي
                </button>
                <button onClick={() => submit(emp.id, 'NEGATIVE')} disabled={saving === emp.id}
                  className="rounded-lg bg-red-500 px-4 py-2 text-xs font-bold text-white hover:bg-red-600 disabled:opacity-50">
                  👎 تقييم سلبي (يرسله للتدريب)
                </button>
              </div>
            </div>
          )
        })}
        {evaluatable.length === 0 && (
          <div className="rounded-2xl bg-white p-10 text-center shadow-sm">
            <p className="text-slate-400">ماكو موظفين تكدر تقيّمهم حالياً.</p>
          </div>
        )}
      </div>
    </div>
  )
}
