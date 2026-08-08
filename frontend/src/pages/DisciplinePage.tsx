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
  MANUAL:                { text: '✋ تعديل يدوي',        cls: 'bg-blue-50 text-blue-700' },
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

  // ── التعديل اليدوي (المالك ومدير النظام) ──
  // النظام يغرّم تلقائياً حتى ما تصير محاباة، بس الآلة ما تعرف كل شي:
  // الموظف ممكن يتأخر لأن الزبون ما كان بالبيت. فلازم مفتاح تصحيح —
  // بشرط إنه ينسجّل باسم الي عدّل وسببه، حتى ما يصير باب خلفي.
  const [adjEmp, setAdjEmp] = useState('')
  const [adjDelta, setAdjDelta] = useState('')
  const [adjReason, setAdjReason] = useState('')
  const [adjBusy, setAdjBusy] = useState(false)
  const [adjMsg, setAdjMsg] = useState<{ ok: boolean; text: string } | null>(null)

  const submitAdjust = async () => {
    const delta = Number(adjDelta)
    if (!adjEmp || !delta || adjReason.trim().length < 3) {
      setAdjMsg({ ok: false, text: 'اختار الموظف، وحدد كم نقطة، واكتب السبب' })
      return
    }
    setAdjBusy(true); setAdjMsg(null)
    try {
      const r = await api.adjustDisciplinePoints(adjEmp, delta, adjReason.trim())
      setAdjMsg({ ok: true, text: `تم — الرصيد الجديد ${r.points} من ١٠٠` })
      setAdjDelta(''); setAdjReason('')
      load()
    } catch (e) {
      setAdjMsg({ ok: false, text: e instanceof Error ? e.message : 'تعذر التعديل' })
    } finally {
      setAdjBusy(false)
    }
  }

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

      {isAdmin && (
        <div className="mt-4 rounded-xl border border-blue-200 bg-blue-50/50 p-4">
          <h3 className="text-sm font-bold text-[#0f2040]">✋ تعديل يدوي على الرصيد</h3>
          <p className="mt-1 text-xs text-slate-500">
            النظام يغرّم تلقائياً، بس أحياناً الغرامة مو بمحلها (الزبون ما كان بالبيت مثلاً).
            كل تعديل هنا ينسجّل بسجل الحركات باسمك وسببه، والموظف يوصله إشعار.
          </p>
          <div className="mt-3 flex flex-wrap items-end gap-2">
            <div>
              <label className="block text-xs text-slate-500">الموظف</label>
              <select
                value={adjEmp}
                onChange={(e) => setAdjEmp(e.target.value)}
                className="mt-1 rounded-lg border border-slate-200 px-3 py-2 text-sm"
              >
                <option value="">— اختار —</option>
                {points.map((p) => (
                  <option key={p.employeeId} value={p.employeeId}>
                    {p.employeeName} ({p.points}/100)
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-slate-500">النقاط (+ يزيد / − ينقص)</label>
              <input
                type="number"
                value={adjDelta}
                onChange={(e) => setAdjDelta(e.target.value)}
                placeholder="مثال: 5 أو 5-"
                className="mt-1 w-32 rounded-lg border border-slate-200 px-3 py-2 text-sm"
              />
            </div>
            <div className="min-w-[220px] flex-1">
              <label className="block text-xs text-slate-500">السبب (إلزامي)</label>
              <input
                value={adjReason}
                onChange={(e) => setAdjReason(e.target.value)}
                placeholder="مثال: الغرامة انزلت غلط، الزبون ما كان بالبيت"
                className="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm"
              />
            </div>
            <button
              onClick={submitAdjust}
              disabled={adjBusy}
              className="rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-50"
            >
              {adjBusy ? 'جاري...' : 'طبّق التعديل'}
            </button>
          </div>
          {adjMsg && (
            <p className={`mt-2 text-sm font-bold ${adjMsg.ok ? 'text-emerald-700' : 'text-red-600'}`}>
              {adjMsg.text}
            </p>
          )}
        </div>
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
