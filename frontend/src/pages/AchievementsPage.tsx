import { useEffect, useState } from 'react'
import { api, type Achievement } from '../api'

// ═══ الإنجازات (ماتركس) — للمالك ومدير النظام حصراً ═══
//
// كل تقرير يومي رفعه أي موظف بأي دور — الهدف يخلي ماتركس يتعرف
// أكثر على الموظفين وسلوكهم. المراجعة هنا تقييم بس («زين» أو «يحتاج
// مراجعة» + ملاحظة)، مو غرامة ولا قرار تلقائي.
export default function AchievementsPage() {
  const [day, setDay] = useState(() => new Date().toISOString().slice(0, 10))
  const [rows, setRows] = useState<Achievement[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    queueMicrotask(() => { if (alive) setLoading(true) })
    void (async () => {
      try {
        const data = await api.getAchievements({ day, limit: 300 })
        if (alive) setRows(data)
      } catch (e) {
        if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [day])

  const review = async (id: string, status: 'GOOD' | 'NEEDS_REVIEW') => {
    let note = ''
    if (status === 'NEEDS_REVIEW') {
      note = window.prompt('ملاحظتك على هذا التقرير؟') || ''
      if (!note.trim()) return
    }
    setBusyId(id)
    try {
      const updated = await api.reviewAchievement(id, status, note.trim())
      setRows((prev) => prev.map((r) => (r.id === id ? updated : r)))
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📋 الإنجازات</h2>
        <p className="mt-1 text-sm text-slate-500">
          تقارير يومية حرة من كل الموظفين — للمالك ومدير النظام حصراً. التقييم يذكّر الموظف، ما يغرّمه.
        </p>
      </div>

      <div className="flex items-center gap-2">
        <label className="text-sm font-bold text-slate-600">اليوم:</label>
        <input
          type="date"
          value={day}
          onChange={(e) => setDay(e.target.value)}
          className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm"
        />
      </div>

      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}
      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">ماكو تقارير بهذا اليوم.</p>
      )}

      <div className="space-y-3">
        {rows.map((a) => (
          <div key={a.id} className="rounded-2xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <span className="font-bold text-[#0f2040]">{a.employeeName || 'موظف'}</span>
                <span className="mr-2 rounded-full bg-slate-100 px-2 py-0.5 text-[11px] text-slate-500">{a.role}</span>
                {a.bookingCode && <span className="mr-2 text-[11px] text-slate-400">حجز: {a.bookingCode}</span>}
              </div>
              <span className="text-[11px] text-slate-400">{new Date(a.createdAt).toLocaleString('en-GB')}</span>
            </div>
            <p className="mt-2 whitespace-pre-wrap text-sm text-slate-700">{a.reportText}</p>

            {a.reviewStatus === 'PENDING' ? (
              <div className="mt-3 flex justify-end gap-2">
                <button
                  onClick={() => review(a.id, 'GOOD')}
                  disabled={busyId === a.id}
                  className="rounded-lg bg-emerald-100 px-3 py-1.5 text-xs font-bold text-emerald-800 hover:bg-emerald-200 disabled:opacity-50"
                >
                  ✓ زين
                </button>
                <button
                  onClick={() => review(a.id, 'NEEDS_REVIEW')}
                  disabled={busyId === a.id}
                  className="rounded-lg bg-amber-100 px-3 py-1.5 text-xs font-bold text-amber-800 hover:bg-amber-200 disabled:opacity-50"
                >
                  ⚠️ يحتاج مراجعة
                </button>
              </div>
            ) : (
              <div className="mt-3 flex items-center justify-end gap-2 text-[12px]">
                {a.reviewStatus === 'GOOD' ? (
                  <span className="font-bold text-emerald-700">✓ تمت المراجعة — زين</span>
                ) : (
                  <span className="font-bold text-amber-700">⚠️ {a.reviewNote}</span>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
