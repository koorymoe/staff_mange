import { useEffect, useState } from 'react'
import { api, type PendingStaffRating } from '../api'

// ═══ تقييمات البشر بمحطاتها (قرار (ع) 10-05) ═══
// الإداري يقيّم الليدر لمن يرجع · المراقب يقيّم الليدر والإداري · الجودة بعد
// الاتصال بالزبون · والمراقب يقيّم المكاتب كل نص شهر. التقييم = ٤٠٪ من
// التقييم النهائي، والـ٦٠٪ الباقية من نقاط ماتركس.

const ROLE: Record<string, string> = { LEADER: 'الليدر', COORDINATOR: 'الإداري' }
const WORD = ['', 'سيء', 'ضعيف', 'مقبول', 'زين', 'ممتاز']

export function Stars({ value, onPick, disabled }: { value: number; onPick: (n: number) => void; disabled?: boolean }) {
  return (
    <span className="inline-flex items-center gap-0.5" dir="ltr">
      {[1, 2, 3, 4, 5].map((n) => (
        <button key={n} type="button" disabled={disabled} onClick={() => onPick(n)} title={WORD[n]}
          className={`text-xl leading-none transition ${n <= value ? 'text-amber-400' : 'text-slate-300 hover:text-amber-200'} disabled:opacity-60`}>★</button>
      ))}
    </span>
  )
}

type Kind = 'coord' | 'audit'

/** طابور «قيّم» — حجوزات رجعت وتنتظر تقييمك. يختفي إذا ماكو شي. */
export function StaffRatingQueue({ kind }: { kind: Kind }) {
  const [rows, setRows] = useState<PendingStaffRating[] | null>(null)
  const [done, setDone] = useState<Record<string, number>>({})
  const [err, setErr] = useState('')
  useEffect(() => {
    const f = kind === 'coord' ? api.getCoordPendingRatings : api.getAuditPendingRatings
    void f().then(setRows).catch(() => setRows([]))
  }, [kind])
  if (!rows || rows.length === 0) return null
  const left = rows.filter((r) => !done[r.bookingId + r.rateeId])
  if (left.length === 0) return <p className="rounded-xl bg-emerald-50 p-3 text-sm font-bold text-emerald-800">✅ خلّصت التقييمات. شكراً.</p>

  const rate = async (r: PendingStaffRating, score: number) => {
    setErr('')
    try {
      await api.rateStaff(kind, { bookingId: r.bookingId, rateeId: r.rateeId, score })
      setDone((d) => ({ ...d, [r.bookingId + r.rateeId]: score }))
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }

  return (
    <div dir="rtl" className="rounded-2xl border border-amber-200 bg-amber-50/60 p-3">
      <p className="text-sm font-extrabold text-amber-900">⭐ {kind === 'coord' ? 'قيّم الليدرية الي رجعوا من حجوزاتك' : 'قيّم الليدر والإداري للحجوزات الي خلصت'}</p>
      <p className="mb-2 text-[11px] text-amber-800">تقييمك يدخل بالتقييم النهائي للموظف (٤٠٪ من البشر و٦٠٪ من ماتركس).</p>
      {err && <p className="text-xs text-red-600">{err}</p>}
      <div className="space-y-1.5">
        {left.slice(0, 12).map((r) => (
          <div key={r.bookingId + r.rateeId} className="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-white px-3 py-2 text-sm">
            <span><b>{r.rateeName}</b> <span className="text-xs text-slate-500">({ROLE[r.role]}) · حجز {r.code}</span></span>
            <Stars value={0} onPick={(n) => void rate(r, n)} />
          </div>
        ))}
      </div>
    </div>
  )
}

/** تقييم أطراف حجز واحد (للجودة بعد الاتصال بالزبون). */
export function BookingStaffRating({ bookingId }: { bookingId: string }) {
  const [st, setSt] = useState<{ parties: PendingStaffRating[]; mine: Record<string, number> } | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { void api.getBookingRatingState(bookingId, 'QUALITY_CALL').then(setSt).catch(() => {}) }, [bookingId])
  if (!st || st.parties.length === 0) return null
  const rate = async (id: string, score: number) => {
    setErr('')
    try {
      await api.rateStaff('quality', { bookingId, rateeId: id, score })
      setSt((s) => s && { ...s, mine: { ...s.mine, [id]: score } })
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }
  return (
    <div dir="rtl" className="mt-3 rounded-xl border border-sky-200 bg-sky-50/60 p-2.5">
      <p className="text-xs font-extrabold text-sky-900">⭐ بعد ما حچيت ويا الزبون: شلون شغلهم؟</p>
      {err && <p className="text-xs text-red-600">{err}</p>}
      {st.parties.map((p) => (
        <div key={p.rateeId} className="mt-1 flex flex-wrap items-center justify-between gap-2 text-sm">
          <span><b>{p.rateeName}</b> <span className="text-xs text-slate-500">({ROLE[p.role]})</span></span>
          <Stars value={st.mine[p.rateeId] ?? 0} onPick={(n) => void rate(p.rateeId, n)} />
        </div>
      ))}
    </div>
  )
}

const ROLE_AR: Record<string, string> = {
  FINANCE: 'محاسب', DESIGNER: 'مصمم', IT_SUPPORT: 'آيتي', TECHNICIAN: 'فني', ENGINEER: 'مهندس', HR_COORDINATOR: 'إداري',
  SALES: 'مبيعات', MEDIA: 'إعلام', QUALITY_ENGINEER: 'جودة', MONITOR: 'مراقب', PROJECT_MANAGER: 'مدير مشاريع',
  SERVICE_MANAGER: 'مسؤول خدمة', GPS_ADMIN: 'جي بي اس', GPS_ENGINEER: 'جي بي اس',
}

/** تقييم المراقب الدوري — كل نص شهر لكل الموظفين حسب شغلهم. */
export function PeriodicRatings() {
  const [st, setSt] = useState<{ period: string; people: { id: string; name: string; role: string }[]; mine: Record<string, number> } | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { void api.getPeriodicRatings().then(setSt).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])
  if (err) return <p className="text-sm text-red-600">{err}</p>
  if (!st) return <p className="text-xs text-slate-400">…</p>
  const rated = Object.keys(st.mine).length
  const rate = async (id: string, score: number) => {
    setErr('')
    try {
      await api.rateStaff('periodic', { rateeId: id, score })
      setSt((s) => s && { ...s, mine: { ...s.mine, [id]: score } })
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }
  return (
    <div dir="rtl" className="space-y-3">
      <StaffRatingQueue kind="audit" />
      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <p className="text-sm font-extrabold text-slate-800">⭐ تقييمك لشغل الموظفين بهالنص شهر</p>
        <p className="mb-2 text-[11px] text-slate-500">قيّمت {rated} من {st.people.length}. قيّم كل واحد حسب شغله (المحاسب بالتدقيق، المصممة بالتصاميم، التقنيين بالخدمات…). تنعاد كل نص شهر.</p>
        <div className="grid gap-1.5 sm:grid-cols-2">
          {st.people.map((p) => (
            <div key={p.id} className={`flex items-center justify-between gap-2 rounded-xl border px-3 py-2 text-sm ${st.mine[p.id] ? 'border-emerald-200 bg-emerald-50/50' : 'border-slate-200'}`}>
              <span><b>{p.name}</b> <span className="text-xs text-slate-500">{ROLE_AR[p.role] ?? p.role}</span></span>
              <Stars value={st.mine[p.id] ?? 0} onPick={(n) => void rate(p.id, n)} />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
