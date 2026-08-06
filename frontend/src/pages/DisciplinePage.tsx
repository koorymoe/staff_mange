import { useEffect, useState } from 'react'
import { api, type DisciplinePoints, type DisciplineEvent } from '../api'
import { useSession } from '../session'

// ═══ نقاط الانضباط والغرامات ═══
//
// كل موظف يبدي بـ١٠٠ نقطة، والنقطة بعشر آلاف دينار. النظام هو الي
// يغرّم تلقائياً — مو المدير. والصفحة مفتوحة لكل الموظفين عمداً:
// الشفافية جزء من العقوبة، وهي الي تخلي الناس تنتبه.
const KIND_LABELS: Record<string, { text: string; cls: string }> = {
  LATE_PAPERWORK:        { text: 'تأخر فاتورة/تقرير', cls: 'bg-red-50 text-red-700' },
  UNBALANCED_ASSIGNMENT: { text: 'توزيع غير عادل',    cls: 'bg-amber-50 text-amber-800' },
  RESTORE:               { text: 'رجوع نقطة',          cls: 'bg-emerald-50 text-emerald-700' },
}

export default function DisciplinePage() {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'
  const [points, setPoints] = useState<DisciplinePoints[]>([])
  const [events, setEvents] = useState<DisciplineEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(false)

  const load = () => {
    Promise.all([api.getDisciplinePoints(), api.getDisciplineEvents()])
      .then(([p, e]) => { setPoints(p); setEvents(e) })
      .catch(() => {})
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const run = async () => {
    setRunning(true)
    try { await api.runDisciplineSweep(); load() } finally { setRunning(false) }
  }

  const totalDeducted = points.reduce((sum, p) => sum + p.deductedDinar, 0)

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">⚖️ نقاط الانضباط</h2>
      <p className="mt-1 text-slate-500">
        كل موظف يبدي بـ١٠٠ نقطة، والنقطة بعشر آلاف دينار. النظام يغرّم تلقائياً —
        والنقطة ترجع لحالها بعد ثلاثة أيام شغل بلا أي غرامة.
      </p>

      {isAdmin && (
        <button
          onClick={run}
          disabled={running}
          className="mt-4 rounded-lg border border-brand-300 bg-brand-50 px-4 py-2 text-sm font-bold text-brand-700 disabled:opacity-50"
        >
          {running ? 'جاري الفحص...' : '🔄 شغّل الفحص الآن'}
        </button>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <>
          {points.length === 0 ? (
            <p className="mt-6 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm font-bold text-emerald-800">
              ✅ ماكو ولا غرامة — كل الموظفين على رصيدهم الكامل ١٠٠ نقطة.
            </p>
          ) : (
            <>
              <div className="mt-6 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <p className="text-sm text-slate-500">
                  مجموع المخصوم: <b className="text-red-600">{totalDeducted.toLocaleString()} د.ع</b>
                </p>
                <div className="mt-3 space-y-2">
                  {points.map((p) => (
                    <div key={p.employeeId} className="flex flex-wrap items-center gap-3 rounded-lg border border-slate-100 px-3 py-2">
                      <span className="font-bold text-slate-800">{p.employeeName}</span>
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${
                          p.points >= 95 ? 'bg-emerald-50 text-emerald-700'
                            : p.points >= 80 ? 'bg-amber-50 text-amber-800'
                              : 'bg-red-50 text-red-700'
                        }`}
                      >
                        {p.points} / 100 نقطة
                      </span>
                      {p.deductedDinar > 0 && (
                        <span className="text-xs text-red-600">−{p.deductedDinar.toLocaleString()} د.ع</span>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div className="mt-4 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <h3 className="mb-3 text-sm font-bold text-[#0f2040]">📜 سجل الحركات</h3>
                <div className="space-y-1.5">
                  {events.map((e) => {
                    const k = KIND_LABELS[e.kind] || { text: e.kind, cls: 'bg-slate-100 text-slate-600' }
                    return (
                      <div key={e.id} className="flex flex-wrap items-start gap-2 rounded-lg border border-slate-100 px-3 py-2 text-xs">
                        <span className={`shrink-0 rounded-full px-2 py-0.5 font-bold ${k.cls}`}>{k.text}</span>
                        <span className="font-bold text-slate-700">{e.employeeName}</span>
                        <span className={`font-bold ${e.delta < 0 ? 'text-red-600' : 'text-emerald-600'}`}>
                          {e.delta > 0 ? `+${e.delta}` : e.delta}
                        </span>
                        <span className="text-slate-500">{e.reason}</span>
                        <span className="mr-auto shrink-0 text-slate-400">
                          {new Date(e.createdAt).toLocaleDateString('ar-IQ')}
                        </span>
                      </div>
                    )
                  })}
                  {events.length === 0 && <p className="text-xs text-slate-400">ماكو حركات بعد.</p>}
                </div>
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}
