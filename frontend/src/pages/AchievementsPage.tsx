import { useEffect, useState } from 'react'
import { api, type Achievement } from '../api'
import { VoicePlayer } from '../components/VoiceNote'

// ═══ الإنجازات (ماتركس) — للمالك والمدير والمراقب (ع) 10-23 ═══
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

  // رد ماتركس المقترح — قابل للتعديل قبل الإرسال (ع) 10-23. بلا window.prompt (محجوب ببعض الأجهزة).
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const review = async (a: Achievement, status: 'GOOD' | 'NEEDS_REVIEW') => {
    const note = (drafts[a.id] ?? a.suggestedReply ?? '').trim()
    if (status === 'NEEDS_REVIEW' && !note) { setErr('اكتب ملاحظتك للموظف — شنو يحتاج يصلّح'); return }
    setBusyId(a.id)
    setErr(null)
    try {
      const updated = await api.reviewAchievement(a.id, status, note)
      setRows((prev) => prev.map((r) => (r.id === a.id ? updated : r)))
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
          تقارير يومية من الموظفين — ماتركس يقترح الرد وأنت تعدّله وتبعثه. الرد يوصل للموظف كإشعار، ما يغرّمه.
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
            {a.voiceSeconds != null && <VoicePlayer achievementId={a.id} seconds={a.voiceSeconds} />}

            {a.reviewStatus === 'PENDING' ? (
              <div className="mt-3 rounded-xl border border-violet-100 bg-violet-50/60 p-3">
                <label className="mb-1 block text-xs font-bold text-violet-800">
                  🤖 رد ماتركس المقترح {a.suggestedStatus === 'NEEDS_REVIEW' ? '(يقترح «يحتاج مراجعة»)' : '(يقترح «زين»)'} — عدّله براحتك
                </label>
                <textarea
                  value={drafts[a.id] ?? a.suggestedReply ?? ''}
                  onChange={(e) => setDrafts((d) => ({ ...d, [a.id]: e.target.value }))}
                  rows={2}
                  className="w-full rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-400"
                />
                <div className="mt-2 flex flex-wrap justify-end gap-2">
                  <button
                    onClick={() => review(a, 'GOOD')}
                    disabled={busyId === a.id}
                    className="min-h-9 rounded-lg bg-emerald-100 px-3 text-xs font-bold text-emerald-800 hover:bg-emerald-200 disabled:opacity-50"
                  >
                    ✓ زين — ابعث الرد
                  </button>
                  <button
                    onClick={() => review(a, 'NEEDS_REVIEW')}
                    disabled={busyId === a.id}
                    className="min-h-9 rounded-lg bg-amber-100 px-3 text-xs font-bold text-amber-800 hover:bg-amber-200 disabled:opacity-50"
                  >
                    ⚠️ يحتاج مراجعة — ابعث الرد
                  </button>
                </div>
              </div>
            ) : (
              <div className="mt-3 flex items-center justify-end gap-2 text-[12px]">
                {a.reviewStatus === 'GOOD' ? (
                  <span className="font-bold text-emerald-700">✓ تمت المراجعة — زين{a.reviewNote ? ` · ${a.reviewNote}` : ''}</span>
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
