import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixSuggestAccuracy } from '../api'
import MatrixNote from './MatrixNote'

// ═══ دقة اقتراحات ماتركس — المرحلة الثانية ═══
// كم مرة المنسقين مشوا على اقتراحه، وهل الحجوزات الي مشت عليه طلعت بوقتها
// أكثر من الي انغيّرت. هذا المقياس الي على أساسه (ع) يقرر يوم ينطي ماتركس
// صلاحية تنفيذ — مو قبل.
export default function MatrixSuggestAccuracyPanel() {
  const [a, setA] = useState<MatrixSuggestAccuracy | null>(null)
  const [days, setDays] = useState(30)
  useEffect(() => {
    let alive = true
    void api.getMatrixSuggestAccuracy(days).then((r) => { if (alive) setA(r) }).catch(() => {})
    return () => { alive = false }
  }, [days])
  if (!a) return null
  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-violet-200 bg-white p-3 text-slate-800">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-extrabold text-violet-900">🎯 دقة اقتراحات ماتركس <span className="text-[11px] font-normal text-slate-500">ماتركس يقترح الموعد والكادر، والمنسق يقرر</span></p>
        <select value={days} onChange={(e) => setDays(Number(e.target.value))} className="rounded-lg border border-slate-200 px-2 py-1 text-xs">
          {[7, 30, 90].map((d) => <option key={d} value={d}>آخر {d} يوم</option>)}
        </select>
      </div>
      <MatrixNote>{a.insights.join(' ')}{a.pending ? ` و${a.pending} اقتراح ينتظر قرار المنسق.` : ''}</MatrixNote>
      <div className="grid gap-2 sm:grid-cols-2">
        {a.kinds.map((k) => (
          <div key={k.kind} className="rounded-xl border border-slate-200 bg-slate-50 p-3 text-xs">
            <b className="text-sm">{k.title}</b>
            <p className="mt-1 text-2xl font-black text-violet-700">{k.decided ? `${k.acceptPct}%` : <span className="text-sm font-bold text-slate-400">بعد ماكو قرارات</span>}</p>
            <p>✅ اعتمده {k.accepted} · 🟡 اعتمد جزء منه {k.partial} · ✏️ غيّره {k.changed}</p>
            {(k.onTimeAccepted != null || k.onTimeChanged != null) && (
              <p className="mt-1 text-slate-600">طلع بوقته: لمن اعتمدوه {k.onTimeAccepted ?? '—'}% ({k.sampleAccepted}) · لمن انغيّر {k.onTimeChanged ?? '—'}% ({k.sampleChanged})</p>
            )}
          </div>
        ))}
      </div>
      {a.recent.length > 0 && (
        <div className="text-xs">
          <b>آخر القرارات:</b>
          {a.recent.map((r, i) => <p key={i}>{r.outcome} — <Link to={`/bookings?focus=${r.bookingId}`} className="font-bold text-sky-700 underline">{r.code}</Link> · {r.kind}</p>)}
        </div>
      )}
    </div>
  )
}
