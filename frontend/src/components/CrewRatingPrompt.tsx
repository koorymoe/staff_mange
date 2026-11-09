import { useEffect, useState } from 'react'
import { api, type PendingCrewRating } from '../api'
import MatrixNote from './MatrixNote'

// ═══ «قيّم فنيّيك» — طلب (ع) 10-05 ═══
// بعد ما يخلص الليدر حجز يطلعله تقييم الفنيين الي طلعوا وياه (١–٥ وملاحظة).
// ماتركس يذكّره كل يوم لحد ما يقيّم. التقييم يطلع بتقرير الفني بس — ماكو نقاط.
export default function CrewRatingPrompt() {
  const [rows, setRows] = useState<PendingCrewRating[]>([])
  const [scores, setScores] = useState<Record<string, number>>({})
  const [notes, setNotes] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState<string | null>(null)
  const [err, setErr] = useState('')

  const load = () => { void api.getPendingCrewRatings().then(setRows).catch(() => setRows([])) }
  useEffect(load, [])
  if (rows.length === 0) return null

  const save = async (b: PendingCrewRating) => {
    const ratings = b.techs.map((t) => ({ technicianId: t.id, score: scores[b.bookingId + t.id] ?? 0, note: notes[b.bookingId + t.id] || undefined }))
    if (ratings.some((r) => r.score < 1)) { setErr('قيّم كل فني (نجمة وحدة على الأقل).'); return }
    setBusy(b.bookingId); setErr('')
    try { await api.rateCrew(b.bookingId, ratings); window.dispatchEvent(new Event('matrix-refresh')); load() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') }
    finally { setBusy(null) }
  }

  return (
    <div dir="rtl" className="mb-4 rounded-2xl border border-amber-200 bg-amber-50/60 p-4">
      <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">⭐ قيّم الفنيين الي طلعوا وياك</h3>
      <MatrixNote className="mb-3">تقييمك يبين بتقرير الفني بس، وما يصير بيه نقاط ولا غرامة. اكتب ملاحظة إذا أكو شي لازم ينعرف.</MatrixNote>
      {err && <p className="mb-2 text-xs text-red-600">{err}</p>}
      <div className="space-y-3">
        {rows.map((b) => (
          <div key={b.bookingId} className="rounded-xl border border-slate-200 bg-white p-3">
            <p className="mb-2 text-sm font-bold">{b.bookingCode} <span className="text-[11px] font-normal text-slate-500">· خلص {new Date(b.completedAt).toLocaleDateString('en-GB')}</span></p>
            {b.techs.map((t) => {
              const k = b.bookingId + t.id
              return (
                <div key={t.id} className="mb-2 flex flex-wrap items-center gap-2">
                  <span className="min-w-[8rem] text-sm">{t.name}</span>
                  <span>
                    {[1, 2, 3, 4, 5].map((n) => (
                      <button key={n} type="button" aria-label={`${n} نجوم`} onClick={() => setScores((s) => ({ ...s, [k]: n }))}
                        className={`text-xl ${n <= (scores[k] ?? 0) ? 'text-amber-500' : 'text-slate-300'}`}>★</button>
                    ))}
                  </span>
                  <input value={notes[k] ?? ''} onChange={(e) => setNotes((s) => ({ ...s, [k]: e.target.value }))}
                    placeholder="ملاحظة (اختياري)" className="min-w-0 flex-1 rounded-lg border border-slate-200 px-2 py-1 text-xs" />
                </div>
              )
            })}
            <button type="button" disabled={busy === b.bookingId} onClick={() => void save(b)}
              className="rounded-lg bg-[#0f2040] px-4 py-1.5 text-xs font-bold text-white disabled:opacity-50">{busy === b.bookingId ? 'يحفظ…' : 'حفظ التقييم'}</button>
          </div>
        ))}
      </div>
    </div>
  )
}
