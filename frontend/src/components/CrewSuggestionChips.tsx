import { useEffect, useState } from 'react'
import { api, type CrewRecommendation, type CrewSuggestion } from '../api'

// ═══ ماتركس — «مقترح ماتركس» للكادر ═══
// ⚠️ الأب يرسمه بس لمن يعبر حارس التكليف (ADMIN/OWNER أو coordinator
// أو crew_management) — نفس حارس GET /api/ai/crew-recommendation.
// الضغط على الاقتراح ما يكلّف: بس يجهّز الاختيار والمنسّق يثبّت.

export default function CrewSuggestionChips({ bookingId, onPick }: { bookingId: string; onPick: (s: CrewSuggestion) => void }) {
  const [data, setData] = useState<CrewRecommendation | null | undefined>(undefined)

  useEffect(() => {
    let alive = true
    api.getCrewRecommendation(bookingId)
      .then((r) => { if (alive) setData(r) })
      .catch(() => { if (alive) setData(null) })
    return () => { alive = false }
  }, [bookingId])

  if (data === undefined || data === null) return null
  if (data.insufficient) {
    return <p className="mt-2 text-[11px] text-slate-400">🧠 مقترح ماتركس: {data.note || 'ماكو بيانات كافية'}</p>
  }

  const row = (label: string, list: CrewSuggestion[]) => list.length > 0 && (
    <div className="flex flex-wrap items-center gap-1.5">
      <span className="text-[11px] font-bold text-slate-500">{label}</span>
      {list.map((s) => (
        <button key={s.employeeId} type="button" onClick={() => onPick(s)}
          title={s.reasons.join(' · ')}
          className="rounded-full bg-violet-50 px-2.5 py-1 text-[11px] font-bold text-violet-700 ring-1 ring-violet-200 hover:bg-violet-100">
          {s.name}
          <span className="mr-1 font-normal text-violet-500">
            {s.doneCount > 0 ? `${s.doneCount} منجز` : 'بلا خبرة مسجّلة'}
            {s.problemRatePct != null ? ` · مشاكل ${s.problemRatePct}٪` : ''}
            {` · ${s.dayLoad ? `${s.dayLoad} حجز باليوم` : 'فاضي'}`}
          </span>
        </button>
      ))}
    </div>
  )

  return (
    <div className="mt-2 space-y-1.5 rounded-xl border border-violet-100 bg-violet-50/40 px-3 py-2">
      <p className="text-[11px] font-extrabold text-violet-800">🧠 مقترح ماتركس{data.serviceName ? ` لـ«${data.serviceName}»` : ''} — اضغط حتى تجهّزه، والتثبيت إلك</p>
      {row('ليدر:', data.leaders)}
      {row('فني:', data.technicians)}
    </div>
  )
}
